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

package codegen

import (
	"context"

	"github.com/longxi-go/nemo/internal/compiler"
)

// Options selects caller-owned packages; Check rejects stale output without writes.
type Options struct {
	Dir   string
	Roots []string
	Check bool
}

// File describes a selected package output and its annotated types.
type File struct {
	Path    string
	Package string
	Types   []string
}

// Report identifies outputs after complete generated-source validation.
type Report struct {
	Files []File
}

// Run generates only after all selected packages pass full overlay type validation.
func Run(ctx context.Context, options Options) (Report, error) {
	return run(ctx, options, false)
}

// Inspect validates and describes prospective output without mutating source or modules.
func Inspect(ctx context.Context, options Options) (Report, error) {
	return run(ctx, options, true)
}

func run(ctx context.Context, options Options, inspect bool) (Report, error) {
	result, err := compiler.Run(ctx, compiler.Options{
		Dir:   options.Dir,
		Roots: options.Roots,
		Check: options.Check,
	}, inspect)
	report := Report{
		Files: make([]File, 0, len(result.Files)),
	}
	for _, file := range result.Files {
		report.Files = append(report.Files, File{
			Path:    file.Path,
			Package: file.Package,
			Types:   file.Types,
		})
	}
	return report, err
}
