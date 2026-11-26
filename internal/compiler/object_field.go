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
	nestedStructDefaultsBelongOnScalarFieldsMessage = "nested struct defaults belong on scalar fields"
)

func (m *modeler) objectField(declared types.Type, value *fieldModel) (bool, error) {
	typ := types.Unalias(declared)
	inner := typ
	renderType := declared
	if ptr, ok := typ.(*types.Pointer); ok {
		inner = types.Unalias(ptr.Elem())
		renderType = ptr.Elem()
		value.Pointer = true
	}
	if nested, ok := inner.(*types.Named); ok {
		if _, isStruct := nested.Underlying().(*types.Struct); isStruct &&
			externalBindable(nested, m.pkg.Types, m.bindable) &&
			!hasText(nested) {
			if value.HasDefault {
				return true, errors.New(nestedStructDefaultsBelongOnScalarFieldsMessage)
			}
			value.Mode = fieldModeExternal
			if !value.Pointer {
				value.Type = m.typeName(declared)
			}
			if value.Pointer {
				value.Helper = m.pointerHelper()
			}
			return true, nil
		}
	}
	if _, isStruct := inner.Underlying().(*types.Struct); isStruct && !hasTextType(inner) {
		if value.HasDefault {
			return true, errors.New(nestedStructDefaultsBelongOnScalarFieldsMessage)
		}
		helper, err := m.objectHelper(renderType)
		if err != nil {
			return true, err
		}
		value.Mode = fieldModeLocal
		value.Type = m.typeName(renderType)
		value.Helper = helper
		return true, nil
	}

	return false, nil
}

func hasTextType(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	return ok && hasText(named)
}
