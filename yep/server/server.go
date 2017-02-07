// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"encoding/json"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/npiganeau/yep/yep/tools/generate"
	"github.com/npiganeau/yep/yep/tools/logging"
)

// A Server is the http server of the application
// It is internally a wrapper around a gin.Engine
type Server struct {
	*gin.Engine
}

// Group creates a new router group. You should add all the routes that have common middlwares or the same path prefix.
// For example, all the routes that use a common middlware for authorization could be grouped.
func (s *Server) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		RouterGroup: *s.Engine.Group(relativePath, wrapContextFuncs(handlers...)...),
	}
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in github.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (s *Server) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.Handle(httpMethod, relativePath, wrapContextFuncs(handlers...)...)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (s *Server) POST(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.POST(relativePath, wrapContextFuncs(handlers...)...)
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (s *Server) GET(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.GET(relativePath, wrapContextFuncs(handlers...)...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (s *Server) DELETE(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.DELETE(relativePath, wrapContextFuncs(handlers...)...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (s *Server) PATCH(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.PATCH(relativePath, wrapContextFuncs(handlers...)...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (s *Server) PUT(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.PUT(relativePath, wrapContextFuncs(handlers...)...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (s *Server) OPTIONS(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.OPTIONS(relativePath, wrapContextFuncs(handlers...)...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (s *Server) HEAD(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.HEAD(relativePath, wrapContextFuncs(handlers...)...)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE
func (s *Server) Any(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return s.RouterGroup.Any(relativePath, wrapContextFuncs(handlers...)...)
}

// A RequestRPC is the message format expected from a client
type RequestRPC struct {
	JsonRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// A ResponseRPC is the message format sent back to a client
// in case of success
type ResponseRPC struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Result  interface{} `json:"result"`
}

// A ResponseError is the message format sent back to a
// client in case of failure
type ResponseError struct {
	JsonRPC string       `json:"jsonrpc"`
	ID      int64        `json:"id"`
	Error   JSONRPCError `json:"error"`
}

// JSONRPCErrorData is the format of the Data field of an Error Response
type JSONRPCErrorData struct {
	Arguments string `json:"arguments"`
	Debug     string `json:"debug"`
}

// JSONRPCError is the format of an Error in a ResponseError
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var yepServer *Server
var log log15.Logger

// GetServer return the http server instance
func GetServer() *Server {
	return yepServer
}

func init() {
	log = logging.GetLogger("server")
	gin.SetMode(gin.ReleaseMode)
	yepServer = &Server{gin.New()}
	store := sessions.NewCookieStore([]byte(">r&5#5T/sG-jnf=EW8$(WQX'-m2R6Gk*^qqr`CxEtG'wQ[/'G@`NYn^on?b!4G`9"),
		[]byte("!WY9Q|}09!4Ke=@w0HS|]$u,p1f^k(5T"))
	yepServer.Use(gin.Recovery())
	yepServer.Use(sessions.Sessions("yep-session", store))
	yepServer.Use(logging.Log15ForGin(log))
	cleanModuleSymlinks()
}

// PostInit runs all actions that need to be done after all modules have been loaded.
// This is typically all actions that need to be done after bootstrapping the models.
// This function:
// - loads the data from the data files of all modules,
// - runs successively all PostInit() func of all modules,
// - loads html templates from all modules.
func PostInit() {
	for _, module := range Modules {
		module.PostInit()
	}

	yepServer.LoadHTMLGlob(generate.YEPDir + "/yep/server/templates/**/*.html")
}
