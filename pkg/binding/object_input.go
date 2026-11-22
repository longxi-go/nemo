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
	"errors"
	"slices"
	"sort"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/propertypath"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// ObjectInput isolates one collection element for a statically generated binder.
// Resolver exposes only this object's projected addresses; Resolve retains the
// enclosing environment's absolute-reference semantics. Prefix is a binding-only
// address, never an index written back into the source.
type ObjectInput struct {
	Resolver  gosr.PropertyResolver
	Prefix    string
	sourceKey string
	origin    gosr.Origin
	path      []string
}

// Wrap adds element/field context without modifying the original diagnostic or
// printing values. Absolute Resolve errors and unclassified external errors keep
// their original provenance. Causes remain accessible through errors.Is/As.
func (input ObjectInput) Wrap(err error) error {
	if err == nil {
		return nil
	}
	var absolute *absoluteResolutionError
	if errors.As(err, &absolute) {
		return err
	}
	var original *gosr.Error
	if !errors.As(err, &original) {
		return err
	}
	if original.Key != input.Prefix && !strings.HasPrefix(original.Key, input.Prefix+propertypath.Separator) {
		return err
	}
	copy := *original
	copy.Cause = err
	copy.Path = slices.Clone(input.path)
	if relative := strings.TrimPrefix(strings.TrimPrefix(original.Key, input.Prefix), propertypath.Separator); relative != "" {
		copy.Path = append(copy.Path, strings.Split(relative, propertypath.Separator)...)
	}
	copy.Path = append(copy.Path, original.Path...)
	if input.sourceKey != "" {
		copy.Key = input.sourceKey
		copy.Origin = input.origin
	}
	return &copy
}

func newObjectInput(
	base gosr.PropertyResolver,
	prefix string,
	value gosr.Value,
	key string,
	origin gosr.Origin,
	path string,
) (ObjectInput, error) {
	input := ObjectInput{
		Prefix:    prefix,
		sourceKey: key,
		origin:    origin,
		path:      []string{path},
	}
	if value.Kind() != gosr.ObjectKind {
		return input, &gosr.Error{
			Code:    gosr.ConversionFailed,
			Key:     key,
			Origin:  origin,
			Path:    []string{path},
			Message: "expected object element",
		}
	}
	properties := maptypes.PropertyMap{
		prefix: {
			Value:  value,
			Origin: origin,
		},
	}
	if err := projectObject(prefix, value, origin, properties); err != nil {
		return input, err
	}
	input.Resolver = elementResolver{
		base:       base,
		properties: properties,
	}
	return input, nil
}

type elementResolver struct {
	base       gosr.PropertyResolver
	properties maptypes.PropertyMap
}

func (r elementResolver) Lookup(key string) (gosr.Property, bool, error) {
	if err := gosr.ValidateKey(key); err != nil {
		return gosr.Property{}, false, err
	}
	property, found := r.properties[key]
	return property, found, nil
}

func (r elementResolver) Keys() []string {
	keys := make([]string, 0, len(r.properties))
	for key := range r.properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (r elementResolver) Resolve(expression string) (string, error) {
	value, err := r.base.Resolve(expression)
	if err != nil {
		return "", &absoluteResolutionError{
			cause: err,
		}
	}
	return value, nil
}

type absoluteResolutionError struct {
	cause error
}

func (e *absoluteResolutionError) Error() string {
	return e.cause.Error()
}

func (e *absoluteResolutionError) Unwrap() error {
	return e.cause
}

func objectFailure(key string, origin gosr.Origin) error {
	return &gosr.Error{
		Code:    gosr.ConversionFailed,
		Key:     key,
		Origin:  origin,
		Message: "expected object collection",
	}
}
