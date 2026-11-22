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

package resolution

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSyntaxScansObserveCancellation(t *testing.T) {
	for _, scanningBounds := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		p := parser{
			check: func() error {
				calls++
				if calls == 3 {
					cancel()
				}
				return ctx.Err()
			},
		}
		var err error
		text := strings.Repeat("x", 8192)
		if scanningBounds {
			_, _, err = p.bounds(text+"}", 0)
		} else {
			err = p.validateSyntax(text, 0)
		}
		cancel()
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("scan bounds=%t ignored cancellation: %v", scanningBounds, err)
		}
	}
}
