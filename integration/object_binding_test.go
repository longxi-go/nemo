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
	"path/filepath"
	"testing"

	"github.com/longxi-go/nemo/pkg/codegen"
)

func TestObjectBindingAnonymous(t *testing.T) {
	checkObjectFixture(t, "anonymous-object")
}

func TestObjectBindingCollections(t *testing.T) {
	checkObjectFixture(t, "object-collections")
}

func TestObjectBindingRecursive(t *testing.T) {
	checkObjectFixture(t, "recursive-object")
}

func TestObjectBindingAcyclicDepth(t *testing.T) {
	checkObjectFixture(t, "acyclic-depth")
}

func checkObjectFixture(t *testing.T, scenario string) {
	t.Helper()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), consumerModule(t, "consumer"))
	fixture := filepath.Join("testdata", scenario)
	write(t, filepath.Join(dir, "config.go"), read(t, filepath.Join(fixture, "config.go.tmpl")))
	command(t, dir, "go", "mod", "tidy")
	options := codegen.Options{
		Dir: dir,
	}
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	first := read(t, filepath.Join(dir, "zz_nemo.gen.go"))
	write(t, filepath.Join(dir, "config_test.go"), read(t, filepath.Join(fixture, "runtime_test.go.tmpl")))
	command(t, dir, "go", "mod", "tidy")
	command(t, dir, "go", "test", "-timeout=30s", "./...")
	options.Check = true
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if read(t, filepath.Join(dir, "zz_nemo.gen.go")) != first {
		t.Fatal("check changed generated output")
	}
	options.Check = false
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if read(t, filepath.Join(dir, "zz_nemo.gen.go")) != first {
		t.Fatal("repeated generation changed output")
	}
}
