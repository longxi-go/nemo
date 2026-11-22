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

import gosr "github.com/longxi-go/gosr/pkg/environment"

// Snapshot returns the current immutable snapshot. A nil or uninitialized
// Environment returns revision zero; Lookup and Resolve return InvalidState.
func (e *Environment) Snapshot() gosr.Snapshot {
	if e != nil {
		if s := e.current.Load(); s != nil {
			return s
		}
	}
	return absentSnapshot{}
}

// Lookup reads one current generation; capture Snapshot for related reads.
func (e *Environment) Lookup(key string) (gosr.Property, bool, error) {
	return e.Snapshot().Lookup(key)
}

// Resolve expands against one current generation without source I/O.
func (e *Environment) Resolve(text string) (string, error) {
	return e.Snapshot().Resolve(text)
}

// Keys detaches the current generation's sorted addresses.
func (e *Environment) Keys() []string {
	return e.Snapshot().Keys()
}

// ActiveProfiles returns detached bootstrap names.
func (e *Environment) ActiveProfiles() []string {
	return e.Snapshot().ActiveProfiles()
}

// DefaultProfiles returns detached bootstrap fallback names.
func (e *Environment) DefaultProfiles() []string {
	return e.Snapshot().DefaultProfiles()
}

// AcceptsProfiles validates every predicate and ORs names or !name tests.
func (e *Environment) AcceptsProfiles(predicates ...string) (bool, error) {
	return e.Snapshot().AcceptsProfiles(predicates...)
}

var (
	_ gosr.Environment      = (*Environment)(nil)
	_ gosr.SnapshotProvider = (*Environment)(nil)
)
