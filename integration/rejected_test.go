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
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/longxi-go/nemo/pkg/codegen"
)

func TestRejectedDeclarations(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), consumerModule(t, "rejected"))
	// Retain the runtime dependency while preparing the fixture once.
	write(t, filepath.Join(dir, "dependency.go"), "package rejected\nimport _ \"github.com/longxi-go/nemo/pkg/binding\"\n")
	command(t, dir, "go", "mod", "tidy")
	for _, test := range []struct {
		name, fixture string
	}{
		{
			"anonymous-private",
			"rejected/anonymous-private.go.tmpl",
		},
		{
			"channel",
			"rejected/channel.go.tmpl",
		},
		{
			"map keys",
			"rejected/map-keys.go.tmpl",
		},
		{
			"nested collection",
			"rejected/nested-collection.go.tmpl",
		},
		{
			"interface",
			"rejected/interface.go.tmpl",
		},
		{
			"generic",
			"rejected/generic.go.tmpl",
		},
		{
			"embedded",
			"rejected/embedded.go.tmpl",
		},
		{
			"empty property",
			"rejected/empty-property.go.tmpl",
		},
		{
			"required syntax",
			"rejected/required-syntax.go.tmpl",
		},
		{
			"default overflow",
			"rejected/default-overflow.go.tmpl",
		},
		{
			"collision",
			"rejected/collision.go.tmpl",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			write(t, filepath.Join(dir, "config.go"), read(t, filepath.Join("testdata", test.fixture)))
			if _, err := codegen.Run(context.Background(), codegen.Options{
				Dir: dir,
			}); err == nil {
				t.Fatalf("accepted %s", test.name)
			}
			if _, err := os.Stat(filepath.Join(dir, "zz_nemo.gen.go")); !os.IsNotExist(err) {
				t.Fatal("invalid output written")
			}
		})
	}
	write(t, filepath.Join(dir, "config.go"), read(t, "testdata/unsupported-binding/rejected.go.tmpl"))
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "BindProperties") {
		t.Fatalf("unsupported target admission: %v %s", err, out)
	}
}
