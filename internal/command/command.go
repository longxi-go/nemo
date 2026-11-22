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

package command

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/longxi-go/nemo/pkg/codegen"
)

const (
	generateCommand = "generate"
	checkCommand    = "check"
)

// Run implements a testable CLI with caller-owned cancellation and diagnostics.
func Run(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 || (args[0] != generateCommand && args[0] != checkCommand) {
		return fmt.Errorf("usage: nemoctl generate|check [--roots pattern] [--dir directory]")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(out)
	var roots rootFlags
	var options codegen.Options
	flags.Var(&roots, "roots", "package patterns, repeatable or comma-separated")
	flags.StringVar(&options.Dir, "dir", "", "module working directory")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments")
	}
	options.Roots = roots
	options.Check = args[0] == checkCommand
	report, err := codegen.Run(ctx, options)
	if err != nil {
		return err
	}
	for _, file := range report.Files {
		if _, err := fmt.Fprintln(out, file.Path); err != nil {
			return err
		}
	}
	return nil
}
