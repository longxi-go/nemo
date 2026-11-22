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
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/gosr/pkg/event"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

type burstWatcher struct {
	calls    atomic.Int32
	version  atomic.Int32
	callback chan source.InvalidationFunc
	entered  chan struct{}
	release  chan struct{}
	followed chan struct{}
}

func (p *burstWatcher) Fetch(ctx context.Context) (gosr.PropertySource, error) {
	call := p.calls.Add(1)
	version := p.version.Load()
	if call == 2 {
		close(p.entered)
		select {
		case <-p.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if call == 3 {
		close(p.followed)
	}
	return source.NewMap("burst", maptypes.ValueMap{
		"value": gosr.String(strconv.Itoa(int(version))),
	})
}

func (p *burstWatcher) Watch(ctx context.Context, signal source.InvalidationFunc) error {
	p.callback <- signal
	<-ctx.Done()
	return ctx.Err()
}

func TestWatchCoalescesWithoutBlockingCallback(t *testing.T) {
	p := &burstWatcher{
		callback: make(chan source.InvalidationFunc, 1),
		entered:  make(chan struct{}),
		release:  make(chan struct{}),
		followed: make(chan struct{}),
	}
	p.version.Store(1)
	committed := make(chan struct{}, 1)
	publisher := publisherFunc(func(context.Context, event.Event) error {
		committed <- struct{}{}
		return nil
	})
	owner, err := environment.New(context.Background(), environment.WithProviders(p), environment.WithPublisher(publisher))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- owner.Watch(ctx)
	}()
	signal := <-p.callback
	firstReturned := make(chan struct{})
	go func() {
		signal()
		close(firstReturned)
	}()
	<-p.entered
	select {
	case <-firstReturned:
	case <-time.After(time.Second):
		close(p.release)
		t.Fatal("invalidation callback blocked on refresh")
	}
	p.version.Store(2)
	for range 100 {
		signal()
	}
	close(p.release)
	// A follow-up refresh proves the dirty notification was not lost.
	select {
	case <-p.followed:
	case <-time.After(time.Second):
		t.Fatal("dirty update lost")
	}
	// Fetch entering alone is not a commit. Observe the worker's publication
	// before cancellation, which may reject an in-flight candidate.
	select {
	case <-committed:
	case <-time.After(time.Second):
		t.Fatal("follow-up refresh did not publish the last update")
	}
	property, _, err := owner.Lookup("value")
	if err != nil || !property.Value.Equal(gosr.String("2")) {
		t.Fatal("last update not retained", err)
	}
	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatal(err)
	}
	if calls := p.calls.Load(); calls != 3 {
		t.Fatalf("refreshes not coalesced: %d", calls)
	}
}
