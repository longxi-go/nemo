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

package environment

import (
	"slices"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/internal/resolution"
)

const profileNegationPrefix = "!"

type snapshot struct {
	resolver *resolution.Resolver
	profiles []string
	defaults []string
	revision uint64
}

func (s *snapshot) Lookup(key string) (gosr.Property, bool, error) {
	return s.resolver.Lookup(key)
}

func (s *snapshot) Resolve(text string) (string, error) {
	return s.resolver.Resolve(text)
}

func (s *snapshot) Keys() []string {
	return s.resolver.Keys()
}

func (s *snapshot) Revision() uint64 {
	return s.revision
}

func (s *snapshot) ActiveProfiles() []string {
	return slices.Clone(s.profiles)
}

func (s *snapshot) DefaultProfiles() []string {
	return slices.Clone(s.defaults)
}

func (s *snapshot) AcceptsProfiles(predicates ...string) (bool, error) {
	effective := s.profiles
	if len(effective) == 0 {
		effective = s.defaults
	}
	matches := false
	for _, predicate := range predicates {
		name, negated := strings.CutPrefix(predicate, profileNegationPrefix)
		if err := gosr.ValidateProfile(name); err != nil {
			return false, err
		}
		present := slices.Contains(effective, name)
		matches = matches || (present != negated)
	}
	return matches, nil
}
