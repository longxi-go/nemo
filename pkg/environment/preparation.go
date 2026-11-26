// Copyright © 2026-present the longxi-go/nemo authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package environment

import (
	"context"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/admission"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	duplicateSourceIdentityMessage   = "duplicate source identity"
	enumeratedPropertyMissingMessage = "enumerated property missing"
)

// fetchProperties admits sources in precedence order before resolution. The
// shared budget covers the complete candidate, including earlier providers.
func (e *Environment) fetchProperties(ctx context.Context) (maptypes.PropertyMap, error) {
	properties := maptypes.PropertyMap{}
	admitted := sourceAdmission{
		policy:     e.limits,
		budget:     admission.New(ctx, e.limits, false),
		properties: properties,
	}
	seen := maptypes.StringBoolMap{}
	for _, provider := range e.providers {
		loaded, err := provider.Fetch(ctx)
		if err != nil {
			return nil, &gosr.Error{
				Code:  gosr.SourceFailed,
				Cause: err,
			}
		}
		if loaded == nil {
			return nil, &gosr.Error{
				Code: gosr.SourceFailed,
			}
		}
		name := loaded.Name()
		if err := gosr.ValidateKey(name); err != nil {
			return nil, &gosr.Error{
				Code:  gosr.SourceFailed,
				Cause: err,
			}
		}
		if seen[name] {
			return nil, &gosr.Error{
				Code: gosr.SourceFailed,
				Origin: gosr.Origin{
					Source: name,
				},
				Message: duplicateSourceIdentityMessage,
			}
		}
		seen[name] = true
		if err := admitted.merge(ctx, loaded, name); err != nil {
			return nil, err
		}
	}
	return properties, nil
}

// sourceAdmission owns the shared candidate budget and first-source-wins table.
// Shadowed values are skipped before Lookup, as required by source precedence.
type sourceAdmission struct {
	policy     limits.Policy
	budget     *admission.Budget
	properties maptypes.PropertyMap
}

func (a *sourceAdmission) merge(ctx context.Context, loaded gosr.PropertySource, name string) error {
	keys := loaded.Keys()
	if err := limits.Check(limits.Properties, int64(len(keys)), int64(a.policy.MaxProperties)); err != nil {
		return err
	}
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := gosr.ValidateKey(key); err != nil {
			return err
		}
		if _, present := a.properties[key]; present {
			continue
		}
		property, found, err := loaded.Lookup(key)
		if err != nil {
			return &gosr.Error{
				Code:  gosr.SourceFailed,
				Key:   key,
				Cause: err,
			}
		}
		if !found {
			return &gosr.Error{
				Code:    gosr.SourceFailed,
				Key:     key,
				Message: enumeratedPropertyMissingMessage,
			}
		}
		if property.Origin.Source == "" {
			property.Origin.Source = name
		}
		if property.Origin.Key == "" {
			property.Origin.Key = key
		}
		if err := limits.Check(limits.Properties, int64(len(a.properties))+1, int64(a.policy.MaxProperties)); err != nil {
			return err
		}
		if err := a.budget.Property(key, property.Value); err != nil {
			return err
		}
		a.properties[key] = property
	}
	return nil
}
