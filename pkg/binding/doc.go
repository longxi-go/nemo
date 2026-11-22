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

// Package binding provides typed conversion and compile-time constraints for
// configuration binders. BindablePointer and TextUnmarshalerPointer support pointer
// receivers and generic type inference without runtime reflection.
//
// Bind captures one snapshot when given a gosr.SnapshotProvider and uses
// PropertyPrefixProvider metadata when its prefix is empty. Direct BindProperties
// calls interpret an empty prefix as the root.
// Defaults apply only to missing properties; null and invalid values fail binding.
// Scalar pointers use new storage, custom scalars use encoding.TextUnmarshaler,
// and scalar list defaults use CSV.
//
// Generated binders assign all fields after successful conversion. Local named and
// anonymous nested structs retain skipped fields. External objects and collection
// elements start from zero to isolate the previous target's references. ObjectInput
// provides a separate property view for each collection element. Optional recursive
// pointers terminate on absent input; MaxObjectDepth bounds each generated private
// helper tree.
//
// Conversion errors retain parser and custom-converter causes for errors.Is and
// errors.As, but omit their messages from the printable error.
package binding
