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

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/gosr/pkg/event"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

// Option configures an Environment during construction. Options must not retain
// or later mutate the Environment. Custom typed-nil dependencies must handle nil
// receivers; the Environment does not use reflection to inspect them.
type Option func(*Environment) error

// WithLimits sets fixed resource limits during construction. It cannot restrict allocations
// already performed inside an explicitly supplied provider or decoder.
func WithLimits(policy limits.Policy) Option {
	return func(e *Environment) error {
		if err := policy.Validate(); err != nil {
			return err
		}
		e.limits = policy
		return nil
	}
}

// WithProviders appends highest-first providers, ahead of discovered files.
func WithProviders(providers ...source.Provider) Option {
	providers = slices.Clone(providers)
	return func(e *Environment) error {
		for _, p := range providers {
			if p == nil {
				return invalid()
			}
		}
		e.providers = append(e.providers, providers...)
		return nil
	}
}

// WithProfiles freezes unique active names; an empty list selects defaults.
func WithProfiles(names ...string) Option {
	names = slices.Clone(names)
	return func(e *Environment) error {
		profiles, err := uniqueProfiles(names)
		if err != nil {
			return err
		}
		e.profiles = profiles
		return nil
	}
}

// WithDefaultProfiles freezes fallback names. Empty selects the standard default.
func WithDefaultProfiles(names ...string) Option {
	names = slices.Clone(names)
	return func(e *Environment) error {
		profiles, err := uniqueProfiles(names)
		if err != nil {
			return err
		}
		if len(profiles) == 0 {
			profiles = []string{gosr.DefaultProfile}
		}
		e.defaults = profiles
		return nil
	}
}

// WithValidator appends a candidate validator. It must not call Refresh on the
// same Environment because candidate preparation is serialized.
func WithValidator(validate SnapshotValidatorFunc) Option {
	return func(e *Environment) error {
		if validate == nil {
			return invalid()
		}
		e.validators = append(e.validators, validate)
		return nil
	}
}

// WithPublisher injects synchronous notifications after successful changed commits.
func WithPublisher(publisher event.Publisher) Option {
	return func(e *Environment) error {
		if publisher == nil {
			return invalid()
		}
		e.publisher = publisher
		return nil
	}
}

// WithErrorHandler handles reload errors during Watch, outside the refresh gate.
func WithErrorHandler(handler ErrorHandlerFunc) Option {
	return func(e *Environment) error {
		if handler == nil {
			return invalid()
		}
		e.errorHandler = handler
		return nil
	}
}

// WithDiscovery selects files from the Environment's configured profiles and
// appends them after explicit providers. Separate discovery profiles are rejected.
func WithDiscovery(discovery source.Discovery) Option {
	discovery.Paths = slices.Clone(discovery.Paths)
	discovery.Names = slices.Clone(discovery.Names)
	discovery.Formats = slices.Clone(discovery.Formats)
	discovery.Profiles = slices.Clone(discovery.Profiles)
	return func(e *Environment) error {
		if len(discovery.Profiles) > 0 {
			return invalid()
		}
		e.discoveries = append(e.discoveries, discovery)
		return nil
	}
}

func uniqueProfiles(names []string) ([]string, error) {
	out := []string{}
	seen := maptypes.StringBoolMap{}
	for _, name := range names {
		if err := gosr.ValidateProfile(name); err != nil {
			return nil, err
		}
		if !seen[name] {
			out = append(out, name)
			seen[name] = true
		}
	}
	return out, nil
}
