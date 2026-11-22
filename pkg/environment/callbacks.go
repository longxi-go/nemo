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

// SnapshotValidatorFunc validates a candidate before publication. It must not
// modify the snapshot or call Refresh on the same Environment.
type SnapshotValidatorFunc = func(gosr.Snapshot) error

// ErrorHandlerFunc receives Watch refresh errors after the refresh gate is
// released. A handler shared by Environment instances must support concurrent calls.
type ErrorHandlerFunc = func(error)
