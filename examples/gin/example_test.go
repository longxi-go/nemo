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

package main

import (
	"context"
	"fmt"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func ExampleConfig() {
	properties, err := source.NewMap("application", maptypes.ValueMap{
		"company.platform.service.name":    gosr.String("alpha"),
		"company.platform.service.address": gosr.String("127.0.0.1:8081"),
	})
	if err != nil {
		panic(err)
	}
	owner, err := environment.New(context.Background(),
		environment.WithProviders(source.Static(properties)),
		environment.WithValidator(validateSnapshot),
	)
	if err != nil {
		panic(err)
	}
	config, err := binding.Bind[Config](owner.Snapshot(), "")
	if err != nil {
		panic(err)
	}
	fmt.Println(config.Name, config.Address)
	// Output: alpha 127.0.0.1:8081
}
