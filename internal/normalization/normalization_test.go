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

package normalization_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/normalization"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestFlatten(t *testing.T) {
	got, err := normalization.Flatten(maptypes.StringAnyMap{
		"app": maptypes.StringAnyMap{
			"ports": []any{
				int64(80),
				true,
				nil,
				maptypes.StringAnyMap{
					"x": "y",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["app.ports"].Kind() != gosr.ListKind {
		t.Fatal(got)
	}
}

func TestInvalid(t *testing.T) {
	for _, v := range []any{
		maptypes.StringAnyMap{
			"a.b": 1,
			"a": maptypes.StringAnyMap{
				"b": 2,
			},
		},
		map[any]any{1: "x"},
		maptypes.StringAnyMap{
			"x": math.Inf(1),
		},
		maptypes.StringAnyMap{
			"x": make(chan int),
		},
		[]any{1},
		maptypes.StringAnyMap{
			" ": 1,
		},
	} {
		if _, err := normalization.Flatten(v); err == nil {
			t.Fatal("invalid accepted")
		}
	}
}

func TestPortableNativeScalars(t *testing.T) {
	for _, input := range []any{int(1), int8(1), int16(1), int32(1), int64(1), uint(1), uint8(1), uint16(1), uint32(1), uint64(1), float32(1), float64(1), json.Number("1")} {
		values, err := normalization.Flatten(maptypes.StringAnyMap{
			"n": input,
		})
		if err != nil {
			t.Fatal(err)
		}
		text, _ := values["n"].Scalar()
		if values["n"].Kind() != gosr.NumberKind || text != "1" {
			t.Fatal(values)
		}
	}
	stamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	values, err := normalization.Flatten(maptypes.StringAnyMap{
		"date": stamp,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !values["date"].Equal(gosr.String("2026-01-02T03:04:05Z")) {
		t.Fatal(values)
	}
}

func TestCyclesAndNestedFailures(t *testing.T) {
	cycle := maptypes.StringAnyMap{}
	cycle["self"] = cycle
	for _, input := range []maptypes.StringAnyMap{
		cycle,
		{
			"list": []any{make(chan int)},
		},
		{
			"object": []any{
				maptypes.StringAnyMap{
					"bad": make(chan int),
				},
			},
		},
		{
			"list": []maptypes.StringAnyMap{
				{
					"bad": make(chan int),
				},
			},
		},
	} {
		if _, err := normalization.Flatten(input); err == nil {
			t.Fatal("invalid nested input accepted")
		}
	}
}
