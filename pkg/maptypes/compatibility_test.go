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

package maptypes_test

import (
	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

// Function signatures require identical types, beyond map assignability.
// Keep the raw forms independent so changing an alias into a defined type
// cannot silently break existing source constructors or decoder implementations.
type (
	rawSourceFunc = func(string, map[string]gosr.Value) (gosr.PropertySource, error)

	rawPropertiesFunc = func(string, map[string]gosr.Property) (gosr.PropertySource, error)

	rawMapDecoderFunc = func(gosr.Value) (map[string]int, error)

	rawStringMapFunc = func(map[string]int)

	rawIntMapFunc = func(map[int]string)

	rawMapFunc = func(map[[2]int]string)

	rawBoolMapFunc = func(map[string]bool)

	rawBytesMapFunc = func(map[string][]byte)

	rawAnyMapFunc = func(map[string]any)
)

var (
	_ rawSourceFunc     = source.NewMap
	_ rawPropertiesFunc = source.NewProperties
	_ rawMapDecoderFunc = binding.MapDecoder(binding.Signed[int])
	_ rawStringMapFunc  = acceptStringMap
	_ rawIntMapFunc     = acceptIntMap
	_ rawMapFunc        = acceptMap
	_ rawBoolMapFunc    = acceptBoolMap
	_ rawBytesMapFunc   = acceptBytesMap
	_ rawAnyMapFunc     = acceptAnyMap
	_ codec.Decoder     = rawDecoder{}
)

func acceptStringMap(maptypes.StringMap[int]) {}

func acceptIntMap(maptypes.IntMap[string]) {}

func acceptMap(maptypes.Map[[2]int, string]) {}

func acceptBoolMap(maptypes.StringBoolMap) {}

func acceptBytesMap(maptypes.StringBytesMap) {}

func acceptAnyMap(maptypes.StringAnyMap) {}

type rawDecoder struct{}

func (rawDecoder) Decode([]byte) (map[string]gosr.Value, error) {
	return maptypes.ValueMap{
		"app.name": gosr.String("demo"),
	}, nil
}
