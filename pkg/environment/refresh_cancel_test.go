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
	"sync/atomic"
	"testing"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func TestRefreshQueuedCancellation(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	provider := providerFunc(func(ctx context.Context) (gosr.PropertySource, error) {
		if calls.Add(1) == 2 {
			close(entered)
			select {
			case <-release:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return source.NewMap("blocking", maptypes.ValueMap{
			"value": gosr.String("one"),
		})
	})
	owner, err := environment.New(context.Background(), environment.WithProviders(provider))
	if err != nil {
		t.Fatal(err)
	}
	first := make(chan error, 1)
	go func() {
		first <- owner.Refresh(context.Background())
	}()
	<-entered
	defer func() {
		close(release)
		if err := <-first; err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	second := make(chan error, 1)
	go func() {
		second <- owner.Refresh(ctx)
	}()
	cancel()
	select {
	case err := <-second:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("queued refresh ignored cancellation")
	}
	if calls.Load() != 2 {
		t.Fatal("canceled refresh fetched sources")
	}
	var absentContext context.Context
	if err := owner.Refresh(absentContext); err == nil {
		t.Fatal("nil context accepted")
	}
}

func TestRefreshCanceledHandoffKeepsGateUsable(t *testing.T) {
	owner, err := environment.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range 100 {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := owner.Refresh(ctx); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if err := owner.Refresh(context.Background()); err != nil {
			t.Fatal("gate leaked", err)
		}
	}
}
