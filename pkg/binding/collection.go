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

package binding

import (
	"encoding/csv"
	"fmt"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/propertypath"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// List accepts portable lists or a single CSV record for scalar defaults.
func List[T any](decode gosr.ValueDecoderFunc[T], v gosr.Value) ([]T, error) {
	if decode == nil {
		return nil, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	var items []gosr.Value
	switch v.Kind() {
	case gosr.ListKind:
		items = v.Items()
	case gosr.StringKind:
		s, _ := v.Scalar()
		if s == "" {
			return []T{}, nil
		}
		fields, err := csv.NewReader(strings.NewReader(s)).ReadAll()
		if err != nil {
			return nil, conversionFailure(err)
		}
		if len(fields) != 1 {
			return nil, conversionFailure(fmt.Errorf("expected one CSV record"))
		}
		for _, field := range fields[0] {
			items = append(items, gosr.String(field))
		}
	default:
		return nil, conversionFailure(fmt.Errorf("expected list"))
	}
	out := make([]T, 0, len(items))
	for _, item := range items {
		value, err := decode(item)
		if err != nil {
			return nil, conversionFailure(err)
		}
		out = append(out, value)
	}
	return out, nil
}

// Map decodes object members without reflection.
func Map[T any](decode gosr.ValueDecoderFunc[T], v gosr.Value) (maptypes.StringMap[T], error) {
	if decode == nil {
		return nil, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	if v.Kind() != gosr.ObjectKind {
		return nil, conversionFailure(fmt.Errorf("expected object"))
	}
	out := make(maptypes.StringMap[T])
	for key, item := range v.Fields() {
		value, err := decode(item)
		if err != nil {
			return nil, conversionFailure(err)
		}
		out[key] = value
	}
	return out, nil
}

// ReadMap decodes the exact property when present. Otherwise it assembles
// descendant scalar properties, retaining their relative keys as map keys.
func ReadMap[T any](r gosr.PropertyResolver, key string, options ReadOptions, decode gosr.ValueDecoderFunc[T]) (maptypes.StringMap[T], error) {
	if r == nil {
		return nil, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	_, found, err := r.Lookup(key)
	if err != nil {
		return nil, err
	}
	if found {
		return Read(r, key, options, MapDecoder(decode))
	}
	out := make(maptypes.StringMap[T])
	for _, k := range r.Keys() {
		if !strings.HasPrefix(k, key+propertypath.Separator) {
			continue
		}
		member := strings.TrimPrefix(k, key+propertypath.Separator)
		value, err := Read(r, k, ReadOptions{}, decode)
		if err != nil {
			return nil, err
		}
		out[member] = value
	}
	if len(out) > 0 {
		return out, nil
	}
	return Read(r, key, options, MapDecoder(decode))
}
