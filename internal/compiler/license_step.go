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
	"fmt"
	"os"
	"path/filepath"

	"github.com/longxi-go/golayout/pkg/header"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	sourceHeaderDirectory = "codequality"
	sourceHeaderFileName  = "HEADER"
)

// A workspace can generate several modules at once. Cache notices by the
// output owner's module directory so the generator's own notice cannot leak.
func licenseGeneration(ctx context.Context, g *generation) error {
	modules := make(maptypes.StringMap[string], len(g.analysis.packages))
	for _, pkg := range g.analysis.packages {
		if pkg.Module != nil {
			modules[pkg.PkgPath] = pkg.Module.Dir
		}
	}
	notices := make(maptypes.StringBytesMap)
	for i := range g.outputs {
		if err := ctx.Err(); err != nil {
			return err
		}
		out := &g.outputs[i]
		dir := modules[out.file.Package]
		if dir == "" {
			return fmt.Errorf("source header %s: output module is unavailable", out.file.Path)
		}
		notice, read := notices[dir]
		if !read {
			var err error
			notice, err = os.ReadFile(filepath.Join(dir, sourceHeaderDirectory, sourceHeaderFileName))
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			notices[dir] = notice
		}
		if notice == nil {
			continue
		}
		updated, err := header.Source(out.Content, notice)
		if err != nil {
			return fmt.Errorf("source header %s: %w", out.file.Path, err)
		}
		out.Content = updated
	}
	return nil
}
