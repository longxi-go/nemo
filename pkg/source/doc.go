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

// Package source supplies immutable property snapshots from maps, files and the
// process environment. Providers own retrieval; codecs own document syntax.
// Sources have sorted detached keys, exact lookup, and retained provenance.
// NewMap and NewProperties copy caller maps; supplied gosr.Value trees are already
// immutable. Custom PropertySource implementations used with Static must honor
// the gosr immutability contract, and custom decoders must be concurrency-safe
// when their containing provider is fetched concurrently. Custom typed-nil
// implementations must handle their own nil receivers, as no reflection is used.
//
// Providers are safe for concurrent calls when their injected dependencies are.
// File reads accept regular files, including symlink targets, within the
// configured byte limit. Special files are rejected. Fetch checks cancellation
// before and after synchronous file reads and decoding;
// an in-flight operating-system read or decoder call cannot itself be interrupted.
// Each call produces a full snapshot. Built-in providers start no background goroutines.
// Optional files ignore only not-exist. Classified load errors preserve causes
// without printing decoder payloads. Watcher is an extension contract: transports
// own client/subscription lifetime, block in Watch, call back outside locks, and
// cease callbacks before returning.
//
// Callers assemble providers highest-first. Discover supplies an explicit file
// policy: later profiles, earlier profiles, then base files; within each group,
// later directories precede earlier directories, then names and formats follow
// their listed order. Discovery.Decoders selects an application decoder registry;
// registration order does not set source precedence. It never inserts an OS source.
// Discovered source names are file: followed by the native cleaned absolute path,
// not file URIs. Origin.Resource retains that path for filesystem diagnostics.
// Discovery deduplicates lexical paths, without resolving symlinks or case aliases.
package source
