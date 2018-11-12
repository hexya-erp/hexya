// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/hexya/tools/exceptions"
	"github.com/hexya-erp/hexya/hexya/tools/hweb"
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
		if err2 := c.BindJSON(&req); err2 != nil {
			c.AbortWithError(http.StatusBadRequest, err2)
			return
		}
		id = req.ID
	}
	if len(err) > 0 && err[0] != nil {
		userError, ok2 := err[0].(exceptions.UserError)
		if !ok2 {
			c.AbortWithError(http.StatusInternalServerError, errors.New("error is of unknown type"))
			return
		}
		respErr := ResponseError{
			JsonRPC: "2.0",
			ID:      id.(int64),
			Error: JSONRPCError{
				Code:    code,
				Message: "Hexya Server Error",
				Data: JSONRPCErrorData{
					Arguments:     []string{userError.Message},
					ExceptionType: "user_error",
					Debug:         userError.Debug,
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
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	sanitizedURI, _ := url.ParseRequestURI(uri)
	targetUrl := fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, sanitizedURI.RequestURI())

	req, _ := http.NewRequest(http.MethodGet, targetUrl, nil)
	sessionCookie, _ := c.Cookie("hexya-session")
	req.AddCookie(&http.Cookie{
		Name:  "hexya-session",
		Value: sessionCookie,
	})
	client := http.Client{}
	return client.Do(req)
}

// HTML renders the HTTP template specified by its file name.
// It also updates the HTTP code and sets the Content-Type as "text/html".
// See http://golang.org/doc/articles/wiki/
func (c *Context) HTML(code int, name string, context hweb.Context) {
	c.Context.HTML(code, name, context)
}
