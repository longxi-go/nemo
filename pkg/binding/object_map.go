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

package binding

import (
	"context"
	"sort"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/admission"
	"github.com/longxi-go/nemo/internal/propertypath"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// MapObjects reads an exact object first, retaining literal map member names.
// Otherwise flat descendants are grouped at their first dot. Flat addresses
// cannot recover dots that belonged to an original map key before normalization.
func MapObjects(r gosr.PropertyResolver, key string, required bool) (maptypes.StringMap[ObjectInput], error) {
	if r == nil {
		return nil, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	property, found, err := r.Lookup(key)
	if err != nil {
		return nil, err
	}
	if found {
		if property.Value.Kind() != gosr.ObjectKind {
			return nil, objectFailure(key, property.Origin)
		}
		if err := admitComposite(key, property); err != nil {
			return nil, err
		}
		fields := property.Value.Fields()
		members := make([]string, 0, len(fields))
		for member := range fields {
			members = append(members, member)
		}
		sort.Strings(members)
		out := make(maptypes.StringMap[ObjectInput], len(fields))
		for _, member := range members {
			// Each element has a separate resolver, so its map key need not become an address.
			input, err := newObjectInput(r, key, fields[member], key, property.Origin, member)
			if err != nil {
				return nil, err
			}
			out[member] = input
		}
		return out, nil
	}
	groups := make(maptypes.StringMap[maptypes.PropertyMap])
	keys := r.Keys()
	if err := limits.Check(limits.Properties, int64(len(keys)), int64(limits.DefaultProperties)); err != nil {
		return nil, err
	}
	budget := admission.New(context.Background(), limits.Default(), false)
	for _, address := range keys {
		if !strings.HasPrefix(address, key+propertypath.Separator) {
			continue
		}
		relative := strings.TrimPrefix(address, key+propertypath.Separator)
		member, _, _ := strings.Cut(relative, propertypath.Separator)
		value, exists, err := r.Lookup(address)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		if err := budget.Property(address, value.Value); err != nil {
			return nil, err
		}
		if groups[member] == nil {
			groups[member] = make(maptypes.PropertyMap)
		}
		groups[member][address] = value
	}
	if len(groups) == 0 {
		if required {
			return nil, &gosr.Error{
				Code: gosr.MissingProperty,
				Key:  key,
			}
		}
		return nil, nil
	}
	out := make(maptypes.StringMap[ObjectInput], len(groups))
	for member, properties := range groups {
		prefix := Join(key, member)
		if exact, exists := properties[prefix]; exists && exact.Value.Kind() != gosr.ObjectKind {
			return nil, objectFailure(prefix, exact.Origin)
		}
		view := elementResolver{
			base:       r,
			properties: properties,
		}
		scoped, err := Nested(view, prefix, true)
		if err != nil {
			return nil, err
		}
		out[member] = ObjectInput{
			Resolver: scoped,
			Prefix:   prefix,
			path:     []string{member},
		}
	}
	return out, nil
}
