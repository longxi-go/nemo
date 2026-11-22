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

package command

import (
	"fmt"
	"strings"
)

const rootListSeparator = ","

type rootFlags []string

func (r *rootFlags) String() string {
	return strings.Join(*r, rootListSeparator)
}

func (r *rootFlags) Set(value string) error {
	for _, root := range strings.Split(value, rootListSeparator) {
		if strings.TrimSpace(root) == "" {
			return fmt.Errorf("empty root")
		}
		*r = append(*r, strings.TrimSpace(root))
	}
	return nil
}
