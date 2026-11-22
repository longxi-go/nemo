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
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/gosr/tools/pkg/annotation"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

type scannedSource struct {
	declarations []declaration
	names        []string
}

func scanDeclarations(files []string, registry *annotation.Registry) (scannedSource, error) {
	var source scannedSource
	for _, path := range files {
		if filepath.Base(path) == generatedFileName {
			continue
		}
		fs := token.NewFileSet()
		file, err := parser.ParseFile(fs, path, nil, parser.ParseComments)
		if err != nil {
			return source, err
		}
		consumed := make(maptypes.Map[*ast.CommentGroup, bool])
		for _, node := range file.Decls {
			if function, ok := node.(*ast.FuncDecl); ok && function.Recv == nil {
				source.names = append(source.names, function.Name.Name)
			}
			gen, ok := node.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					source.names = append(source.names, spec.Name.Name)
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						source.names = append(source.names, name.Name)
					}
				}
			}
			if gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				comments := annotation.TypeComments(gen, ts)
				consumed[comments] = true
				values, err := registry.Parse(fs, comments, markers.DescribesType)
				if err != nil {
					return source, err
				}
				for _, value := range values {
					prefix := value.Value
					if text := prefix.(string); text != "" {
						if err := gosr.ValidateKey(text); err != nil {
							return source, fmt.Errorf("%s: invalid config prefix", path)
						}
					}
					if ts.TypeParams != nil || ts.Assign.IsValid() {
						return source, fmt.Errorf("%s: generic or alias config %s unsupported", path, ts.Name)
					}
					if _, ok := ts.Type.(*ast.StructType); !ok {
						return source, fmt.Errorf("%s: config %s must be a struct", path, ts.Name)
					}
					source.declarations = append(source.declarations, declaration{
						name:   ts.Name.Name,
						prefix: prefix.(string),
					})
				}
			}
		}
		for _, comments := range file.Comments {
			if consumed[comments] {
				continue
			}
			if _, err := registry.Parse(fs, comments, markers.DescribesField); err != nil {
				return source, err
			}
		}
	}
	return source, nil
}
