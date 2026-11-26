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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/gosr/pkg/event"
	env "github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

type provider struct {
	mu       sync.Mutex
	values   maptypes.ValueMap
	err      error
	changes  chan chan struct{}
	started  chan struct{}
	stopped  chan struct{}
	terminal chan error
}

func (p *provider) Fetch(ctx context.Context) (gosr.PropertySource, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return nil, p.err
	}
	return source.NewMap("mutable", p.values)
}

func (p *provider) set(values maptypes.ValueMap, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.values = values
	p.err = err
}

// Use the raw function signature to verify compatibility with the public alias.
func (p *provider) Watch(ctx context.Context, changed func()) error {
	close(p.started)
	defer close(p.stopped)
	for {
		select {
		case ack := <-p.changes:
			changed()
			close(ack)
		case err := <-p.terminal:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func static(t *testing.T, name string, values maptypes.ValueMap) source.Provider {
	t.Helper()
	s, err := source.NewMap(name, values)
	if err != nil {
		t.Fatal(err)
	}
	return source.Static(s)
}

func value(t *testing.T, s gosr.PropertyResolver, key, want string) {
	t.Helper()
	p, ok, err := s.Lookup(key)
	if err != nil || !ok || !p.Value.Equal(gosr.String(want)) {
		t.Fatalf("%s: %v %v %v", key, p, ok, err)
	}
}

func TestGenerationSemantics(t *testing.T) {
	p := &provider{
		values: maptypes.ValueMap{
			"a":     gosr.String("first"),
			"gone":  gosr.String("bye"),
			"null":  {},
			"empty": gosr.String(""),
		},
	}
	profiles := []string{
		"dev",
		"dev",
	}
	owner, err := env.New(context.Background(), env.WithProviders(p, static(t, "low", maptypes.ValueMap{
		"a":    gosr.String("low"),
		"null": gosr.String("low"),
	})), env.WithProfiles(profiles...))
	if err != nil {
		t.Fatal(err)
	}
	profiles[0] = "changed"
	old := owner.Snapshot()
	value(t, old, "a", "first")
	if old.Revision() != 1 || len(old.ActiveProfiles()) != 1 || old.ActiveProfiles()[0] != "dev" {
		t.Fatal("bootstrap")
	}
	if ok, err := old.AcceptsProfiles("dev", "invalid profile"); ok || err == nil {
		t.Fatal("predicate validation", ok, err)
	}
	got := old.Keys()
	got[0] = "changed"
	if old.Keys()[0] == "changed" {
		t.Fatal("keys alias")
	}
	if prop, ok, err := old.Lookup("null"); err != nil || !ok || prop.Value.Kind() != gosr.NullKind {
		t.Fatal(prop, ok, err)
	}
	value(t, old, "empty", "")
	if _, ok, err := old.Lookup("missing"); err != nil || ok {
		t.Fatal(ok, err)
	}
	if err := owner.Refresh(context.Background()); err != nil || owner.Snapshot().Revision() != 1 {
		t.Fatal("unchanged", err)
	}
	p.set(maptypes.ValueMap{
		"a": gosr.String("second"),
	}, nil)
	if err := owner.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	value(t, old, "a", "first")
	value(t, owner.Snapshot(), "a", "second")
	if _, ok, _ := owner.Snapshot().Lookup("gone"); ok {
		t.Fatal("deletion")
	}
	p.set(maptypes.ValueMap{
		"unused": gosr.String("${absent}"),
	}, nil)
	if err := owner.Refresh(context.Background()); err == nil || owner.Snapshot().Revision() != 2 {
		t.Fatal("rollback", err)
	}
	cause := errors.New("private source payload")
	p.set(nil, cause)
	err = owner.Refresh(context.Background())
	if !errors.Is(err, cause) || strings.Contains(err.Error(), cause.Error()) {
		t.Fatal(err)
	}
}

type publisherFunc func(context.Context, event.Event) error

func (f publisherFunc) PublishEvent(ctx context.Context, e event.Event) error {
	return f(ctx, e)
}

func TestPublicationAndValidator(t *testing.T) {
	p := &provider{
		values: maptypes.ValueMap{
			"a":    gosr.String("first"),
			"gone": gosr.String("bye"),
		},
	}
	var owner *env.Environment
	var changes []env.Changed
	cause := errors.New("secret publication data")
	pub := publisherFunc(func(ctx context.Context, e event.Event) error {
		change, ok := e.(env.Changed)
		if !ok {
			t.Fatal(e)
		}
		changes = append(changes, change)
		value(t, owner.Snapshot(), "a", "second")
		if err := owner.Refresh(ctx); err != nil {
			t.Fatal(err)
		}
		return cause
	})
	rejected := errors.New("secret validation data")
	validator := func(s gosr.Snapshot) error {
		p, _, _ := s.Lookup("a")
		if p.Value.Equal(gosr.String("bad")) {
			return rejected
		}
		return nil
	}
	var err error
	owner, err = env.New(context.Background(), env.WithProviders(p), env.WithPublisher(pub), env.WithValidator(validator))
	if err != nil {
		t.Fatal(err)
	}
	p.set(maptypes.ValueMap{
		"a": gosr.String("bad"),
	}, nil)
	if err := owner.Refresh(context.Background()); !errors.Is(err, rejected) ||
		strings.Contains(err.Error(), rejected.Error()) || owner.Snapshot().Revision() != 1 {
		t.Fatal(err)
	}
	p.set(maptypes.ValueMap{
		"a": gosr.String("second"),
	}, nil)
	done := make(chan error, 1)
	go func() {
		done <- owner.Refresh(context.Background())
	}()
	select {
	case err = <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("publisher reentry deadlock")
	}
	var notification *env.NotificationError
	if !errors.As(err, &notification) ||
		!errors.Is(err, cause) ||
		notification.Revision != 2 ||
		strings.Contains(err.Error(), cause.Error()) {
		t.Fatal(err)
	}
	if len(changes) != 1 ||
		changes[0].Old.Revision() != 1 ||
		changes[0].New.Revision() != 2 ||
		strings.Join(changes[0].Keys, ",") != "a,gone" {
		t.Fatal(changes)
	}
	changes[0].Keys[0] = "mutated"
	value(t, changes[0].Old, "a", "first")
}

func TestWatchRecoveryAndReaders(t *testing.T) {
	p := &provider{
		values: maptypes.ValueMap{
			"a": gosr.String("first"),
		},
		changes:  make(chan chan struct{}),
		started:  make(chan struct{}),
		stopped:  make(chan struct{}),
		terminal: make(chan error),
	}
	reported := make(chan error, 2)
	committed := make(chan struct{}, 8)
	pub := publisherFunc(func(context.Context, event.Event) error {
		committed <- struct{}{}
		return nil
	})
	var owner *env.Environment
	var err error
	owner, err = env.New(context.Background(), env.WithProviders(p), env.WithPublisher(pub), env.WithErrorHandler(func(err error) {
		_ = owner.Snapshot()
		if e := owner.Refresh(context.Background()); e == nil {
			t.Error("bad update unexpectedly valid")
		}
		reported <- err
	}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- owner.Watch(ctx)
	}()
	<-p.started
	if err := owner.Watch(ctx); err == nil {
		t.Fatal("second Watch")
	}
	var readers sync.WaitGroup
	for range 8 {
		readers.Go(func() {
			for range 100 {
				s := owner.Snapshot()
				_, _, _ = s.Lookup("a")
				_ = s.Keys()
			}
		})
	}
	update := func(text string) {
		t.Helper()
		p.set(maptypes.ValueMap{
			"a": gosr.String(text),
		}, nil)
		ack := make(chan struct{})
		p.changes <- ack
		select {
		case <-ack:
		case <-time.After(3 * time.Second):
			t.Fatal("callback deadlock")
		}
	}
	update("second")
	select {
	case <-committed:
	case <-time.After(3 * time.Second):
		t.Fatal("no commit")
	}
	value(t, owner.Snapshot(), "a", "second")
	update("${missing}")
	select {
	case err := <-reported:
		if !errors.Is(err, &gosr.Error{
			Code: gosr.MissingProperty,
		}) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no error handler")
	}
	if owner.Snapshot().Revision() != 2 {
		t.Fatal("bad update committed")
	}
	update("recovered")
	select {
	case <-committed:
	case <-time.After(3 * time.Second):
		t.Fatal("no recovery commit")
	}
	value(t, owner.Snapshot(), "a", "recovered")
	readers.Wait()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not stop")
	}
	select {
	case <-p.stopped:
	default:
		t.Fatal("watch not joined")
	}
	select {
	case err := <-reported:
		t.Fatal("cancellation reported", err)
	default:
	}
}

func TestDiscoveryProfiles(t *testing.T) {
	dir := t.TempDir()
	fixtureFiles := maptypes.StringMap[string]{
		"application.json":          `{"a":"base"}`,
		"application-dev.json":      `{"a":"active"}`,
		"application-fallback.json": `{"a":"default"}`,
	}
	for name, data := range fixtureFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, active := range []bool{
		false,
		true,
	} {
		opts := []env.Option{
			env.WithDiscovery(source.Discovery{
				Paths:   []string{dir},
				Formats: []string{"json"},
			}),
			env.WithDefaultProfiles("fallback"),
		}
		want := "default"
		if active {
			opts = append(opts, env.WithProfiles("dev"))
			want = "active"
		}
		owner, err := env.New(context.Background(), opts...)
		if err != nil {
			t.Fatal(err)
		}
		value(t, owner.Snapshot(), "a", want)
	}
}

func TestInvalidConfiguration(t *testing.T) {
	for _, opts := range [][]env.Option{
		{
			nil,
		},
		{
			env.WithProviders(nil),
		},
		{
			env.WithValidator(nil),
		},
		{
			env.WithPublisher(nil),
		},
		{
			env.WithErrorHandler(nil),
		},
		{
			env.WithProfiles("bad profile"),
		},
		{
			env.WithDiscovery(source.Discovery{
				Profiles: []string{"dev"},
			}),
		},
		{
			env.WithProviders(static(t, "same", nil), static(t, "same", nil)),
		},
	} {
		if _, err := env.New(context.Background(), opts...); err == nil {
			t.Fatal("invalid options admitted")
		}
	}
	owner, err := env.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := owner.Snapshot().DefaultProfiles(); len(got) != 1 || got[0] != "default" {
		t.Fatal(got)
	}
	if err := owner.Watch(context.Background()); err != nil {
		t.Fatal(err)
	}
	var absent *env.Environment
	if absent.Refresh(context.Background()) == nil || absent.Watch(context.Background()) == nil {
		t.Fatal("nil receiver")
	}
}

type providerFunc func(context.Context) (gosr.PropertySource, error)

func (f providerFunc) Fetch(ctx context.Context) (gosr.PropertySource, error) {
	return f(ctx)
}

type rawSource struct {
	name       string
	properties maptypes.PropertyMap
	err        error
}

func (s *rawSource) Name() string {
	return s.name
}

func (s *rawSource) Keys() []string {
	keys := []string{}
	for k := range s.properties {
		keys = append(keys, k)
	}
	return keys
}

func (s *rawSource) Lookup(k string) (gosr.Property, bool, error) {
	p, ok := s.properties[k]
	return p, ok, s.err
}

func TestSourceDetachmentAndOriginOnlyChange(t *testing.T) {
	raw := &rawSource{
		name: "raw",
		properties: maptypes.PropertyMap{
			"a": {
				Value: gosr.String("one"),
				Origin: gosr.Origin{
					Resource: "resource",
					Line:     1,
				},
			},
		},
	}
	owner, err := env.New(context.Background(), env.WithProviders(source.Static(raw)))
	if err != nil {
		t.Fatal(err)
	}
	first := owner.Snapshot()
	p, _, _ := first.Lookup("a")
	if p.Origin.Source != "raw" || p.Origin.Key != "a" || p.Origin.Resource != "resource" {
		t.Fatal(p)
	}
	raw.properties["a"] = gosr.Property{
		Value: gosr.String("one"),
		Origin: gosr.Origin{
			Resource: "resource",
			Line:     2,
		},
	}
	p, _, _ = first.Lookup("a")
	if p.Origin.Line != 1 {
		t.Fatal("source mutation leaked")
	}
	if err := owner.Refresh(context.Background()); err != nil || owner.Snapshot().Revision() != 2 {
		t.Fatal("origin change ignored", err)
	}
	cause := errors.New("secret lookup payload")
	raw.err = cause
	if err := owner.Refresh(context.Background()); !errors.Is(err, cause) ||
		strings.Contains(err.Error(), cause.Error()) || owner.Snapshot().Revision() != 2 {
		t.Fatal(err)
	}
	for _, raw := range []*rawSource{
		{
			name:       "",
			properties: nil,
		},
		{
			name: "raw",
			properties: maptypes.PropertyMap{
				" ": {},
			},
		},
	} {
		if _, err := env.New(context.Background(), env.WithProviders(source.Static(raw))); err == nil {
			t.Fatal("bad source admitted")
		}
	}
	emptyProvider := providerFunc(func(context.Context) (gosr.PropertySource, error) {
		return nil, nil
	})
	if _, err := env.New(context.Background(), env.WithProviders(emptyProvider)); err == nil {
		t.Fatal("nil source")
	}
}

func TestConcurrentRefreshCoherentReaders(t *testing.T) {
	p := &provider{
		values: maptypes.ValueMap{
			"a": gosr.String("0"),
			"b": gosr.String("0"),
		},
	}
	owner, err := env.New(context.Background(), env.WithProviders(p))
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			for range 50 {
				if err := owner.Refresh(context.Background()); err != nil {
					t.Error(err)
				}
			}
		})
	}
	for range 4 {
		group.Go(func() {
			for range 100 {
				s := owner.Snapshot()
				a, _, _ := s.Lookup("a")
				b, _, _ := s.Lookup("b")
				if !a.Value.Equal(b.Value) {
					t.Error("mixed generation")
				}
			}
		})
	}
	for i := range 30 {
		text := strings.Repeat("x", i)
		p.set(maptypes.ValueMap{
			"a": gosr.String(text),
			"b": gosr.String(text),
		}, nil)
	}
	group.Wait()
}

