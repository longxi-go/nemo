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
	"net/http"

	"github.com/gin-gonic/gin"
	gosr "github.com/longxi-go/gosr/pkg/environment"
	"github.com/longxi-go/nemo/pkg/environment"
	"github.com/longxi-go/nemo/pkg/source"
)

// Application owns configuration; readers share snapshots rather than mutable structs.
type Application struct {
	environment *environment.Environment
	router      *gin.Engine
}

// New validates typed configuration before exposing the router. Providers retain
// ownership of external clients; this example starts no background subscriptions.
func New(ctx context.Context, provider source.Provider) (*Application, error) {
	owner, err := environment.New(ctx,
		environment.WithProviders(provider),
		environment.WithValidator(validateSnapshot),
	)
	if err != nil {
		return nil, err
	}
	app := &Application{
		environment: owner,
		router:      gin.New(),
	}
	app.router.GET(configurationPath, app.configuration)
	return app, nil
}

// Router exposes ordinary net/http composition without a framework adapter.
func (a *Application) Router() http.Handler {
	return a.router
}

// Snapshot pins a generation that remains readable after later refreshes.
func (a *Application) Snapshot() gosr.Snapshot {
	return a.environment.Snapshot()
}

// Refresh rejects invalid candidates before publication. Existing Config values
// stay unchanged; subsequent request handlers explicitly bind the new snapshot.
func (a *Application) Refresh(ctx context.Context) error {
	return a.environment.Refresh(ctx)
}
