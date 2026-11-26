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

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/longxi-go/nemo/pkg/codec"
	"github.com/longxi-go/nemo/pkg/source"
)

const (
	configurationSource = "application"
	demonstrationFile   = "testdata/application.yaml"
)

func run(ctx context.Context, output io.Writer) error {
	decoder, err := codec.DefaultRegistry().Lookup(codec.FormatYAML)
	if err != nil {
		return err
	}
	provider, err := source.NewFile(configurationSource, demonstrationFile, decoder, false)
	if err != nil {
		return err
	}
	app, err := New(ctx, provider)
	if err != nil {
		return err
	}
	response := httptest.NewRecorder()
	app.Router().ServeHTTP(response, httptest.NewRequestWithContext(ctx, http.MethodGet, configurationPath, nil))
	if response.Code != http.StatusOK {
		return errors.New(configurationUnavailable)
	}
	_, err = io.Copy(output, response.Body)
	return err
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
