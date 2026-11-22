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
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func TestEnvironmentRejectsCompositeBudget(t *testing.T) {
	value := gosr.String("leaf")
	for range 80 {
		value = gosr.Object(maptypes.ValueMap{
			"child": value,
		})
	}
	raw, err := source.NewMap("custom", maptypes.ValueMap{
		"app": value,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = environment.New(context.Background(), environment.WithProviders(source.Static(raw)))
	if !errors.Is(err, &gosr.Error{
		Code: gosr.ResolutionLimit,
	}) {
		t.Fatalf("deep custom source admitted: %v", err)
	}
}

func TestEnvironmentAggregateBudget(t *testing.T) {
	raw, _ := source.NewMap("custom", maptypes.ValueMap{
		"a": gosr.String("123"),
		"b": gosr.String("456"),
	})
	p := limits.Default()
	p.MaxValueBytes = 7
	_, err := environment.New(context.Background(), environment.WithProviders(source.Static(raw)), environment.WithLimits(p))
	if !errors.Is(err, &gosr.Error{
		Code: gosr.ResolutionLimit,
	}) {
		t.Fatal(err)
	}
}

func TestEnvironmentResolvedBudgetRollback(t *testing.T) {
	raw, _ := source.NewMap("initial", maptypes.ValueMap{
		"a": gosr.String("x"),
		"b": gosr.String("${a}${a}"),
	})
	current := raw
	provider := providerFunc(func(context.Context) (gosr.PropertySource, error) {
		return current, nil
	})
	policy := limits.Default()
	policy.MaxResolvedBytes = 6
	owner, err := environment.New(context.Background(), environment.WithProviders(provider), environment.WithLimits(policy))
	if err != nil {
		t.Fatal(err)
	}
	old := owner.Snapshot()
	current, _ = source.NewMap("initial", maptypes.ValueMap{
		"a": gosr.String("xxx"),
		"b": gosr.String("${a}${a}"),
	})
	if err := owner.Refresh(context.Background()); !errors.Is(err, &gosr.Error{
		Code: gosr.ResolutionLimit,
	}) {
		t.Fatal("aggregate expanded output admitted", err)
	}
	if owner.Snapshot().Revision() != old.Revision() {
		t.Fatal("invalid generation committed")
	}
	property, _, err := owner.Lookup("a")
	if err != nil || !property.Value.Equal(gosr.String("x")) {
		t.Fatal("old value not retained", err)
	}
}

func TestEnvironmentSourceAndKeyBudgets(t *testing.T) {
	raw, _ := source.NewMap("one", maptypes.ValueMap{
		"a": gosr.String("v"),
		"b": gosr.String("v"),
	})
	policy := limits.Default()
	policy.MaxProperties = 1
	if _, err := environment.New(context.Background(), environment.WithProviders(source.Static(raw)), environment.WithLimits(policy)); err == nil {
		t.Fatal("source keys admitted")
	}
	policy = limits.Default()
	policy.MaxSources = 1
	if _, err := environment.New(context.Background(), environment.WithProviders(source.Static(raw), source.Static(raw)), environment.WithLimits(policy)); err == nil {
		t.Fatal("source count admitted")
	}
}
