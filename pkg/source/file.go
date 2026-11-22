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

package source

import (
	"context"
	"errors"
	"os"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/limits"
	"github.com/longxi-go/nemo/pkg/maptypes"
)

// fileSourceNamePrefix namespaces discovered source identities. It is not a URI
// scheme: the suffix retains the native absolute path used by the file provider.
const fileSourceNamePrefix = "file:"

type fileProvider struct {
	name     string
	path     string
	decoder  codec.Decoder
	optional bool
	limits   limits.Policy
}

// FileOptions configures optional files and resource limits for file reads.
type FileOptions struct {
	Optional bool
	Limits   limits.Policy
}

// NewFile creates a file provider using the default limits. An optional file
// may be absent; permission errors, directories and malformed documents fail.
// Each Fetch closes its file before returning; the decoder holds no file handle.
func NewFile(name, path string, decoder codec.Decoder, optional bool) (Provider, error) {
	return NewFileWithOptions(name, path, decoder, FileOptions{
		Optional: optional,
		Limits:   limits.Default(),
	})
}

// NewFileWithOptions bounds document reads. Only regular files (including
// symlink targets) are accepted. Context cannot interrupt arbitrary OS opens.
func NewFileWithOptions(name, path string, decoder codec.Decoder, options FileOptions) (Provider, error) {
	if err := options.Limits.Validate(); err != nil {
		return nil, err
	}
	if err := validName(name); err != nil {
		return nil, err
	}
	if path == "" || strings.TrimSpace(path) != path || strings.ContainsRune(path, 0) {
		return nil, failure(name, path, nil)
	}
	if decoder == nil {
		return nil, failure(name, path, nil)
	}
	if f, ok := decoder.(codec.DecoderFunc); ok && f == nil {
		return nil, failure(name, path, nil)
	}
	return fileProvider{
		name:     name,
		path:     path,
		decoder:  decoder,
		optional: options.Optional,
		limits:   options.Limits,
	}, nil
}

func (p fileProvider) Fetch(ctx context.Context) (gosr.PropertySource, error) {
	if err := ctx.Err(); err != nil {
		return nil, failure(p.name, p.path, err)
	}
	data, err := p.read(ctx)
	if err != nil {
		if p.optional && errors.Is(err, os.ErrNotExist) {
			return NewMap(p.name, nil)
		}
		return nil, failure(p.name, p.path, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, failure(p.name, p.path, err)
	}
	values, err := p.decoder.Decode(data)
	if err != nil {
		return nil, failure(p.name, p.path, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, failure(p.name, p.path, err)
	}
	properties := make(maptypes.PropertyMap, len(values))
	for key, value := range values {
		properties[key] = gosr.Property{
			Value: value,
			Origin: gosr.Origin{
				Source:   p.name,
				Resource: p.path,
				Key:      key,
			},
		}
	}
	s, err := NewProperties(p.name, properties)
	if err != nil {
		return nil, failure(p.name, p.path, err)
	}
	return s, nil
}
