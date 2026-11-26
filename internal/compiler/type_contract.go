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

import (
	"go/types"
)

const (
	durationTypeName    = "Duration"
	errorTypeName       = "error"
	timePackagePath     = "time"
	unmarshalTextMethod = "UnmarshalText"
)

func hasText(t *types.Named) bool {
	obj, _, _ := types.LookupFieldOrMethod(types.NewPointer(t), true, t.Obj().Pkg(), unmarshalTextMethod)
	if _, ok := obj.(*types.Func); !ok {
		return false
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok ||
		sig.Variadic() ||
		sig.Params().Len() != 1 ||
		sig.Results().Len() != 1 {
		return false
	}
	return types.Identical(sig.Params().At(0).Type(), types.NewSlice(types.Typ[types.Byte])) &&
		types.Identical(sig.Results().At(0).Type(), types.Universe.Lookup(errorTypeName).Type())
}

func externalBindable(t *types.Named, pkg *types.Package, contract *types.Interface) bool {
	if t.Obj().Pkg() == pkg {
		return false
	}
	return contract != nil && types.Implements(types.NewPointer(t), contract)
}

// Both default validation and decoder selection must recognize aliases of the
// same standard-library type rather than its underlying integer representation.
func isDurationType(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	return ok && named.Obj().Pkg() != nil &&
		named.Obj().Pkg().Path() == timePackagePath && named.Obj().Name() == durationTypeName
}
