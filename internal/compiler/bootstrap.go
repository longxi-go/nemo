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
	"fmt"
	"strings"
)

const (
	bootstrapImportFormat  = "import %s %q\n"
	bootstrapPackageFormat = "package %s\n"
	bootstrapMethodsFormat = `func (*%s) BindProperties(%s.PropertyResolver, string) error {
    return nil
}

func (*%s) PropertyPrefix() string {
    return ""
}
`
)

// bootstrapSource exposes only generated method signatures, permitting first
// generation without suppressing unrelated errors in the authored package.
func bootstrapSource(packageName string, source scannedSource) []byte {
	var stub strings.Builder
	fmt.Fprintf(&stub, bootstrapPackageFormat, packageName)
	if len(source.declarations) == 0 {
		return []byte(stub.String())
	}
	alias := newIdentifiers(source.names).take(environmentImportAlias)
	fmt.Fprintf(&stub, bootstrapImportFormat, alias, environmentImportPath)
	for _, declaration := range source.declarations {
		fmt.Fprintf(&stub, bootstrapMethodsFormat, declaration.name, alias, declaration.name)
	}
	return []byte(stub.String())
}
