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

package codec

import (
	"github.com/longxi-go/nemo/internal/normalization"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// Decoding bounds parser-owned documents before normalization applies its own
// independent value-tree limit.
const maxDocumentDepth = 64

func normalize(input any) (maptypes.ValueMap, error) {
	v, err := normalization.Flatten(input)
	if err != nil {
		return nil, decodeError(err)
	}
	return v, nil
}
