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
	"context"

	"github.com/longxi-go/nemo/internal/admission"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// boundedDecoder applies document and value limits to direct Decode calls.
type boundedDecoder struct {
	decoder Decoder
	policy  limits.Policy
}

func (d boundedDecoder) Decode(data []byte) (maptypes.ValueMap, error) {
	if err := limits.Check(limits.DocumentBytes, int64(len(data)), d.policy.MaxDocumentBytes); err != nil {
		return nil, err
	}
	values, err := d.decoder.Decode(data)
	if err != nil {
		return nil, err
	}
	if err := limits.Check(limits.Properties, int64(len(values)), int64(d.policy.MaxProperties)); err != nil {
		return nil, err
	}
	budget := admission.New(context.Background(), d.policy, false)
	for key, value := range values {
		if err := budget.Property(key, value); err != nil {
			return nil, err
		}
	}
	return values, nil
}

// NewBuiltinRegistry creates a built-in decoder registry using the supplied
// limits. Parsers also enforce an independent 64-level depth limit.
func NewBuiltinRegistry(policy limits.Policy) (DecoderLookup, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	builder := NewRegistryBuilder()
	if err := registerBuiltins(builder, policy); err != nil {
		return nil, err
	}
	return builder.Build(), nil
}
