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

func TestFormatGenerationUsesEachPackageModuleVersion(t *testing.T) {
	// Octal spelling exposes version selection without depending on template text.
	dir := t.TempDir()
	t.Setenv("GOWORK", filepath.Join(dir, "go.work"))
	fixtureFiles := maptypes.StringMap[string]{
		"go.work":         "go 1.18\nuse (\n./older\n./newer\n)\n",
		"older/go.mod":    "module example.org/older\ngo 1.12\n",
		"older/config.go": "package older\n",
		"newer/go.mod":    "module example.org/newer\ngo 1.13.0\n",
		"newer/config.go": "package newer\n",
	}
	for name, source := range fixtureFiles {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	config := loadConfig(context.Background(), Options{
		Dir: dir,
	}, nil)
	loaded, err := packages.Load(config, "./older", "./newer")
	if err != nil {
		t.Fatal(err)
	}
	if err := packageErrors(loaded); err != nil {
		t.Fatal(err)
	}
	g := generation{
		analysis: analysis{
			packages: loaded,
			overlay:  make(maptypes.StringBytesMap),
		},
		outputs: []output{
			{
				file: File{
					Path:    "older/zz_nemo.gen.go",
					Package: "example.org/older",
				},
				File: artifact.File{
					Content: []byte("package older\nconst Mode = 0755\n"),
				},
			},
			{
				file: File{
					Path:    "newer/zz_nemo.gen.go",
					Package: "example.org/newer",
				},
				File: artifact.File{
					Content: []byte("package newer\nconst Mode = 0755\n"),
				},
			},
		},
	}
	if err := formatGeneration(context.Background(), &g); err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"const Mode = 0755", "const Mode = 0o755"} {
		out := g.outputs[i]
		if !strings.Contains(string(out.Content), want) {
			t.Errorf("%s: want %q, got %s", out.file.Package, want, out.Content)
		}
		if string(g.analysis.overlay[out.file.Path]) != string(out.Content) {
			t.Errorf("%s: validation overlay differs from formatted output", out.file.Package)
		}
	}
}
