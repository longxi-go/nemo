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

// Package maptypes provides map aliases for Nemo APIs and application code.
// Map supports any comparable key. StringMap and IntMap fix the key type;
// StringBoolMap, StringBytesMap and StringAnyMap also fix common value types.
// ValueMap and PropertyMap name configuration entries. Aliases retain map identity, zero values,
// ownership and concurrency requirements. They add no validation or copying.
package maptypes
