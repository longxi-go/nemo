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

	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestAllModesRejectSymlinkOutput(t *testing.T) {
	t.Setenv("GOWORK", "off")
	for _, mode := range []string{"generate", "check", "inspect"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			fixtureFiles := maptypes.StringMap[string]{
				"go.mod":    "module fixture\ngo 1.27.0\n",
				"config.go": "package fixture\n",
			}
			for name, content := range fixtureFiles {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			target := filepath.Join(t.TempDir(), "generated.go")
			original := generatedHeader + "package fixture\n"
			if err := os.WriteFile(target, []byte(original), 0o600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, generatedFileName)
			if err := os.Symlink(target, path); err != nil {
				t.Fatal(err)
			}
			_, err := Run(context.Background(), Options{
				Dir:   dir,
				Check: mode == "check",
			}, mode == "inspect")
			if err == nil || !strings.Contains(err.Error(), "not a regular file") {
				t.Fatalf("%s accepted symlink output: %v", mode, err)
			}
			actual, err := os.ReadFile(target)
			if err != nil || string(actual) != original {
				t.Fatalf("%s changed symlink target: %v", mode, err)
			}
			if info, err := os.Lstat(path); err != nil || info.Mode()&os.ModeSymlink == 0 {
				t.Fatalf("%s replaced symlink: %v", mode, err)
			}
		})
	}
}
