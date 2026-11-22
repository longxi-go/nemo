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
	"encoding/json"
	"errors"
	"io"

	"github.com/longxi-go/nemo/pkg/maptypes"
)

// Keep delimiters typed: decoder.Token returns json.Delim inside an interface.
const (
	jsonObjectStart = json.Delim('{')
	jsonArrayStart  = json.Delim('[')
)

func decodeJSON(data []byte) (maptypes.ValueMap, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	input, err := jsonValue(decoder, 0)
	if err != nil {
		return nil, decodeError(err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, decodeError(err)
	}
	return normalize(input)
}

// Token parsing rejects duplicate object members before a map could erase them.
func jsonValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > maxDocumentDepth {
		return nil, decodeError(nil)
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch token {
	case jsonObjectStart:
		object := make(maptypes.StringAnyMap)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, decodeError(nil)
			}
			if _, exists := object[key]; exists {
				return nil, decodeError(nil)
			}
			child, err := jsonValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			object[key] = child
		}
		_, err := decoder.Token()
		return object, err
	case jsonArrayStart:
		list := []any{}
		for decoder.More() {
			child, err := jsonValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			list = append(list, child)
		}
		_, err := decoder.Token()
		return list, err
	default:
		return token, nil
	}
}
