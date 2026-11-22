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
	"errors"
	"reflect"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func objectResolver(t *testing.T, values maptypes.ValueMap) gosr.PropertyResolver {
	t.Helper()
	snapshot, err := source.NewMap("objects", values)
	if err != nil {
		t.Fatal(err)
	}
	env, err := environment.New(context.Background(), environment.WithProviders(source.Static(snapshot)))
	if err != nil {
		t.Fatal(err)
	}
	return env.Snapshot()
}

func TestObjectListScopeIsolation(t *testing.T) {
	resolver := objectResolver(t, maptypes.ValueMap{
		"app.servers": gosr.List(
			gosr.Object(maptypes.ValueMap{
				"port": gosr.String("80"),
			}),
			gosr.Object(maptypes.ValueMap{
				"port": gosr.String("secret invalid"),
			}),
		),
		"app.servers.port": gosr.String("999"),
	})
	inputs, err := binding.ListObjects(resolver, "app.servers", true)
	if err != nil || len(inputs) != 2 {
		t.Fatalf("inputs: %v %v", inputs, err)
	}
	port, err := binding.Read(inputs[0].Resolver, binding.Join(inputs[0].Prefix, "port"), binding.ReadOptions{}, binding.Signed[int])
	if err != nil || port != 80 {
		t.Fatalf("mixed element scope: %v %v", port, err)
	}
	_, err = binding.Read(inputs[1].Resolver, binding.Join(inputs[1].Prefix, "port"), binding.ReadOptions{}, binding.Signed[int])
	wrapped := inputs[1].Wrap(err)
	var detail *gosr.Error
	if !errors.As(wrapped, &detail) || detail.Key != "app.servers" || detail.Origin.Source != "objects" || !reflect.DeepEqual(detail.Path, []string{"[1]", "port"}) {
		t.Fatalf("element diagnostic: %#v", detail)
	}
	var original *gosr.Error
	if !errors.As(err, &original) || original.Key != "app.servers.port" || len(original.Path) != 0 {
		t.Fatal("Wrap modified its input error")
	}
	for _, expression := range []string{"${global.port}", "${app.servers.port}"} {
		_, absolute := inputs[1].Resolver.Resolve(expression)
		// The second expression succeeds; a distinct failing resolver below covers the collision.
		if absolute != nil && inputs[1].Wrap(absolute) != absolute {
			t.Fatal("absolute error replaced")
		}
	}
	unknown := errors.New("external failure")
	if inputs[1].Wrap(unknown) != unknown {
		t.Fatal("unknown external error changed")
	}
	if inputs[1].Wrap(nil) != nil {
		t.Fatal("nil wrapped")
	}
}

func TestObjectInputAbsoluteErrorProvenance(t *testing.T) {
	resolver := objectResolver(t, maptypes.ValueMap{
		"app.servers": gosr.List(gosr.Object(nil)),
	})
	inputs, err := binding.ListObjects(resolver, "app.servers", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, expression := range []string{"${global.port}", "${app.servers.port}"} {
		_, err := inputs[0].Resolver.Resolve(expression)
		var before *gosr.Error
		if !errors.As(err, &before) {
			t.Fatal(err)
		}
		wrapped := inputs[0].Wrap(err)
		var after *gosr.Error
		if !errors.As(wrapped, &after) || !reflect.DeepEqual(before, after) || after.Origin.Source != "" {
			t.Fatalf("absolute provenance changed: %#v", after)
		}
	}
}

func TestObjectMapSources(t *testing.T) {
	resolver := objectResolver(t, maptypes.ValueMap{
		"exact": gosr.Object(maptypes.ValueMap{
			"us.east": gosr.Object(maptypes.ValueMap{
				"port": gosr.String("80"),
			}),
		}),
		"exact.us.east.port": gosr.String("999"),
		"flat.east.port":     gosr.String("81"),
		"flat.west.port":     gosr.String("82"),
		"emptylist":          gosr.List(),
		"emptymap":           gosr.Object(nil),
		"null":               {},
		"scalar":             gosr.String("bad"),
		"badlist":            gosr.List(gosr.Value{}),
	})
	exact, err := binding.MapObjects(resolver, "exact", true)
	if err != nil || len(exact) != 1 {
		t.Fatalf("exact map: %v %v", exact, err)
	}
	port, err := binding.Read(exact["us.east"].Resolver, binding.Join(exact["us.east"].Prefix, "port"), binding.ReadOptions{}, binding.Signed[int])
	if err != nil || port != 80 {
		t.Fatalf("literal member/exact precedence: %v %v", port, err)
	}
	flat, err := binding.MapObjects(resolver, "flat", true)
	if err != nil || len(flat) != 2 {
		t.Fatalf("flat map: %v %v", flat, err)
	}
	expectedValues := maptypes.StringMap[int]{
		"east": 81,
		"west": 82,
	}
	for name, expected := range expectedValues {
		input := flat[name]
		port, err := binding.Read(input.Resolver, binding.Join(input.Prefix, "port"), binding.ReadOptions{}, binding.Signed[int])
		if err != nil || port != expected {
			t.Fatalf("flat %s: %v %v", name, port, err)
		}
	}
	list, err := binding.ListObjects(resolver, "emptylist", false)
	if err != nil || list == nil || len(list) != 0 {
		t.Fatal("explicit empty list lost")
	}
	object, err := binding.MapObjects(resolver, "emptymap", false)
	if err != nil || object == nil || len(object) != 0 {
		t.Fatal("explicit empty map lost")
	}
	list, err = binding.ListObjects(resolver, "missing", false)
	if err != nil || list != nil {
		t.Fatal("missing list not nil")
	}
	object, err = binding.MapObjects(resolver, "missing", false)
	if err != nil || object != nil {
		t.Fatal("missing map not nil")
	}
	for _, key := range []string{"null", "scalar", "badlist"} {
		if _, err := binding.ListObjects(resolver, key, false); !errors.Is(err, &gosr.Error{
			Code: gosr.ConversionFailed,
		}) {
			t.Fatalf("%s: %v", key, err)
		}
	}
	if _, err := binding.ListObjects(resolver, "missing", true); !errors.Is(err, &gosr.Error{
		Code: gosr.MissingProperty,
	}) {
		t.Fatal(err)
	}
	if _, err := binding.MapObjects(resolver, "missing", true); !errors.Is(err, &gosr.Error{
		Code: gosr.MissingProperty,
	}) {
		t.Fatal(err)
	}
}
