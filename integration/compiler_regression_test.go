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

func TestGeneratedExternalValueAndIdentifierCollisions(t *testing.T) {
	for _, scenario := range []string{"external-value", "identifier-collisions", "public-aliases"} {
		t.Run(scenario, func(t *testing.T) {
			checkExternalFixture(t, scenario)
		})
	}
}

func TestRejectInvalidExternalBinders(t *testing.T) {
	for _, scenario := range []string{"function-field", "broad-parameter", "variadic"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			write(t, filepath.Join(dir, "go.mod"), consumerModule(t, "consumer"))
			if err := os.Mkdir(filepath.Join(dir, "child"), 0o755); err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(dir, "config.go"), read(t, "testdata/function-field-binder/config.go.tmpl"))
			write(t, filepath.Join(dir, "child/config.go"), read(t, "testdata/function-field-binder/"+scenario+".go.tmpl"))
			command(t, dir, "go", "mod", "tidy")
			_, err := codegen.Run(context.Background(), codegen.Options{
				Dir: dir,
			})
			if err == nil || !strings.Contains(err.Error(), "needs BindProperties or UnmarshalText") {
				t.Fatalf("invalid external binder admitted: %v", err)
			}
			if _, err := os.Stat(filepath.Join(dir, "zz_nemo.gen.go")); !os.IsNotExist(err) {
				t.Fatal("invalid binder wrote generated output")
			}
		})
	}
}

func TestDefaultsUseTargetIntegerWidth(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), consumerModule(t, "consumer"))
	write(t, filepath.Join(dir, "config.go"), read(t, "testdata/target-width/config.go.tmpl"))
	command(t, dir, "go", "mod", "tidy")
	options := codegen.Options{
		Dir: dir,
	}
	t.Setenv("GOARCH", "386")
	if _, err := codegen.Run(context.Background(), options); err == nil || !strings.Contains(err.Error(), "invalid default") {
		t.Fatalf("32-bit target accepted overflowing default: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "zz_nemo.gen.go")); !os.IsNotExist(err) {
		t.Fatal("invalid default wrote generated output")
	}
	t.Setenv("GOARCH", "amd64")
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatalf("64-bit target rejected representable default: %v", err)
	}
}

func TestExternalObjectDepthBoundary(t *testing.T) {
	checkExternalFixture(t, "external-depth")
}

func TestExternalBinderLateParentFailure(t *testing.T) {
	checkExternalFixture(t, "external-ownership")
}

func checkExternalFixture(t *testing.T, scenario string) {
	t.Helper()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), consumerModule(t, "consumer"))
	if err := os.Mkdir(filepath.Join(dir, "child"), 0o755); err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join("testdata", scenario)
	write(t, filepath.Join(dir, "config.go"), read(t, filepath.Join(fixture, "config.go.tmpl")))
	write(t, filepath.Join(dir, "child/config.go"), read(t, filepath.Join(fixture, "child.go.tmpl")))
	command(t, dir, "go", "mod", "tidy")
	options := codegen.Options{
		Dir:   dir,
		Roots: []string{"./..."},
	}
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	first := read(t, filepath.Join(dir, "zz_nemo.gen.go"))
	write(t, filepath.Join(dir, "config_test.go"), read(t, filepath.Join(fixture, "runtime_test.go.tmpl")))
	command(t, dir, "go", "mod", "tidy")
	command(t, dir, "go", "test", "./...")
	options.Check = true
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if read(t, filepath.Join(dir, "zz_nemo.gen.go")) != first {
		t.Fatal("check changed generated output")
	}
}
