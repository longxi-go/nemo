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

package environment_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	env "github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestDynamicKeyDiagnosticsAtEnvironmentBoundaries(t *testing.T) {
	const sentinel = "private_value_8f96"
	for _, entry := range []string{"resolve", "bootstrap", "refresh"} {
		t.Run(entry, func(t *testing.T) {
			values := maptypes.ValueMap{
				"selector": gosr.String(sentinel),
			}
			if entry == "bootstrap" {
				values["a.entry"] = gosr.String("${${selector}}")
			}
			provider := &provider{
				values: values,
			}
			owner, err := env.New(context.Background(), env.WithProviders(provider))
			var before gosr.Snapshot
			if entry != "bootstrap" {
				if err != nil {
					t.Fatal(err)
				}
				before = owner.Snapshot()
			}
			switch entry {
			case "resolve":
				text, resolveErr := owner.Resolve("${${selector}}")
				if text != "" {
					t.Fatal("failed resolution returned partial text")
				}
				err = resolveErr
			case "refresh":
				provider.set(maptypes.ValueMap{
					"selector": gosr.String(sentinel),
					"a.entry":  gosr.String("${${selector}}"),
				}, nil)
				err = owner.Refresh(context.Background())
			}
			if !errors.Is(err, &gosr.Error{
				Code: gosr.MissingProperty,
			}) ||
				strings.Contains(err.Error(), sentinel) {
				t.Fatalf("unsafe or incorrectly classified environment failure: %v", err)
			}
			if entry == "bootstrap" {
				if owner != nil {
					t.Fatal("failed bootstrap exposed an environment")
				}
				return
			}
			if owner.Snapshot() != before || owner.Snapshot().Revision() != before.Revision() {
				t.Fatal("failed resolution published a candidate snapshot")
			}
			value(t, owner, "selector", sentinel)
		})
	}
}
