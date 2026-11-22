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
	"maps"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// RegistryBuilder registers application decoders during initialization. Build copies the name
// mapping; subsequent registrations cannot affect existing environments. The zero
// value is usable. Registration and Build must not run concurrently on a builder.
type RegistryBuilder struct {
	decoders maptypes.StringMap[Decoder]
}

// NewRegistryBuilder starts an empty application decoder registry.
func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{}
}

// Register validates the entire definition before adding any names. A duplicate
// name is an error even when both definitions use the same decoder.
func (b *RegistryBuilder) Register(definition Definition) error {
	return b.register([]Definition{definition})
}

func (b *RegistryBuilder) register(definitions []Definition) error {
	if b == nil {
		return registryError(gosr.InvalidState)
	}
	additions := make(maptypes.StringMap[Decoder])
	for _, definition := range definitions {
		if definition.Decoder == nil {
			return registryError(gosr.InvalidValue)
		}
		if f, ok := definition.Decoder.(DecoderFunc); ok && f == nil {
			return registryError(gosr.InvalidValue)
		}
		names := append([]string{definition.Name}, definition.Aliases...)
		for _, candidate := range names {
			name, valid := formatName(candidate)
			if !valid {
				return registryError(gosr.InvalidValue)
			}
			if _, exists := b.decoders[name]; exists {
				return registryError(gosr.InvalidValue)
			}
			if _, exists := additions[name]; exists {
				return registryError(gosr.InvalidValue)
			}
			additions[name] = definition.Decoder
		}
	}
	if b.decoders == nil {
		b.decoders = make(maptypes.StringMap[Decoder])
	}
	maps.Copy(b.decoders, additions)
	return nil
}

// Build returns an immutable name mapping. It does not clone decoder objects;
// custom typed-nil decoders must handle their own nil receivers.
// A nil builder produces an empty registry.
func (b *RegistryBuilder) Build() DecoderLookup {
	if b == nil {
		return registry{}
	}
	return registry{
		decoders: maps.Clone(b.decoders),
	}
}
