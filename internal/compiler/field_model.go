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

package compiler

import (
	"errors"
	"go/types"
	"reflect"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

const (
	bindingRead                         = "Read"
	bindingReadMap                      = "ReadMap"
	requiredFalse                       = "false"
	requiredTrue                        = "true"
	unexportedBoundFieldMessage         = "unexported bound field; mark property:\"-\" to preserve"
	embeddedFieldsAreUnsupportedMessage = "embedded fields are unsupported"
	invalidDefaultForFieldTypeMessage   = "invalid default for field type"
	invalidPropertyTagMessage           = "invalid property tag"
	mapsRequireStringKeysMessage        = "maps require string keys"
	requiredMustBeTrueOrFalseMessage    = "required must be true or false"
)

func (m *modeler) field(field *types.Var, rawTag string) (fieldModel, bool, error) {
	value := fieldModel{
		Name:   field.Name(),
		Reader: m.names.Binding + goSelector + bindingRead,
		Mode:   fieldModeRead,
	}
	if err := validateTags(rawTag); err != nil {
		return value, false, err
	}
	tag := reflect.StructTag(rawTag)
	key, hasKey := tag.Lookup(propertyTag)
	if key == skippedProperty {
		return value, true, nil
	}
	if field.Embedded() {
		return value, false, errors.New(embeddedFieldsAreUnsupportedMessage)
	}
	if !field.Exported() {
		return value, false, errors.New(unexportedBoundFieldMessage)
	}
	if !hasKey {
		key = strings.ToLower(field.Name())
	}
	if err := gosr.ValidateKey(key); err != nil {
		return value, false, errors.New(invalidPropertyTagMessage)
	}
	value.Key = key
	value.Default, value.HasDefault = tag.Lookup(defaultTag)
	if value.HasDefault {
		if err := validateDefault(field.Type(), value.Default, m.pkg.TypesSizes); err != nil {
			return value, false, errors.New(invalidDefaultForFieldTypeMessage)
		}
	}
	if raw, ok := tag.Lookup(requiredTag); ok {
		if raw != requiredTrue && raw != requiredFalse {
			return value, false, errors.New(requiredMustBeTrueOrFalseMessage)
		}
		value.Required = raw == requiredTrue
	}
	if object, err := m.objectField(field.Type(), &value); object || err != nil {
		return value, false, err
	}
	declared := field.Type()
	typ := types.Unalias(declared)

	if collection, err := m.objectCollection(declared, &value); collection || err != nil {
		return value, false, err
	}
	var err error
	if named, ok := typ.(*types.Named); ok && hasText(named) {
		value.Decoder, err = m.decoder(declared, false)
	} else if mp, ok := typ.Underlying().(*types.Map); ok {
		if !types.Identical(mp.Key(), types.Typ[types.String]) {
			return value, false, errors.New(mapsRequireStringKeysMessage)
		}
		value.Reader = m.names.Binding + goSelector + bindingReadMap
		value.Decoder, err = m.decoder(mp.Elem(), true)
	} else {
		value.Decoder, err = m.decoder(declared, false)
	}
	return value, false, err
}
