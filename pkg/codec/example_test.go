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
	"fmt"

	"github.com/longxi-go/nemo/pkg/codec"
)

func ExampleForFormat() {
	decoder, err := codec.ForFormat("json")
	if err != nil {
		panic(err)
	}
	values, err := decoder.Decode([]byte(`{"app":{"port":8080}}`))
	if err != nil {
		panic(err)
	}
	text, _ := values["app.port"].Scalar()
	fmt.Println(text)
	// Output: 8080
}

func ExampleRegistryBuilder() {
	builder := codec.NewRegistryBuilder()
	if err := codec.RegisterBuiltins(builder); err != nil {
		panic(err)
	}
	registry := builder.Build()
	decoder, err := registry.Lookup("JSON")
	if err != nil {
		panic(err)
	}
	values, err := decoder.Decode([]byte(`{"port":8080}`))
	if err != nil {
		panic(err)
	}
	text, _ := values["port"].Scalar()
	fmt.Println(text)
	// Output: 8080
}
