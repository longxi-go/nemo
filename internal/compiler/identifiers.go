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
	"strconv"

	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	preferredBinding  = "binding"
	preferredCurrent  = "current"
	preferredDepth    = "depth"
	preferredErr      = "err"
	preferredIndex    = "index"
	preferredInputs   = "inputs"
	preferredItem     = "item"
	preferredKeys     = "keys"
	preferredNemoSort = "nemoSort"
	preferredNext     = "next"
	preferredPrefix   = "prefix"
	preferredPresent  = "present"
	preferredResolver = "resolver"
	preferredScoped   = "scoped"
	preferredTarget   = "target"
	preferredValue    = "value"
)

// identifiers reserves the package namespace before allocating imports, helpers
// and function locals. The same policy protects bootstrap and final overlays.
type identifiers struct {
	used maptypes.StringBoolMap
}

func newIdentifiers(names []string) *identifiers {
	set := &identifiers{
		used: make(maptypes.StringBoolMap, len(names)),
	}
	for _, name := range names {
		set.used[name] = true
	}
	return set
}

func (s *identifiers) take(preferred string) string {
	name := preferred
	for index := 1; s.used[name]; index++ {
		name = preferred + strconv.Itoa(index)
	}
	s.used[name] = true
	return name
}

type generatedNames struct {
	Binding     string
	Environment string
	Target      string
	Resolver    string
	Prefix      string
	Next        string
	Err         string
	Value       string
	Scoped      string
	Current     string
	Inputs      string
	Index       string
	Item        string
	Keys        string
	Sort        string
	Depth       string
	Present     string
}

func newGeneratedNames(set *identifiers) generatedNames {
	return generatedNames{
		Binding:     set.take(preferredBinding),
		Environment: set.take(environmentImportAlias),
		Target:      set.take(preferredTarget),
		Resolver:    set.take(preferredResolver),
		Prefix:      set.take(preferredPrefix),
		Next:        set.take(preferredNext),
		Err:         set.take(preferredErr),
		Value:       set.take(preferredValue),
		Scoped:      set.take(preferredScoped),
		Current:     set.take(preferredCurrent),
		Inputs:      set.take(preferredInputs),
		Index:       set.take(preferredIndex),
		Item:        set.take(preferredItem),
		Keys:        set.take(preferredKeys),
		Sort:        set.take(preferredNemoSort),
		Depth:       set.take(preferredDepth),
		Present:     set.take(preferredPresent),
	}
}
