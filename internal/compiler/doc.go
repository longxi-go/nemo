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

// Package compiler owns the private staged annotation pipeline. Invocation-owned
// typed state flows through fixed analyze, model, render, format, overlay-validation
// and output steps. AST discovery creates method stubs before typed analysis.
// Semantic binder, helper and field models feed embedded Go templates; the model
// phase supplies typed decoder expressions but never assembles statement bodies.
// Dependencies are never generation destinations. Unrelated source errors remain
// fatal, and semantic failures never reach output.
//
// The shared gosr pipeline executes the steps; this package owns their order,
// invocation state, semantic modeling, and output ownership policy. Templates
// and decoder expressions supply structural layout before the formatter runs.
// Shared codegen packages own imports, artifact snapshots, diagnostic causes,
// strict templates and filesystem commit. The separate gosr/tools module owns
// copied package-loading inputs and controller-tools comment adaptation.
// Domain schemas, models and module/discovery policy remain in this package.
package compiler
