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

import "github.com/longxi-go/gosr/pkg/codegen/artifact"

// output associates a shared artifact with nemo's domain-specific report.
type output struct {
	file File
	artifact.File
}

func outputArtifacts(outputs []output) (*artifact.Set, error) {
	files := make([]artifact.File, 0, len(outputs))
	for _, out := range outputs {
		file := out.File
		file.Path = out.file.Path
		files = append(files, file)
	}
	return artifact.New(files...)
}
