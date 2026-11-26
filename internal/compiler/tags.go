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

package compiler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	duplicateTagFormat           = "duplicate tag %s"
	invalidQuotedTagMessage      = "invalid quoted tag"
	malformedStructTagMessage    = "malformed struct tag"
	malformedTagNameMessage      = "malformed tag name"
	tagValueMustBeQuotedMessage  = "tag value must be quoted"
	tagsRequireSeparationMessage = "tags require separation"
	unterminatedTagValueMessage  = "unterminated tag value"
)

// validateTags rejects malformed or duplicate keys rather than silently treating
// an unreadable framework tag as absent. Unrelated well-formed tags are retained.
func validateTags(raw string) error {
	seen := make(maptypes.StringBoolMap)
	for raw != "" {
		raw = strings.TrimLeft(raw, tagSeparator)
		if raw == "" {
			break
		}
		colon := strings.Index(raw, tagValueSeparator)
		if colon < 1 {
			return errors.New(malformedStructTagMessage)
		}
		name := raw[:colon]
		if strings.ContainsAny(name, invalidTagNameCharacters) {
			return errors.New(malformedTagNameMessage)
		}
		if seen[name] {
			return fmt.Errorf(duplicateTagFormat, name)
		}
		seen[name] = true
		rest := raw[colon+len(tagValueSeparator):]
		if len(rest) == 0 || rest[0] != tagQuote {
			return errors.New(tagValueMustBeQuotedMessage)
		}
		end := 1
		for end < len(rest) {
			if rest[end] == tagEscape {
				end += 2
				continue
			}
			if rest[end] == tagQuote {
				break
			}
			end++
		}
		if end >= len(rest) {
			return errors.New(unterminatedTagValueMessage)
		}
		if _, err := strconv.Unquote(rest[:end+1]); err != nil {
			return errors.New(invalidQuotedTagMessage)
		}
		raw = rest[end+1:]
		if raw != "" && raw[0] != tagSeparator[0] {
			return errors.New(tagsRequireSeparationMessage)
		}
	}
	return nil
}
