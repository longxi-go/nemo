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
	"slices"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

type mapSource struct {
	name       string
	properties maptypes.PropertyMap
	keys       []string
}

// NewMap detaches the map; gosr.Value already owns immutable nested storage.
// A nil map is a valid empty source. Names and exact keys must be nonempty and trimmed.
func NewMap(name string, values maptypes.ValueMap) (gosr.PropertySource, error) {
	properties := make(maptypes.PropertyMap, len(values))
	for key, value := range values {
		properties[key] = gosr.Property{
			Value: value,
		}
	}
	return NewProperties(name, properties)
}

// NewProperties preserves supplied provenance and fills missing Source/Key fields.
// The detached source can be shared by readers without application synchronization.
func NewProperties(name string, values maptypes.PropertyMap) (gosr.PropertySource, error) {
	if err := validName(name); err != nil {
		return nil, err
	}
	properties := make(maptypes.PropertyMap, len(values))
	keys := make([]string, 0, len(values))
	for key, property := range values {
		if err := gosr.ValidateKey(key); err != nil {
			return nil, err
		}
		if property.Origin.Source == "" {
			property.Origin.Source = name
		}
		if property.Origin.Key == "" {
			property.Origin.Key = key
		}
		properties[key] = property
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return &mapSource{
		name:       name,
		properties: properties,
		keys:       keys,
	}, nil
}

func (s *mapSource) Name() string {
	return s.name
}

func (s *mapSource) Keys() []string {
	return slices.Clone(s.keys)
}

func (s *mapSource) Lookup(key string) (gosr.Property, bool, error) {
	if err := gosr.ValidateKey(key); err != nil {
		return gosr.Property{}, false, err
	}
	p, ok := s.properties[key]
	return p, ok, nil
}
