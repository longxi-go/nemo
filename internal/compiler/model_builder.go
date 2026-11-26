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
	"path/filepath"
	"sort"
	"strconv"

	"github.com/longxi-go/gosr/pkg/codegen/imports"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	bindableTypeName       = "Bindable"
	goImportAliasSeparator = " "
	invalidConfigFormat    = "invalid config %s"
)

func buildModels(a analysis) ([]preparedPackage, error) {
	var models []preparedPackage
	for _, pkg := range a.packages {
		decls := a.declarations[pkg.PkgPath]
		if len(decls) == 0 && !a.owned[pkg.PkgPath] {
			continue
		}
		m := modeler{
			pkg:         pkg,
			identifiers: newIdentifiers(pkg.Types.Scope().Names()),
			helpers:     make(maptypes.Map[types.Type, string]),
		}
		if len(decls) > 0 {
			environment := pkg.Imports[environmentImportPath]
			m.bindable = environment.Types.Scope().Lookup(bindableTypeName).Type().Underlying().(*types.Interface)
		}
		m.names = newGeneratedNames(m.identifiers)
		m.imports = imports.New(pkg.PkgPath, m.importAlias)
		m.imports.Add(bindingImportPath, m.names.Binding)
		m.imports.Add(environmentImportPath, m.names.Environment)
		model := packageModel{
			Names: m.names,
			Name:  pkg.Name,
		}
		info := File{
			Path:    filepath.Join(filepath.Dir(pkg.GoFiles[0]), generatedFileName),
			Package: pkg.PkgPath,
		}
		for _, d := range decls {
			named, ok := pkg.Types.Scope().Lookup(d.name).Type().(*types.Named)
			if !ok {
				return nil, fmt.Errorf(invalidConfigFormat, d.name)
			}
			helper, err := m.objectHelper(named)
			if err != nil {
				return nil, err
			}
			model.Binders = append(model.Binders, binderModel{
				Name:   d.name,
				Prefix: d.prefix,
				Helper: helper,
			})
			info.Types = append(info.Types, d.name)
		}
		markRecursivePointers(m.models)
		model.Helpers = m.models
		model.NewPointer = m.newPointer
		if len(decls) > 0 {
			entries, err := m.imports.List()
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				model.Imports = append(model.Imports, entry.Alias+goImportAliasSeparator+strconv.Quote(entry.Path))
			}
		}
		sort.Strings(model.Imports)
		models = append(models, preparedPackage{
			model: model,
			file:  info,
		})
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].file.Path < models[j].file.Path
	})
	return models, nil
}
