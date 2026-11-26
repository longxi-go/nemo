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

// Package resolution resolves all selected properties before a snapshot is
// published. It parses placeholders, expands references and enforces depth and
// output limits independently of transport, profiles and notifications.
//
// Unused properties are resolved too. Escaped placeholders become literal text
// and are not expanded again by Lookup or later substitution. Cycles and reference
// paths beyond 64 levels fail resolution; each expanded string is limited to 1 MiB.
// Nested placeholder keys expand before lookup, sharing depth and output limits;
// an escaped expansion remains literal when used as a key.
// Diagnostics omit expanded lookup addresses, including those in operand error
// paths, while preserving classification, safe owning origins and error causes.
// Immutable resolvers support concurrent reads.
package resolution
