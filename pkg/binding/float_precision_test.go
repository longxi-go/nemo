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

package binding_test

import (
	"errors"
	"strconv"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

type precisionFloat32 float32

type precisionFloat64 float64

func TestFloatDestinationPrecision(t *testing.T) {
	t.Run("float32", testFloat32Precision[float32])
	t.Run("named float32", testFloat32Precision[precisionFloat32])
	t.Run("float64", testFloat64Precision[float64])
	t.Run("named float64", testFloat64Precision[precisionFloat64])
}

func testFloat32Precision[T ~float32](t *testing.T) {
	t.Helper()
	for _, text := range []string{"1.0000000596046448", "1.0000000596046447753906250000000001"} {
		got, err := binding.Float[T](gosr.String(text))
		if err != nil || float64(got) != 1.00000011920928955078125 {
			t.Errorf("%s: got %.25g, error %v; want next float32 above 1", text, got, err)
		}
		negative, err := binding.Float[T](gosr.String("-" + text))
		if err != nil || float64(negative) != -1.00000011920928955078125 {
			t.Errorf("-%s: got %.25g, error %v; want next float32 below -1", text, negative, err)
		}
	}
	_, err := binding.Float[T](gosr.String("3.5e38"))
	if !errors.Is(err, strconv.ErrRange) {
		t.Errorf("float32 overflow: got %v, want preserved ErrRange", err)
	}
}

func testFloat64Precision[T ~float64](t *testing.T) {
	t.Helper()
	expectedValues := maptypes.StringMap[float64]{
		"1.0000000596046448":  1.0000000596046448,
		"-1.0000000596046448": -1.0000000596046448,
		"3.5e38":              3.5e38,
	}
	for text, want := range expectedValues {
		got, err := binding.Float[T](gosr.String(text))
		if err != nil || float64(got) != want {
			t.Errorf("%s: got %.25g, error %v; want %.25g", text, got, err, want)
		}
	}
	_, err := binding.Float[T](gosr.String("1e309"))
	if !errors.Is(err, strconv.ErrRange) {
		t.Errorf("float64 overflow: got %v, want preserved ErrRange", err)
	}
}
