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

import (
	"context"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/admission"
	"github.com/longxi-go/nemo/pkg/limits"
)

// admitComposite applies default limits before projecting values supplied by
// external resolvers. It cannot limit allocations already made by those resolvers.
func admitComposite(key string, property gosr.Property) error {
	if err := admission.New(context.Background(), limits.Default(), false).Property(key, property.Value); err != nil {
		return &gosr.Error{
			Code:   gosr.ConversionFailed,
			Key:    key,
			Origin: property.Origin,
			Cause:  err,
		}
	}
	return nil
}
