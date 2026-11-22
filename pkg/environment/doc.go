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

// Package environment implements gosr configuration environments with explicit
// provider order, fixed profiles and immutable snapshots. New loads revision 1.
// Refresh resolves every selected property and validates the candidate before
// publication. Equal values and origins do not advance the revision. Property
// lookup performs no external I/O.
//
// Providers are ordered from highest to lowest precedence; null and empty values
// override lower sources. WithDiscovery appends files selected using the active
// profiles, or default profiles when none are active. OS sources require explicit
// configuration. Custom Option functions must not retain or later mutate the
// Environment. Injected dependencies must support their documented concurrency
// and nil-receiver behavior.
//
// WithLimits sets source, property and value limits during construction. A queued
// Refresh can return on context cancellation without waiting for another refresh.
//
// Watch starts one loop per source.Watcher and one refresh worker. Repeated
// invalidations coalesce into at most one pending refresh request. An invalidation
// callback returns after signaling without waiting for publication. Callers observe
// revisions or Changed events to detect a commit. Failed preparation preserves
// the current snapshot and invokes the configured error handler. A terminal watcher
// exit cancels the group; Watch
// waits for the provider loops, callbacks and refresh worker before returning.
//
// Notifications run synchronously after commit, outside the refresh gate. Publishers
// and error handlers may read or refresh the Environment. Validators run inside
// serialized preparation and must not call Refresh on the same instance.
// Notifications may overlap, so revisions identify commits rather than delivery
// order. NotificationError reports a committed revision whose delivery failed and
// omits the underlying cause's message from its printable error.
package environment
