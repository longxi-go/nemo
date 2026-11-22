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

import "github.com/longxi-go/nemo/pkg/limits"

// RegisterBuiltins adds every built-in format atomically. Applications may start
// with an empty builder instead when only selected or custom formats are allowed.
func RegisterBuiltins(builder *RegistryBuilder) error {
	return registerBuiltins(builder, limits.Default())
}

func registerBuiltins(builder *RegistryBuilder, policy limits.Policy) error {
	return builder.register([]Definition{
		{
			Name:    FormatYAML,
			Aliases: []string{FormatYML},
			Decoder: boundedDecoder{
				decoder: DecoderFunc(decodeYAML),
				policy:  policy,
			},
		},
		{
			Name: FormatJSON,
			Decoder: boundedDecoder{
				decoder: DecoderFunc(decodeJSON),
				policy:  policy,
			},
		},
		{
			Name: FormatTOML,
			Decoder: boundedDecoder{
				decoder: DecoderFunc(decodeTOML),
				policy:  policy,
			},
		},
		{
			Name: FormatProperties,
			Decoder: boundedDecoder{
				decoder: DecoderFunc(decodeProperties),
				policy:  policy,
			},
		},
		{
			Name: FormatINI,
			Decoder: boundedDecoder{
				decoder: DecoderFunc(decodeINI),
				policy:  policy,
			},
		},
	})
}

// DefaultRegistry returns a read-only registry of built-ins. Each call owns its
// mapping; no process-wide mutable registration state participates in selection.
func DefaultRegistry() DecoderLookup {
	builder := NewRegistryBuilder()
	// This closed, static definition set is validated by registry tests.
	if err := RegisterBuiltins(builder); err != nil {
		panic(err)
	}
	return builder.Build()
}
