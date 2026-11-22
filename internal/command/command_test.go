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

package command

import (
	"context"
	"io"
	"testing"
)

func TestArguments(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{
			"other",
		},
		{
			"generate",
			"--roots",
			",",
		},
		{
			"check",
			"--unknown",
		},
		{
			"generate",
			"unexpected",
		},
	} {
		if err := Run(context.Background(), args, io.Discard); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	var roots rootFlags
	if err := roots.Set("./a,..."); err != nil || len(roots) != 2 || roots.String() != "./a,..." {
		t.Fatal(roots, err)
	}
}
