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

	"github.com/longxi-go/gosr/pkg/pipeline"
)

const (
	stageAnalyze       = "analyze"
	stageModel         = "model"
	stageRender        = "render"
	stageSourceHeaders = "apply source headers"
	stageFormat        = "format Go"
	stageValidate      = "validate generated overlay"
	stageFinish        = "finish output"
)

func generationSteps() []pipeline.Step[*generation] {
	return []pipeline.Step[*generation]{
		pipeline.NewStep(stageAnalyze, analyzeGeneration),
		pipeline.NewStep(stageModel, modelGeneration),
		pipeline.NewStep(stageRender, renderGeneration),
		pipeline.NewStep(stageSourceHeaders, licenseGeneration),
		pipeline.NewStep(stageFormat, formatGeneration),
		pipeline.NewStep(stageValidate, validateGeneration),
		pipeline.NewStep(stageFinish, finishGeneration),
	}
}

// Run executes fixed private stages; Inspect follows identical validation without writes.
func Run(ctx context.Context, options Options, inspect bool) (Report, error) {
	g := generation{
		options: options,
		inspect: inspect,
	}
	err := pipeline.Run(ctx, &g, generationSteps()...)
	return g.report, err
}
