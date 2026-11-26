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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/longxi-go/gosr/tools/pkg/annotation"
	"github.com/longxi-go/gosr/tools/pkg/packageload"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"golang.org/x/tools/go/packages"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

const (
	goEnvCommand                              = "env"
	moduleCacheVariable                       = "GOMODCACHE"
	locateModuleCacheFormat                   = "locate module cache: %w"
	packageIsADependencyGenerateItInItsFormat = "package %s is a dependency; generate it in its owning module"
	packageNoGoFilesFormat                    = "package %s: no Go files"
	refusingOutputInsideModuleCacheFormat     = "refusing output inside module cache: %s"
)

func analyze(ctx context.Context, o Options) (analysis, error) {
	a := analysis{
		declarations: make(maptypes.StringMap[[]declaration]),
		owned:        make(maptypes.StringBoolMap),
		overlay:      make(maptypes.StringBytesMap),
	}
	var err error
	a.overlay, err = discoveryOverlay(ctx, o)
	if err != nil {
		return a, err
	}
	// Discovery deliberately requests no type information: first-generation callers
	// can refer to methods that are provided only by the forthcoming overlay.
	config := loadConfig(ctx, o, a.overlay)
	config.Mode = packages.NeedModule | packages.NeedName | packages.NeedFiles
	initial, err := packageload.Load(config, roots(o)...)
	if err != nil {
		return a, err
	}
	definition, err := annotation.New(annotationPrefix, []annotation.Definition{
		{
			Name:   configPrefixMarker,
			Target: markers.DescribesType,
			Sample: "",
		},
	})
	if err != nil {
		return a, err
	}
	for _, pkg := range initial {
		if len(pkg.GoFiles) == 0 {
			return a, fmt.Errorf(packageNoGoFilesFormat, pkg.PkgPath)
		}
		dir := filepath.Dir(pkg.GoFiles[0])
		generated := filepath.Join(dir, generatedFileName)
		_, owned, err := readGeneratedOutput(generated)
		if err != nil {
			return a, err
		}
		a.owned[pkg.PkgPath] = owned

		source, err := scanDeclarations(pkg.GoFiles, definition)
		if err != nil {
			return a, err
		}
		a.declarations[pkg.PkgPath] = source.declarations
		if len(source.declarations) > 0 || a.owned[pkg.PkgPath] {
			if pkg.Module != nil && !pkg.Module.Main {
				return a, fmt.Errorf(packageIsADependencyGenerateItInItsFormat, pkg.PkgPath)
			}
			cache, err := exec.CommandContext(ctx, goExecutable, goEnvCommand, moduleCacheVariable).Output()
			if err != nil {
				return a, fmt.Errorf(locateModuleCacheFormat, err)
			}
			relative, err := filepath.Rel(strings.TrimSpace(string(cache)), dir)
			if err != nil {
				return a, err
			}
			outsidePrefix := parentDirectory + string(filepath.Separator)
			if relative == currentDirectory ||
				(relative != parentDirectory && !strings.HasPrefix(relative, outsidePrefix)) {
				return a, fmt.Errorf(refusingOutputInsideModuleCacheFormat, dir)
			}

			a.overlay[generated] = bootstrapSource(pkg.Name, source)
		}
	}
	a.packages, err = packageload.Load(loadConfig(ctx, o, a.overlay), roots(o)...)
	if err != nil {
		return a, err
	}
	if err = checkLoadedOutputs(a.packages, a.overlay); err != nil {
		return a, err
	}
	if err = packageErrors(a.packages); err != nil {
		return a, err
	}
	return a, nil
}
