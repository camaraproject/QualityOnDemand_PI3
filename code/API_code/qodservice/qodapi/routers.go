// Copyright 2023 Spry Fox Networks
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qodapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CONTENT_TYPE_PROBLEM = "application/problem+json"
	CONTENT_TYPE_DATA    = "application/json"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// Routes is the list of the generated Route.
type Routes []Route

// AddService adds the routes
func AddService(engine *gin.Engine) *gin.RouterGroup {
	group := engine.Group("/qod/v0")

	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			group.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			group.POST(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			group.DELETE(route.Pattern, route.HandlerFunc)
		}
	}
	return group
}

// Index is the index handler.
func Index(c *gin.Context) {
	c.String(http.StatusOK, "Hello World!")
}

var routes = Routes{
	{
		"Index",
		http.MethodGet,
		"/",
		Index,
	},

	{
		"DeleteIndPCFBinding",
		http.MethodDelete,
		"/sessions/:sessionId",
		DeleteSession,
	},

	{
		"CreatePCFBinding",
		http.MethodPost,
		"/sessions",
		CreateSession,
	},

	{
		"GetPCFBindings",
		http.MethodGet,
		"/sessions/:sessionId",
		GetSession,
	},
}
