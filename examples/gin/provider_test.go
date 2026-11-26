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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

type mutableProvider struct {
	mu      sync.RWMutex
	current gosr.PropertySource
}

func (p *mutableProvider) Fetch(ctx context.Context) (gosr.PropertySource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current, nil
}

func (p *mutableProvider) replace(t *testing.T, name, address string) {
	t.Helper()
	properties, err := source.NewMap("test", maptypes.ValueMap{
		"company.platform.service.name":    gosr.String(name),
		"company.platform.service.address": gosr.String(address),
	})
	if err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	p.current = properties
	p.mu.Unlock()
}

func newApplication(t *testing.T) (*Application, *mutableProvider) {
	t.Helper()
	p := &mutableProvider{}
	p.replace(t, "alpha", "127.0.0.1:8081")
	app, err := New(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}
	return app, p
}

func query(ctx context.Context, router http.Handler) (View, error) {
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequestWithContext(ctx, http.MethodGet, "/configuration", nil))
	if response.Code != http.StatusOK {
		return View{}, fmt.Errorf("unexpected HTTP status %d", response.Code)
	}
	var view View
	err := json.NewDecoder(response.Body).Decode(&view)
	return view, err
}
