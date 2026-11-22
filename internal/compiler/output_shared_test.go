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
	"testing"

	"github.com/longxi-go/gosr/pkg/codegen/artifact"
)

func TestOutputSetPreflightPreservesEarlierFiles(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "a.go")
	last := filepath.Join(dir, "z.go")
	original := generatedHeader + "package original\n"
	if err := os.WriteFile(first, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "target.go")
	if err := os.WriteFile(target, []byte(generatedHeader+"package target\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, last); err != nil {
		t.Fatal(err)
	}
	g := generation{
		validated: true,
		outputs: []output{
			{
				file: File{
					Path: first,
				},
				File: artifact.File{
					Content: []byte(generatedHeader + "package changed\n"),
				},
			},
			{
				file: File{
					Path: last,
				},
				File: artifact.File{
					Content: []byte(generatedHeader + "package last\n"),
				},
			},
		},
	}
	if err := finishGeneration(context.Background(), &g); err == nil {
		t.Fatal("symlink destination accepted")
	}
	current, err := os.ReadFile(first)
	if err != nil || string(current) != original {
		t.Fatal("changed earlier output before set preflight", err)
	}
}
