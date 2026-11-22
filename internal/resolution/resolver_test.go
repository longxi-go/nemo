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

package resolution

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestStrictResolution(t *testing.T) {
	raw := maptypes.PropertyMap{
		"name": {
			Value: gosr.String("world"),
		},

		"literal": {
			Value: gosr.String(`\${absent}`),
		},

		"greeting": {
			Value: gosr.String("hello ${name}"),
			Origin: gosr.Origin{
				Source: "own",
			},
		},

		"nested": {
			Value: gosr.List(gosr.Object(maptypes.ValueMap{
				"a": gosr.String("${missing:${name}}"),
			})),
		},

		"empty": {
			Value: gosr.String(""),
		},
		"null": {},
	}
	r, err := New(raw)
	if err != nil {
		t.Fatal(err)
	}
	expectedValues := maptypes.StringMap[string]{
		"${greeting}":        "hello world",
		"${literal}":         "${absent}",
		"${empty:fallback}":  "",
		"${missing:${name}}": "world",
	}
	for text, want := range expectedValues {
		got, err := r.Resolve(text)
		if err != nil || got != want {
			t.Fatalf("%q: got %q err %v want %q", text, got, err, want)
		}
	}
	p, ok, err := r.Lookup("greeting")
	if !ok || err != nil || p.Origin.Source != "own" {
		t.Fatal(p, ok, err)
	}
	nested, _, _ := r.Lookup("nested")
	if !nested.Value.Items()[0].Fields()["a"].Equal(gosr.String("world")) {
		t.Fatal(nested)
	}
	failures := maptypes.StringMap[gosr.Code]{
		"${null:fallback}": gosr.InvalidValue,
		"${nested}":        gosr.InvalidValue,
		"${missing}":       gosr.MissingProperty,
		"${name":           gosr.InvalidSyntax,
		"${}":              gosr.InvalidKey,
	}
	for text, code := range failures {
		_, err := r.Resolve(text)
		if !errors.Is(err, &gosr.Error{
			Code: code,
		}) {
			t.Fatalf("%s: %v", text, err)
		}
	}
	delete(raw, "name")
	if got, _ := r.Resolve("${name}"); got != "world" {
		t.Fatal("input alias")
	}
	keys := r.Keys()
	keys[0] = "changed"
	if r.Keys()[0] == "changed" {
		t.Fatal("keys alias")
	}
}

func TestStrictUnusedAndLimits(t *testing.T) {
	for _, raw := range []maptypes.PropertyMap{
		{
			"a": {
				Value: gosr.String("${absent}"),
			},
		},
		{
			"a": {
				Value: gosr.String("${b}"),
			},
			"b": {
				Value: gosr.String("${a}"),
			},
		},
	} {
		if _, err := New(raw); err == nil {
			t.Fatal("bad unused property admitted")
		}
	}
	r, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{
		64,
		65,
	} {
		text := "ok"
		for range n {
			text = "${missing:" + text + "}"
		}
		got, err := r.Resolve(text)
		if n == 64 && (err != nil || got != "ok") || n == 65 && !errors.Is(err, &gosr.Error{
			Code: gosr.ResolutionLimit,
		}) {
			t.Fatal(n, got, err)
		}
	}
	if _, err := r.Resolve(strings.Repeat("x", 1<<20)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve(strings.Repeat("x", (1<<20)+1)); !errors.Is(err, &gosr.Error{
		Code: gosr.ResolutionLimit,
	}) {
		t.Fatal(err)
	}
	raw := maptypes.PropertyMap{}
	for i := 0; i < 66; i++ {
		text := "ok"
		if i < 65 {
			text = fmt.Sprintf("${k%02d}", i+1)
		}
		raw[fmt.Sprintf("k%02d", i)] = gosr.Property{
			Value: gosr.String(text),
		}
	}
	if _, err := New(raw); err == nil {
		t.Fatal("reference depth limit")
	}
}

func TestEscapesAndExpansionOverflow(t *testing.T) {
	r, err := New(maptypes.PropertyMap{
		"a": {
			Value: gosr.String(`\${missing}`),
		},
		"b": {
			Value: gosr.String("${a}"),
		},
		"c": {
			Value: gosr.String("${b}"),
		},
		"huge": {
			Value: gosr.String(strings.Repeat("x", 1<<20)),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"${a}",
		"${b}",
		"${c}",
		"${none:${c}}",
	} {
		got, err := r.Resolve(text)
		if err != nil || got != "${missing}" {
			t.Fatal(got, err)
		}
	}
	if _, err := r.Resolve("${huge}x"); !errors.Is(err, &gosr.Error{
		Code: gosr.ResolutionLimit,
	}) {
		t.Fatal(err)
	}
}

func TestUnusedDefaultSyntax(t *testing.T) {
	r, err := New(maptypes.PropertyMap{
		"present": {
			Value: gosr.String("ok"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve("${present:${}}"); !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidKey,
	}) {
		t.Fatal("malformed unused default admitted", err)
	}
}

func TestNilResolver(t *testing.T) {
	var resolver *Resolver
	if _, _, err := resolver.Lookup("key"); !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatal(err)
	}
	if _, err := resolver.Resolve("text"); !errors.Is(err, &gosr.Error{
		Code: gosr.InvalidState,
	}) {
		t.Fatal(err)
	}
	if len(resolver.Keys()) != 0 {
		t.Fatal("nil resolver keys")
	}
}

func TestResolutionDiagnosticPath(t *testing.T) {
	_, err := New(maptypes.PropertyMap{
		"a": {
			Value: gosr.String("${b}"),
			Origin: gosr.Origin{
				Source: "top",
			},
		},
		"b": {
			Value: gosr.String("${a}"),
		},
	})
	var detail *gosr.Error
	if !errors.As(err, &detail) || strings.Join(detail.Path, " -> ") != "a -> b -> a" || detail.Origin.Source != "top" {
		t.Fatal("outer diagnostic lost path", err)
	}
}

func TestResolveFailureReturnsNoPartialText(t *testing.T) {
	r, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"prefix ${missing}",
		"prefix ${missing:${alsoMissing}}",
		strings.Repeat("x", maxOutput+1),
	} {
		out, err := r.Resolve(text)
		if err == nil || out != "" {
			t.Fatalf("failed expansion returned partial output (bytes=%d, err=%v)", len(out), err)
		}
	}
}
