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

// Package normalization owns the parser-to-property boundary. Parser-native
// scalar types become closed gosr values without reflection; nested objects
// flatten into dotted addresses while objects in lists stay structured. It
// rejects ambiguous addresses and unsupported/nonfinite values before sources
// can publish them. Collection conversion detaches parser-owned storage, and a
// finite depth bound protects against cyclic or excessively nested inputs.
package normalization
