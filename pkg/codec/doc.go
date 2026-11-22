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

// Package codec translates complete configuration documents into portable gosr
// values. It owns syntax and normalization; source providers own transport.
// ForFormat selects built-ins. RegistryBuilder assembles an application-owned
// registry with atomic registration and detached read-only name mappings. Inject
// the registry into source.Discovery to discover custom formats. A registry fixes
// decoder bindings, not decoder state: shared decoders must be concurrency-safe.
// Builders are serial bootstrap objects; no global registration is performed.
// Decoders return detached maps.
//
// YAML/YML, JSON and TOML retain native scalar types. Java properties and INI
// retain strings because their syntax has no portable typed scalar contract.
// Objects flatten into dotted keys; arrays retain typed objects and lists. Empty
// objects remain atomic values. Keys remain case-sensitive. Ambiguous flattened
// addresses, nonstring YAML keys, nonfinite numbers and nesting beyond 64 levels
// are rejected. Untagged YAML integer text outside int64/uint64 is rejected before the
// parser can round it through float64. Explicit finite floats retain float
// semantics; quoted numeric text remains a string. Built-in registries limit
// document bytes, normalized values and parser depth. Dates parsed by YAML/TOML
// become RFC3339Nano strings.
//
// JSON, YAML and TOML reject duplicate declarations according to their parsers
// (JSON additionally checks members before map construction). YAML merge keys
// follow YAML parser precedence. Properties uses the last declaration for a key;
// INI merges repeated sections and uses the last declaration for a key. Distinct
// INI section/key declarations that flatten to one address are rejected.
// Properties interpolation is disabled: placeholder resolution belongs to the
// environment, and INI values are read without parser interpolation.
//
// Built-in errors retain parser causes through Unwrap, but their printable
// diagnostics omit document content. Callers must not print unwrapped causes
// when configuration may contain secrets.
package codec
