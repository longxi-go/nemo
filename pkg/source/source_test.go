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

package source_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func TestImmutableMap(t *testing.T) {
	values := maptypes.ValueMap{
		"z": gosr.String("original"),
		"a": {},
	}
	s, err := source.NewMap("test", values)
	if err != nil {
		t.Fatal(err)
	}
	values["z"] = gosr.String("changed")
	p, ok, err := s.Lookup("z")
	if err != nil || !ok || !p.Value.Equal(gosr.String("original")) || p.Origin.Source != "test" || p.Origin.Key != "z" {
		t.Fatal(p, ok, err)
	}
	keys := s.Keys()
	keys[0] = "changed"
	if s.Keys()[0] != "a" {
		t.Fatal("aliased keys")
	}
	if _, _, err := s.Lookup(" "); err == nil {
		t.Fatal("invalid lookup")
	}
	if _, ok, err := s.Lookup("missing"); ok || err != nil {
		t.Fatal(ok, err)
	}
	for _, name := range []string{"", " bad"} {
		if _, err := source.NewMap(name, nil); err == nil {
			t.Fatal("invalid name")
		}
	}
	if _, err := source.NewMap("x", maptypes.ValueMap{
		"": {},
	}); err == nil {
		t.Fatal("invalid key")
	}
	if _, err := source.Static(nil).Fetch(context.Background()); err == nil {
		t.Fatal("nil static")
	}
}

