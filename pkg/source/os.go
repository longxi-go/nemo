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

package source

import (
	"context"
	"os"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/propertypath"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	osEnvironmentResource      = "os.Environ"
	osAssignmentSeparator      = "="
	osPropertySegmentSeparator = "_"
	invalidOSPrefixCharacters  = osAssignmentSeparator + "\x00"
)

type osProvider struct {
	name   string
	prefix string
}

// NewOS snapshots os.Environ on each Fetch. Prefix matching is case-sensitive;
// matching names lose the prefix, become lowercase, and replace every underscore
// with a dot. Values stay strings, including empty strings. Aliases are errors.
// An empty prefix deliberately includes the entire process environment.
func NewOS(name, prefix string) (Provider, error) {
	if err := validName(name); err != nil {
		return nil, err
	}
	if strings.ContainsAny(prefix, invalidOSPrefixCharacters) {
		return nil, failure(name, osEnvironmentResource, nil)
	}
	return osProvider{
		name:   name,
		prefix: prefix,
	}, nil
}

func (p osProvider) Fetch(ctx context.Context) (gosr.PropertySource, error) {
	if err := ctx.Err(); err != nil {
		return nil, failure(p.name, osEnvironmentResource, err)
	}
	values := make(maptypes.PropertyMap)
	for _, entry := range os.Environ() {
		name, value, ok := strings.Cut(entry, osAssignmentSeparator)
		if !ok {
			continue
		}
		key, matched := strings.CutPrefix(name, p.prefix)
		if !matched {
			continue
		}
		key = strings.ToLower(key)
		key = strings.ReplaceAll(key, osPropertySegmentSeparator, propertypath.Separator)
		if err := gosr.ValidateKey(key); err != nil {
			return nil, failure(p.name, osEnvironmentResource, err)
		}
		if _, exists := values[key]; exists {
			return nil, failure(p.name, osEnvironmentResource, nil)
		}
		values[key] = gosr.Property{
			Value: gosr.String(value),
			Origin: gosr.Origin{
				Source:   p.name,
				Resource: osEnvironmentResource,
				Key:      key,
			},
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, failure(p.name, osEnvironmentResource, err)
	}
	return NewProperties(p.name, values)
}
