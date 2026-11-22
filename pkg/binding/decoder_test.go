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
	"slices"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestDecoderComposition(t *testing.T) {
	value := gosr.String("42")
	decodePointer := binding.PointerDecoder(binding.Signed[int])
	first, err := decodePointer(value)
	if err != nil || first == nil || *first != 42 {
		t.Fatal(first, err)
	}
	second, err := decodePointer(value)
	if err != nil || second == first {
		t.Fatal("pointer storage was shared", err)
	}
	decodeList := binding.ListDecoder(binding.Signed[int])
	list, err := decodeList(gosr.String("1,2"))
	if err != nil || !slices.Equal(list, []int{1, 2}) {
		t.Fatal(list, err)
	}
	decodeMap := binding.MapDecoder(binding.Signed[int])
	object, err := decodeMap(gosr.Object(maptypes.ValueMap{
		"port": value,
	}))
	if err != nil || object["port"] != 42 {
		t.Fatal(object, err)
	}
	sentinel := errors.New("custom conversion")
	decodeFailure := func(gosr.Value) (int, error) {
		return 0, sentinel
	}
	if _, err := binding.PointerDecoder(decodeFailure)(value); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if _, err := binding.ListDecoder(decodeFailure)(gosr.List(value)); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if _, err := binding.MapDecoder(decodeFailure)(gosr.Object(maptypes.ValueMap{
		"bad": value,
	})); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
}

func TestNilDecoderComposition(t *testing.T) {
	if binding.PointerDecoder[int](nil) != nil || binding.ListDecoder[int](nil) != nil || binding.MapDecoder[int](nil) != nil {
		t.Fatal("nil decoder must remain diagnosable")
	}
}

func TestReadRejectsNilDecoder(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"present": gosr.String("value"),
	})
	_, err := binding.Read[string](r, "present", binding.ReadOptions{}, nil)
	if !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatal(err)
	}
	_, err = binding.Read(r, "present", binding.ReadOptions{}, binding.ListDecoder[int](nil))
	if !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatal("composed nil decoder must retain the binding diagnostic", err)
	}
}
