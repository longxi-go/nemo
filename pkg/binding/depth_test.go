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
)

func TestObjectDepthBoundary(t *testing.T) {
	for _, depth := range []int{0, 64} {
		if err := binding.CheckObjectDepth(depth); err != nil {
			t.Fatalf("depth %d: %v", depth, err)
		}
	}
	for _, depth := range []int{-1, 65} {
		if err := binding.CheckObjectDepth(depth); !errors.Is(err, &gosr.Error{
			Code: gosr.ConversionFailed,
		}) {
			t.Fatalf("depth %d: %v", depth, err)
		}
	}
	if binding.MaxObjectDepth != 64 {
		t.Fatal("binding budget changed")
	}
}
