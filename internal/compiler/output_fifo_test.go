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

//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package compiler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestAllModesRejectFIFOOutputBeforeDiscovery(t *testing.T) {
	t.Setenv("GOWORK", "off")
	for _, mode := range []string{"generate", "check", "inspect"} {
		for _, linked := range []bool{false, true} {
			name := mode
			if linked {
				name += " symlink"
			}
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				writeDiscoveryFixture(t, dir, "fixture")
				output := filepath.Join(dir, generatedFileName)
				fifo := output
				if linked {
					fifo = filepath.Join(t.TempDir(), "pipe")
				}
				if err := syscall.Mkfifo(fifo, 0o600); err != nil {
					t.Fatal(err)
				}
				if linked {
					if err := os.Symlink(fifo, output); err != nil {
						t.Fatal(err)
					}
				}
				// The package loader's subprocess belongs to this context and is reaped
				// before Run returns, including when directory discovery blocks.
				ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
				defer cancel()
				_, err := Run(ctx, Options{
					Dir:   dir,
					Check: mode == "check",
				}, mode == "inspect")
				if err == nil || !strings.Contains(err.Error(), "not a regular file") {
					t.Fatalf("%s did not reject the unsafe destination before discovery: %v", mode, err)
				}
				if info, err := os.Lstat(fifo); err != nil || info.Mode()&os.ModeNamedPipe == 0 {
					t.Fatalf("%s changed FIFO destination: %v", mode, err)
				}
			})
		}
	}
}

func TestDiscoveryDoesNotRejectUnselectedFIFO(t *testing.T) {
	t.Setenv("GOWORK", "off")
	for _, root := range []string{"./chosen", "./.../chosen", "./..."} {
		t.Run(root, func(t *testing.T) {
			dir := t.TempDir()
			writeDiscoveryFixture(t, dir, "fixture")
			for _, child := range []string{"chosen", "testdata/ignored", "vendor/ignored", "_ignored", ".ignored", "nested"} {
				path := filepath.Join(dir, child)
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(path, "config.go"), []byte("package fixture\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				if child == "nested" {
					writeDiscoveryFixture(t, path, "nested")
				}
				if child != "chosen" {
					if err := syscall.Mkfifo(filepath.Join(path, generatedFileName), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			if root != "./..." {
				path := filepath.Join(dir, "unmatched")
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := syscall.Mkfifo(filepath.Join(path, generatedFileName), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if _, err := Run(ctx, Options{
				Dir:   dir,
				Roots: []string{root},
			}, true); err != nil {
				t.Fatalf("unselected FIFO affected discovery: %v", err)
			}
		})
	}
}

func writeDiscoveryFixture(t *testing.T, dir, module string) {
	t.Helper()
	fixtureFiles := maptypes.StringMap[string]{
		"go.mod":    "module " + module + "\ngo 1.27.0\n",
		"config.go": "package fixture\n",
	}
	for name, content := range fixtureFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDiscoveryResolvesWorkspaceAndImportRoots(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{"first", "second"} {
		dir := filepath.Join(base, name)
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeDiscoveryFixture(t, dir, "example.org/"+name)
	}
	workspace := filepath.Join(base, "go.work")
	if err := os.WriteFile(workspace, []byte("go 1.27.0\nuse (\n./first\n./second\n)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOWORK", workspace)
	if err := syscall.Mkfifo(filepath.Join(base, "second", generatedFileName), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name   string
		roots  []string
		unsafe bool
	}{
		{
			"local only",
			[]string{"./..."},
			false,
		},
		{
			"import only",
			[]string{"example.org/first/..."},
			false,
		},
		{
			"import selected",
			[]string{"example.org/second"},
			true,
		},
		{
			"import wildcard",
			[]string{"example.org/..."},
			true,
		},
		{
			"relative sibling",
			[]string{"../second"},
			true,
		},
		{
			"multiple roots",
			[]string{".", "../second"},
			true,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_, err := Run(ctx, Options{
				Dir:   filepath.Join(base, "first"),
				Roots: scenario.roots,
			}, true)
			if scenario.unsafe {
				if err == nil || !strings.Contains(err.Error(), "not a regular file") {
					t.Fatalf("selected unsafe workspace destination was not rejected: %v", err)
				}
			} else if err != nil {
				t.Fatalf("unselected workspace destination affected discovery: %v", err)
			}
		})
	}
}

func TestDiscoveryRejectsFIFOOnlyPackageWithoutBlocking(t *testing.T) {
	t.Setenv("GOWORK", "off")
	for _, pattern := range []string{".", "./..."} {
		t.Run(pattern, func(t *testing.T) {
			dir := t.TempDir()
			writeDiscoveryFixture(t, dir, "fixture")
			selected := dir
			if pattern == "./..." {
				selected = filepath.Join(dir, "only-output")
				if err := os.Mkdir(selected, 0o755); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Remove(filepath.Join(dir, "config.go")); err != nil {
				t.Fatal(err)
			}
			if err := syscall.Mkfifo(filepath.Join(selected, generatedFileName), 0o600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			defer cancel()
			if _, err := Run(ctx, Options{
				Dir:   dir,
				Roots: []string{pattern},
			}, true); err == nil || !strings.Contains(err.Error(), "not a regular file") {
				t.Fatalf("FIFO-only package was accepted or blocked until cancellation: %v", err)
			}
		})
	}
}

func TestDiscoveryMasksRequiredLocalDependencies(t *testing.T) {
	t.Setenv("GOWORK", "off")
	for _, scenario := range []struct {
		name     string
		root     string
		imported bool
	}{
		{
			"selected by all",
			"all",
			true,
		},
		{
			"required dependency",
			".",
			true,
		},
		{
			"unselected dependency",
			".",
			false,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			base := t.TempDir()
			for _, name := range []string{"main", "dependency"} {
				dir := filepath.Join(base, name)
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				writeDiscoveryFixture(t, dir, name+".review")
			}
			main := filepath.Join(base, "main")
			if err := os.WriteFile(filepath.Join(main, "go.mod"), []byte(`module main.review

go 1.27.0

require dependency.review v0.0.0

replace dependency.review => ../dependency
`), 0o600); err != nil {
				t.Fatal(err)
			}
			if scenario.imported {
				if err := os.WriteFile(filepath.Join(main, "config.go"), []byte("package fixture\nimport _ \"dependency.review\"\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := syscall.Mkfifo(filepath.Join(base, "dependency", generatedFileName), 0o600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_, err := Run(ctx, Options{
				Dir:   main,
				Roots: []string{scenario.root},
			}, true)
			if scenario.imported {
				if err == nil || !strings.Contains(err.Error(), "not a regular file") {
					t.Fatalf("unsafe required dependency blocked discovery: %v", err)
				}
			} else if err != nil {
				t.Fatalf("unselected local dependency affected discovery: %v", err)
			}
		})
	}
}
