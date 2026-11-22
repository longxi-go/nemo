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
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/longxi-go/gosr/pkg/codegen/artifact"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"golang.org/x/tools/go/packages"
)

func TestFormatAndOutputSafety(t *testing.T) {
	path := filepath.Join(t.TempDir(), "zz_nemo.gen.go")
	g := generation{
		analysis: analysis{
			packages: []*packages.Package{
				{
					PkgPath: "example.org/fixture",
					Module: &packages.Module{
						GoVersion: "1.13.0",
					},
				},
			},
			overlay: make(maptypes.StringBytesMap),
		},
		outputs: []output{
			{
				file: File{
					Path:    path,
					Package: "example.org/fixture",
				},
				File: artifact.File{
					Content: []byte("package fixture\nfunc Value( ) int {return 1}\n"),
				},
			},
		},
	}
	ctx := context.Background()
	if err := finishGeneration(ctx, &g); err == nil {
		t.Fatal("output admitted before validation")
	}
	if err := formatGeneration(ctx, &g); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(g.outputs[0].Content), "func Value() int") || string(g.analysis.overlay[path]) != string(g.outputs[0].Content) {
		t.Fatal("formatted bytes not used in overlay")
	}
	if !strings.Contains(string(g.outputs[0].Content), "func Value() int {\n\treturn 1\n}") {
		t.Fatal("format step did not expand compact function body")
	}
	g.validated = true
	g.options.Check = true
	if err := finishGeneration(ctx, &g); err == nil {
		t.Fatal("missing check output accepted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("check wrote source")
	}
	g.options.Check = false
	g.inspect = true
	if err := finishGeneration(ctx, &g); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("inspect wrote source")
	}
	g.inspect = false
	if err := finishGeneration(ctx, &g); err != nil {
		t.Fatal(err)
	}
	bytes, err := os.ReadFile(path)
	if err != nil || string(bytes) != string(g.outputs[0].Content) {
		t.Fatal(string(bytes), err)
	}
}
