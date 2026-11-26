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
	"io"
	"os"

	"github.com/longxi-go/nemo/pkg/limits"
)

const (
	configurationIsNotARegularFileMessage = "configuration is not a regular file"
)

// read rejects special files before opening and checks the opened descriptor too.
// Concurrent replacement between Stat and Open remains possible.
func (p fileProvider) read(ctx context.Context) (data []byte, err error) {
	info, err := os.Stat(p.path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New(configurationIsNotARegularFileMessage)
	}
	file, err := os.Open(p.path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()
	info, err = file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New(configurationIsNotARegularFileMessage)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err = io.ReadAll(io.LimitReader(file, p.limits.MaxDocumentBytes+1))
	if err != nil {
		return nil, err
	}
	if err := limits.Check(limits.DocumentBytes, int64(len(data)), p.limits.MaxDocumentBytes); err != nil {
		return nil, err
	}
	return data, ctx.Err()
}
