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
	"testing"

	"github.com/longxi-go/nemo/pkg/codegen"
)

func TestDiscoverRegisteredDSL(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), consumerModule(t, "consumer"))
	for _, name := range []string{"config", "decoder"} {
		write(t, filepath.Join(dir, name+".go"), read(t, filepath.Join("testdata/codec-extension", name+".go.tmpl")))
	}
	command(t, dir, "go", "mod", "tidy")
	options := codegen.Options{
		Dir: dir,
	}
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	original := read(t, filepath.Join(dir, "zz_nemo.gen.go"))
	write(t, filepath.Join(dir, "config_test.go"), read(t, "testdata/codec-extension/runtime_test.go.tmpl"))
	if err := os.Mkdir(filepath.Join(dir, "formats"), 0o755); err != nil {
		t.Fatal(err)
	}
	formats, err := filepath.Glob("../pkg/codec/testdata/formats/*")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range formats {
		write(t, filepath.Join(dir, "formats", filepath.Base(path)), read(t, path))
	}
	command(t, dir, "go", "mod", "tidy")
	command(t, dir, "go", "test", "-v", "./...")
	options.Check = true
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if read(t, filepath.Join(dir, "zz_nemo.gen.go")) != original {
		t.Fatal("read-only check changed generated output")
	}
}
