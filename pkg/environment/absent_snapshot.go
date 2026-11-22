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

type absentSnapshot struct{}

func (absentSnapshot) Lookup(string) (gosr.Property, bool, error) {
	return gosr.Property{}, false, invalid()
}

func (absentSnapshot) Resolve(string) (string, error) {
	return "", invalid()
}

func (absentSnapshot) Keys() []string {
	return nil
}

func (absentSnapshot) Revision() uint64 {
	return 0
}

func (absentSnapshot) ActiveProfiles() []string {
	return nil
}

func (absentSnapshot) DefaultProfiles() []string {
	return nil
}

func (absentSnapshot) AcceptsProfiles(...string) (bool, error) {
	return false, invalid()
}
