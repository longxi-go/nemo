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
	"fmt"
	"math"
	"strconv"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

// conversionFailure keeps application-provided parsing details reachable only
// through Unwrap: Error's printable form must never include source values.
func conversionFailure(cause error) error {
	if cause == nil {
		return nil
	}
	return &gosr.Error{
		Code:    gosr.ConversionFailed,
		Message: "typed conversion failed",
		Cause:   cause,
	}
}

func scalar(v gosr.Value) (string, error) {
	s, ok := v.Scalar()
	if !ok {
		return "", conversionFailure(fmt.Errorf("expected scalar"))
	}
	return s, nil
}

// String accepts only strings, preventing implicit object or null coercion.
func String[T ~string](v gosr.Value) (T, error) {
	if v.Kind() != gosr.StringKind {
		return "", conversionFailure(fmt.Errorf("expected string"))
	}
	s, _ := v.Scalar()
	return T(s), nil
}

// Boolean parses canonical and strconv boolean spellings.
func Boolean[T ~bool](v gosr.Value) (T, error) {
	s, err := scalar(v)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseBool(s)
	return T(b), conversionFailure(err)
}

// Signed checks destination width before conversion, including named integer types.
func Signed[T ~int | ~int8 | ~int16 | ~int32 | ~int64](v gosr.Value) (T, error) {
	s, err := scalar(v)
	if err != nil {
		return 0, conversionFailure(err)
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, conversionFailure(err)
	}
	result := T(n)
	if int64(result) != n {
		return 0, conversionFailure(strconv.ErrRange)
	}
	return result, nil
}

// Unsigned rejects negatives and overflow before publishing a typed value.
func Unsigned[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr](v gosr.Value) (T, error) {
	s, err := scalar(v)
	if err != nil {
		return 0, conversionFailure(err)
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, conversionFailure(err)
	}
	result := T(n)
	if uint64(result) != n {
		return 0, conversionFailure(strconv.ErrRange)
	}
	return result, nil
}

// Float rejects overflow and nonfinite values for both supported widths.
func Float[T ~float32 | ~float64](v gosr.Value) (T, error) {
	s, err := scalar(v)
	if err != nil {
		return 0, conversionFailure(err)
	}
	bitSize := 64
	// 2^24+1 is exact in float64 but rounds in float32. Conversion through T
	// also detects named types without reflection or an unsafe size query.
	const precisionProbe = 1<<24 + 1
	if float64(T(precisionProbe)) != precisionProbe {
		bitSize = 32
	}
	n, err := strconv.ParseFloat(s, bitSize)
	result := T(n)
	if err != nil {
		return 0, conversionFailure(err)
	}
	if math.IsNaN(float64(result)) || math.IsInf(float64(result), 0) {
		return 0, conversionFailure(fmt.Errorf("expected finite float"))
	}
	return result, nil
}

// Duration uses Go duration syntax rather than treating durations as bare integers.
func Duration(v gosr.Value) (time.Duration, error) {
	s, err := scalar(v)
	if err != nil {
		return 0, conversionFailure(err)
	}
	value, err := time.ParseDuration(s)
	return value, conversionFailure(err)
}

// Text delegates custom parsing to a fresh typed value, never an existing aliased pointer.
func Text[T any, P TextUnmarshalerPointer[T]](v gosr.Value) (T, error) {
	var value T
	s, err := scalar(v)
	if err != nil {
		return value, conversionFailure(err)
	}
	err = P(&value).UnmarshalText([]byte(s))
	return value, conversionFailure(err)
}

// Pointer allocates independent storage only after successful conversion.
func Pointer[T any](decode gosr.ValueDecoderFunc[T], v gosr.Value) (*T, error) {
	if decode == nil {
		return nil, &gosr.Error{
			Code: gosr.InvalidState,
		}
	}
	value, err := decode(v)
	if err != nil {
		return nil, conversionFailure(err)
	}
	return &value, nil
}
