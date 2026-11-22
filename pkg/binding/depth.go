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

package binding

import gosr "github.com/longxi-go/gosr/pkg/environment"

// MaxObjectDepth bounds one generated private helper tree. The root is depth 0;
// each instantiated child object adds one, independently of pointer/container
// wrappers. Opaque external binders own their internal limits; a generated public
// entry in another package begins its own budget.
const MaxObjectDepth = 64

// CheckObjectDepth bounds both recursive and acyclic generated object traversal.
// It is independent of document-parser and placeholder-resolution limits.
func CheckObjectDepth(depth int) error {
	if depth < 0 || depth > MaxObjectDepth {
		return &gosr.Error{
			Code:    gosr.ConversionFailed,
			Message: "object binding depth exceeded",
		}
	}
	return nil
}
