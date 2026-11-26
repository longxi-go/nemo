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
	"slices"
	"sort"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/admission"
	"github.com/longxi-go/nemo/internal/propertypath"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	ambiguousObjectMemberAddressMessage = "ambiguous object member address"
	invalidObjectMemberAddressMessage   = "invalid object member address"
)

// Nested validates composite presence and projects object members into descendant
// addresses. Literal dots remain part of member names, while object nesting adds
// separators. Two object paths producing one address are ambiguous and rejected,
// even when their values agree. Independently addressed source properties retain
// exact Lookup precedence over this explicit binding-only projection.
func Nested(r gosr.PropertyResolver, key string, required bool) (gosr.PropertyResolver, error) {
	if r == nil {
		return nil, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	property, found, err := r.Lookup(key)
	if err != nil {
		return nil, err
	}
	if found && property.Value.Kind() != gosr.ObjectKind {
		return nil, &gosr.Error{
			Code:   gosr.ConversionFailed,
			Key:    key,
			Origin: property.Origin,
		}
	}
	present, err := Present(r, key)
	if err != nil {
		return nil, err
	}
	if required && !present {
		return nil, &gosr.Error{
			Code: gosr.MissingProperty,
			Key:  key,
		}
	}
	if !found {
		return r, nil
	}
	projected := make(maptypes.PropertyMap)
	base := r
	if parent, ok := r.(objectResolver); ok {
		base = parent.PropertyResolver
		// A selected object replaces its earlier projected branch. Keep the
		// original resolver separate so true exact addresses still win.
		for address, property := range parent.projected {
			if !strings.HasPrefix(address, key+propertypath.Separator) {
				projected[address] = property
			}
		}
	}
	if err := projectObject(key, property.Value, property.Origin, projected); err != nil {
		return nil, err
	}
	return objectResolver{
		PropertyResolver: base,
		projected:        projected,
	}, nil
}

// projectObject constructs the single address table shared by Keys and Lookup.
// Sorted traversal makes ambiguous-address diagnostics independent of map order.
func projectObject(prefix string, value gosr.Value, origin gosr.Origin, projected maptypes.PropertyMap) error {
	if err := admitComposite(prefix, gosr.Property{
		Value:  value,
		Origin: origin,
	}); err != nil {
		return err
	}
	budget := admission.New(context.Background(), limits.Default(), false)
	projection := objectProjection{
		origin:     origin,
		properties: projected,
		budget:     budget,
	}
	return projection.members(prefix, value)
}

// objectProjection owns one bounded traversal and its collision table.
type objectProjection struct {
	origin     gosr.Origin
	properties maptypes.PropertyMap
	budget     *admission.Budget
}

func (p *objectProjection) members(prefix string, value gosr.Value) error {
	fields := value.Fields()
	members := make([]string, 0, len(fields))
	for member := range fields {
		members = append(members, member)
	}
	sort.Strings(members)
	for _, member := range members {
		address := Join(prefix, member)
		if err := gosr.ValidateKey(address); err != nil {
			return &gosr.Error{
				Code:    gosr.ConversionFailed,
				Key:     prefix,
				Origin:  p.origin,
				Message: invalidObjectMemberAddressMessage,
				Cause:   err,
			}
		}
		if _, exists := p.properties[address]; exists {
			return &gosr.Error{
				Code:    gosr.ConversionFailed,
				Key:     address,
				Origin:  p.origin,
				Message: ambiguousObjectMemberAddressMessage,
			}
		}
		item := fields[member]
		if err := limits.Check(limits.Properties, int64(len(p.properties))+1, int64(limits.DefaultProperties)); err != nil {
			return err
		}
		if err := p.budget.Bytes(len(address)); err != nil {
			return err
		}
		p.properties[address] = gosr.Property{
			Value:  item,
			Origin: p.origin,
		}
		if item.Kind() == gosr.ObjectKind {
			if err := p.members(address, item); err != nil {
				return err
			}
		}
	}
	return nil
}

type objectResolver struct {
	gosr.PropertyResolver
	projected maptypes.PropertyMap
}

func (r objectResolver) Lookup(key string) (gosr.Property, bool, error) {
	property, found, err := r.PropertyResolver.Lookup(key)
	if found || err != nil {
		return property, found, err
	}
	property, found = r.projected[key]
	return property, found, nil
}

func (r objectResolver) Keys() []string {
	keys := r.PropertyResolver.Keys()
	for key := range r.projected {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return slices.Compact(keys)
}
