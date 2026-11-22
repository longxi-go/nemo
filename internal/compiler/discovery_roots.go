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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

type discoveryRoot struct {
	dir       string
	recursive bool
	vendor    bool
}

type discoveryModule struct {
	Path    string
	Dir     string
	Version string
	Main    bool
	Replace *discoveryModule
}

// Only filesystem prefixes implied by requested roots are inspected. Module
// metadata resolves import-path roots without reading any Go source; wildcard
// matching itself remains the Go command's responsibility.
func discoveryRoots(ctx context.Context, options Options) ([]discoveryRoot, error) {
	base := options.Dir
	if base == "" {
		base = currentDirectory
	}
	base, err := filepath.Abs(base)
	if err != nil {
		return nil, err
	}
	var selected []discoveryRoot
	var imports []string
	for _, pattern := range roots(options) {
		switch {
		case strings.HasPrefix(pattern, fileQueryPrefix):
			path := strings.TrimPrefix(pattern, fileQueryPrefix)
			if !filepath.IsAbs(path) {
				path = filepath.Join(base, path)
			}
			selected = append(selected, discoveryRoot{
				dir: filepath.Dir(path),
			})
		case strings.HasSuffix(pattern, goSourceExtension) && !strings.Contains(pattern, packageWildcard):
			path := pattern
			if !filepath.IsAbs(path) {
				path = filepath.Join(base, path)
			}
			selected = append(selected, discoveryRoot{
				dir: filepath.Dir(path),
			})
		case build.IsLocalImport(pattern) || filepath.IsAbs(pattern):
			selected = append(selected, discoveryPatternRoot(base, pattern))
		case pattern == "std" || pattern == "cmd":
			continue
		default:
			imports = append(imports, pattern)
		}
	}
	if len(imports) == 0 && len(selected) == 0 {
		return selected, nil
	}
	modules, err := discoveryModules(ctx, options.Dir)
	if err != nil {
		return nil, err
	}
	// Local workspace and replacement modules may become dependencies during
	// the later type load, even when they are not generation destinations.
	for _, module := range modules {
		if module.Dir != "" && (module.Main || module.Replace != nil && module.Replace.Version == "") {
			selected = append(selected, discoveryRoot{
				dir:       module.Dir,
				recursive: true,
			})
		}
	}
	for _, pattern := range imports {
		prefix, _, wildcard := strings.Cut(pattern, packageWildcard)
		for _, module := range modules {
			if module.Dir == "" {
				continue
			}
			switch {
			case pattern == "all":
				selected = append(selected, discoveryRoot{
					dir:       module.Dir,
					recursive: true,
				})
			case pattern == module.Path || strings.HasPrefix(pattern, module.Path+importPathSeparator):
				suffix := strings.TrimPrefix(strings.TrimPrefix(pattern, module.Path), importPathSeparator)
				selected = append(selected, discoveryPatternRoot(module.Dir, filepath.FromSlash(suffix)))
			case wildcard && strings.HasPrefix(module.Path, prefix):
				selected = append(selected, discoveryRoot{
					dir:       module.Dir,
					recursive: true,
					vendor:    strings.Contains(pattern, vendorPathSegment),
				})
			}
		}
	}
	return selected, nil
}

func discoveryPatternRoot(base, pattern string) discoveryRoot {
	root := discoveryRoot{
		dir: pattern,
	}
	if !filepath.IsAbs(root.dir) {
		root.dir = filepath.Join(base, root.dir)
	}
	root.dir = filepath.Clean(root.dir)
	prefix, _, recursive := strings.Cut(pattern, packageWildcard)
	if recursive {
		dir, _ := filepath.Split(prefix)
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(base, dir)
		}
		root.dir = filepath.Clean(dir)
		root.recursive = true
		root.vendor = strings.Contains(filepath.ToSlash(pattern), vendorPathSegment) ||
			strings.HasSuffix(filepath.ToSlash(pattern), vendorPathSuffix)
	}
	return root
}

func discoveryModules(ctx context.Context, dir string) ([]discoveryModule, error) {
	command := exec.CommandContext(ctx, goExecutable, "list", "-m", "-json", readOnlyModuleFlag, "all")
	command.Dir = dir
	data, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("discover module roots: %w", errors.Join(ctx.Err(), err))
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var modules []discoveryModule
	for {
		var module discoveryModule
		if err := decoder.Decode(&module); err != nil {
			if errors.Is(err, io.EOF) {
				return modules, nil
			}
			return nil, fmt.Errorf("discover module roots: %w", err)
		}
		modules = append(modules, module)
	}
}
