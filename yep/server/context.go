// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package server

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/npiganeau/yep/yep/tools"
)

// The Context allows to pass data across controller layers
// and middlewares.
type Context struct {
	*gin.Context
}

// RPC serializes the given struct as JSON-RPC into the response body.
func (c *Context) RPC(code int, obj interface{}, err ...error) {
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
				Data: JSONRPCErrorData{
					Arguments: "Internal Server Error",
					Debug:     err[0].Error(),
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
func (c *Context) BindRPCParams(data interface{}) {
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

// Session returns the current Session instance
func (c *Context) Session() sessions.Session {
	return sessions.Default(c.Context)
}

// Super calls the next middleware / handler layer
// It is an alias for Next
func (c *Context) Super() {
	c.Next()
}

// HTTPGet makes an http GET request to this server with the context's session cookie
func (c *Context) HTTPGet(uri string) (*http.Response, error) {
	url := tools.AbsolutizeURL(c.Request, uri)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	sessionCookie, _ := c.Cookie("yep-session")
	req.AddCookie(&http.Cookie{
		Name:  "yep-session",
		Value: sessionCookie,
	})
	client := http.Client{}
	return client.Do(req)
}
