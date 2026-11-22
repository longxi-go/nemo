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

package integration_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// Check delivered source rather than template spelling: formatting must preserve
// structural layout for methods and converter closures in actual consumer modules.
func assertGeneratedLayout(t *testing.T, path string) {
	t.Helper()
	source := read(t, path)
	positions := token.NewFileSet()
	file, err := parser.ParseFile(positions, path, source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		var body *ast.BlockStmt
		switch function := node.(type) {
		case *ast.FuncDecl:
			body = function.Body
		case *ast.FuncLit:
			body = function.Body
		}
		if body != nil && positions.Position(body.Lbrace).Line == positions.Position(body.Rbrace).Line {
			t.Errorf("%s: generated function body must span multiple lines", positions.Position(body.Lbrace))
		}
		return true
	})
	lines := strings.Split(source, "\n")
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			start := function.Pos()
			if function.Doc != nil {
				start = function.Doc.Pos()
			}
			line := positions.Position(start).Line
			if line > 1 && strings.TrimSpace(lines[line-2]) != "" {
				t.Errorf("%s: generated declaration needs a preceding blank line", positions.Position(start))
			}
		}
	}
}
