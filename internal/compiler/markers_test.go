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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/longxi-go/gosr/tools/pkg/annotation"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func TestRejectUnconsumedConfigurationMarkers(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		source string
	}{
		{
			"package",
			"// +nemo:config:prefix=app\npackage fixture\ntype Config struct{}\n",
		},
		{
			"function",
			"package fixture\n// +nemo:config:prefix=app\nfunc Configure() {}\n",
		},
		{
			"field",
			"package fixture\n// +nemo:config:prefix=app\ntype Config struct {\n// +nemo:config:unknown\nValue int\n}\n",
		},
		{
			"unknown type marker",
			"package fixture\n// +nemo:config:unknown\ntype Config struct{}\n",
		},
		{
			"unknown namespace",
			"package fixture\n// +nemo:unknown\ntype Config struct{}\n",
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.go")
			if err := os.WriteFile(path, []byte(scenario.source), 0o600); err != nil {
				t.Fatal(err)
			}
			registry, err := annotation.New(annotationPrefix, []annotation.Definition{
				{
					Name:   configPrefixMarker,
					Target: markers.DescribesType,
					Sample: "",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := scanDeclarations([]string{path}, registry); err == nil || !strings.Contains(err.Error(), "unknown or misplaced marker") {
				t.Fatalf("configuration marker was silently ignored: %v", err)
			}
		})
	}
}
