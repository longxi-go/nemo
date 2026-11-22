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
	"fmt"

	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
)

func ExampleSigned() {
	n, err := binding.Signed[int16](gosr.String("8080"))
	fmt.Println(n, err) // Output: 8080 <nil>
}

func ExampleListDecoder() {
	decode := binding.ListDecoder(binding.Signed[int])
	values, err := decode(gosr.String("8080,9090"))
	fmt.Println(values, err)
	// Output: [8080 9090] <nil>
}

func ExampleListObjects() {
	properties, err := source.NewMap("example", maptypes.ValueMap{
		"servers": gosr.List(
			gosr.Object(maptypes.ValueMap{
				"port": gosr.String("8080"),
			}),
		),
	})
	if err != nil {
		panic(err)
	}
	env, err := environment.New(context.Background(), environment.WithProviders(source.Static(properties)))
	if err != nil {
		panic(err)
	}
	inputs, err := binding.ListObjects(env.Snapshot(), "servers", true)
	if err != nil {
		panic(err)
	}
	port, err := binding.Read(inputs[0].Resolver, binding.Join(inputs[0].Prefix, "port"), binding.ReadOptions{}, binding.Signed[int])
	if err != nil {
		panic(inputs[0].Wrap(err))
	}
	fmt.Println(port)
	// Output: 8080
}
