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

package resolution

import gosr "github.com/longxi-go/gosr/pkg/environment"

const expandedReferenceFailureMessage = "expanded property reference failed"

// propertyReference separates a lookup address from printable identity. A key
// produced by substitution may contain secrets even when it names no property.
type propertyReference struct {
	key      string
	expanded bool
}

func (r propertyReference) failure(code gosr.Code, cause error) error {
	if !r.expanded && cause != nil {
		return cause
	}
	detail := &gosr.Error{
		Code:  code,
		Cause: cause,
	}
	if r.expanded {
		// Operand failures can already contain the expanded key in their paths.
		// Preserve their causes without promoting that metadata into outer errors.
		detail.Message = expandedReferenceFailureMessage
	} else {
		detail.Key = r.key
	}
	return detail
}
