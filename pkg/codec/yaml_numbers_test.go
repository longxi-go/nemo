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

package codec_test

import (
	"testing"

	"github.com/longxi-go/nemo/pkg/codec"
)

func TestYAMLExplicitFloatAndInteger(t *testing.T) {
	decoder, err := codec.ForFormat(codec.FormatYAML)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name, scalar string
		valid        bool
	}{
		{
			"explicit-float",
			"!!float 18446744073709551616",
			true,
		},
		{
			"scientific",
			"1.8446744073709552e19",
			true,
		},
		{
			"quoted-large-text",
			"\"18446744073709551616\"",
			true,
		},
		{
			"explicit-string",
			"!!str 18446744073709551616",
			true,
		},
		{
			"float-alias",
			"&number !!float 18446744073709551616",
			true,
		},
		{
			"uint64",
			"18446744073709551615",
			true,
		},
		{
			"int64-min",
			"-9223372036854775808",
			true,
		},
		{
			"implicit-overflow",
			"18446744073709551616",
			false,
		},
		{
			"hex-overflow",
			"0x10000000000000000",
			false,
		},
		{
			"explicit-int-overflow",
			"!!int 18446744073709551616",
			false,
		},
		{
			"infinity",
			".inf",
			false,
		},
		{
			"nan",
			".nan",
			false,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			_, err := decoder.Decode([]byte("value: " + scenario.scalar + "\n"))
			if (err == nil) != scenario.valid {
				t.Fatalf("valid=%t error=%v", scenario.valid, err)
			}
		})
	}
}
