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
	"errors"
	"net"
	"strconv"
	"strings"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
)

const maximumPort = 65535

var errInvalidConfiguration = errors.New("service name and host:port address are required")

func validateSnapshot(snapshot gosr.Snapshot) error {
	config, err := binding.Bind[Config](snapshot, "")
	if err != nil {
		return err
	}
	if strings.TrimSpace(config.Name) == "" {
		return errInvalidConfiguration
	}
	host, port, err := net.SplitHostPort(config.Address)
	if err != nil || strings.TrimSpace(host) == "" {
		return errInvalidConfiguration
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > maximumPort {
		return errInvalidConfiguration
	}
	return nil
}
