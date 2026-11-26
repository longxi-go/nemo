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
	"slices"
	"sync/atomic"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/gosr/pkg/event"
	"github.com/longxi-go/nemo/internal/resolution"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/source"
)

// Environment publishes immutable snapshots and manages Watch subscriptions.
// Construct it with New and do not copy it after use. Reads perform no external I/O.
type Environment struct {
	writer       refreshGate
	watching     atomic.Bool
	current      atomic.Pointer[snapshot]
	providers    []source.Provider
	profiles     []string
	defaults     []string
	discoveries  []source.Discovery
	validators   []SnapshotValidatorFunc
	publisher    event.Publisher
	errorHandler ErrorHandlerFunc
	limits       limits.Policy
}

// New freezes bootstrap inputs and loads revision 1 without emitting a change.
func New(ctx context.Context, options ...Option) (*Environment, error) {
	e := &Environment{
		writer:   make(refreshGate, 1),
		limits:   limits.Default(),
		defaults: []string{gosr.DefaultProfile},
	}
	for _, option := range options {
		if option == nil {
			return nil, invalid()
		}
		if err := option(e); err != nil {
			return nil, err
		}
	}
	effective := e.profiles
	if len(effective) == 0 {
		effective = e.defaults
	}
	for _, discovery := range e.discoveries {
		discovery.Profiles = slices.Clone(effective)
		discovery.Limits = e.limits
		providers, err := source.Discover(discovery)
		if err != nil {
			return nil, err
		}
		e.providers = append(e.providers, providers...)
	}
	if err := limits.Check(limits.Sources, int64(len(e.providers)), int64(e.limits.MaxSources)); err != nil {
		return nil, err
	}
	candidate, err := e.prepare(ctx, 1)
	if err != nil {
		return nil, err
	}
	e.current.Store(candidate)
	return e, nil
}

// Refresh serializes preparation and commits only complete valid generations.
// Notification executes after releasing the refresh gate and may overlap other
// Refresh notifications. Revisions describe commits, not global delivery order.
func (e *Environment) Refresh(ctx context.Context) error {
	if e == nil || e.current.Load() == nil {
		return invalid()
	}
	old, next, keys, err := e.refresh(ctx)
	if err != nil || next == nil {
		return err
	}
	if e.publisher != nil {
		if err := e.publisher.PublishEvent(ctx, Changed{
			Old:  old,
			New:  next,
			Keys: keys,
		}); err != nil {
			return &NotificationError{
				Revision: next.Revision(),
				Cause:    err,
			}
		}
	}
	return nil
}

func (e *Environment) refresh(ctx context.Context) (*snapshot, *snapshot, []string, error) {
	if err := e.writer.acquire(ctx); err != nil {
		return nil, nil, nil, err
	}
	defer e.writer.release()
	old := e.current.Load()
	next, err := e.prepare(ctx, old.revision+1)
	if err != nil {
		return nil, nil, nil, err
	}
	keys := changedKeys(old, next)
	if len(keys) == 0 {
		return nil, nil, nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	e.current.Store(next)
	return old, next, keys, nil
}

func (e *Environment) prepare(ctx context.Context, revision uint64) (*snapshot, error) {
	if ctx == nil {
		return nil, invalid()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	properties, err := e.fetchProperties(ctx)
	if err != nil {
		return nil, err
	}
	resolved, err := resolution.NewWithPolicy(ctx, properties, e.limits)
	if err != nil {
		return nil, err
	}
	candidate := &snapshot{
		resolver: resolved,
		profiles: e.profiles,
		defaults: e.defaults,
		revision: revision,
	}
	for _, validate := range e.validators {
		if err := validate(candidate); err != nil {
			return nil, &gosr.Error{
				Code:  gosr.ValidationFailed,
				Cause: err,
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return candidate, nil
}

func invalid() error {
	return &gosr.Error{
		Code: gosr.InvalidState,
	}
}
