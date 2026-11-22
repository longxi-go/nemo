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

package environment

import (
	"context"
	"sync"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/source"
)

// Watch runs at most one subscription group per Environment, returning only after all
// provider loops, their callbacks and the refresh worker stop. Invalidation
// returns when its coalesced signal is accepted, not when a refresh commits.
// Invalid reloads reach the error handler and permit recovery; terminal provider
// exits cancel the remaining loops.
// Publisher and error-handler callbacks run outside the refresh gate and may call
// Snapshot, Lookup or Refresh. Concurrent callback invocations may overlap.
func (e *Environment) Watch(ctx context.Context) error {
	if e == nil || e.current.Load() == nil || ctx == nil {
		return invalid()
	}
	if !e.watching.CompareAndSwap(false, true) {
		return invalid()
	}
	defer e.watching.Store(false)
	watchers := []source.Watcher{}
	for _, provider := range e.providers {
		if watcher, ok := provider.(source.Watcher); ok {
			watchers = append(watchers, watcher)
		}
	}
	if len(watchers) == 0 {
		return nil
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan error, len(watchers))
	var group sync.WaitGroup
	coordinator := newWatchCoordinator(child, e)
	invalidate := coordinator.signal
	for _, watcher := range watchers {
		watch := func() {
			err := watcher.Watch(child, invalidate)
			results <- err
		}
		group.Go(watch)
	}
	var result error
	select {
	case result = <-results:
	case <-ctx.Done():
		result = ctx.Err()
	}
	cancel()
	group.Wait()
	<-coordinator.done
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if result != nil {
		return &gosr.Error{
			Code:  gosr.SourceFailed,
			Cause: result,
		}
	}
	return nil
}
