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
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

func TestEquivalentFormats(t *testing.T) {
	cases := maptypes.StringMap[string]{
		"yaml":       "app:\n  name: demo\n  enabled: true\n",
		"yml":        "app: {name: demo, enabled: true}",
		"json":       `{"app":{"name":"demo","enabled":true}}`,
		"toml":       "[app]\nname='demo'\nenabled=true",
		"ini":        "[app]\nname=demo\nenabled=true",
		"properties": "app.name=demo\napp.enabled=true",
	}
	for format, input := range cases {
		t.Run(format, func(t *testing.T) {
			d, err := codec.ForFormat(format)
			if err != nil {
				t.Fatal(err)
			}
			v, err := d.Decode([]byte(input))
			if err != nil {
				t.Fatal(err)
			}
			if !v["app.name"].Equal(gosr.String("demo")) {
				t.Fatal(v)
			}
			if s, _ := v["app.enabled"].Scalar(); s != "true" {
				t.Fatal(v)
			}
		})
	}
}

func TestTypedCollectionsAndNull(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			d, _ := codec.ForFormat(format)
			v, err := d.Decode([]byte(`{"app":{"port":8080,"items":[true,null,{"x":2}],"nil":null,"empty":{}}}`))
			if err != nil {
				t.Fatal(err)
			}
			if v["app.port"].Kind() != gosr.NumberKind || v["app.nil"].Kind() != gosr.NullKind || v["app.empty"].Kind() != gosr.ObjectKind {
				t.Fatal(v)
			}
			items := v["app.items"].Items()
			if len(items) != 3 || items[0].Kind() != gosr.BoolKind || items[2].Fields()["x"].Kind() != gosr.NumberKind {
				t.Fatal(items)
			}
		})
	}
}

func TestRejectedDocuments(t *testing.T) {
	cases := []struct {
		format, input string
	}{
		{
			"json",
			`{"a.b":1,"a":{"b":2}}`,
		},
		{
			"yaml",
			"a.b: 1\na: {b: 2}",
		},
		{
			"yaml",
			"1: secret",
		},
		{
			"yaml",
			"a: .nan",
		},
		{
			"toml",
			"a=inf",
		},
		{
			"json",
			`{"a":1,"a":2}`,
		},
		{
			"yaml",
			"a: 1\na: 2",
		},
		{
			"yaml",
			"a: &a [*a]",
		},
		{
			"json",
			`[]`,
		},
		{
			"json",
			`{} {}`,
		},
		{
			"yaml",
			"a: 1\n---\nb: 2",
		},
		{
			"properties",
			"=secret",
		},
		{
			"ini",
			"a.b=1\n[a]\nb=2\n",
		},
	}
	for _, tt := range cases {
		t.Run(tt.format+tt.input, func(t *testing.T) {
			d, _ := codec.ForFormat(tt.format)
			if _, err := d.Decode([]byte(tt.input)); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
	if _, err := codec.ForFormat("custom"); err == nil {
		t.Fatal("unsupported format accepted")
	}
	var f codec.DecoderFunc
	if _, err := f.Decode(nil); err == nil {
		t.Fatal("nil function accepted")
	}
}

func TestCustomDecoder(t *testing.T) {
	d := codec.DecoderFunc(func(b []byte) (map[string]gosr.Value, error) {
		return maptypes.ValueMap{
			"custom": gosr.String(string(b)),
		}, nil
	})
	v, err := d.Decode([]byte("ok"))
	if err != nil || !v["custom"].Equal(gosr.String("ok")) {
		t.Fatal(v, err)
	}
}

func TestYAMLIntegerPrecision(t *testing.T) {
	d, err := codec.ForFormat("yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"18446744073709551616", "-9223372036854775809", "0x10000000000000000", "999999999999999999999999999999999999999"} {
		values, err := d.Decode([]byte("number: " + text))
		if err == nil {
			scalar, _ := values["number"].Scalar()
			if scalar != text {
				t.Fatalf("silently rounded %s to %s", text, scalar)
			}
		}
	}
	for _, text := range []string{"18446744073709551615", "-9223372036854775808", "9007199254740993"} {
		values, err := d.Decode([]byte("number: " + text))
		if err != nil {
			t.Fatal(err)
		}
		scalar, _ := values["number"].Scalar()
		if scalar != text {
			t.Fatalf("integer precision lost: %s", scalar)
		}
	}
}

func TestPlaceholdersRemainRaw(t *testing.T) {
	documents := maptypes.StringMap[string]{
		"properties": "name=demo\nref=${name}\n",
		"ini":        "name=demo\nref=%(name)s\n",
	}
	for format, input := range documents {
		d, err := codec.ForFormat(format)
		if err != nil {
			t.Fatal(err)
		}
		values, err := d.Decode([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		want := "${name}"
		if format == "ini" {
			want = "%(name)s"
		}
		if !values["ref"].Equal(gosr.String(want)) {
			t.Fatal(values)
		}
	}
}

func TestNativeDuplicates(t *testing.T) {
	documents2 := maptypes.StringMap[string]{
		"properties": "key=first\nkey=second",
		"ini":        "[app]\nkey=first\n[app]\nkey=second",
	}
	for format, input := range documents2 {
		d, err := codec.ForFormat(format)
		if err != nil {
			t.Fatal(err)
		}
		values, err := d.Decode([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		key := "key"
		if format == "ini" {
			key = "app.key"
		}
		if !values[key].Equal(gosr.String("second")) {
			t.Fatal(values)
		}
	}
}

func TestTOMLArraysAndYAMLAliases(t *testing.T) {
	cases := maptypes.StringMap[string]{
		"toml": "[[app.members]]\nname='demo'\n",
		"yaml": "defaults: &defaults\n  name: demo\napp:\n  members: [*defaults]\n",
	}
	for format, input := range cases {
		d, err := codec.ForFormat(format)
		if err != nil {
			t.Fatal(err)
		}
		values, err := d.Decode([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		items := values["app.members"].Items()
		if len(items) != 1 || !items[0].Fields()["name"].Equal(gosr.String("demo")) {
			t.Fatal(values)
		}
	}
}
