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
	"crypto/tls"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
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

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (s *Server) Run(addr string) (err error) {
	defer func() { log.Error("HTTP server stopped", err) }()

	log.Info("Hexya is up and running HTTP", "address", addr)
	err = http.ListenAndServe(addr, s)
	return
}

// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (s *Server) RunTLS(addr string, certFile string, keyFile string) (err error) {
	defer func() { log.Error("HTTPS server stopped", err) }()

	log.Info("Hexya is up and running HTTPS", "address", addr, "cert", certFile, "key", keyFile)
	err = http.ListenAndServeTLS(addr, certFile, keyFile, s)
	return
}

// RunAutoTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests on port 443
// for all interfaces.
// It automatically gets certificate for the given domain from Letsencrypt.
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (s *Server) RunAutoTLS(domain string) (err error) {
	defer func() { log.Error("HTTPS server stopped", err) }()

	log.Info("Hexya is up and running HTTPS auto", "domain", domain)

	cacheDir := filepath.Join(viper.GetString("DataDir"), "autotls")
	m := &autocert.Manager{
		Cache:      autocert.DirCache(cacheDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
	}
	go http.ListenAndServe(":http", m.HTTPHandler(nil))
	srv := &http.Server{
		Addr:      ":https",
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		Handler:   s,
	}
	err = srv.ListenAndServeTLS("", "")
	return
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
	Arguments     []string `json:"arguments"`
	ExceptionType string   `json:"exception_type"`
	Debug         string   `json:"debug"`
}

// JSONRPCError is the format of an Error in a ResponseError
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var hexyaServer *Server
var log logging.Logger

// GetServer return the http server instance
func GetServer() *Server {
	return hexyaServer
}

func init() {
	log = logging.GetLogger("server")
	// Set to ReleaseMode now for tests and is overridden later (hexya/cmd/server.go)
	gin.SetMode(gin.ReleaseMode)
	hexyaServer = &Server{gin.New()}
	store := sessions.NewCookieStore([]byte(">r&5#5T/sG-jnf=EW8$(WQX'-m2R6Gk*^qqr`CxEtG'wQ[/'G@`NYn^on?b!4G`9"),
		[]byte("!WY9Q|}09!4Ke=@w0HS|]$u,p1f^k(5T"))
	hexyaServer.Use(gin.Recovery())
	hexyaServer.Use(sessions.Sessions("hexya-session", store))
	hexyaServer.Use(logging.LogForGin(log))
}

// PreInit runs all actions that need to be done after we get the configuration,
// but before bootstrap.
//
// This function runs successively all PreInit() func of modules
func PreInit() {
	PreInitModules()
}

// PreInitModules calls successively all PreInit functions of all installed modules
func PreInitModules() {
	for _, module := range Modules {
		if module.PreInit != nil {
			module.PreInit()
		}
	}
}

// PostInit runs all actions that need to be done after all modules have been loaded.
// This is typically all actions that need to be done after bootstrapping the models.
// This function:
// - runs successively all PostInit() func of all modules,
// - loads html templates from all modules.
func PostInit() {
	PostInitModules()
	hexyaServer.LoadHTMLGlob(generate.HexyaDir + "/hexya/server/templates/**/*.html")
}

// PostInitModules calls successively all PostInit functions of all installed modules
func PostInitModules() {
	for _, module := range Modules {
		if module.PostInit != nil {
			module.PostInit()
		}
	}
}
