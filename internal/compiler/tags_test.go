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

package compiler

import (
	"go/types"
	"testing"
)

func TestTags(t *testing.T) {
	for _, tag := range []string{`property:"a" required:"true"`, `json:"x,omitempty" default:""`} {
		if err := validateTags(tag); err != nil {
			t.Fatal(err)
		}
	}
	for _, tag := range []string{`property:a`, `property:"a" property:"b"`, `default:"unterminated`, `property:"x"default:"1"`} {
		if err := validateTags(tag); err == nil {
			t.Fatalf("accepted %s", tag)
		}
	}
}

func TestDefaults(t *testing.T) {
	for _, raw := range []string{"128", "not-a-number"} {
		if err := validateDefault(types.Typ[types.Int8], raw, types.SizesFor("gc", "amd64")); err == nil {
			t.Fatal(raw)
		}
	}
	if err := validateDefault(types.NewSlice(types.Typ[types.Int]), "1,2", types.SizesFor("gc", "amd64")); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultTargetSizes(t *testing.T) {
	for _, test := range []struct {
		name string
		typ  types.Type
		raw  string
	}{
		{
			"int",
			types.Typ[types.Int],
			"2147483648",
		},
		{
			"uint",
			types.Typ[types.Uint],
			"4294967296",
		},
		{
			"uintptr",
			types.Typ[types.Uintptr],
			"4294967296",
		},
		{
			"pointer",
			types.NewPointer(types.Typ[types.Int]),
			"2147483648",
		},
		{
			"list",
			types.NewSlice(types.Typ[types.Int]),
			"1,2147483648",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := validateDefault(test.typ, test.raw, types.SizesFor("gc", "386")); err == nil {
				t.Fatal("32-bit target accepted overflowing default")
			}
			if err := validateDefault(test.typ, test.raw, types.SizesFor("gc", "amd64")); err != nil {
				t.Fatalf("64-bit target rejected representable default: %v", err)
			}
		})
	}
}
