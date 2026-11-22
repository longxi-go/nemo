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

package compiler

import "fmt"

// Validate semantic choices before a template can interpret an unknown mode as
// another strategy. Model diagnostics identify the originating type and field.
func (m packageModel) validate() error {
	for _, helper := range m.Helpers {
		for _, field := range helper.Fields {
			if !field.Mode.valid() {
				return fmt.Errorf("field %s.%s: unsupported binding mode %q", helper.Type, field.Name, field.Mode)
			}
		}
	}
	return nil
}
