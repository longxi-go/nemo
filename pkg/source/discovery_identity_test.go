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

package source_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/longxi-go/nemo/pkg/source"
)

func TestDiscoveryNativeFileIdentity(t *testing.T) {
	t.Chdir(t.TempDir())
	directory := "config #100% 中文"
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(directory, "application.json")
	if err := os.WriteFile(filename, []byte(`{"port":8080}`), 0o600); err != nil {
		t.Fatal(err)
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		t.Fatal(err)
	}
	// ToSlash exercises alternate Windows separators on a Windows test runner.
	// On Unix the same policy preserves native path spelling without URI escaping.
	providers, err := source.Discover(source.Discovery{
		Paths: []string{
			directory,
			absolute,
			directory + string(filepath.Separator) + ".",
			filepath.ToSlash(absolute),
		},
		Formats: []string{"json"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 {
		t.Fatalf("equivalent native paths produced %d providers", len(providers))
	}
	snapshot, err := providers[0].Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	resource := filepath.Join(absolute, "application.json")
	expectedName := "file:" + resource
	property, found, err := snapshot.Lookup("port")
	if err != nil || !found {
		t.Fatalf("lookup: found=%v, error=%v", found, err)
	}
	if snapshot.Name() != expectedName ||
		property.Origin.Source != expectedName ||
		property.Origin.Resource != resource {
		t.Fatalf("native identity or resource rewritten: name=%q, origin=%#v", snapshot.Name(), property.Origin)
	}
	if text, scalar := property.Value.Scalar(); !scalar || text != "8080" {
		t.Fatalf("native resource not decoded: %#v", property.Value)
	}
}
