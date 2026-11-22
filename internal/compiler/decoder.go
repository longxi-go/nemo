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
	"fmt"
	"go/types"
)

func (m *modeler) decoder(t types.Type, scalarOnly bool) (string, error) {
	declared := t
	t = types.Unalias(t)
	if isDurationType(t) {
		return m.names.Binding + goSelector + "Duration", nil
	}
	if named, ok := t.(*types.Named); ok {
		if hasText(named) {
			function := m.names.Binding + goSelector + "Text"
			return function + goTypeArgumentsOpen + m.typeName(declared) + goTypeArgumentsClose, nil
		}
	}
	switch typ := t.Underlying().(type) {
	case *types.Basic:
		fn := ""
		switch {
		case typ.Info()&types.IsString != 0:
			fn = "String"
		case typ.Info()&types.IsBoolean != 0:
			fn = "Boolean"
		case typ.Info()&types.IsUnsigned != 0:
			fn = "Unsigned"
		case typ.Info()&types.IsInteger != 0:
			fn = "Signed"
		case typ.Info()&types.IsFloat != 0:
			fn = "Float"
		default:
			return "", fmt.Errorf("unsupported scalar %s", types.TypeString(t, nil))
		}
		// Imports belong to emitted type expressions. Container types adapted by
		// combinators do not appear in the output and must not register imports.
		function := m.names.Binding + goSelector + fn
		return function + goTypeArgumentsOpen + m.typeName(declared) + goTypeArgumentsClose, nil
	case *types.Pointer:
		if scalarOnly {
			return "", fmt.Errorf("collections admit scalar elements only")
		}
		decode, err := m.decoder(typ.Elem(), true)
		if err != nil {
			return "", err
		}
		return m.names.Binding + goSelector + "PointerDecoder" + goCallOpen + decode + goCallClose, nil
	case *types.Slice:
		if scalarOnly {
			return "", fmt.Errorf("nested collections unsupported")
		}
		decode, err := m.decoder(typ.Elem(), true)
		if err != nil {
			return "", err
		}
		return m.names.Binding + goSelector + "ListDecoder" + goCallOpen + decode + goCallClose, nil
	case *types.Map:
		if scalarOnly {
			return "", fmt.Errorf("nested collections unsupported")
		}
		if !types.Identical(typ.Key(), types.Typ[types.String]) {
			return "", fmt.Errorf("maps require string keys")
		}
		decode, err := m.decoder(typ.Elem(), true)
		if err != nil {
			return "", err
		}
		return m.names.Binding + goSelector + "MapDecoder" + goCallOpen + decode + goCallClose, nil
	default:
		return "", fmt.Errorf("unsupported type %s", types.TypeString(t, nil))
	}
}
