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
	"net/netip"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

var _ binding.PropertyPrefixProvider = (*handwritten)(nil)

func TestBindPrefixSelection(t *testing.T) {
	r := resolver(t, maptypes.ValueMap{
		"name":          gosr.String("root"),
		"app.name":      gosr.String("metadata"),
		"explicit.name": gosr.String("explicit"),
	})
	metadata, err := binding.Bind[handwritten](r, "")
	if err != nil || metadata.Value != "metadata" {
		t.Fatalf("inferred target and metadata prefix: %#v %v", metadata, err)
	}
	explicit, err := binding.Bind[handwritten, *handwritten](r, "explicit")
	if err != nil || explicit.Value != "explicit" {
		t.Fatalf("explicit target and prefix override: %#v %v", explicit, err)
	}
	root, err := binding.Bind[prefixlessConfig](r, "")
	if err != nil || root.Value != "root" {
		t.Fatalf("binder without prefix metadata: %#v %v", root, err)
	}
	var direct handwritten
	if err := direct.BindProperties(r, ""); err != nil || direct.Value != "root" {
		t.Fatalf("direct binding must retain root semantics: %#v %v", direct, err)
	}
}

func TestTextPointerInference(t *testing.T) {
	input := gosr.String("192.0.2.1")
	want := netip.MustParseAddr("192.0.2.1")
	inferred, err := binding.Text[netip.Addr](input)
	if err != nil || inferred != want {
		t.Fatalf("inferred text pointer: %v %v", inferred, err)
	}
	explicit, err := binding.Text[netip.Addr, *netip.Addr](input)
	if err != nil || explicit != want {
		t.Fatalf("explicit text pointer: %v %v", explicit, err)
	}
}

type prefixlessConfig struct {
	Value string
}

func (c *prefixlessConfig) BindProperties(r gosr.PropertyResolver, prefix string) error {
	value, err := binding.Read(r, binding.Join(prefix, "name"), binding.ReadOptions{}, binding.String[string])
	if err == nil {
		c.Value = value
	}
	return err
}
