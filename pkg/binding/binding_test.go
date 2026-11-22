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

package binding_test

import (
	"context"
	"encoding/csv"
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func TestCheckedConversions(t *testing.T) {
	if _, err := binding.Signed[int8](gosr.String("128")); err == nil {
		t.Fatal("overflow accepted")
	}
	if _, err := binding.Float[float32](gosr.String("NaN")); err == nil {
		t.Fatal("nonfinite accepted")
	}
	value, err := binding.List(binding.Signed[int], gosr.String("1,2"))
	if err != nil || len(value) != 2 || value[1] != 2 {
		t.Fatalf("CSV: %v %v", value, err)
	}
	if _, err := binding.String[string](gosr.Value{}); err == nil {
		t.Fatal("null accepted")
	}
}

func TestMissingDefaultsNullAndProvenance(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"empty": gosr.String(""),
		"null":  {},
		"bad":   gosr.String("secret"),
		"map.a": gosr.String("1"),
	})
	opts := binding.ReadOptions{
		Required:   true,
		HasDefault: true,
		Default:    "fallback",
	}
	if _, err := binding.Read(r, "map", opts, binding.String[string]); !errors.Is(err, &gosr.Error{
		Code: gosr.ConversionFailed,
	}) {
		t.Fatalf("descendant data defaulted: %v", err)
	}
	v, err := binding.Read(r, "empty", opts, binding.String[string])
	if err != nil || v != "" {
		t.Fatalf("empty defaulted: %q %v", v, err)
	}
	v, err = binding.Read(r, "missing", opts, binding.String[string])
	if err != nil || v != "fallback" {
		t.Fatal(v, err)
	}
	_, err = binding.Read(r, "missing", binding.ReadOptions{
		Required: true,
	}, binding.String[string])
	if !errors.Is(err, &gosr.Error{
		Code: gosr.MissingProperty,
	}) {
		t.Fatal(err)
	}
	_, err = binding.Read(r, "null", opts, binding.String[string])
	if !errors.Is(err, &gosr.Error{
		Code: gosr.ConversionFailed,
	}) {
		t.Fatal(err)
	}
	_, err = binding.Read(r, "bad", binding.ReadOptions{}, binding.Signed[int])
	var failure *gosr.Error
	if !errors.As(err, &failure) || failure.Origin.Source != "test" || strings.Contains(err.Error(), "secret") {
		t.Fatal(err)
	}
	mp, err := binding.ReadMap(r, "map", binding.ReadOptions{
		Required: true,
	}, binding.Signed[int])
	if err != nil || mp["a"] != 1 {
		t.Fatal(mp, err)
	}
}

func TestConverters(t *testing.T) {
	if _, err := binding.Unsigned[uint8](gosr.String("256")); err == nil {
		t.Fatal("overflow")
	}
	if _, err := binding.Unsigned[uint](gosr.String("-1")); err == nil {
		t.Fatal("negative")
	}
	if _, err := binding.Float[float32](gosr.String("1e100")); err == nil {
		t.Fatal("overflow")
	}
	if b, err := binding.Boolean[bool](gosr.Bool(true)); err != nil || !b {
		t.Fatal(b, err)
	}
	if d, err := binding.Duration(gosr.String("2s")); err != nil || d != 2*time.Second {
		t.Fatal(d, err)
	}
	p, err := binding.Pointer(binding.Signed[int], gosr.String("3"))
	if err != nil || *p != 3 {
		t.Fatal(p, err)
	}
	list, err := binding.List(binding.Boolean[bool], gosr.List(gosr.Bool(true)))
	if err != nil || len(list) != 1 || !list[0] {
		t.Fatal(list, err)
	}
	mp, err := binding.Map(binding.String[string], gosr.Object(maptypes.ValueMap{
		"a": gosr.String("b"),
	}))
	if err != nil || mp["a"] != "b" {
		t.Fatal(mp, err)
	}
}

func TestNestedObject(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"db": gosr.Object(maptypes.ValueMap{
			"host": gosr.String("local"),
		}),
	})
	nested, err := binding.Nested(r, "db", true)
	if err != nil {
		t.Fatal(err)
	}
	host, err := binding.Read(nested, "db.host", binding.ReadOptions{}, binding.String[string])
	if err != nil || host != "local" {
		t.Fatal(host, err)
	}
	if !slices.Contains(nested.Keys(), "db.host") {
		t.Fatal(nested.Keys())
	}
}

