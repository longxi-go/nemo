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
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/gosr/pkg/event"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

// TestGeneratedConfigWatch crosses the actual provider, codec, profile,
// generated binder, candidate validator and notification boundaries.
func TestGeneratedConfigWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	decoder, err := codec.ForFormat("yaml")
	if err != nil {
		t.Fatal(err)
	}
	remote := &yamlRemote{
		decoder: decoder,
		data:    []byte("app:\n  name: remote-one\n  port: 9001\n  timeout: 1s\n"),
		updates: make(chan []byte),
		ready:   make(chan struct{}),
		stopped: make(chan struct{}),
	}
	directory := t.TempDir()
	fixtureFiles := maptypes.StringMap[string]{
		"application.yml":            "app:\n  name: local\n  database:\n    host: base-db\n",
		"application-production.yml": "app:\n  database:\n    host: production-db\n",
	}
	for name, content := range fixtureFiles {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	deliveryFailure := errors.New("sensitive delivery detail")
	publisher := &transitionPublisher{
		events: make(chan environment.Changed, 4),
		cause:  deliveryFailure,
	}
	reloads := make(chan error, 4)
	owner, err := environment.New(ctx,
		environment.WithProviders(remote),
		environment.WithProfiles("production"),
		environment.WithDiscovery(source.Discovery{
			Paths: []string{directory},
		}),
		environment.WithValidator(func(candidate gosr.Snapshot) error {
			_, err := binding.Bind[Config](candidate, "")
			return err
		}),
		environment.WithPublisher(publisher),
		environment.WithErrorHandler(func(err error) {
			reloads <- err
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	initial := owner.Snapshot()
	assertBoundConfig(t, initial, 1, "remote-one", 9001)
	done := make(chan error, 1)
	go func() {
		done <- owner.Watch(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-remote.stopped:
		case <-time.After(10 * time.Second):
			t.Error("remote subscription did not stop")
		}
	})
	await(t, remote.ready)
	sendDocument(t, remote, "app:\n  name: remote-two\n  port: 9002\n  timeout: 1s\n")
	changed := await(t, publisher.events)
	assertBoundConfig(t, changed.Old, 1, "remote-one", 9001)
	assertBoundConfig(t, changed.New, 2, "remote-two", 9002)
	assertBoundConfig(t, owner.Snapshot(), 2, "remote-two", 9002)

	sendDocument(t, remote, "app:\n  name: invalid-update\n  port: not-an-integer\n  timeout: 1s\n")
	err = await(t, reloads)
	if !errors.Is(err, &gosr.Error{
		Code: gosr.ValidationFailed,
	}) ||
		!errors.Is(err, &gosr.Error{
			Code: gosr.ConversionFailed,
		}) {
		t.Fatalf("typed candidate error = %v", err)
	}
	assertBoundConfig(t, owner.Snapshot(), 2, "remote-two", 9002)
	select {
	case <-publisher.events:
		t.Fatal("invalid candidate emitted a committed event")
	default:
	}

	publisher.fail.Store(true)
	sendDocument(t, remote, "app:\n  name: remote-three\n  port: 9003\n  timeout: 1s\n")
	changed = await(t, publisher.events)
	assertBoundConfig(t, changed.New, 3, "remote-three", 9003)
	err = await(t, reloads)
	var notification *environment.NotificationError
	if !errors.As(err, &notification) || notification.Revision != 3 || !errors.Is(err, deliveryFailure) {
		t.Fatalf("committed notification error = %v", err)
	}
	if strings.Contains(err.Error(), deliveryFailure.Error()) {
		t.Fatal("notification printed its sensitive cause")
	}
	assertBoundConfig(t, owner.Snapshot(), 3, "remote-three", 9003)
	assertBoundConfig(t, initial, 1, "remote-one", 9001)
	cancel()
	if err := await(t, done); !errors.Is(err, context.Canceled) {
		t.Fatalf("Watch cancellation = %v", err)
	}
	await(t, remote.stopped)
	select {
	case err := <-reloads:
		t.Fatalf("normal cancellation reported as reload error: %v", err)
	default:
	}
}

func assertBoundConfig(t *testing.T, view gosr.Snapshot, revision uint64, name string, port int) {
	t.Helper()
	config, err := binding.Bind[Config](view, "")
	if err != nil {
		t.Fatal(err)
	}
	if view.Revision() != revision || config.Name != name || config.Port != port ||
		config.Timeout != time.Second || config.Database.Host != "production-db" {
		t.Fatalf("generation %d config = %+v", view.Revision(), config)
	}
}

func await[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(10 * time.Second):
		t.Fatal("configuration flow timed out")
		var zero T
		return zero
	}
}

func sendDocument(t *testing.T, remote *yamlRemote, document string) {
	t.Helper()
	select {
	case remote.updates <- []byte(document):
	case <-time.After(10 * time.Second):
		t.Fatal("remote update was not consumed")
	}
}

// yamlRemote owns mutable transport content, exposing only complete immutable
// sources. Callbacks happen after releasing its transport lock.
type yamlRemote struct {
	mu      sync.RWMutex
	decoder codec.Decoder
	data    []byte
	updates chan []byte
	ready   chan struct{}
	stopped chan struct{}
}

func (r *yamlRemote) Fetch(ctx context.Context) (gosr.PropertySource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	data := slices.Clone(r.data)
	r.mu.RUnlock()
	values, err := r.decoder.Decode(data)
	if err != nil {
		return nil, err
	}
	return source.NewMap("remote", values)
}

func (r *yamlRemote) Watch(ctx context.Context, changed source.InvalidationFunc) error {
	defer close(r.stopped)
	close(r.ready)
	changed()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data := <-r.updates:
			r.mu.Lock()
			r.data = slices.Clone(data)
			r.mu.Unlock()
			changed()
		}
	}
}

type transitionPublisher struct {
	events chan environment.Changed
	fail   atomic.Bool
	cause  error
}

func (p *transitionPublisher) PublishEvent(_ context.Context, value event.Event) error {
	change, ok := value.(environment.Changed)
	if !ok {
		return errors.New("unexpected event type")
	}
	p.events <- change
	if p.fail.Load() {
		return p.cause
	}
	return nil
}
