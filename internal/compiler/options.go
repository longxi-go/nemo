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

// Options controls only explicitly selected writable package roots.
type Options struct {
	Dir   string
	Roots []string
	Check bool
}

// File describes deterministic output without exposing compiler implementation state.
type File struct {
	Path    string
	Package string
	Types   []string
}

// Report identifies selected outputs, including unchanged files.
type Report struct {
	Files []File
}
