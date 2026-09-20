// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/tools/exceptions"
)

func TestGetServer(t *testing.T) {
	srv := GetServer()
	assert.NotNil(t, srv)
	assert.Equal(t, hexyaServer, srv)
	assert.NotNil(t, srv.Engine)
	assert.Equal(t, gin.ReleaseMode, gin.Mode())
}

func TestWrapContextFuncs(t *testing.T) {
	t.Run("Without any handler", func(t *testing.T) {
		assert.Len(t, wrapContextFuncs(), 0)
	})
	t.Run("Handlers should be wrapped in order", func(t *testing.T) {
		var calls []string
		wrapped := wrapContextFuncs(
			func(c *Context) { calls = append(calls, "first"); c.Super() },
			func(c *Context) { calls = append(calls, "second") })
		assert.Len(t, wrapped, 2)
		for _, hf := range wrapped {
			hf(&gin.Context{})
		}
		assert.Equal(t, []string{"first", "second"}, calls)
	})
}

func TestRouterGroups(t *testing.T) {
	srv := &Server{gin.New()}
	var middlewareCalled bool
	group := srv.Group("/test", func(c *Context) {
		middlewareCalled = true
		c.Super()
	})
	assert.NotNil(t, group)
	group.Use(func(c *Context) { c.Super() })
	subGroup := group.Group("/sub")
	assert.NotNil(t, subGroup)

	handler := func(c *Context) { c.String(http.StatusOK, "%s %s", c.Request.Method, c.Request.URL.Path) }
	group.GET("/get", handler)
	group.POST("/post", handler)
	group.PUT("/put", handler)
	group.PATCH("/patch", handler)
	group.DELETE("/delete", handler)
	group.OPTIONS("/options", handler)
	group.HEAD("/head", handler)
	group.Handle(http.MethodGet, "/handle", handler)
	subGroup.Any("/any", handler)

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/test/get"},
		{http.MethodPost, "/test/post"},
		{http.MethodPut, "/test/put"},
		{http.MethodPatch, "/test/patch"},
		{http.MethodDelete, "/test/delete"},
		{http.MethodOptions, "/test/options"},
		{http.MethodHead, "/test/head"},
		{http.MethodGet, "/test/handle"},
		{http.MethodGet, "/test/sub/any"},
		{http.MethodPost, "/test/sub/any"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			middlewareCalled = false
			resp := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			srv.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusOK, resp.Code)
			assert.True(t, middlewareCalled)
			if tc.method != http.MethodHead {
				assert.Equal(t, tc.method+" "+tc.path, resp.Body.String())
			}
		})
	}
}

// testRequest runs the given handler on the test server at /rpc with the given body
// and returns the response recorder.
func testRequest(handler HandlerFunc, body string) *httptest.ResponseRecorder {
	srv := &Server{gin.New()}
	srv.Group("/").POST("rpc", handler)
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(resp, req)
	return resp
}

