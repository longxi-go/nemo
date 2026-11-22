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

package compiler

type fieldModel struct {
	Name       string
	Key        string
	Required   bool
	HasDefault bool
	Default    string
	Mode       fieldMode
	Pointer    bool
	Recursive  bool
	Type       string
	Helper     string
	Reader     string
	Decoder    string
	Element    *objectModel
}

type helperModel struct {
	Name   string
	Type   string
	Fields []fieldModel
}

type binderModel struct {
	Name   string
	Prefix string
	Helper string
}

type packageModel struct {
	Names      generatedNames
	NewPointer string
	Header     string
	Name       string
	Imports    []string
	Binders    []binderModel
	Helpers    []helperModel
}

type preparedPackage struct {
	model packageModel
	file  File
}

type objectModel struct {
	Type     string
	Base     string
	Pointer  bool
	External bool
	Helper   string
}
