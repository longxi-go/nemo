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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func command(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", args, err, out)
	}
	return string(out)
}

// consumerModule keeps consumer tests independent of the local ecosystem layout.
func consumerModule(t *testing.T, name string) string {
	t.Helper()
	repo, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(read(t, filepath.Join(repo, "go.mod")))
	var version string
	for i := 0; i+1 < len(fields); i++ {
		if fields[i] == "go" {
			version = fields[i+1]
			break
		}
	}
	if version == "" {
		t.Fatal("module Go version is unavailable")
	}
	module := fmt.Sprintf(`module %s

go %s

require github.com/longxi-go/nemo v0.0.0

replace github.com/longxi-go/nemo => %s
`, name, version, strconv.Quote(repo))
	return workspaceConsumerModule(t, repo, module)
}

// workspaceConsumerModule opts isolated consumers into an explicitly selected
// workspace. Default tests continue to verify the declared published graph.
func workspaceConsumerModule(t *testing.T, repo, content string) string {
	t.Helper()
	workspace := os.Getenv("NEMO_TEST_WORKSPACE")
	if workspace == "" {
		return content
	}
	if !filepath.IsAbs(workspace) {
		t.Fatal("NEMO_TEST_WORKSPACE must be absolute")
	}
	work, err := modfile.ParseWork(workspace, []byte(read(t, workspace)), nil)
	if err != nil {
		t.Fatal(err)
	}
	consumer, err := modfile.Parse("go.mod", []byte(content), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, use := range work.Use {
		dir := use.Path
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(filepath.Dir(workspace), dir)
		}
		if filepath.Clean(dir) == repo {
			continue
		}
		path := filepath.Join(dir, "go.mod")
		dependency, err := modfile.Parse(path, []byte(read(t, path)), nil)
		if err != nil || dependency.Module == nil {
			t.Fatalf("workspace module %s: %v", path, err)
		}
		if err := consumer.AddReplace(dependency.Module.Mod.Path, "", dir, ""); err != nil {
			t.Fatal(err)
		}
	}
	data, err := consumer.Format()
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
