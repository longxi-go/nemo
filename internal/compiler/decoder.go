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
	"fmt"
	"go/types"
)

const (
	bindingBoolean                            = "Boolean"
	bindingDuration                           = "Duration"
	bindingFloat                              = "Float"
	bindingListDecoder                        = "ListDecoder"
	bindingMapDecoder                         = "MapDecoder"
	bindingPointerDecoder                     = "PointerDecoder"
	bindingSigned                             = "Signed"
	bindingString                             = "String"
	bindingText                               = "Text"
	bindingUnsigned                           = "Unsigned"
	collectionsAdmitScalarElementsOnlyMessage = "collections admit scalar elements only"
	nestedCollectionsUnsupportedMessage       = "nested collections unsupported"
	unsupportedScalarFormat                   = "unsupported scalar %s"
	unsupportedTypeFormat                     = "unsupported type %s"
)

func (m *modeler) decoder(t types.Type, scalarOnly bool) (string, error) {
	declared := t
	t = types.Unalias(t)
	if isDurationType(t) {
		return m.names.Binding + goSelector + bindingDuration, nil
	}
	if named, ok := t.(*types.Named); ok {
		if hasText(named) {
			function := m.names.Binding + goSelector + bindingText
			return function + goTypeArgumentsOpen + m.typeName(declared) + goTypeArgumentsClose, nil
		}
	}
	switch typ := t.Underlying().(type) {
	case *types.Basic:
		fn := ""
		switch {
		case typ.Info()&types.IsString != 0:
			fn = bindingString
		case typ.Info()&types.IsBoolean != 0:
			fn = bindingBoolean
		case typ.Info()&types.IsUnsigned != 0:
			fn = bindingUnsigned
		case typ.Info()&types.IsInteger != 0:
			fn = bindingSigned
		case typ.Info()&types.IsFloat != 0:
			fn = bindingFloat
		default:
			return "", fmt.Errorf(unsupportedScalarFormat, types.TypeString(t, nil))
		}
		// Imports belong to emitted type expressions. Container types adapted by
		// combinators do not appear in the output and must not register imports.
		function := m.names.Binding + goSelector + fn
		return function + goTypeArgumentsOpen + m.typeName(declared) + goTypeArgumentsClose, nil
	case *types.Pointer:
		if scalarOnly {
			return "", errors.New(collectionsAdmitScalarElementsOnlyMessage)
		}
		decode, err := m.decoder(typ.Elem(), true)
		if err != nil {
			return "", err
		}
		return m.names.Binding + goSelector + bindingPointerDecoder + goCallOpen + decode + goCallClose, nil
	case *types.Slice:
		if scalarOnly {
			return "", errors.New(nestedCollectionsUnsupportedMessage)
		}
		decode, err := m.decoder(typ.Elem(), true)
		if err != nil {
			return "", err
		}
		return m.names.Binding + goSelector + bindingListDecoder + goCallOpen + decode + goCallClose, nil
	case *types.Map:
		if scalarOnly {
			return "", errors.New(nestedCollectionsUnsupportedMessage)
		}
		if !types.Identical(typ.Key(), types.Typ[types.String]) {
			return "", errors.New(mapsRequireStringKeysMessage)
		}
		decode, err := m.decoder(typ.Elem(), true)
		if err != nil {
			return "", err
		}
		return m.names.Binding + goSelector + bindingMapDecoder + goCallOpen + decode + goCallClose, nil
	default:
		return "", fmt.Errorf(unsupportedTypeFormat, types.TypeString(t, nil))
	}
}
