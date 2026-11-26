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
	"strings"

	"github.com/longxi-go/gosr/pkg/codegen/diagnostic"
	"github.com/longxi-go/gosr/pkg/codegen/imports"
	"github.com/longxi-go/nemo/internal/propertypath"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"golang.org/x/tools/go/packages"
)

const (
	anonymousObjectHelper                                        = "nemoBindObject"
	fieldDiagnosticSeparator                                     = ": "
	fieldLabel                                                   = "field "
	namedObjectHelperPrefix                                      = "nemoBind"
	conflictingPropertyPathsMessage                              = "conflicting property paths"
	expectedStructMessage                                        = "expected struct"
	externalNestedStructNeedsBindPropertiesOrUnmarshalTextFormat = "external nested struct %s needs BindProperties or UnmarshalText"
	genericNestedStructsUnsupportedMessage                       = "generic nested structs unsupported"
	objectTypeAliasFormat                                        = "nemoType%d"
)

type modeler struct {
	pkg         *packages.Package
	bindable    *types.Interface
	identifiers *identifiers
	names       generatedNames
	newPointer  string
	imports     *imports.Set
	helpers     maptypes.Map[types.Type, string]
	models      []helperModel
}

func (m *modeler) typeName(t types.Type) string {
	return m.imports.Type(t)
}

func (m *modeler) importAlias(_ string, _ string, index int) string {
	return m.identifiers.take(fmt.Sprintf(objectTypeAliasFormat, index))
}

func (m *modeler) diagnostic(field *types.Var, message string) error {
	return diagnostic.Wrap(stageAnalyze, m.pkg.Fset.Position(field.Pos()), fieldLabel+field.Name()+fieldDiagnosticSeparator+message, nil)
}

func (m *modeler) objectHelper(declared types.Type) (string, error) {
	typ := types.Unalias(declared)
	named, isNamed := typ.(*types.Named)
	if isNamed && named.TypeArgs().Len() > 0 {
		return "", errors.New(genericNestedStructsUnsupportedMessage)
	}
	if helper, ok := m.helpers[typ]; ok {
		return helper, nil
	}
	if isNamed && named.Obj().Pkg() != m.pkg.Types {
		return "", fmt.Errorf(externalNestedStructNeedsBindPropertiesOrUnmarshalTextFormat, named)
	}
	st, ok := typ.Underlying().(*types.Struct)
	if !ok {
		return "", errors.New(expectedStructMessage)
	}
	preferred := anonymousObjectHelper
	if isNamed {
		preferred = namedObjectHelperPrefix + named.Obj().Name()
	}
	helper := m.identifiers.take(preferred)
	m.helpers[typ] = helper
	model := helperModel{
		Name: helper,
		Type: m.typeName(declared),
	}
	keys := make(maptypes.StringBoolMap)
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		value, skip, err := m.field(field, st.Tag(i))
		if err != nil {
			return "", m.diagnostic(field, err.Error())
		}
		if skip {
			continue
		}
		for key := range keys {
			if key == value.Key ||
				strings.HasPrefix(key, value.Key+propertypath.Separator) ||
				strings.HasPrefix(value.Key, key+propertypath.Separator) {
				return "", m.diagnostic(field, conflictingPropertyPathsMessage)
			}
		}
		keys[value.Key] = true
		model.Fields = append(model.Fields, value)
	}
	m.models = append(m.models, model)
	return helper, nil
}
