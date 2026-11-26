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
	"context"

	gosr "github.com/longxi-go/gosr/pkg/environment"
)

const (
	cannotLoadConfigurationSourceMessage = "cannot load configuration source"
)

// InvalidationFunc requests a refresh without carrying transport-owned data.
// A watcher must finish all callback invocations before Watch returns.
type InvalidationFunc = func()

// Provider owns retrieval and returns an immutable source. Implementations must
// respect cancellation, own transport/client lifetimes, and keep I/O out of Lookup.
type (
	Provider interface {
		Fetch(context.Context) (gosr.PropertySource, error)
	}

	// Watcher owns a blocking subscription. Callbacks run outside transport locks,
	// signal invalidation only, and stop before Watch returns. The caller owns joining it.
	// After establishing the subscription, signal an initial invalidation or perform
	// equivalent resynchronization so an update between an earlier Fetch and Watch
	// startup is not lost. The adapter owns that transport-specific guarantee.
	Watcher interface {
		Provider
		Watch(context.Context, InvalidationFunc) error
	}
)

func failure(name, path string, cause error) error {
	return &gosr.Error{
		Code: gosr.SourceFailed,
		Origin: gosr.Origin{
			Source:   name,
			Resource: path,
		},
		Message: cannotLoadConfigurationSourceMessage,
		Cause:   cause,
	}
}

func validName(name string) error {
	return gosr.ValidateKey(name)
}
