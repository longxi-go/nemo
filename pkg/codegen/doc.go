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

// Package codegen generates configuration binding methods and validates their
// output. Controller-tools parses +nemo:config:prefix markers; AST and Go type
// analysis validate named, non-generic configuration structs and their field tags.
// Temporary method stubs permit calls to not-yet-generated methods during type
// analysis. Unrelated type errors fail generation.
//
// Supported fields include named scalars, durations, TextUnmarshaler scalars,
// scalar pointers, scalar lists, string-keyed maps, and named or anonymous local
// nested structs. Object collections and recursive object pointers use typed
// helpers with bounded traversal. External nested structs must implement
// BindProperties or UnmarshalText. Generic structs, bound embedded fields,
// interfaces, channels, functions, arrays and nested collections are rejected.
// Unexported fields require property:"-".
//
// Generated binders prepare a separate candidate and assign it after all fields
// succeed. Local nested structs retain skipped fields. External nested binders and
// object collection elements start from zero; their skipped fields do not inherit
// old state. Bound reference fields use replacement storage. Handwritten external
// binders must preserve the target on failure and avoid modifying shared state.
//
// Run writes selected main-module packages after all generated overlays pass type
// checking. Output files use atomic per-file rename; filesystem failures may leave
// earlier packages updated. Semantic failures write nothing. Inspect and check
// mode do not modify source or module metadata. Package loading uses readonly
// module mode. Dependencies and the module cache are not output destinations.
//
// Only regular files with the standard nemo-generated header may be replaced;
// handwritten files and symlinks are rejected. Removing a final annotation replaces
// owned output with a package-only generated file after validation removes obsolete
// methods. Check mode rejects stale methods until regeneration. Generated source
// is formatted with golayout before validation and writing.
package codegen
