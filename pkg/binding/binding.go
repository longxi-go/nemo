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
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/propertypath"
)

// Bind requires *T to implement Bindable. It captures one snapshot when the
// resolver implements gosr.SnapshotProvider.
func Bind[T any, P BindablePointer[T]](resolver gosr.PropertyResolver, prefix string) (T, error) {
	var target T
	binder := P(&target)
	resolver = Pin(resolver)
	if prefix == "" {
		if p, ok := any(binder).(PropertyPrefixProvider); ok {
			prefix = p.PropertyPrefix()
		}
	}
	err := binder.BindProperties(resolver, prefix)
	return target, err
}

// Pin calls Snapshot once on a SnapshotProvider; snapshots pass through unchanged.
func Pin(resolver gosr.PropertyResolver) gosr.PropertyResolver {
	if _, ok := resolver.(gosr.Snapshot); ok {
		return resolver
	}
	if p, ok := resolver.(gosr.SnapshotProvider); ok {
		return p.Snapshot()
	}
	return resolver
}

// ReadOptions separates a missing default from an explicitly empty default.
type ReadOptions struct {
	Required   bool
	HasDefault bool
	Default    string
}

// Join preserves the shared empty-prefix root convention.
func Join(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + propertypath.Separator + key
}

// Present checks exact and descendant presence without changing exact Lookup semantics.
func Present(r gosr.PropertyResolver, key string) (bool, error) {
	if r == nil {
		return false, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	_, ok, err := r.Lookup(key)
	if err != nil || ok {
		return ok, err
	}
	for _, k := range r.Keys() {
		if strings.HasPrefix(k, key+propertypath.Separator) {
			return true, nil
		}
	}
	return false, nil
}

// Read decodes a property or its missing-value default. Errors retain source
// information without printing the property value.
func Read[T any](r gosr.PropertyResolver, key string, options ReadOptions, decode gosr.ValueDecoderFunc[T]) (T, error) {
	var zero T
	if r == nil || decode == nil {
		return zero, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	p, found, err := r.Lookup(key)
	if err != nil {
		return zero, err
	}
	if !found {
		for _, child := range r.Keys() {
			if !strings.HasPrefix(child, key+propertypath.Separator) {
				continue
			}
			property, _, err := r.Lookup(child)
			if err != nil {
				return zero, err
			}
			return zero, &gosr.Error{
				Code:    gosr.ConversionFailed,
				Key:     key,
				Origin:  property.Origin,
				Message: "expected exact scalar or collection property",
			}
		}
	}
	if !found && options.HasDefault {
		p.Value = gosr.String(options.Default)
		found = true
	}
	if !found {
		if options.Required {
			return zero, &gosr.Error{
				Code: gosr.MissingProperty,
				Key:  key,
			}
		}
		return zero, nil
	}
	value, err := decode(p.Value)
	if err != nil {
		return zero, &gosr.Error{
			Code:   gosr.ConversionFailed,
			Key:    key,
			Origin: p.Origin,
			Cause:  err,
		}
	}
	return value, nil
}
