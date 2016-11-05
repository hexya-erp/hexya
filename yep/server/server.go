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
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/npiganeau/yep/yep/tools/logging"
)

// A Server is the http server of the application
// It is internally a wrapper around a gin.Engine
type Server struct {
	*gin.Engine
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

// JSONRPCError is the format of an Error in a ResponseError
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// RPC serializes the given struct as JSON-RPC into the response body.
func RPC(c *gin.Context, code int, obj interface{}, err ...error) {
	id, ok := c.Get("id")
	if !ok {
		var req RequestRPC
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		id = req.ID
	}
	if len(err) > 0 && err[0] != nil {
		respErr := ResponseError{
			JsonRPC: "2.0",
			ID:      id.(int64),
			Error: JSONRPCError{
				Code:    code,
				Message: "YEP Server Error",
				Data: map[string]string{
					"debug": err[0].Error(),
				},
			},
		}
		c.JSON(code, respErr)
		return
	}
	resp := ResponseRPC{
		JsonRPC: "2.0",
		ID:      id.(int64),
		Result:  obj,
	}
	c.JSON(code, resp)
}

// BindRPCParams binds the RPC parameters to the given data object.
func BindRPCParams(c *gin.Context, data interface{}) {
	var req RequestRPC
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.Set("id", req.ID)
	if err := json.Unmarshal(req.Params, data); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
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

	yepServer.LoadHTMLGlob("yep/server/templates/**/*.html")
}
