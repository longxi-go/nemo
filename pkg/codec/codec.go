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
	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// Decoder turns one document into normalized portable values, without transport ownership.
type (
	Decoder interface {
		Decode([]byte) (maptypes.ValueMap, error)
	}

	// DecoderFunc adapts a function to Decoder.
	DecoderFunc func([]byte) (maptypes.ValueMap, error)
)

// Decode calls the function, or returns InvalidValue if it is nil.
func (f DecoderFunc) Decode(b []byte) (maptypes.ValueMap, error) {
	if f == nil {
		return nil, decodeError(nil)
	}
	return f(b)
}

func decodeError(cause error) error {
	return &gosr.Error{
		Code:    gosr.InvalidValue,
		Message: "cannot decode configuration document",
		Cause:   cause,
	}
}
