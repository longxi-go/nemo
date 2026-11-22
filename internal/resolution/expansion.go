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

import (
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

// expansion shares one builder across nested defaults. Admission precedes each
// append, so intermediate defaults cannot bypass the candidate output budget.
type expansion struct {
	text   strings.Builder
	charge byteAdmissionFunc
}

func (e *expansion) append(part string) error {
	if len(part) > maxOutput-e.text.Len() {
		return failure(gosr.ResolutionLimit)
	}
	if e.charge != nil {
		if err := e.charge(len(part)); err != nil {
			return err
		}
	}
	e.text.WriteString(part)
	return nil
}