func TestPropertiesAndContext(t *testing.T) {
	values := maptypes.PropertyMap{
		"a": {
			Value: gosr.String("value"),
			Origin: gosr.Origin{
				Resource: "custom",
				Line:     3,
			},
		},
	}
	s, err := source.NewProperties("props", values)
	if err != nil {
		t.Fatal(err)
	}
	delete(values, "a")
	p, ok, err := s.Lookup("a")
	if err != nil || !ok || p.Origin.Resource != "custom" || p.Origin.Line != 3 {
		t.Fatal(p, ok, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := source.Static(s).Fetch(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestFileLifecycle(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "app.custom")
	d := codec.DecoderFunc(func(b []byte) (map[string]gosr.Value, error) {
		return maptypes.ValueMap{
			"value": gosr.String(string(b)),
		}, nil
	})
	p, err := source.NewFile("file", path, d, true)
	if err != nil {
		t.Fatal(err)
	}
	s, err := p.Fetch(ctx)
	if err != nil || len(s.Keys()) != 0 {
		t.Fatal(s, err)
	}
	if err := os.WriteFile(path, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err = p.Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	prop, _, _ := s.Lookup("value")
	if prop.Origin.Resource != path {
		t.Fatal(prop)
	}
	if err := os.WriteFile(path, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	next, err := p.Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	n, _, _ := next.Lookup("value")
	if !n.Value.Equal(gosr.String("second")) || !prop.Value.Equal(gosr.String("first")) {
		t.Fatal(n, prop)
	}
	required, _ := source.NewFile("required", path+"missing", d, false)
	if _, err := required.Fetch(ctx); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	bad := codec.DecoderFunc(func([]byte) (map[string]gosr.Value, error) {
		return nil, errors.New("sensitive payload")
	})
	bp, _ := source.NewFile("bad", path, bad, true)
	if _, err := bp.Fetch(ctx); err == nil || strings.Contains(err.Error(), "sensitive payload") || !errors.Is(err, &gosr.Error{
		Code: gosr.SourceFailed,
	}) {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, path string
		decoder    codec.Decoder
	}{
		{
			"",
			path,
			d,
		},
		{
			"x",
			"",
			d,
		},
		{
			"x",
			path,
			nil,
		},
		{
			"x",
			path,
			codec.DecoderFunc(nil),
		},
	} {
		if _, err := source.NewFile(tt.name, tt.path, tt.decoder, false); err == nil {
			t.Fatal("invalid constructor")
		}
	}
	directory, _ := source.NewFile("dir", filepath.Dir(path), d, true)
	if _, err := directory.Fetch(ctx); err == nil {
		t.Fatal("optional directory accepted")
	}
}

func TestOSSnapshot(t *testing.T) {
	t.Setenv("NEMOTEST_APP_PORT", "8080")
	t.Setenv("NEMOTEST_EMPTY", "")
	p, err := source.NewOS("os", "NEMOTEST_")
	if err != nil {
		t.Fatal(err)
	}
	s, err := p.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	prop, ok, _ := s.Lookup("app.port")
	if !ok || !prop.Value.Equal(gosr.String("8080")) {
		t.Fatal(prop)
	}
	t.Setenv("NEMOTEST_APP_PORT", "changed")
	old, _, _ := s.Lookup("app.port")
	if !old.Value.Equal(prop.Value) {
		t.Fatal("snapshot mutated")
	}
	// Dots and underscores are distinct environment names on every supported
	// platform but collapse to the same property key under this provider's policy.
	t.Setenv("NEMOTEST_APP.PORT", "collision secret")
	if _, err := p.Fetch(context.Background()); err == nil || strings.Contains(err.Error(), "collision secret") {
		t.Fatal(err)
	}
}

func TestDiscoveryPriorityAndValidation(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	d := source.Discovery{
		Paths:    []string{a, b},
		Profiles: []string{"dev", "prod"},
		Formats:  []string{"json", "yaml"},
	}
	providers, err := source.Discover(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 12 {
		t.Fatal(len(providers))
	}
	want := []string{
		filepath.Join(b, "application-prod.json"),
		filepath.Join(b, "application-prod.yaml"),
		filepath.Join(a, "application-prod.json"),
		filepath.Join(a, "application-prod.yaml"),
		filepath.Join(b, "application-dev.json"),
		filepath.Join(b, "application-dev.yaml"),
		filepath.Join(a, "application-dev.json"),
		filepath.Join(a, "application-dev.yaml"),
		filepath.Join(b, "application.json"),
		filepath.Join(b, "application.yaml"),
		filepath.Join(a, "application.json"),
		filepath.Join(a, "application.yaml"),
	}
	for i, path := range want {
		data, err := json.Marshal(maptypes.StringMap[string]{
			"winner": path,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		s, err := providers[i].Fetch(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		p, ok, _ := s.Lookup("winner")
		if !ok || p.Origin.Resource != path || !p.Value.Equal(gosr.String(path)) {
			t.Fatal(i, p)
		}
	}
	seen := maptypes.StringBoolMap{}
	for _, p := range providers {
		s, err := p.Fetch(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if seen[s.Name()] {
			t.Fatal("duplicate source identity")
		}
		seen[s.Name()] = true
	}
	for _, d := range []source.Discovery{
		{
			Names: []string{"../bad"},
		},
		{
			Paths: []string{""},
		},
		{
			Profiles: []string{"!dev"},
		},
		{
			Formats: []string{"bad"},
		},
		{
			Formats: []string{"../json"},
		},
	} {
		if _, err := source.Discover(d); err == nil {
			t.Fatal("invalid discovery", d)
		}
	}
	if _, err := source.Discover(source.Discovery{}); err != nil {
		t.Fatal(err)
	}
}

type customProvider struct {
	snapshot gosr.PropertySource
}

func (p customProvider) Fetch(ctx context.Context) (gosr.PropertySource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return p.snapshot, nil
}

func TestProviderExtension(t *testing.T) {
	s, err := source.NewMap("custom", nil)
	if err != nil {
		t.Fatal(err)
	}
	var p source.Provider = customProvider{
		snapshot: s,
	}
	got, err := p.Fetch(context.Background())
	if err != nil || got.Name() != "custom" {
		t.Fatal(got, err)
	}
}

func TestConcurrentFetch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.json")
	if err := os.WriteFile(path, []byte(`{"key":"value"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	decoder, err := codec.ForFormat("json")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := source.NewFile("file", path, decoder, false)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 16 {
		group.Go(func() {
			for range 20 {
				snapshot, err := provider.Fetch(context.Background())
				if err != nil {
					t.Error(err)
					return
				}
				property, found, err := snapshot.Lookup("key")
				if err != nil || !found || !property.Value.Equal(gosr.String("value")) {
					t.Error(property, found, err)
					return
				}
				keys := snapshot.Keys()
				keys[0] = "detached"
			}
		})
	}
	group.Wait()
}

func TestFileCancellationAndCause(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.custom")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	decoder := codec.DecoderFunc(func([]byte) (map[string]gosr.Value, error) {
		cancel()
		return nil, nil
	})
	provider, err := source.NewFile("canceled", path, decoder, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Fetch(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	cause := errors.New("private parser input")
	provider, err = source.NewFile("error", path, codec.DecoderFunc(func([]byte) (map[string]gosr.Value, error) {
		return nil, cause
	}), false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Fetch(context.Background())
	if !errors.Is(err, cause) || strings.Contains(err.Error(), cause.Error()) {
		t.Fatal(err)
	}
}
