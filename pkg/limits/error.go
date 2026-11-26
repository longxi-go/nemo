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
	"fmt"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

const (
	configurationLimitFormat = "configuration limit: %s (%d > %d)"
)

// Resource identifies a budget without including configuration values.
type Resource string

const (
	DocumentBytes Resource = "document bytes"
	Sources       Resource = "sources"
	Properties    Resource = "properties"
	ValueDepth    Resource = "value depth"
	ValueNodes    Resource = "value nodes"
	ValueBytes    Resource = "value bytes"
	ResolvedBytes Resource = "resolved bytes"
)

// ExceededError retains numerical diagnostics while redacting input data.
type ExceededError struct {
	Resource Resource
	Limit    int64
	Observed int64
}

func (e *ExceededError) Error() string {
	return fmt.Sprintf(configurationLimitFormat, e.Resource, e.Observed, e.Limit)
}

func (e *ExceededError) Unwrap() error {
	return &gosr.Error{
		Code: gosr.ResolutionLimit,
	}
}

// Check compares nonnegative counts without changing their owning budget.
func Check(resource Resource, observed, limit int64) error {
	if observed > limit {
		return &ExceededError{
			Resource: resource,
			Limit:    limit,
			Observed: observed,
		}
	}
	return nil
}
