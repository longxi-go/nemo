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

package codec

import (
	"bytes"
	"errors"
	"io"

	"github.com/longxi-go/nemo/pkg/maptypes"
	"go.yaml.in/yaml/v3"
)

func decodeYAML(data []byte) (maptypes.ValueMap, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		if errors.Is(err, io.EOF) {
			return maptypes.ValueMap{}, nil
		}
		return nil, decodeError(err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, decodeError(err)
	}
	if err := validateYAMLNumbers(document.Content[0], 0); err != nil {
		return nil, err
	}
	var input any
	if err := document.Decode(&input); err != nil {
		return nil, decodeError(err)
	}
	if input == nil {
		return maptypes.ValueMap{}, nil
	}
	return normalize(input)
}
