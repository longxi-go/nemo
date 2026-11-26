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
)

const (
	sortPackagePath                               = "sort"
	objectCollectionDefaultsAreUnsupportedMessage = "object collection defaults are unsupported"
)

func (m *modeler) objectCollection(declared types.Type, field *fieldModel) (bool, error) {
	typ := types.Unalias(declared)
	if hasTextType(typ) {
		return false, nil
	}
	var element types.Type
	var mode fieldMode
	switch collection := typ.Underlying().(type) {
	case *types.Slice:
		element = collection.Elem()
		mode = fieldModeList
	case *types.Map:
		if !types.Identical(collection.Key(), types.Typ[types.String]) {
			return false, nil
		}
		element = collection.Elem()
		mode = fieldModeMap
	default:
		return false, nil
	}
	base := types.Unalias(element)
	pointer := false
	if ptr, ok := base.(*types.Pointer); ok {
		base = types.Unalias(ptr.Elem())
		pointer = true
	}
	if _, ok := base.Underlying().(*types.Struct); !ok || hasTextType(base) {
		return false, nil
	}
	if field.HasDefault {
		return true, errors.New(objectCollectionDefaultsAreUnsupportedMessage)
	}
	object := &objectModel{
		Type:    m.typeName(element),
		Pointer: pointer,
	}
	if named, ok := base.(*types.Named); ok && externalBindable(named, m.pkg.Types, m.bindable) {
		object.External = true
		if pointer {
			object.Helper = m.pointerHelper()
		}
	} else {
		renderBase := element
		if ptr, ok := types.Unalias(element).(*types.Pointer); ok {
			renderBase = ptr.Elem()
		}
		helper, err := m.objectHelper(renderBase)
		if err != nil {
			return true, err
		}
		object.Helper = helper
		object.Base = m.typeName(renderBase)
	}
	field.Mode = mode
	field.Type = m.typeName(declared)
	field.Element = object
	if mode == fieldModeMap {
		m.imports.Add(sortPackagePath, m.names.Sort)
	}
	return true, nil
}
