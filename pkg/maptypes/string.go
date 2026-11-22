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

package maptypes

// StringBoolMap names string-keyed flags without introducing set semantics.
// Absent entries read as false, as with an ordinary map[string]bool.
type StringBoolMap = StringMap[bool]

// StringBytesMap associates names with byte slices, such as source overlays.
// Neither the map nor its slices are copied by this alias.
type StringBytesMap = StringMap[[]byte]

// StringAnyMap represents unnormalized object trees from configuration parsers.
// Use ValueMap after converting entries to gosr.Value.
type StringAnyMap = StringMap[any]
