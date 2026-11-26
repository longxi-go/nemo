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

package main

import (
	"errors"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func TestNestedYAMLBinding(t *testing.T) {
	decoder, err := codec.DefaultRegistry().Lookup(codec.FormatYAML)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := source.NewFile("file", "testdata/application.yaml", decoder, false)
	if err != nil {
		t.Fatal(err)
	}
	app, err := New(t.Context(), provider)
	if err != nil {
		t.Fatal(err)
	}
	view, err := query(t.Context(), app.Router())
	if err != nil || view.Name != "alpha" || view.Address != "127.0.0.1:8081" || view.Revision != 1 {
		t.Fatalf("view=%+v error=%v", view, err)
	}
}

func TestInvalidInitialConfiguration(t *testing.T) {
	for _, values := range []maptypes.ValueMap{
		{},
		{
			"company.platform.service.name": gosr.String("alpha"),
		},
		{
			"company.platform.service.name":    gosr.String(""),
			"company.platform.service.address": gosr.String("127.0.0.1:8081"),
		},
		{
			"company.platform.service.name":    gosr.String("alpha"),
			"company.platform.service.address": gosr.String("invalid"),
		},
	} {
		properties, err := source.NewMap("invalid", values)
		if err != nil {
			t.Fatal(err)
		}
		_, err = New(t.Context(), source.Static(properties))
		if !errors.Is(err, &gosr.Error{
			Code: gosr.ValidationFailed,
		}) {
			t.Fatalf("expected validation failure, got %v", err)
		}
	}
}

func TestExplicitRebindAndFailedRefresh(t *testing.T) {
	app, provider := newApplication(t)
	held := app.Snapshot()
	initial, err := binding.Bind[Config](held, "")
	if err != nil {
		t.Fatal(err)
	}
	provider.replace(t, "beta", "127.0.0.1:8082")
	if err := app.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	current, err := query(t.Context(), app.Router())
	if err != nil || current.Name != "beta" || current.Address != "127.0.0.1:8082" || current.Revision != 2 {
		t.Fatalf("refreshed=%+v error=%v", current, err)
	}
	old, err := binding.Bind[Config](held, "")
	if err != nil || old != initial || initial.Name != "alpha" || held.Revision() != 1 {
		t.Fatalf("held=%+v initial=%+v error=%v", old, initial, err)
	}
	provider.replace(t, "invalid", "bad-address")
	if err := app.Refresh(t.Context()); !errors.Is(err, &gosr.Error{
		Code: gosr.ValidationFailed,
	}) {
		t.Fatalf("expected rejected candidate, got %v", err)
	}
	after, err := query(t.Context(), app.Router())
	if err != nil || after != current {
		t.Fatalf("failed refresh changed response: %+v -> %+v (%v)", current, after, err)
	}
}

func TestAddressValidation(t *testing.T) {
	for _, address := range []string{":8080", "localhost:bad", "localhost:0", "localhost:65536"} {
		t.Run(address, func(t *testing.T) {
			provider := &mutableProvider{}
			provider.replace(t, "alpha", address)
			if _, err := New(t.Context(), provider); !errors.Is(err, &gosr.Error{
				Code: gosr.ValidationFailed,
			}) {
				t.Fatalf("invalid address accepted: %v", err)
			}
		})
	}
	provider := &mutableProvider{}
	provider.replace(t, "alpha", "[::1]:8080")
	if _, err := New(t.Context(), provider); err != nil {
		t.Fatalf("valid IPv6 address rejected: %v", err)
	}
}
