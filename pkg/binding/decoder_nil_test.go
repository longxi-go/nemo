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
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

type nilDecoderCall = func(gosr.Value) error

func TestConvertersRejectNilDecoder(t *testing.T) {
	cases := []struct {
		name   string
		call   nilDecoderCall
		values []gosr.Value
	}{
		{
			name: "pointer",
			call: func(value gosr.Value) error {
				_, err := binding.Pointer[int](nil, value)
				return err
			},
			values: []gosr.Value{gosr.String("1")},
		},
		{
			name: "list",
			call: func(value gosr.Value) error {
				_, err := binding.List[int](nil, value)
				return err
			},
			values: []gosr.Value{gosr.List(), gosr.List(gosr.String("1")), gosr.String(""), gosr.String("1")},
		},
		{
			name: "map",
			call: func(value gosr.Value) error {
				_, err := binding.Map[int](nil, value)
				return err
			},
			values: []gosr.Value{gosr.Object(nil), gosr.Object(maptypes.ValueMap{
				"one": gosr.String("1"),
			})},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("nil decoder panicked: %v", recovered)
				}
			}()
			for _, value := range tc.values {
				err := tc.call(value)
				if !errors.Is(err, &gosr.Error{
					Code: gosr.InvalidState,
				}) {
					t.Errorf("value kind %v: got %v, want InvalidState", value.Kind(), err)
				}
			}
		})
	}
}
