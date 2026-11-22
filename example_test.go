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

package nemo_test

import (
	"context"
	"fmt"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func ExampleNew() {
	properties, err := source.NewMap("bootstrap", maptypes.ValueMap{
		"app.name": gosr.String("longxi"),
	})
	if err != nil {
		panic(err)
	}
	app, err := nemo.New(context.Background(),
		environment.WithProviders(source.Static(properties)),
		environment.WithProfiles("production"),
	)
	if err != nil {
		panic(err)
	}
	view := app.Snapshot()
	name, err := view.Resolve("${app.name}")
	if err != nil {
		panic(err)
	}
	fmt.Println(view.Revision(), name, view.ActiveProfiles())
	// Output: 1 longxi [production]
}
