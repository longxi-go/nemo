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

// Map retains native map identity for caller-selected keys and values.
// Comparable keys include compiler nodes and type descriptors as well as scalars.
type Map[K comparable, V any] = map[K]V

// StringMap is a map with string keys and caller-selected values.
// It is interchangeable with map[string]V, including in function signatures.
type StringMap[V any] = Map[string, V]

// IntMap is a map with int keys and caller-selected values.
// It is interchangeable with map[int]V, including in function signatures.
type IntMap[V any] = Map[int, V]
