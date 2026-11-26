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
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

// validateSyntax checks unused defaults without evaluating their references.
// A default handles missing properties, but does not suppress syntax errors.
func (p parser) validateSyntax(text string, depth int) error {
	for i := 0; i < len(text); i++ {
		if err := p.checkAt(i); err != nil {
			return err
		}
		if strings.HasPrefix(text[i:], gosr.EscapedPlaceholderPrefix) {
			i += len(gosr.EscapedPlaceholderPrefix) - 1
			continue
		}
		if !strings.HasPrefix(text[i:], gosr.PlaceholderPrefix) {
			continue
		}
		if depth >= maxDepth {
			return failure(gosr.ResolutionLimit)
		}
		end, colon, err := p.bounds(text, i+len(gosr.PlaceholderPrefix))
		if err != nil {
			return err
		}
		keyEnd := end
		if colon >= 0 {
			keyEnd = colon
		}
		if err := gosr.ValidateKey(text[i+len(gosr.PlaceholderPrefix) : keyEnd]); err != nil {
			return err
		}
		if err := p.validateSyntax(text[i+len(gosr.PlaceholderPrefix):keyEnd], depth+1); err != nil {
			return err
		}
		if colon >= 0 {
			if err := p.validateSyntax(text[colon+len(gosr.PlaceholderDefaultSeparator):end], depth+1); err != nil {
				return err
			}
		}
		i = end
	}
	return nil
}

// bounds scans balanced placeholders before evaluation, including unused defaults.
func (p parser) bounds(text string, start int) (int, int, error) {
	level, colon := 0, -1
	for i := start; i < len(text); i++ {
		if err := p.checkAt(i - start); err != nil {
			return 0, 0, err
		}
		if strings.HasPrefix(text[i:], gosr.EscapedPlaceholderPrefix) {
			i += len(gosr.EscapedPlaceholderPrefix) - 1
			level++
			continue
		}
		if strings.HasPrefix(text[i:], gosr.PlaceholderPrefix) {
			level++
			if level > maxDepth {
				return 0, 0, failure(gosr.ResolutionLimit)
			}
			i += len(gosr.PlaceholderPrefix) - 1
			continue
		}
		if strings.HasPrefix(text[i:], gosr.PlaceholderSuffix) {
			if level == 0 {
				return i, colon, nil
			}
			level--
		}
		if strings.HasPrefix(text[i:], gosr.PlaceholderDefaultSeparator) && level == 0 && colon < 0 {
			colon = i
		}
	}
	return 0, 0, failure(gosr.InvalidSyntax)
}
