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
	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// PointerDecoder composes typed conversion for generated and handwritten binders.
// A nil input remains nil so Read can diagnose an invalid binding configuration.
func PointerDecoder[T any](decode gosr.ValueDecoderFunc[T]) gosr.ValueDecoderFunc[*T] {
	if decode == nil {
		return nil
	}
	return func(value gosr.Value) (*T, error) {
		return Pointer(decode, value)
	}
}

// ListDecoder adapts one element decoder without duplicating wrappers in generated code.
// A nil input remains nil.
func ListDecoder[T any](decode gosr.ValueDecoderFunc[T]) gosr.ValueDecoderFunc[[]T] {
	if decode == nil {
		return nil
	}
	return func(value gosr.Value) ([]T, error) {
		return List(decode, value)
	}
}

// MapDecoder composes object-member conversion while retaining Map's error policy.
// A nil input remains nil.
func MapDecoder[T any](decode gosr.ValueDecoderFunc[T]) gosr.ValueDecoderFunc[maptypes.StringMap[T]] {
	if decode == nil {
		return nil
	}
	return func(value gosr.Value) (maptypes.StringMap[T], error) {
		return Map(decode, value)
	}
}
