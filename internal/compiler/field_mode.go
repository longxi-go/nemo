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

// fieldMode selects a binding strategy shared by modeling and rendering.
// Templates use predicates so their control flow cannot drift from these values.
type fieldMode string

const (
	fieldModeRead     fieldMode = "read"
	fieldModeLocal    fieldMode = "local"
	fieldModeExternal fieldMode = "external"
	fieldModeList     fieldMode = "list"
	fieldModeMap      fieldMode = "map"
)

func (m fieldMode) valid() bool {
	switch m {
	case fieldModeRead, fieldModeLocal, fieldModeExternal, fieldModeList, fieldModeMap:
		return true
	default:
		return false
	}
}

func (m fieldMode) IsRead() bool {
	return m == fieldModeRead
}

func (m fieldMode) IsLocal() bool {
	return m == fieldModeLocal
}

func (m fieldMode) IsExternal() bool {
	return m == fieldModeExternal
}

func (m fieldMode) IsList() bool {
	return m == fieldModeList
}

func (m fieldMode) IsMap() bool {
	return m == fieldModeMap
}

func (m fieldMode) IsCollection() bool {
	return m.IsList() || m.IsMap()
}
