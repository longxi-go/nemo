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

package maptypes_test

import (
	"fmt"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func ExampleStringMap() {
	ports := maptypes.StringMap[int]{
		"http": 8080,
	}
	fmt.Println(ports["http"])
	// Output: 8080
}

func ExampleIntMap() {
	labels := maptypes.IntMap[string]{
		200: "OK",
	}
	fmt.Println(labels[200])
	// Output: OK
}

func ExampleValueMap() {
	values := maptypes.ValueMap{
		"app.name": gosr.String("demo"),
	}
	properties, err := source.NewMap("defaults", values)
	if err != nil {
		panic(err)
	}
	property, found, err := properties.Lookup("app.name")
	if err != nil {
		panic(err)
	}
	text, scalar := property.Value.Scalar()
	fmt.Println(found, scalar, text)
	// Output: true true demo
}

func ExamplePropertyMap() {
	values := maptypes.PropertyMap{
		"app.name": {
			Value: gosr.String("demo"),
			Origin: gosr.Origin{
				Resource: "defaults.json",
			},
		},
	}
	properties, err := source.NewProperties("defaults", values)
	if err != nil {
		panic(err)
	}
	property, _, err := properties.Lookup("app.name")
	if err != nil {
		panic(err)
	}
	fmt.Println(property.Origin.Source, property.Origin.Resource)
	// Output: defaults defaults.json
}

func ExampleMap() {
	labels := maptypes.Map[[2]int, string]{
		{
			1,
			2,
		}: "origin",
	}
	fmt.Println(labels[[2]int{1, 2}])
	// Output: origin
}

func ExampleStringBoolMap() {
	selected := maptypes.StringBoolMap{
		"primary": true,
	}
	fmt.Println(selected["primary"], selected["missing"])
	// Output: true false
}

func ExampleStringBytesMap() {
	files := maptypes.StringBytesMap{
		"config": []byte("name=nemo"),
	}
	fmt.Println(string(files["config"]))
	// Output: name=nemo
}

func ExampleStringAnyMap() {
	object := maptypes.StringAnyMap{
		"name": "nemo",
		"port": 8080,
	}
	fmt.Println(object["name"], object["port"])
	// Output: nemo 8080
}
