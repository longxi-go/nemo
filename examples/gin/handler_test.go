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
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestRequestsDuringRefreshRemainCoherent(t *testing.T) {
	app, provider := newApplication(t)
	var readers sync.WaitGroup
	failures := make(chan error, 4)
	for range 4 {
		readers.Go(func() {
			for range 40 {
				view, err := query(t.Context(), app.Router())
				if err != nil {
					failures <- err
					return
				}
				alpha := view.Name == "alpha" &&
					view.Address == "127.0.0.1:8081" && view.Revision%2 == 1
				beta := view.Name == "beta" &&
					view.Address == "127.0.0.1:8082" && view.Revision%2 == 0
				if !alpha && !beta {
					failures <- fmt.Errorf("mixed configuration: %+v", view)
					return
				}
			}
		})
	}
	for index := range 30 {
		if index%2 == 0 {
			provider.replace(t, "beta", "127.0.0.1:8082")
		} else {
			provider.replace(t, "alpha", "127.0.0.1:8081")
		}
		if err := app.Refresh(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	readers.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
}

func TestCanceledRequest(t *testing.T) {
	app, _ := newApplication(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	response := httptest.NewRecorder()
	app.Router().ServeHTTP(response, httptest.NewRequestWithContext(ctx, http.MethodGet, "/configuration", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("canceled request status=%d", response.Code)
	}
}
