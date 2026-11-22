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
	"embed"
	"strconv"
	"text/template"

	"github.com/longxi-go/gosr/pkg/codegen/artifact"
	"github.com/longxi-go/gosr/pkg/codegen/render"
)

const (
	templatePattern     = "templates/*.go.tmpl"
	packageTemplateName = "package.go.tmpl"
)

//go:embed templates/*.go.tmpl
var templateSources embed.FS

func renderModels(models []preparedPackage) ([]output, error) {
	t, err := render.ParseFS(templateSources, template.FuncMap{
		"quote":     strconv.Quote,
		"view":      objectView,
		"fieldView": fieldView,
	}, templatePattern)
	if err != nil {
		return nil, err
	}
	var outputs []output
	for _, prepared := range models {
		if err := prepared.model.validate(); err != nil {
			return nil, err
		}
		prepared.model.Header = generatedHeader
		data, err := t.Execute(packageTemplateName, prepared.model)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output{
			file: prepared.file,
			File: artifact.File{
				Path:    prepared.file.Path,
				Content: data,
				Kind:    artifact.Go,
			},
		})
	}
	return outputs, nil
}

type objectTemplateView struct {
	Names  generatedNames
	Helper helperModel
}

func objectView(names generatedNames, helper helperModel) objectTemplateView {
	return objectTemplateView{
		Names:  names,
		Helper: helper,
	}
}

type fieldTemplateView struct {
	Names generatedNames
	Field fieldModel
}

func fieldView(names generatedNames, field fieldModel) fieldTemplateView {
	return fieldTemplateView{
		Names: names,
		Field: field,
	}
}
