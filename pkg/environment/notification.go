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
	"fmt"
	"slices"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

const (
	nilNotificationErrorText  = "<nil>"
	notificationFailureFormat = "environment: notification failed after committed revision %d"
)

// Changed identifies a committed transition. Old/New pin immutable generations;
// Keys is a detached sorted list, read-only during delivery by event contract.
// Concurrent notifications can overlap; use New.Revision to identify commits.
type Changed struct {
	Old  gosr.Snapshot
	New  gosr.Snapshot
	Keys []string
}

// EventMarker admits Changed to the injected gosr publisher.
func (Changed) EventMarker() {
}

// NotificationError means Revision was committed, but delivery failed. Cause
// remains reachable with errors.Is/As and is excluded from printable diagnostics.
type NotificationError struct {
	Revision uint64
	Cause    error
}

func (e *NotificationError) Error() string {
	if e == nil {
		return nilNotificationErrorText
	}
	return fmt.Sprintf(notificationFailureFormat, e.Revision)
}

// Unwrap preserves delivery causes without including secret payloads in Error.
func (e *NotificationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func changedKeys(old, next *snapshot) []string {
	keys := maptypes.StringBoolMap{}
	for _, key := range old.Keys() {
		keys[key] = true
	}
	for _, key := range next.Keys() {
		keys[key] = true
	}
	changed := []string{}
	for key := range keys {
		before, b, _ := old.Lookup(key)
		after, a, _ := next.Lookup(key)
		if b != a || before.Origin != after.Origin || !before.Value.Equal(after.Value) {
			changed = append(changed, key)
		}
	}
	slices.Sort(changed)
	return changed
}