func TestWatchTerminalCancelsAndJoinsAll(t *testing.T) {
	p := &provider{
		changes:  make(chan chan struct{}),
		started:  make(chan struct{}),
		stopped:  make(chan struct{}),
		terminal: make(chan error),
	}
	q := &otherWatcher{
		provider: provider{
			changes:  make(chan chan struct{}),
			started:  make(chan struct{}),
			stopped:  make(chan struct{}),
			terminal: make(chan error),
		},
	}
	owner, err := env.New(context.Background(), env.WithProviders(p, q))
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- owner.Watch(context.Background())
	}()
	<-p.started
	<-q.started
	cause := errors.New("terminal transport failure")
	p.terminal <- cause
	select {
	case err := <-done:
		if !errors.Is(err, cause) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("terminal did not join")
	}
	for _, stopped := range []chan struct{}{
		p.stopped,
		q.stopped,
	} {
		select {
		case <-stopped:
		default:
			t.Fatal("unjoined provider")
		}
	}
}

type otherWatcher struct {
	provider
}

func (p *otherWatcher) Fetch(context.Context) (gosr.PropertySource, error) {
	return source.NewMap("other", nil)
}

func TestPublisherChangedReentry(t *testing.T) {
	p := &provider{
		values: maptypes.ValueMap{
			"a": gosr.String("one"),
		},
	}
	var owner *env.Environment
	var revisions []uint64
	pub := publisherFunc(func(ctx context.Context, e event.Event) error {
		change := e.(env.Changed)
		revisions = append(revisions, change.New.Revision())
		if change.New.Revision() == 2 {
			p.set(maptypes.ValueMap{
				"a": gosr.String("three"),
			}, nil)
			return owner.Refresh(ctx)
		}
		return nil
	})
	var err error
	owner, err = env.New(context.Background(), env.WithProviders(p), env.WithPublisher(pub))
	if err != nil {
		t.Fatal(err)
	}
	p.set(maptypes.ValueMap{
		"a": gosr.String("two"),
	}, nil)
	done := make(chan error, 1)
	go func() {
		done <- owner.Refresh(context.Background())
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("changed reentry deadlock")
	}
	value(t, owner.Snapshot(), "a", "three")
	if len(revisions) != 2 || revisions[0] != 2 || revisions[1] != 3 {
		t.Fatal(revisions)
	}
}

