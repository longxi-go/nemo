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

package limits

import (
	"math"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

const (
	invalidResourceLimitsMessage       = "invalid resource limits"
	maximumValueDepth                  = 1024
	DefaultDocumentBytes         int64 = 8 << 20
	DefaultSources                     = 256
	DefaultProperties                  = 10000
	DefaultValueDepth                  = 64
	DefaultValueNodes                  = 100000
	DefaultValueBytes            int64 = 16 << 20
	DefaultResolvedBytes         int64 = 16 << 20
)

// Policy sets document, source and candidate resource limits. Fields must be
// positive. Constructors with an omitted policy use Default.
type Policy struct {
	MaxDocumentBytes int64
	MaxSources       int
	MaxProperties    int
	MaxValueDepth    int
	MaxValueNodes    int
	MaxValueBytes    int64
	MaxResolvedBytes int64
}

// Default returns the default limits by value.
func Default() Policy {
	return Policy{
		MaxDocumentBytes: DefaultDocumentBytes,
		MaxSources:       DefaultSources,
		MaxProperties:    DefaultProperties,
		MaxValueDepth:    DefaultValueDepth,
		MaxValueNodes:    DefaultValueNodes,
		MaxValueBytes:    DefaultValueBytes,
		MaxResolvedBytes: DefaultResolvedBytes,
	}
}

// Validate rejects unlimited or overflowing budgets. Depth is capped at 1024
// because downstream parsers and binders may use bounded recursive traversal.
func (p Policy) Validate() error {
	if p.MaxDocumentBytes <= 0 || p.MaxDocumentBytes == math.MaxInt64 ||
		p.MaxSources <= 0 ||
		p.MaxProperties <= 0 ||
		p.MaxValueDepth <= 0 || p.MaxValueDepth > maximumValueDepth ||
		p.MaxValueNodes <= 0 ||
		p.MaxValueBytes <= 0 ||
		p.MaxResolvedBytes <= 0 {
		return &gosr.Error{
			Code:    gosr.InvalidValue,
			Message: invalidResourceLimitsMessage,
		}
	}
	return nil
}
