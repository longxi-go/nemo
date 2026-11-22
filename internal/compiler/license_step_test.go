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
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/longxi-go/gosr/pkg/codegen/artifact"
	"golang.org/x/tools/go/packages"
)

func TestLicenseHeadersUseEachOutputModule(t *testing.T) {
	g := generation{}
	for _, owner := range []string{"first", "second"} {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "codequality"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "codequality/HEADER"), []byte("Copyright © 2026 the "+owner+" authors.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		g.analysis.packages = append(g.analysis.packages, &packages.Package{
			PkgPath: owner,
			Module: &packages.Module{
				Dir: dir,
			},
		})
		g.outputs = append(g.outputs, output{
			file: File{
				Package: owner,
				Path:    filepath.Join(dir, generatedFileName),
			},
			File: artifact.File{
				Content: []byte(generatedHeader + "\npackage " + owner + "\n"),
			},
		})
	}
	if err := licenseGeneration(context.Background(), &g); err != nil {
		t.Fatal(err)
	}
	for _, out := range g.outputs {
		if !bytes.HasPrefix(out.Content, []byte(generatedHeader)) || !bytes.Contains(out.Content, []byte("the "+out.file.Package+" authors.")) {
			t.Fatalf("wrong output owner:\n%s", out.Content)
		}
	}
	if err := licenseGeneration(context.Background(), &g); err != nil {
		t.Fatal("second refresh failed", err)
	}
}
