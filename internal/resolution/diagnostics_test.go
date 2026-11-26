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
	"strings"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const diagnosticSentinel = "private_value_8f96"

func TestDynamicKeyDiagnosticsRedactExpandedValues(t *testing.T) {
	for _, scenario := range []struct {
		name string
		text string
		code gosr.Code
	}{
		{
			name: "missing",
			text: "${${selector}}",
			code: gosr.MissingProperty,
		},
		{
			name: "nested",
			text: "${${${route}}}",
			code: gosr.MissingProperty,
		},
		{
			name: "default-derived-key",
			text: "${${absent:" + diagnosticSentinel + "}}",
			code: gosr.MissingProperty,
		},
		{
			name: "non-scalar",
			text: "${${selector}.object}",
			code: gosr.InvalidValue,
		},
		{
			name: "null",
			text: "${${selector}.null}",
			code: gosr.InvalidValue,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resolver, err := New(maptypes.PropertyMap{
				"route": {
					Value: gosr.String("selector"),
				},
				"selector": {
					Value: gosr.String(diagnosticSentinel),
				},
				diagnosticSentinel + ".object": {
					Value: gosr.Object(nil),
				},
				diagnosticSentinel + ".null": {},
			})
			if err != nil {
				t.Fatal(err)
			}
			text, err := resolver.Resolve(scenario.text)
			if text != "" || !errors.Is(err, &gosr.Error{
				Code: scenario.code,
			}) {
				t.Fatalf("expected classified failure without partial output: %q %v", text, err)
			}
			assertSafeReferenceDiagnostic(t, err)
		})
	}
}

func TestDynamicOperandDiagnosticsRetainSafeOuterOrigin(t *testing.T) {
	for _, scenario := range []struct {
		name string
		text string
		code gosr.Code
	}{
		{
			name: "missing",
			text: "${absent}",
			code: gosr.MissingProperty,
		},
		{
			name: "cycle",
			text: "${a.entry}",
			code: gosr.ResolutionCycle,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			_, err := New(maptypes.PropertyMap{
				"a.entry": {
					Value: gosr.String("${${selector}}"),
					Origin: gosr.Origin{
						Source:   "application",
						Resource: "application.yaml",
						Line:     7,
					},
				},
				"selector": {
					Value: gosr.String(diagnosticSentinel),
				},
				diagnosticSentinel: {
					Value: gosr.String(scenario.text),
				},
			})
			if !errors.Is(err, &gosr.Error{
				Code: scenario.code,
			}) {
				t.Fatalf("lost failure classification: %v", err)
			}
			assertSafeReferenceDiagnostic(t, err)
			var detail *gosr.Error
			if !errors.As(err, &detail) || detail.Key != "a.entry" ||
				detail.Origin.Source != "application" || detail.Origin.Resource != "application.yaml" ||
				detail.Origin.Line != 7 {
				t.Fatalf("lost owning declaration provenance: %v", err)
			}
		})
	}
}

func TestDynamicOperandDiagnosticsPreserveOperationalCause(t *testing.T) {
	cause := &gosr.Error{
		Code:    gosr.ResolutionLimit,
		Key:     diagnosticSentinel,
		Path:    []string{diagnosticSentinel},
		Message: diagnosticSentinel,
		Cause:   context.Canceled,
	}
	resolve := func(key string) (gosr.Value, int, bool, error) {
		if key == "selector" {
			return gosr.String(diagnosticSentinel), 0, true, nil
		}
		return gosr.Value{}, 0, false, cause
	}
	p := parser{
		operand: resolve,
	}
	_, _, err := p.expand("${${selector}}")
	if !errors.Is(err, cause) || !errors.Is(err, context.Canceled) {
		t.Fatalf("lost classified or operational cause: %v", err)
	}
	assertSafeReferenceDiagnostic(t, err)
	_, _, err = p.expand("${static}")
	if err != cause {
		t.Fatal("static reference replaced its original diagnostic")
	}
}

func TestStaticReferenceDiagnosticsRemainAvailable(t *testing.T) {
	resolver, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		text string
		key  string
	}{
		{
			text: "${missing}",
			key:  "missing",
		},
		{
			text: `${\${literal}}`,
			key:  "${literal}",
		},
	} {
		_, err := resolver.Resolve(scenario.text)
		var detail *gosr.Error
		if !errors.As(err, &detail) || detail.Key != scenario.key || detail.Code != gosr.MissingProperty {
			t.Fatalf("static key diagnostic changed: %v", err)
		}
	}
}

func assertSafeReferenceDiagnostic(t *testing.T, err error) {
	t.Helper()
	if err == nil || strings.Contains(err.Error(), diagnosticSentinel) {
		t.Fatalf("expanded property value reached printable diagnostic: %v", err)
	}
	var detail *gosr.Error
	if !errors.As(err, &detail) || strings.Contains(detail.Key, diagnosticSentinel) ||
		strings.Contains(strings.Join(detail.Path, "/"), diagnosticSentinel) ||
		strings.Contains(detail.Message, diagnosticSentinel) {
		t.Fatalf("expanded property value reached diagnostic metadata: %#v", detail)
	}
}
