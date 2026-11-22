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

package admission_test

import (
	"context"
	"errors"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/admission"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestCompositeBudgets(t *testing.T) {
	value := gosr.String("leaf")
	for range 80 {
		value = gosr.Object(maptypes.ValueMap{
			"child": value,
		})
	}
	budget := admission.New(context.Background(), limits.Default(), false)
	if err := budget.Property("app", value); !errors.Is(err, &gosr.Error{
		Code: gosr.ResolutionLimit,
	}) {
		t.Fatalf("depth error: %v", err)
	}
	p := limits.Default()
	p.MaxValueNodes = 2
	budget = admission.New(context.Background(), p, false)
	if err := budget.Property("app", gosr.List(gosr.String("one"), gosr.String("two"))); err == nil {
		t.Fatal("nodes not limited")
	}
	p = limits.Default()
	p.MaxValueBytes = 6
	budget = admission.New(context.Background(), p, false)
	if err := budget.Property("k", gosr.String("12345")); err != nil {
		t.Fatal(err)
	}
	if err := budget.Property("x", gosr.String("1")); err == nil {
		t.Fatal("aggregate bytes not limited")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := admission.New(ctx, limits.Default(), false).Property("x", gosr.String("v")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestBudgetBoundariesAndWideInput(t *testing.T) {
	p := limits.Default()
	p.MaxValueDepth = 2
	p.MaxValueNodes = 3
	value := gosr.Object(maptypes.ValueMap{
		"child": gosr.Object(maptypes.ValueMap{
			"leaf": gosr.String("v"),
		}),
	})
	if err := admission.New(context.Background(), p, false).Value(value); err != nil {
		t.Fatal("exact depth/node limit rejected", err)
	}
	p.MaxValueDepth = 1
	if err := admission.New(context.Background(), p, false).Value(value); err == nil {
		t.Fatal("depth overrun admitted")
	}
	wide := make([]gosr.Value, limits.DefaultValueNodes)
	if err := admission.New(context.Background(), limits.Default(), false).Value(gosr.List(wide...)); err == nil {
		t.Fatal("wide list admitted before node budget")
	}
	p = limits.Default()
	p.MaxResolvedBytes = 2
	budget := admission.New(context.Background(), p, true)
	if err := budget.Property("k", gosr.String("v")); err != nil {
		t.Fatal(err)
	}
	if err := budget.Property("k", gosr.String("v")); err == nil {
		t.Fatal("resolved budget not cumulative")
	}
}