func resolver(t *testing.T, values maptypes.ValueMap) gosr.PropertyResolver {
	t.Helper()
	s, err := source.NewMap("test", values)
	if err != nil {
		t.Fatal(err)
	}
	e, err := environment.New(context.Background(), environment.WithProviders(source.Static(s)))
	if err != nil {
		t.Fatal(err)
	}
	return e.Snapshot()
}

type pinned struct {
	gosr.PropertyResolver
	view  gosr.Snapshot
	calls int
}

func (p *pinned) Snapshot() gosr.Snapshot {
	p.calls++
	return p.view
}

type handwritten struct {
	Value string
}

func (*handwritten) PropertyPrefix() string {
	return "app"
}

func (h *handwritten) BindProperties(r gosr.PropertyResolver, prefix string) error {
	var err error
	h.Value, err = binding.Read(r, binding.Join(prefix, "name"), binding.ReadOptions{}, binding.String[string])
	return err
}

func TestBindPinsOnce(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"app.name": gosr.String("pinned"),
	})
	p := &pinned{
		PropertyResolver: r,
		view:             r.(gosr.Snapshot),
	}
	value, err := binding.Bind[handwritten](p, "")
	if err != nil || value.Value != "pinned" || p.calls != 1 {
		t.Fatalf("%v %v calls=%d", value, err, p.calls)
	}
}

func TestNilResolverHelpers(t *testing.T) {
	if _, err := binding.Present(nil, "x"); err == nil {
		t.Fatal("Present")
	}
	if _, err := binding.Nested(nil, "x", false); err == nil {
		t.Fatal("Nested")
	}
	if _, err := binding.ReadMap(nil, "x", binding.ReadOptions{}, binding.String[string]); err == nil {
		t.Fatal("ReadMap")
	}
}

type customConversionError struct {
	message string
}

func (e *customConversionError) Error() string {
	return e.message
}

var customConversionCause = &customConversionError{
	message: "custom secret-value",
}

type failingText string

func (*failingText) UnmarshalText([]byte) error {
	return customConversionCause
}

func TestDirectConversionRedactionAndCauses(t *testing.T) {
	secret := "secret-value"
	type conversionProbeFunc = func() error

	tests := []struct {
		name    string
		convert conversionProbeFunc
		wantNum bool
		wantCSV bool
		cause   error
	}{
		{
			"bool",
			func() error {
				_, err := binding.Boolean[bool](gosr.String(secret))
				return err
			},
			true,
			false,
			strconv.ErrSyntax,
		},
		{
			"signed",
			func() error {
				_, err := binding.Signed[int](gosr.String(secret))
				return err
			},
			true,
			false,
			strconv.ErrSyntax,
		},
		{
			"unsigned",
			func() error {
				_, err := binding.Unsigned[uint](gosr.String(secret))
				return err
			},
			true,
			false,
			strconv.ErrSyntax,
		},
		{
			"float",
			func() error {
				_, err := binding.Float[float64](gosr.String(secret))
				return err
			},
			true,
			false,
			strconv.ErrSyntax,
		},
		{
			"duration",
			func() error {
				_, err := binding.Duration(gosr.String(secret))
				return err
			},
			false,
			false,
			nil,
		},
		{
			"CSV",
			func() error {
				_, err := binding.List(binding.String[string], gosr.String(secret+"\""))
				return err
			},
			false,
			true,
			csv.ErrBareQuote,
		},
		{
			"text",
			func() error {
				_, err := binding.Text[failingText](gosr.String(secret))
				return err
			},
			false,
			false,
			customConversionCause,
		},
		{
			"pointer callback",
			func() error {
				_, err := binding.Pointer(func(gosr.Value) (string, error) {
					return "", customConversionCause
				}, gosr.String(secret))
				return err
			},
			false,
			false,
			customConversionCause,
		},
		{
			"list callback",
			func() error {
				_, err := binding.List(func(gosr.Value) (string, error) {
					return "", customConversionCause
				}, gosr.List(gosr.String(secret)))
				return err
			},
			false,
			false,
			customConversionCause,
		},
		{
			"map callback",
			func() error {
				_, err := binding.Map(func(gosr.Value) (string, error) {
					return "", customConversionCause
				}, gosr.Object(maptypes.ValueMap{
					"item": gosr.String(secret),
				}))
				return err
			},
			false,
			false,
			customConversionCause,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.convert()
			if err == nil || strings.Contains(err.Error(), secret) {
				t.Fatalf("unsafe conversion error: %v", err)
			}
			if test.cause != nil && !errors.Is(err, test.cause) {
				t.Fatalf("cause lost: %v", err)
			}
			if test.cause == customConversionCause {
				var custom *customConversionError
				if !errors.As(err, &custom) || custom != customConversionCause {
					t.Fatalf("original custom error lost: %v", err)
				}
			}
			if test.wantNum {
				var number *strconv.NumError
				if !errors.As(err, &number) || number.Num != secret {
					t.Fatalf("original strconv cause lost: %v", err)
				}
			}
			if test.wantCSV {
				var parse *csv.ParseError
				if !errors.As(err, &parse) {
					t.Fatalf("CSV cause lost: %v", err)
				}
			}
			if errors.Unwrap(err) == nil {
				t.Fatalf("original cause lost: %v", err)
			}
		})
	}
}

