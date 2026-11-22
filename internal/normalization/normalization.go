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

package normalization

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/propertypath"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const maxNestingDepth = 64

type flattenFunc = func(string, any, int) error

// Flatten translates parser-owned maps into atomic dotted addresses. Objects
// inside lists stay objects; empty objects stay addressable values.
func Flatten(input any) (maptypes.ValueMap, error) {
	root, ok := input.(maptypes.StringAnyMap)
	if !ok {
		return nil, invalid()
	}
	result := make(maptypes.ValueMap)
	var walk flattenFunc
	walk = func(prefix string, input any, depth int) error {
		if depth > maxNestingDepth {
			return invalid()
		}
		if object, ok := input.(maptypes.StringAnyMap); ok && len(object) > 0 {
			for key, v := range object {
				if err := gosr.ValidateKey(key); err != nil {
					return err
				}
				path := key
				if prefix != "" {
					path = prefix + propertypath.Separator + key
				}
				if err := walk(path, v, depth+1); err != nil {
					return err
				}
			}
			return nil
		}
		if err := gosr.ValidateKey(prefix); err != nil {
			return err
		}
		if _, exists := result[prefix]; exists {
			return invalid()
		}
		v, err := value(input, depth)
		if err != nil {
			return err
		}
		result[prefix] = v
		return nil
	}
	for key, v := range root {
		if err := gosr.ValidateKey(key); err != nil {
			return nil, err
		}
		if err := walk(key, v, 1); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func invalid() error {
	return &gosr.Error{
		Code:    gosr.InvalidValue,
		Message: "invalid or ambiguous document value",
	}
}

func value(input any, depth int) (gosr.Value, error) {
	if depth > maxNestingDepth {
		return gosr.Value{}, invalid()
	}
	switch v := input.(type) {
	case nil:
		return gosr.Value{}, nil
	case string:
		return gosr.String(v), nil
	case bool:
		return gosr.Bool(v), nil
	case json.Number:
		return gosr.Number(string(v))
	case int:
		return gosr.Number(strconv.FormatInt(int64(v), 10))
	case int8:
		return gosr.Number(strconv.FormatInt(int64(v), 10))
	case int16:
		return gosr.Number(strconv.FormatInt(int64(v), 10))
	case int32:
		return gosr.Number(strconv.FormatInt(int64(v), 10))
	case int64:
		return gosr.Number(strconv.FormatInt(v, 10))
	case uint:
		return gosr.Number(strconv.FormatUint(uint64(v), 10))
	case uint8:
		return gosr.Number(strconv.FormatUint(uint64(v), 10))
	case uint16:
		return gosr.Number(strconv.FormatUint(uint64(v), 10))
	case uint32:
		return gosr.Number(strconv.FormatUint(uint64(v), 10))
	case uint64:
		return gosr.Number(strconv.FormatUint(v, 10))
	case float32:
		return floatValue(float64(v), 32)
	case float64:
		return floatValue(v, 64)
	case time.Time:
		return gosr.String(v.Format(time.RFC3339Nano)), nil
	case []maptypes.StringAnyMap:
		items := make([]gosr.Value, len(v))
		for i, child := range v {
			item, err := value(child, depth+1)
			if err != nil {
				return gosr.Value{}, err
			}
			items[i] = item
		}
		return gosr.List(items...), nil
	case []any:
		items := make([]gosr.Value, len(v))
		for i, child := range v {
			item, err := value(child, depth+1)
			if err != nil {
				return gosr.Value{}, err
			}
			items[i] = item
		}
		return gosr.List(items...), nil
	case maptypes.StringAnyMap:
		fields := make(maptypes.ValueMap, len(v))
		for key, child := range v {
			item, err := value(child, depth+1)
			if err != nil {
				return gosr.Value{}, err
			}
			fields[key] = item
		}
		return gosr.Object(fields), nil
	default:
		return gosr.Value{}, invalid()
	}
}

func floatValue(v float64, bits int) (gosr.Value, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return gosr.Value{}, invalid()
	}
	return gosr.Number(strconv.FormatFloat(v, 'g', -1, bits))
}