type initialWatcher struct {
	provider
	synchronized chan struct{}
}

// Verify the raw function signature on the initial invalidation path too.
func (p *initialWatcher) Watch(ctx context.Context, changed func()) error {
	close(p.started)
	defer close(p.stopped)
	// An adapter subscribes first, then invalidates to reconcile the Fetch/Watch gap.
	changed()
	close(p.synchronized)
	<-ctx.Done()
	return ctx.Err()
}

func TestWatchInitialInvalidationClosesLoadSubscriptionGap(t *testing.T) {
	p := &initialWatcher{
		provider: provider{
			values: maptypes.ValueMap{
				"a": gosr.String("v1"),
			},
			started: make(chan struct{}),
			stopped: make(chan struct{}),
		},
		synchronized: make(chan struct{}),
	}
	committed := make(chan struct{}, 1)
	pub := publisherFunc(func(context.Context, event.Event) error {
		committed <- struct{}{}
		return nil
	})
	owner, err := env.New(context.Background(), env.WithProviders(p), env.WithPublisher(pub))
	if err != nil {
		t.Fatal(err)
	}
	value(t, owner.Snapshot(), "a", "v1")
	p.set(maptypes.ValueMap{
		"a": gosr.String("v2"),
	}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- owner.Watch(ctx)
	}()
	select {
	case <-p.synchronized:
	case <-time.After(3 * time.Second):
		t.Fatal("no initial resync")
	}
	select {
	case <-committed:
	case <-time.After(3 * time.Second):
		t.Fatal("initial refresh not committed")
	}
	value(t, owner.Snapshot(), "a", "v2")
	if owner.Snapshot().Revision() != 2 {
		t.Fatal("initial invalidation did not publish")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watch not joined")
	}
	select {
	case <-p.stopped:
	default:
		t.Fatal("adapter still active")
	}
}

func TestNilOwnerAccessors(t *testing.T) {
	var owner *env.Environment
	if owner.Snapshot().Revision() != 0 || len(owner.Keys()) != 0 ||
		len(owner.ActiveProfiles()) != 0 || len(owner.DefaultProfiles()) != 0 {
		t.Fatal("nil owner state")
	}
	if _, _, err := owner.Lookup("a"); !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatal(err)
	}
	if _, err := owner.Resolve("text"); !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatal(err)
	}
	if _, err := owner.AcceptsProfiles("dev"); !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatal(err)
	}
	var notification *env.NotificationError
	if notification.Error() != "<nil>" || notification.Unwrap() != nil {
		t.Fatal("nil notification")
	}
}
