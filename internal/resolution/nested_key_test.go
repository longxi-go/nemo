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
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/longxi-go/nemo/pkg/limits"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestNestedKeys(t *testing.T) {
	r, err := New(maptypes.PropertyMap{
		"selector": {
			Value: gosr.String("dev"),
		},
		"dev.url": {
			Value: gosr.String("ready"),
		},
		"literal": {
			Value: gosr.String(`\${selector}`),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for input, want := range map[string]string{
		"${${selector}.url}":             "ready",
		"${${missing:dev}.url:fallback}": "ready",
		"${${selector}.absent:fallback}": "fallback",
		`${missing:\${selector}}`:        "${selector}",
	} {
		got, err := r.Resolve(input)
		if err != nil || got != want {
			t.Fatalf("%s: %q %v", input, got, err)
		}
	}
}

func TestNestedKeyFailures(t *testing.T) {
	if _, err := New(maptypes.PropertyMap{
		"a": {
			Value: gosr.String("${${a}}"),
		},
	}); err == nil {
		t.Fatal("nested key cycle admitted")
	}
	r, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	text := "missing"
	for range 65 {
		text = "${" + text + ":missing}"
	}
	if _, err := r.Resolve(text); err == nil {
		t.Fatal("nested key depth admitted")
	}
	if got, err := r.Resolve(`${\${absent}:fallback}`); err != nil || got != "fallback" {
		t.Fatal(got, err)
	}
}

func TestNestedKeyAdmissionBoundary(t *testing.T) {
	raw := maptypes.PropertyMap{
		"a": {
			Value: gosr.String("${${z}:ok}"),
		},
		"z": {
			Value: gosr.String("missing"),
		},
	}
	// Two property keys, cached selector output, intermediate key and result.
	policy := limits.Default()
	policy.MaxResolvedBytes = 2 + 7 + 7 + 2
	if _, err := NewWithPolicy(context.Background(), raw, policy); err != nil {
		t.Fatal(err)
	}
	policy.MaxResolvedBytes--
	_, err := NewWithPolicy(context.Background(), raw, policy)
	var exceeded *limits.ExceededError
	if !errors.As(err, &exceeded) || exceeded.Resource != limits.ResolvedBytes {
		t.Fatalf("intermediate admission: %v", err)
	}
}

func TestNestedKeyIntermediateOutputBoundary(t *testing.T) {
	r, err := New(maptypes.PropertyMap{
		"selector": {
			Value: gosr.String(strings.Repeat("x", maxOutput)),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := r.Resolve("${${selector}:ok}"); err != nil || got != "ok" {
		t.Fatal(got, err)
	}
	if _, err := r.Resolve("${${selector}x:ok}"); !errors.Is(err, &gosr.Error{
		Code: gosr.ResolutionLimit,
	}) {
		t.Fatalf("intermediate output: %v", err)
	}
}

func TestNestedKeyCachedReferenceDepth(t *testing.T) {
	for _, count := range []int{62, 63} {
		raw := maptypes.PropertyMap{}
		for i := range count {
			next := "missing"
			if i+1 < count {
				next = fmt.Sprintf("${k%02d}", i+1)
			}
			raw[fmt.Sprintf("k%02d", i)] = gosr.Property{
				Value: gosr.String(next),
			}
		}
		raw["z"] = gosr.Property{
			Value: gosr.String("${${k00}:ok}"),
		}
		r, err := New(raw)
		if err != nil {
			t.Fatal(err)
		}
		got, err := r.Resolve("${z}")
		if count == 62 && (err != nil || got != "ok") {
			t.Fatal(got, err)
		}
		if count == 63 && !errors.Is(err, &gosr.Error{
			Code: gosr.ResolutionLimit,
		}) {
			t.Fatalf("cached nested depth: %v", err)
		}
	}
}
