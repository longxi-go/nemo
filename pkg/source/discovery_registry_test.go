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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

type literalDecoder string

func (d literalDecoder) Decode(data []byte) (map[string]gosr.Value, error) {
	return maptypes.ValueMap{
		"value": gosr.String(string(d) + string(data)),
	}, nil
}

func TestDiscoveryKeepsSelectedDecoder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "application.CONF.DSL")
	if err := os.WriteFile(path, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "application.other"), []byte("lower"), 0o600); err != nil {
		t.Fatal(err)
	}
	builder := codec.NewRegistryBuilder()
	for _, definition := range []codec.Definition{
		{
			Name:    "other",
			Decoder: literalDecoder("other:"),
		},
		{
			Name:    "acme",
			Aliases: []string{"conf.dsl"},
			Decoder: literalDecoder("acme:"),
		},
	} {
		if err := builder.Register(definition); err != nil {
			t.Fatal(err)
		}
	}
	catalog := builder.Build()
	config := source.Discovery{
		Paths:    []string{dir},
		Formats:  []string{"CONF.DSL", "other"},
		Decoders: catalog,
	}
	env, err := environment.New(context.Background(), environment.WithDiscovery(config))
	if err != nil {
		t.Fatal(err)
	}
	assertDiscovered(t, env, "acme:first", path)
	if err := builder.Register(codec.Definition{
		Name:    "later",
		Decoder: literalDecoder("later:"),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Lookup("later"); err == nil {
		t.Fatal("catalog changed after bootstrap")
	}
	if err := os.WriteFile(path, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := env.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertDiscovered(t, env, "acme:second", path)
	// Create the configured spelling explicitly: native filesystems differ in
	// whether this creates another file or updates the uppercase path above.
	lowerPath := filepath.Join(dir, "application.conf.dsl")
	if err := os.WriteFile(lowerPath, []byte("lowercase"), 0o600); err != nil {
		t.Fatal(err)
	}
	config.Formats = []string{"conf.dsl"}
	lower, err := environment.New(context.Background(), environment.WithDiscovery(config))
	if err != nil {
		t.Fatal(err)
	}
	assertDiscovered(t, lower, "acme:lowercase", lowerPath)
	config.Decoders = nil
	if _, err := source.Discover(config); err == nil {
		t.Fatal("custom format accepted without injected catalog")
	}
}

type propertyLookup interface {
	Lookup(string) (gosr.Property, bool, error)
}

func assertDiscovered(t *testing.T, resolver propertyLookup, expected, resource string) {
	t.Helper()
	value, found, err := resolver.Lookup("value")
	if err != nil || !found || !value.Value.Equal(gosr.String(expected)) {
		t.Fatalf("discovered value: %v, %v, %v", value, found, err)
	}
	if value.Origin.Resource != resource {
		t.Fatalf("resource spelling: got %q, want %q", value.Origin.Resource, resource)
	}
}

type failingDecoder struct {
	cause error
}

func (d *failingDecoder) Decode([]byte) (map[string]gosr.Value, error) {
	if d == nil {
		return maptypes.ValueMap{
			"value": gosr.String("nil handled"),
		}, nil
	}
	return nil, d.cause
}

func TestPluginDecodeFailureOrigin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "application.dsl")
	if err := os.WriteFile(path, []byte("secret payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	secret := errors.New("secret payload")
	builder := codec.NewRegistryBuilder()
	if err := builder.Register(codec.Definition{
		Name: "dsl",
		Decoder: &failingDecoder{
			cause: secret,
		},
	}); err != nil {
		t.Fatal(err)
	}
	providers, err := source.Discover(source.Discovery{
		Paths:    []string{dir},
		Formats:  []string{"dsl"},
		Decoders: builder.Build(),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = providers[0].Fetch(context.Background())
	var detail *gosr.Error
	if !errors.As(err, &detail) ||
		detail.Code != gosr.SourceFailed ||
		detail.Origin.Resource != path ||
		!errors.Is(err, secret) ||
		strings.Contains(err.Error(), "secret payload") {
		t.Fatalf("plugin error provenance/redaction: %v", err)
	}
	var nilDecoder *failingDecoder
	builder = codec.NewRegistryBuilder()
	if err := builder.Register(codec.Definition{
		Name:    "dsl",
		Decoder: nilDecoder,
	}); err != nil {
		t.Fatal(err)
	}
	providers, err = source.Discover(source.Discovery{
		Paths:    []string{dir},
		Formats:  []string{"dsl"},
		Decoders: builder.Build(),
	})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := providers[0].Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertDiscovered(t, loaded, "nil handled", path)
}
