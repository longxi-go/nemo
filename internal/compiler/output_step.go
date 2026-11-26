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
	"errors"
	"fmt"

	fileoutput "github.com/longxi-go/gosr/pkg/codegen/output"
)

const (
	generatedOutputIsStaleFormat                   = "generated output is stale: %s"
	outputRequiresValidatedGeneratedOverlayMessage = "output requires validated generated overlay"
)

func finishGeneration(ctx context.Context, g *generation) error {
	if !g.validated {
		return errors.New(outputRequiresValidatedGeneratedOverlayMessage)
	}
	for _, out := range g.outputs {
		g.report.Files = append(g.report.Files, out.file)
	}
	if g.inspect {
		return nil
	}
	for _, out := range g.outputs {
		old, _, err := readGeneratedOutput(out.file.Path)
		if err != nil {
			return err
		}
		if bytes.Equal(old, out.Content) {
			continue
		}
		if g.options.Check {
			return fmt.Errorf(generatedOutputIsStaleFormat, out.file.Path)
		}
	}
	if g.options.Check {
		return nil
	}
	set, err := outputArtifacts(g.outputs)
	if err != nil {
		return err
	}
	return fileoutput.Commit(ctx, set, fileoutput.Options{
		Owns: ownsGeneratedOutput,
	})
}

func ownsGeneratedOutput(_ string, current []byte) bool {
	return bytes.HasPrefix(current, []byte(generatedHeader))
}
