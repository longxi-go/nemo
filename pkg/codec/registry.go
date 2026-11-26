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
	"strings"
	"unicode"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	invalidOrUnregisteredConfigurationFormatMessage = "invalid or unregistered configuration format"
)

// Format names have their own grammar, independent of property addresses.
const (
	formatSegmentSeparator = '.'
	formatWordSeparator    = '-'
	formatWordUnderscore   = '_'
)

// Definition assigns one decoder to a canonical extension and optional aliases.
// Names use ASCII letters, digits, underscores and hyphens, with optional dot
// separated segments. Lookup ignores case; discovery retains extension spelling.
type Definition struct {
	Name    string
	Aliases []string
	Decoder Decoder
}

// DecoderLookup selects a decoder by format name. A lookup
// shared across providers must be safe for concurrent calls. Decoder state remains
// owned by its implementation and must support concurrent Decode calls when shared.
type DecoderLookup interface {
	Lookup(format string) (Decoder, error)
}

type registry struct {
	decoders maptypes.StringMap[Decoder]
}

func (r registry) Lookup(format string) (Decoder, error) {
	name, ok := formatName(format)
	if !ok {
		return nil, registryError(gosr.InvalidValue)
	}
	decoder, found := r.decoders[name]
	if !found {
		return nil, registryError(gosr.InvalidValue)
	}
	return decoder, nil
}

func formatName(format string) (string, bool) {
	for _, c := range format {
		if c > unicode.MaxASCII {
			return "", false
		}
	}
	name := strings.ToLower(format)
	segmentStart := true
	for _, c := range name {
		alphaNumeric := c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
		if segmentStart && !alphaNumeric {
			return "", false
		}
		if !alphaNumeric &&
			c != formatWordUnderscore &&
			c != formatWordSeparator &&
			c != formatSegmentSeparator {
			return "", false
		}
		segmentStart = c == formatSegmentSeparator
	}
	return name, name != "" && !segmentStart
}

func registryError(code gosr.Code) error {
	return &gosr.Error{
		Code:    code,
		Message: invalidOrUnregisteredConfigurationFormatMessage,
	}
}
