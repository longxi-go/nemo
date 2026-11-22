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

func TestGeneratedImportedDecoderTypes(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), consumerModule(t, "example.org/decoder"))
	if err := os.Mkdir(filepath.Join(dir, "values"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "values/types.go"), "package values\ntype Ports []int\ntype Count *int\n")
	write(t, filepath.Join(dir, "config.go"), read(t, "testdata/imported-decoders/decoder.go.tmpl"))
	command(t, dir, "go", "mod", "tidy")
	options := codegen.Options{
		Dir: dir,
	}
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	generated := read(t, filepath.Join(dir, "zz_nemo.gen.go"))
	write(t, filepath.Join(dir, "config_test.go"), read(t, "testdata/imported-decoders/runtime_test.go.tmpl"))
	command(t, dir, "go", "mod", "tidy")
	command(t, dir, "go", "test", "./...")
	options.Check = true
	if _, err := codegen.Run(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if read(t, filepath.Join(dir, "zz_nemo.gen.go")) != generated {
		t.Fatal("generated output changed during check")
	}
}
