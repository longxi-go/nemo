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
	"slices"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestNestedExactObjectReplacesAncestorProjection(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"app": gosr.Object(maptypes.ValueMap{
			"db": gosr.Object(maptypes.ValueMap{
				"host":    gosr.String("ancestor"),
				"removed": gosr.String("ancestor-only"),
			}),
			"name": gosr.String("application"),
		}),
		"app.db": gosr.Object(maptypes.ValueMap{
			"host": gosr.String("override"),
			"port": gosr.String("object-port"),
		}),
		"app.db.port": gosr.String("exact-port"),
	})
	parent, err := binding.Nested(r, "app", true)
	if err != nil {
		t.Fatal(err)
	}
	child, err := binding.Nested(parent, "app.db", true)
	if err != nil {
		t.Fatal(err)
	}
	expectedValues := maptypes.StringMap[string]{
		"app.db.host": "override",
		"app.db.port": "exact-port",
		"app.name":    "application",
	}
	for key, want := range expectedValues {
		property, found, err := child.Lookup(key)
		if err != nil || !found || !property.Value.Equal(gosr.String(want)) {
			t.Errorf("%s: got %#v, found %v, error %v; want %q", key, property, found, err, want)
		}
	}
	host, _, err := child.Lookup("app.db.host")
	if err != nil || host.Origin.Key != "app.db" {
		t.Errorf("replacement member lost its object's origin: %#v, error %v", host.Origin, err)
	}
	if _, found, err := child.Lookup("app.db.removed"); err != nil || found {
		t.Errorf("ancestor-only member leaked into replacement: found %v, error %v", found, err)
	}
	if slices.Contains(child.Keys(), "app.db.removed") {
		t.Error("replacement advertises ancestor-only member")
	}
	before, found, err := parent.Lookup("app.db.host")
	if err != nil || !found || !before.Value.Equal(gosr.String("ancestor")) {
		t.Errorf("nesting mutated the parent view: %#v, found %v, error %v", before, found, err)
	}
	if _, found, err := parent.Lookup("app.db.removed"); err != nil || !found {
		t.Errorf("nesting removed data from the parent view: found %v, error %v", found, err)
	}
	for _, key := range child.Keys() {
		if _, found, err := child.Lookup(key); err != nil || !found {
			t.Errorf("advertised key %s unavailable: found %v, error %v", key, found, err)
		}
	}
}

func TestNestedEmptyExactObjectClearsAncestorProjection(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"app": gosr.Object(maptypes.ValueMap{
			"db": gosr.Object(maptypes.ValueMap{
				"host": gosr.String("ancestor"),
			}),
		}),
		"app.db":      gosr.Object(nil),
		"app.db.port": gosr.String("exact-port"),
	})
	parent, err := binding.Nested(r, "app", true)
	if err != nil {
		t.Fatal(err)
	}
	child, err := binding.Nested(parent, "app.db", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, found, err := child.Lookup("app.db.host"); err != nil || found {
		t.Fatalf("empty replacement inherited host: found %v, error %v", found, err)
	}
	if slices.Contains(child.Keys(), "app.db.host") {
		t.Error("empty replacement advertises ancestor host")
	}
	port, err := binding.Read(child, "app.db.port", binding.ReadOptions{}, binding.String[string])
	if err != nil || port != "exact-port" {
		t.Fatalf("empty replacement hid independent exact address: %q, error %v", port, err)
	}
}
