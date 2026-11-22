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

import (
	"github.com/longxi-go/nemo/pkg/maptypes"
	"golang.org/x/tools/go/packages"
)

type declaration struct {
	name   string
	prefix string
}

type analysis struct {
	packages     []*packages.Package
	declarations maptypes.StringMap[[]declaration]
	owned        maptypes.StringBoolMap
	overlay      maptypes.StringBytesMap
}

// generation stores the semantic models and rendered output for one compiler
// invocation. Pipeline steps share this typed state.
type generation struct {
	options   Options
	inspect   bool
	analysis  analysis
	models    []preparedPackage
	outputs   []output
	report    Report
	validated bool
}
