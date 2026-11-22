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

package integration_test

import (
	"context"
	"fmt"

	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/source"
)

func Example() {
	decoder, err := codec.ForFormat("json")
	if err != nil {
		panic(err)
	}
	values, err := decoder.Decode([]byte(`{"server":{"port":8080}}`))
	if err != nil {
		panic(err)
	}
	properties, err := source.NewMap("document", values)
	if err != nil {
		panic(err)
	}
	owner, err := environment.New(context.Background(),
		environment.WithProviders(source.Static(properties)),
	)
	if err != nil {
		panic(err)
	}
	port, err := binding.Read(owner.Snapshot(), "server.port", binding.ReadOptions{
		Required: true,
	}, binding.Unsigned[uint16])
	if err != nil {
		panic(err)
	}
	fmt.Println(port)
	// Output: 8080
}
