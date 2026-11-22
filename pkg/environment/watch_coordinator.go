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

import "context"

// watchCoordinator owns one worker and one pending dirty signal. Signals that
// arrive while refreshing remain queued for a follow-up round, without an
// unbounded queue or a goroutine for each notification.
type watchCoordinator struct {
	ctx   context.Context
	owner *Environment
	dirty chan struct{}
	done  chan struct{}
}

func newWatchCoordinator(ctx context.Context, owner *Environment) *watchCoordinator {
	c := &watchCoordinator{
		ctx:   ctx,
		owner: owner,
		dirty: make(chan struct{}, 1),
		done:  make(chan struct{}),
	}
	go c.run()
	return c
}

func (c *watchCoordinator) signal() {
	if c.ctx.Err() != nil {
		return
	}
	select {
	case c.dirty <- struct{}{}:
	default:
	}
}

func (c *watchCoordinator) run() {
	defer close(c.done)
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.dirty:
			if c.ctx.Err() != nil {
				return
			}
			if err := c.owner.Refresh(c.ctx); err != nil && c.ctx.Err() == nil && c.owner.errorHandler != nil {
				c.owner.errorHandler(err)
			}
		}
	}
}