func TestContextRPC(t *testing.T) {
	t.Run("Returning a result", func(t *testing.T) {
		resp := testRequest(func(c *Context) {
			c.RPC(http.StatusOK, map[string]string{"foo": "bar"})
		}, `{"jsonrpc": "2.0", "id": 12, "method": "call", "params": {}}`)
		assert.Equal(t, http.StatusOK, resp.Code)
		var res ResponseRPC
		assert.Nil(t, json.Unmarshal(resp.Body.Bytes(), &res))
		assert.Equal(t, "2.0", res.JsonRPC)
		assert.Equal(t, int64(12), res.ID)
		assert.Equal(t, map[string]any{"foo": "bar"}, res.Result)
	})
	t.Run("Getting the id from the context", func(t *testing.T) {
		resp := testRequest(func(c *Context) {
			c.Set("id", int64(56))
			c.RPC(http.StatusOK, "result")
		}, "")
		assert.Equal(t, http.StatusOK, resp.Code)
		var res ResponseRPC
		assert.Nil(t, json.Unmarshal(resp.Body.Bytes(), &res))
		assert.Equal(t, int64(56), res.ID)
		assert.Equal(t, "result", res.Result)
	})
	t.Run("Returning a user error", func(t *testing.T) {
		resp := testRequest(func(c *Context) {
			c.RPC(http.StatusInternalServerError, nil,
				exceptions.UserError{Message: "Error Message", Debug: "Debug Info"})
		}, `{"jsonrpc": "2.0", "id": 12, "method": "call", "params": {}}`)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		var res ResponseError
		assert.Nil(t, json.Unmarshal(resp.Body.Bytes(), &res))
		assert.Equal(t, int64(12), res.ID)
		assert.Equal(t, http.StatusInternalServerError, res.Error.Code)
		assert.Equal(t, "Hexya Server Error", res.Error.Message)
		data := res.Error.Data.(map[string]any)
		assert.Equal(t, []any{"Error Message"}, data["arguments"])
		assert.Equal(t, "user_error", data["exception_type"])
		assert.Equal(t, "Debug Info", data["debug"])
	})
	t.Run("A nil error should be ignored", func(t *testing.T) {
		resp := testRequest(func(c *Context) {
			c.Set("id", int64(1))
			c.RPC(http.StatusOK, "result", nil)
		}, "")
		assert.Equal(t, http.StatusOK, resp.Code)
	})
	t.Run("Unknown error types should abort", func(t *testing.T) {
		resp := testRequest(func(c *Context) {
			c.Set("id", int64(1))
			c.RPC(http.StatusInternalServerError, nil, errors.New("unknown error"))
		}, "")
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
	t.Run("Invalid JSON body should abort", func(t *testing.T) {
		resp := testRequest(func(c *Context) {
			c.RPC(http.StatusOK, "result")
		}, `not json`)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestContextBindRPCParams(t *testing.T) {
	type params struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	t.Run("Binding valid params", func(t *testing.T) {
		var data params
		var id any
		resp := testRequest(func(c *Context) {
			c.BindRPCParams(&data)
			id, _ = c.Get("id")
		}, `{"jsonrpc": "2.0", "id": 12, "method": "call", "params": {"name": "Jane", "age": 24}}`)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, params{Name: "Jane", Age: 24}, data)
		assert.Equal(t, int64(12), id)
	})
	t.Run("Invalid JSON body should abort", func(t *testing.T) {
		var data params
		resp := testRequest(func(c *Context) { c.BindRPCParams(&data) }, `not json`)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
	t.Run("Invalid params should abort", func(t *testing.T) {
		var data params
		resp := testRequest(func(c *Context) { c.BindRPCParams(&data) },
			`{"jsonrpc": "2.0", "id": 12, "method": "call", "params": "not an object"}`)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestContextSession(t *testing.T) {
	var session any
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	hexyaServer.Group("/").GET("session", func(c *Context) {
		sess := c.Session()
		sess.Set("foo", "bar")
		session = sess.Get("foo")
		c.String(http.StatusOK, "ok")
	})
	hexyaServer.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "bar", session)
}

func TestContextHTTPGet(t *testing.T) {
	srv := &Server{gin.New()}
	group := srv.Group("/")
	group.GET("target", func(c *Context) {
		cookie, err := c.Cookie("hexya-session")
		assert.Nil(t, err)
		c.String(http.StatusOK, "target:%s:%s", c.Query("q"), cookie)
	})
	group.GET("source", func(c *Context) {
		resp, err := c.HTTPGet("/target?q=value")
		assert.Nil(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		c.String(resp.StatusCode, string(body))
	})
	ts := httptest.NewServer(srv)
	defer ts.Close()
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/source", nil)
	req.AddCookie(&http.Cookie{Name: "hexya-session", Value: "mySession"})
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "target:value:mySession", string(body))
}
