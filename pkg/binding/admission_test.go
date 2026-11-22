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
	"errors"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

type admissionResolver struct {
	gosr.PropertySource
}

func (admissionResolver) Resolve(text string) (string, error) {
	return text, nil
}

func TestProjectionRejectsDeepExternalResolver(t *testing.T) {
	value := gosr.String("leaf")
	for range 80 {
		value = gosr.Object(maptypes.ValueMap{
			"child": value,
		})
	}
	raw, _ := source.NewMap("external", maptypes.ValueMap{
		"app":  value,
		"list": gosr.List(value),
		"map": gosr.Object(maptypes.ValueMap{
			"entry": value,
		}),
	})
	r := admissionResolver{
		raw,
	}
	for _, check := range []func() error{
		func() error {
			_, err := binding.Nested(r, "app", true)
			return err
		},
		func() error {
			_, err := binding.ListObjects(r, "list", true)
			return err
		},
		func() error {
			_, err := binding.MapObjects(r, "map", true)
			return err
		},
	} {
		if err := check(); !errors.Is(err, &gosr.Error{
			Code: gosr.ResolutionLimit,
		}) {
			t.Fatal("external deep object admitted", err)
		}
	}
}
