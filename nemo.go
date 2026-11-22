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

package nemo

import (
	"context"

	"github.com/longxi-go/nemo/pkg/environment"
)

// Environment is the configuration environment defined in pkg/environment.
// Both entry points use the same snapshot, refresh and watch implementation.
type Environment = environment.Environment

// Option configures an Environment using pkg/environment options.
type Option = environment.Option

// New loads and validates the initial configuration before returning the Environment.
// Source selection is explicit through pkg/environment options.
func New(ctx context.Context, options ...Option) (*Environment, error) {
	return environment.New(ctx, options...)
}
