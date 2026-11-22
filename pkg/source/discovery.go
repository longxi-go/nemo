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

package source

import (
	"path/filepath"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	defaultConfigName      = "application"
	defaultConfigDirectory = "."
	profileFileSeparator   = "-"
	fileExtensionSeparator = "."
)

// Discovery defines a deterministic highest-first search. Later profiles precede
// earlier profiles, all profiles precede base files, and later directories precede
// earlier directories. Names and Formats break ties in their listed order.
// Empty lists default to current directory, application, and yml, respectively.
// All candidates are optional files; equivalent cleaned absolute paths are deduplicated.
type Discovery struct {
	Paths, Names, Formats, Profiles []string
	// Decoders is an application decoder registry. Nil selects the built-in formats.
	Decoders codec.DecoderLookup
	// Limits selects finite file budgets. An omitted policy uses defaults.
	Limits limits.Policy
}

// Discover constructs providers without reading configuration files. Names must be profile-grammar filename
// stems; formats must be registered extension names. Extension spelling is used
// unchanged for filesystem paths; registry lookup does not broaden file search.
// Source names namespace native absolute paths; they are opaque identities, not
// file URIs. Discovery does not resolve symlinks or filesystem case aliases.
func Discover(config Discovery) ([]Provider, error) {
	policy := config.Limits
	if policy == (limits.Policy{}) {
		policy = limits.Default()
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	paths := config.Paths
	if len(paths) == 0 {
		paths = []string{defaultConfigDirectory}
	}
	names := config.Names
	if len(names) == 0 {
		names = []string{defaultConfigName}
	}
	formats := config.Formats
	if len(formats) == 0 {
		formats = []string{codec.FormatYML}
	}
	for _, path := range paths {
		if path == "" || strings.TrimSpace(path) != path || strings.ContainsRune(path, 0) {
			return nil, failure("", "", nil)
		}
	}
	for _, name := range names {
		if err := gosr.ValidateProfile(name); err != nil {
			return nil, err
		}
	}
	for _, profile := range config.Profiles {
		if err := gosr.ValidateProfile(profile); err != nil {
			return nil, err
		}
	}
	lookup := config.Decoders
	if lookup == nil {
		var err error
		lookup, err = codec.NewBuiltinRegistry(policy)
		if err != nil {
			return nil, err
		}
	}
	decoders := make([]codec.Decoder, len(formats))
	for i, format := range formats {
		// Validate filename safety even for application-supplied lookup implementations.
		if err := gosr.ValidateProfile(format); err != nil {
			return nil, err
		}
		decoder, err := lookup.Lookup(format)
		if err != nil {
			return nil, err
		}
		decoders[i] = decoder
	}
	profiles := make([]string, 0, len(config.Profiles)+1)
	for i := len(config.Profiles) - 1; i >= 0; i-- {
		profiles = append(profiles, config.Profiles[i])
	}
	profiles = append(profiles, "")
	providers := []Provider{}
	seen := maptypes.StringBoolMap{}
	for _, profile := range profiles {
		for i := len(paths) - 1; i >= 0; i-- {
			for _, name := range names {
				stem := name
				if profile != "" {
					stem += profileFileSeparator + profile
				}
				for j, format := range formats {
					path, err := filepath.Abs(filepath.Join(paths[i], stem+fileExtensionSeparator+format))
					if err != nil {
						return nil, failure("", "", err)
					}
					if seen[path] {
						continue
					}
					seen[path] = true
					if err := limits.Check(limits.Sources, int64(len(providers))+1, int64(policy.MaxSources)); err != nil {
						return nil, err
					}
					p, err := NewFileWithOptions(fileSourceNamePrefix+path, path, decoders[j], FileOptions{
						Optional: true,
						Limits:   policy,
					})
					if err != nil {
						return nil, err
					}
					providers = append(providers, p)
				}
			}
		}
	}
	return providers, nil
}
