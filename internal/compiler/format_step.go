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

	"github.com/longxi-go/golayout/pkg/format"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	formatOutputErrorFormat      = "format %s: %w"
	missingOutputGoVersionFormat = "format %s: target module Go version is unavailable for %s"
)

func formatGeneration(_ context.Context, g *generation) error {
	versions := make(maptypes.StringMap[string], len(g.analysis.packages))
	for _, pkg := range g.analysis.packages {
		if pkg.Module != nil && pkg.Module.GoVersion != "" {
			versions[pkg.PkgPath] = goLanguageVersionPrefix + pkg.Module.GoVersion
		}
	}
	for i := range g.outputs {
		out := &g.outputs[i]
		langVersion, ok := versions[out.file.Package]
		if !ok {
			return fmt.Errorf(missingOutputGoVersionFormat, out.file.Path, out.file.Package)
		}
		// A workspace may contain modules with different Go directives. Use the
		// output owner's version rather than the generator's language baseline.
		formatted, err := format.Source(out.Content, format.Options{
			LangVersion: langVersion,
		})
		if err != nil {
			return fmt.Errorf(formatOutputErrorFormat, out.file.Path, err)
		}
		out.GoVersion = langVersion
		out.Content = formatted
		g.analysis.overlay[out.file.Path] = formatted
	}
	return nil
}
