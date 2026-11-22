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

package binding

import gosr "github.com/longxi-go/gosr/pkg/environment"

// BindablePointer constrains *T to implement gosr.Bindable. Bind allocates T
// directly, and callers may omit the pointer type because it is inferred from T.
type BindablePointer[T any] interface {
	*T
	gosr.Bindable
}
