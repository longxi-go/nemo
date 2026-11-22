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
	"fmt"
	"strings"
	"testing"

	"github.com/longxi-go/nemo/pkg/codec"
)

func TestDocumentDepthBoundary(t *testing.T) {
	for _, depth := range []int{63, 64, 65} {
		document := strings.Repeat(`{"child":`, depth) + `"leaf"` + strings.Repeat(`}`, depth)
		for _, format := range []string{codec.FormatJSON, codec.FormatYAML} {
			t.Run(fmt.Sprintf("%s/%d", format, depth), func(t *testing.T) {
				decoder, err := codec.ForFormat(format)
				if err != nil {
					t.Fatal(err)
				}
				_, err = decoder.Decode([]byte(document))
				if (err == nil) != (depth <= 64) {
					t.Fatalf("depth %d: %v", depth, err)
				}
			})
		}
	}
}

func TestDocumentMixedDepthBoundary(t *testing.T) {
	for _, depth := range []int{63, 64, 65} {
		// The property object is root depth zero; each list adds one value edge.
		document := `{"items":` + strings.Repeat("[", depth-1) + `"leaf"` + strings.Repeat("]", depth-1) + `}`
		for _, format := range []string{codec.FormatJSON, codec.FormatYAML} {
			decoder, _ := codec.ForFormat(format)
			_, err := decoder.Decode([]byte(document))
			if (err == nil) != (depth <= 64) {
				t.Fatalf("%s mixed depth %d: %v", format, depth, err)
			}
		}
		toml := strings.Repeat("child.", depth-1) + "value = 1"
		decoder, _ := codec.ForFormat(codec.FormatTOML)
		_, err := decoder.Decode([]byte(toml))
		if (err == nil) != (depth <= 64) {
			t.Fatalf("toml depth %d: %v", depth, err)
		}
	}
}