func TestNestedDottedObjectMembers(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"app": gosr.Object(maptypes.ValueMap{
			"db.host": gosr.String("actual"),
			"nested": gosr.Object(maptypes.ValueMap{
				"leaf.name": gosr.String("deep"),
			}),
		}),
	})
	nested, err := binding.Nested(r, "app", true)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range nested.Keys() {
		if _, found, err := nested.Lookup(key); err != nil || !found {
			t.Fatalf("advertised key %s unavailable: %v", key, err)
		}
	}
	for _, options := range []binding.ReadOptions{
		{
			HasDefault: true,
			Default:    "fallback",
		},
		{
			Required: true,
		},
	} {
		host, err := binding.Read(nested, "app.db.host", options, binding.String[string])
		if err != nil || host != "actual" {
			t.Fatalf("present dotted member ignored: %q %v", host, err)
		}
	}
	db, err := binding.Nested(nested, "app.db", true)
	if err != nil {
		t.Fatal(err)
	}
	host, err := binding.Read(db, "app.db.host", binding.ReadOptions{
		HasDefault: true,
		Default:    "fallback",
	}, binding.String[string])
	if err != nil || host != "actual" {
		t.Fatal(host, err)
	}
	deep, err := binding.Read(nested, "app.nested.leaf.name", binding.ReadOptions{
		Required: true,
	}, binding.String[string])
	if err != nil || deep != "deep" {
		t.Fatal(deep, err)
	}
	// Exact source addresses retain precedence over the projected object view.
	exact := resolver(t, maptypes.ValueMap{
		"app": gosr.Object(maptypes.ValueMap{
			"db.host": gosr.String("object"),
		}),
		"app.db.host": gosr.String("exact"),
	})
	view, err := binding.Nested(exact, "app", true)
	if err != nil {
		t.Fatal(err)
	}
	host, err = binding.Read(view, "app.db.host", binding.ReadOptions{
		Required: true,
	}, binding.String[string])
	if err != nil || host != "exact" {
		t.Fatal(host, err)
	}
}

func TestNestedAmbiguousObjectProjection(t *testing.T) {
	for _, literal := range []gosr.Value{gosr.String("literal"), gosr.String("nested")} {
		r := resolver(t, maptypes.ValueMap{
			"app": gosr.Object(maptypes.ValueMap{
				"db.host": literal,
				"db": gosr.Object(maptypes.ValueMap{
					"host": gosr.String("nested"),
				}),
			}),
		})
		_, err := binding.Nested(r, "app", true)
		var failure *gosr.Error
		if !errors.Is(err, &gosr.Error{
			Code: gosr.ConversionFailed,
		}) || !errors.As(err, &failure) || failure.Origin.Source != "test" || failure.Key != "app.db.host" {
			t.Fatalf("ambiguous projection admitted or provenance lost: %v", err)
		}
	}
}

func TestNestedInvalidProjectedKey(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"app": gosr.Object(maptypes.ValueMap{
			"trailing ": gosr.String("value"),
		}),
	})
	_, err := binding.Nested(r, "app", true)
	if !errors.Is(err, &gosr.Error{
		Code: gosr.ConversionFailed,
	}) || !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidKey,
	}) {
		t.Fatalf("invalid projected address admitted: %v", err)
	}
}
