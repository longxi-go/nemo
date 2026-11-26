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

	"github.com/longxi-go/nemo/pkg/maptypes"
	"golang.org/x/tools/go/packages"
)

const (
	discoveryOverlaySource = "package nemoDiscovery\n"
)

// discoveryOverlay prevents Go's own directory discovery from opening a pipe or
// device at the generated filename. It does not decide which packages match:
// go/packages remains authoritative, and selected packages are checked against
// the original destination after discovery. Unmatched paths remain in-memory
// placeholders during later loads.
func discoveryOverlay(ctx context.Context, options Options) (maptypes.StringBytesMap, error) {
	selected, err := discoveryRoots(ctx, options)
	if err != nil {
		return nil, err
	}
	overlay := make(maptypes.StringBytesMap)
	for _, root := range selected {
		if err := maskDiscoveryTree(ctx, root, root.dir, overlay); err != nil {
			return nil, err
		}
	}
	return overlay, nil
}

func maskDiscoveryTree(ctx context.Context, root discoveryRoot, dir string, overlay maptypes.StringBytesMap) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if dir != root.dir {
		if _, err := os.Stat(filepath.Join(dir, moduleFileName)); err == nil {
			return nil
		}
	}
	path := filepath.Join(dir, generatedFileName)
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		overlay[path] = []byte(discoveryOverlaySource)
	}
	if !root.recursive {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		// An unmatched wildcard subtree must not become a new error source. Go
		// reports directory-access failures for the packages it actually selects.
		return nil
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() ||
			strings.HasPrefix(name, hiddenDirectoryPrefix) ||
			strings.HasPrefix(name, ignoredDirectoryPrefix) || name == testdataDirectoryName {
			continue
		}
		child := root
		if name == vendorDirectoryName && !root.vendor {
			child.recursive = false
		}
		if err := maskDiscoveryTree(ctx, child, filepath.Join(dir, name), overlay); err != nil {
			return err
		}
	}
	return nil
}

// A safety placeholder cannot authorize an imported dependency's real special
// file. Check only packages actually loaded, leaving unrelated masked paths alone.
func checkLoadedOutputs(packagesLoaded []*packages.Package, overlay maptypes.StringBytesMap) error {
	var failure error
	packages.Visit(packagesLoaded, func(pkg *packages.Package) bool {
		if failure != nil {
			return false
		}
		for _, path := range pkg.GoFiles {
			if _, masked := overlay[path]; !masked {
				continue
			}
			if _, _, err := readGeneratedOutput(path); err != nil {
				failure = err
				return false
			}
		}
		return true
	}, nil)
	return failure
}
