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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/longxi-go/nemo/pkg/binding"
)

const (
	configurationPath        = "/configuration"
	configurationUnavailable = "configuration unavailable"
)

// View couples the response fields to their single source revision.
type View struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Revision uint64 `json:"revision"`
}

type failureResponse struct {
	Error string `json:"error"`
}

func (a *Application) configuration(ctx *gin.Context) {
	if ctx.Request.Context().Err() != nil {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, failureResponse{
			Error: configurationUnavailable,
		})
		return
	}
	snapshot := a.Snapshot()
	config, err := binding.Bind[Config](snapshot, "")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, failureResponse{
			Error: configurationUnavailable,
		})
		return
	}
	ctx.JSON(http.StatusOK, View{
		Name:     config.Name,
		Address:  config.Address,
		Revision: snapshot.Revision(),
	})
}
