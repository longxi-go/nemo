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
	"errors"
	"os"
	"path/filepath"
	"testing"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

func TestFileDocumentLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{"a":1}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	decoder, _ := codec.ForFormat(codec.FormatJSON)
	p := limits.Default()
	p.MaxDocumentBytes = int64(len(data))
	provider, err := source.NewFileWithOptions("file", path, decoder, source.FileOptions{
		Limits: p,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.MaxDocumentBytes--
	provider, err = source.NewFileWithOptions("file", path, decoder, source.FileOptions{
		Limits: p,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Fetch(context.Background()); !errors.Is(err, &gosr.Error{
		Code: gosr.ResolutionLimit,
	}) {
		t.Fatal(err)
	}
}

func TestFileLimitBeforeCustomDecoder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom")
	if err := os.WriteFile(path, []byte("too large"), 0o600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	decode := codec.DecoderFunc(func([]byte) (maptypes.ValueMap, error) {
		calls++
		return nil, nil
	})
	policy := limits.Default()
	policy.MaxDocumentBytes = 2
	provider, err := source.NewFileWithOptions("custom", path, decode, source.FileOptions{
		Limits: policy,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Fetch(context.Background()); err == nil || calls != 0 {
		t.Fatal("oversized input reached decoder")
	}
	provider, err = source.NewFileWithOptions("optional", path+".absent", decode, source.FileOptions{
		Limits:   policy,
		Optional: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Fetch(context.Background()); err != nil || calls != 0 {
		t.Fatal("optional absence decoded", err)
	}
	provider, err = source.NewFileWithOptions("directory", filepath.Dir(path), decode, source.FileOptions{
		Limits: policy,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Fetch(context.Background()); err == nil {
		t.Fatal("directory accepted")
	}
}
