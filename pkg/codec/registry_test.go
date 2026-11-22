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

package codec_test

import (
	"errors"
	"sync"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

type markerDecoder string

func (d markerDecoder) Decode([]byte) (map[string]gosr.Value, error) {
	return maptypes.ValueMap{
		"marker": gosr.String(string(d)),
	}, nil
}

func TestRegistryRegisterAtomic(t *testing.T) {
	builder := codec.NewRegistryBuilder()
	if err := builder.Register(codec.Definition{
		Name:    "yaml",
		Aliases: []string{"yml"},
		Decoder: markerDecoder("original"),
	}); err != nil {
		t.Fatal(err)
	}
	if err := builder.Register(codec.Definition{
		Name:    "dsl",
		Aliases: []string{"acme", "yml"},
		Decoder: markerDecoder("new"),
	}); err == nil {
		t.Fatal("accepted conflicting alias")
	}
	registry := builder.Build()
	for _, name := range []string{"dsl", "acme"} {
		if _, err := registry.Lookup(name); err == nil {
			t.Fatalf("failed registration leaked %q", name)
		}
	}
	assertDecoder(t, registry, "yml", "original")

	// A conflict late in the built-in batch must not leave earlier formats behind.
	other := codec.NewRegistryBuilder()
	if err := other.Register(codec.Definition{
		Name:    "ini",
		Decoder: markerDecoder("custom"),
	}); err != nil {
		t.Fatal(err)
	}
	if err := codec.RegisterBuiltins(other); err == nil {
		t.Fatal("built-ins replaced an existing decoder")
	}
	for _, name := range []string{"yaml", "yml", "json", "toml", "properties"} {
		if _, err := other.Build().Lookup(name); err == nil {
			t.Fatalf("failed built-in batch leaked %q", name)
		}
	}
	assertDecoder(t, other.Build(), "ini", "custom")
}

func TestRegistrySnapshotIsolation(t *testing.T) {
	builder := codec.NewRegistryBuilder()
	empty := builder.Build()
	aliases := []string{"conf.dsl"}
	if err := builder.Register(codec.Definition{
		Name:    "acme",
		Aliases: aliases,
		Decoder: markerDecoder("acme"),
	}); err != nil {
		t.Fatal(err)
	}
	aliases[0] = "changed"
	if _, err := empty.Lookup("acme"); err == nil {
		t.Fatal("builder mutated an earlier catalog")
	}
	assertDecoder(t, builder.Build(), "CONF.DSL", "acme")
}

func TestRegistryNamesAndAliases(t *testing.T) {
	for _, name := range []string{"", " yaml", "yaml ", "/tmp/a", `a\b`, "a\x00b", ".yaml", "a..b", "a.", "a!b", "格式", "K"} {
		builder := codec.NewRegistryBuilder()
		if err := builder.Register(codec.Definition{
			Name:    name,
			Decoder: markerDecoder("invalid"),
		}); !errors.Is(err, &gosr.Error{
			Code: gosr.InvalidValue,
		}) {
			t.Fatalf("name %q: %v", name, err)
		}
	}
	for _, aliases := range [][]string{
		{
			"Acme",
		},
		{
			"alias",
			"ALIAS",
		},
		{
			"good",
			"bad/alias",
		},
	} {
		builder := codec.NewRegistryBuilder()
		if err := builder.Register(codec.Definition{
			Name:    "acme",
			Aliases: aliases,
			Decoder: markerDecoder("invalid"),
		}); err == nil {
			t.Fatalf("accepted aliases %q", aliases)
		}
		if _, err := builder.Build().Lookup("acme"); err == nil {
			t.Fatal("invalid definition left its canonical name")
		}
	}
	for _, decoder := range []codec.Decoder{nil, codec.DecoderFunc(nil)} {
		if err := codec.NewRegistryBuilder().Register(codec.Definition{
			Name:    "acme",
			Decoder: decoder,
		}); err == nil {
			t.Fatal("accepted nil decoder")
		}
	}
	var nilBuilder *codec.RegistryBuilder
	if err := nilBuilder.Register(codec.Definition{
		Name:    "acme",
		Decoder: markerDecoder("test"),
	}); !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatalf("nil builder: %v", err)
	}
	if err := codec.RegisterBuiltins(nilBuilder); !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatalf("nil built-in builder: %v", err)
	}
	if _, err := nilBuilder.Build().Lookup("yaml"); err == nil {
		t.Fatal("nil builder yielded a populated registry")
	}
	builder := codec.NewRegistryBuilder()
	if err := codec.RegisterBuiltins(builder); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"YAML", "yml", "JSON", "toml", "properties", "ini"} {
		if _, err := builder.Build().Lookup(name); err != nil {
			t.Fatalf("built-in %q: %v", name, err)
		}
	}
}

func TestRegistryConcurrentLookup(t *testing.T) {
	registry := codec.DefaultRegistry()
	var group sync.WaitGroup
	for range 16 {
		group.Go(func() {
			for range 30 {
				decoder, err := registry.Lookup("JsOn")
				if err != nil {
					t.Error(err)
					return
				}
				values, err := decoder.Decode([]byte(`{"port":8080}`))
				if err != nil || values["port"].Kind() != gosr.NumberKind {
					t.Errorf("concurrent decode: %v, %v", values, err)
				}
			}
		})
	}
	group.Wait()
}

func assertDecoder(t *testing.T, registry codec.DecoderLookup, name, marker string) {
	t.Helper()
	decoder, err := registry.Lookup(name)
	if err != nil {
		t.Fatal(err)
	}
	values, err := decoder.Decode(nil)
	if err != nil || !values["marker"].Equal(gosr.String(marker)) {
		t.Fatalf("%q decoder: %v, %v", name, values, err)
	}
}
