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
	"fmt"
	"time"

	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/binding"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/maptypes"
	"github.com/longxi-go/nemo/pkg/source"
)

// Config declares application properties for generated binding.
// +nemo:config:prefix=app
type Config struct {
	Name     string        `required:"true"`
	Port     int           `default:"8080"`
	Timeout  time.Duration `default:"3s"`
	Database Database
}

// Database declares the nested database properties.
type Database struct {
	Host string `default:"localhost"`
}

func run() error {
	properties, err := source.NewMap("application", maptypes.ValueMap{
		"app.name": gosr.String("nemo"),
	})
	if err != nil {
		return err
	}
	owner, err := environment.New(context.Background(), environment.WithProviders(source.Static(properties)))
	if err != nil {
		return err
	}
	config, err := binding.Bind[Config](owner, "")
	if err != nil {
		return err
	}
	fmt.Println(config.Name, config.Port, config.Timeout, config.Database.Host)
	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
