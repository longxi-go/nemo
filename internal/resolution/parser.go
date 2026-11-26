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

package resolution

import (
	"errors"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

// Check cancellation periodically without adding a callback to every byte.
const cancellationCheckBytes = 1024

type operandResolverFunc = func(string) (gosr.Value, int, bool, error)

type cancellationFunc = func() error

type byteAdmissionFunc = func(int) error

type parser struct {
	operand operandResolverFunc
	check   cancellationFunc
	charge  byteAdmissionFunc
}

func failure(code gosr.Code) error {
	return &gosr.Error{
		Code: code,
	}
}

func codeOf(err error) gosr.Code {
	var detail *gosr.Error
	if errors.As(err, &detail) {
		return detail.Code
	}
	return gosr.InvalidValue
}

func (p parser) expand(text string) (string, int, error) {
	if err := p.checkAt(0); err != nil {
		return "", 0, err
	}
	if err := p.validateSyntax(text, 0); err != nil {
		return "", 0, err
	}
	out := expansion{
		charge: p.charge,
	}
	height, err := p.expandAt(text, 0, &out)
	if err != nil {
		return "", 0, err
	}
	return out.text.String(), height, nil
}

func (p parser) checkAt(offset int) error {
	if offset%cancellationCheckBytes == 0 && p.check != nil {
		return p.check()
	}
	return nil
}

func (p parser) expandAt(text string, depth int, out *expansion) (int, error) {
	if depth > maxDepth {
		return 0, failure(gosr.ResolutionLimit)
	}
	height := 0
	for i := 0; i < len(text); {
		if err := p.checkAt(i); err != nil {
			return 0, err
		}
		var part string
		if strings.HasPrefix(text[i:], gosr.EscapedPlaceholderPrefix) {
			part = gosr.PlaceholderPrefix
			i += len(gosr.EscapedPlaceholderPrefix)
		} else if strings.HasPrefix(text[i:], gosr.PlaceholderPrefix) {
			end, colon, err := p.bounds(text, i+len(gosr.PlaceholderPrefix))
			if err != nil {
				return 0, err
			}
			keyEnd := end
			if colon >= 0 {
				keyEnd = colon
			}
			keyText := text[i+len(gosr.PlaceholderPrefix) : keyEnd]
			keyOutput := expansion{}
			if strings.Contains(keyText, gosr.PlaceholderPrefix) {
				keyOutput.charge = p.charge
			}
			keyHeight, err := p.expandAt(keyText, depth+1, &keyOutput)
			if err != nil {
				return 0, err
			}
			reference := propertyReference{
				key:      keyOutput.text.String(),
				expanded: keyHeight > 0,
			}
			if err := gosr.ValidateKey(reference.key); err != nil {
				return 0, err
			}
			if depth == maxDepth {
				return 0, failure(gosr.ResolutionLimit)
			}
			value, h, found, err := p.operand(reference.key)
			if err != nil {
				return 0, reference.failure(codeOf(err), err)
			}
			if !found {
				if colon < 0 {
					return 0, reference.failure(gosr.MissingProperty, nil)
				}
				h, err = p.expandAt(text[colon+len(gosr.PlaceholderDefaultSeparator):end], depth+1, out)
				if err != nil {
					return 0, err
				}
			} else {
				var scalar bool
				part, scalar = value.Scalar()
				if !scalar {
					return 0, reference.failure(gosr.InvalidValue, nil)
				}
			}
			h = max(h, keyHeight) + 1
			if depth+h > maxDepth {
				return 0, failure(gosr.ResolutionLimit)
			}
			height = max(height, h)
			i = end + len(gosr.PlaceholderSuffix)
		} else {
			part = text[i : i+1]
			i++
		}
		if err := out.append(part); err != nil {
			return 0, err
		}
	}
	return height, nil
}
