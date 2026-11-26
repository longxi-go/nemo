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

	"github.com/longxi-go/gosr/tools/pkg/packageload"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"golang.org/x/tools/go/packages"
)

const (
	phaseTypeValidation = "type validation"
)

func loadConfig(ctx context.Context, o Options, overlay maptypes.StringBytesMap) *packages.Config {
	config := packageload.Config(ctx, o.Dir, overlay)
	config.Mode = packages.NeedModule | packages.NeedName | packages.NeedFiles |
		packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes |
		packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedImports | packages.NeedDeps
	config.BuildFlags = []string{readOnlyModuleFlag}
	return config
}

func roots(o Options) []string {
	if len(o.Roots) == 0 {
		return []string{currentDirectory}
	}
	return o.Roots
}

func packageErrors(pkgs []*packages.Package) error {
	return packageload.Errors(phaseTypeValidation, pkgs)
}

func validate(ctx context.Context, o Options, overlay maptypes.StringBytesMap) error {
	pkgs, err := packageload.Load(loadConfig(ctx, o, overlay), roots(o)...)
	if err != nil {
		return err
	}
	if err := checkLoadedOutputs(pkgs, overlay); err != nil {
		return err
	}
	return packageErrors(pkgs)
}
