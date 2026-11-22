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
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/longxi-go/nemo/pkg/codegen"
)

func TestExternalFirstGeneration(t *testing.T) {
	dir := t.TempDir()
	repo, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	module := consumerModule(t, "consumer")
	write(t, filepath.Join(dir, "go.mod"), module)
	write(t, filepath.Join(dir, "main.go"), read(t, "testdata/consumer/main.go.tmpl"))
	write(t, filepath.Join(dir, "config.go"), read(t, "testdata/consumer/config.go.tmpl"))
	if err := os.Mkdir(filepath.Join(dir, "child"), 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "child/config.go"), read(t, "testdata/consumer/child.go.tmpl"))
	// Fixture dependency preparation is explicit, outside generator execution.
	command(t, dir, "go", "mod", "tidy")
	before := read(t, filepath.Join(dir, "go.mod"))
	sum := read(t, filepath.Join(dir, "go.sum"))
	report, err := codegen.Inspect(context.Background(), codegen.Options{
		Dir:   dir,
		Roots: []string{"./..."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Files) != 2 {
		t.Fatal(report)
	}
	generated := filepath.Join(dir, "zz_nemo.gen.go")
	if _, err := os.Stat(generated); !os.IsNotExist(err) {
		t.Fatal("Inspect wrote output")
	}
	if _, err = codegen.Run(context.Background(), codegen.Options{
		Dir:   dir,
		Roots: []string{"./..."},
		Check: true,
	}); err == nil {
		t.Fatal("missing output check accepted")
	}
	if _, err = codegen.Run(context.Background(), codegen.Options{
		Dir:   dir,
		Roots: []string{"./..."},
	}); err != nil {
		t.Fatal(err)
	}
	first := read(t, generated)
	for _, file := range report.Files {
		assertGeneratedLayout(t, file.Path)
	}
	command(t, repo, "go", "run", "./cmd/nemoctl", "check", "--dir", dir, "--roots", ".", "--roots", "./child")
	if _, err = codegen.Run(context.Background(), codegen.Options{
		Dir:   dir,
		Roots: []string{"./..."},
		Check: true,
	}); err != nil {
		t.Fatal(err)
	}
	if output := command(t, dir, "go", "run", "."); strings.TrimSpace(output) != "OK" {
		t.Fatal(output)
	}
	if read(t, filepath.Join(dir, "go.mod")) != before || read(t, filepath.Join(dir, "go.sum")) != sum {
		t.Fatal("generator modified module metadata")
	}
	write(t, generated, first+"\n// stale\n")
	if _, err = codegen.Run(context.Background(), codegen.Options{
		Dir:   dir,
		Roots: []string{"./..."},
		Check: true,
	}); err == nil {
		t.Fatal("stale accepted")
	}
	if read(t, generated) != first+"\n// stale\n" {
		t.Fatal("check mutated output")
	}
	if _, err = codegen.Run(context.Background(), codegen.Options{
		Dir:   dir,
		Roots: []string{"./..."},
	}); err != nil {
		t.Fatal(err)
	}
	if read(t, generated) != first {
		t.Fatal("nondeterministic")
	}
	write(t, filepath.Join(dir, "broken.go"), "package main\nvar broken int = \"wrong\"\n")
	if _, err = codegen.Run(context.Background(), codegen.Options{
		Dir:   dir,
		Roots: []string{"./..."},
	}); err == nil {
		t.Fatal("unrelated type error ignored")
	}
	if read(t, generated) != first {
		t.Fatal("failed analysis wrote output")
	}
}
