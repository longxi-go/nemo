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

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestResolvedBudgetStopsBeforeLaterOperands(t *testing.T) {
	policy := limits.Default()
	policy.MaxResolvedBytes = 128
	items := make([]gosr.Value, 10)
	for i := range items {
		items[i] = gosr.String("${z}")
	}
	items = append(items, gosr.String("${missing}"))
	_, err := NewWithPolicy(context.Background(), maptypes.PropertyMap{
		"a": {
			Value: gosr.List(items...),
		},
		"z": {
			Value: gosr.String(strings.Repeat("x", 32)),
		},
	}, policy)
	var exceeded *limits.ExceededError
	if !errors.As(err, &exceeded) || exceeded.Resource != limits.ResolvedBytes {
		t.Fatalf("output budget must reject before evaluating the later missing operand: %v", err)
	}
}

func TestResolvedBudgetCountsCachedValuesAndDefaultsOnce(t *testing.T) {
	policy := limits.Default()
	// Property keys (2) plus z (3) plus both list items (6).
	policy.MaxResolvedBytes = 11
	raw := maptypes.PropertyMap{
		"a": {
			Value: gosr.List(gosr.String("${z}"), gosr.String("${missing:${z}}")),
		},
		"z": {
			Value: gosr.String("abc"),
		},
	}
	if _, err := NewWithPolicy(context.Background(), raw, policy); err != nil {
		t.Fatal("exact output budget rejected", err)
	}
	policy.MaxResolvedBytes--
	if _, err := NewWithPolicy(context.Background(), raw, policy); err == nil {
		t.Fatal("output overrun accepted")
	}
}
