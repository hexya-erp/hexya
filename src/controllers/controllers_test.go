// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/src/server"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func newServer() *server.Server {
	gin.SetMode(gin.ReleaseMode)
	return &server.Server{Engine: gin.New()}
}

func newRegistry() *Group {
	registry := newGroup("/")
	registry.AddGroup("/test")
	return registry
}

func TestControllers(t *testing.T) {
	t.Run("Testing inheritable controllers", func(t *testing.T) {
		t.Run("Testing GetGroup", func(t *testing.T) {
			registry := newRegistry()
			grp, err := registry.GetGroup("/test")
			assert.EqualValues(t, grp, registry.groups["/test"])
			assert.NotNil(t, err)
		})
		t.Run("Testing simple addition of controllers", func(t *testing.T) {
			registry := newRegistry()
			grp := registry.MustGetGroup("/test")
			grp.AddController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "pong")
			})
			assert.True(t, grp.HasController(http.MethodGet, "/ping"))
			srv := newServer()
			registry.createRoutes(srv.Group("/"))
			r := performRequest(srv, http.MethodGet, "/test/ping")
			assert.EqualValues(t, r.Code, http.StatusOK)
			assert.EqualValues(t, r.Body.String(), "pong")
		})
		t.Run("Testing inheritance of controller", func(t *testing.T) {
			registry := newRegistry()
			grp := registry.MustGetGroup("/test")
			grp.AddController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "pong")
			})
			grp.ExtendController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "before*")
				ctx.String(http.StatusOK, "*after")
			})
			grp.ExtendController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.Super()
				ctx.String(http.StatusOK, "/after2")
			})
			srv := newServer()
			registry.createRoutes(srv.Group("/"))
			r := performRequest(srv, http.MethodGet, "/test/ping")
			assert.EqualValues(t, r.Code, http.StatusOK)
			assert.EqualValues(t, r.Body.String(), "before**afterpong/after2")
		})
		t.Run("Testing overriding of controller", func(t *testing.T) {
			registry := newRegistry()
			grp := registry.MustGetGroup("/test")
			grp.AddController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "pong")
			})
			grp.OverrideController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "before*")
				ctx.String(http.StatusOK, "*after")
			})
			grp.ExtendController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.Super()
				ctx.String(http.StatusOK, "/after2")
			})
			srv := newServer()
			registry.createRoutes(srv.Group("/"))
			r := performRequest(srv, http.MethodGet, "/test/ping")
			assert.EqualValues(t, r.Code, http.StatusOK)
			assert.EqualValues(t, r.Body.String(), "before**after/after2")
		})
		t.Run("Testing group middlewares", func(t *testing.T) {
			registry := newRegistry()
			grp := registry.MustGetGroup("/test")
			grp.AddMiddleWare(func(ctx *server.Context) {
				ctx.String(http.StatusOK, "middleware-")
				ctx.Next()
				ctx.String(http.StatusOK, "-middleware")
			})
			grp.AddMiddleWare(func(ctx *server.Context) {
				ctx.String(http.StatusOK, "hexya-")
			})
			grp.AddController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "pong")
			})
			grp.ExtendController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "before/")
			})
			srv := newServer()
			registry.createRoutes(srv.Group("/"))
			r := performRequest(srv, http.MethodGet, "/test/ping")
			assert.EqualValues(t, r.Code, http.StatusOK)
			assert.EqualValues(t, r.Body.String(), "hexya-middleware-before/pong-middleware")
		})
		t.Run("Testing static dir controller", func(t *testing.T) {
			registry := newRegistry()
			grp := registry.MustGetGroup("/test")
			grp.AddStatic("/static", "testdata")
			srv := newServer()
			registry.createRoutes(srv.Group("/"))
			r := performRequest(srv, http.MethodGet, "/test/static/testfile.js")
			assert.EqualValues(t, r.Code, http.StatusOK)
			assert.EqualValues(t, r.Body.String(), `window.alert("Test message");`)
		})
		t.Run("Getting a group that does not exist should fail", func(t *testing.T) {
			registry := newRegistry()
			assert.Panics(t, func() { registry.MustGetGroup("/nonexistent") })
		})
		t.Run("Adding an already existing static dir should fail", func(t *testing.T) {
			registry := newRegistry()
			grp := registry.MustGetGroup("/test")
			grp.AddStatic("/static", "testdata")
			assert.Panics(t, func() { grp.AddStatic("/static", "testdata") })
		})
		t.Run("Adding an already existing group should fail", func(t *testing.T) {
			registry := newRegistry()
			assert.Panics(t, func() { registry.AddGroup("/test") })
		})
		t.Run("Adding an already existing controller should fail", func(t *testing.T) {
			registry := newRegistry()
			grp := registry.MustGetGroup("/test")
			grp.AddController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "pong")
			})
			assert.Panics(t, func() { grp.AddController(http.MethodGet, "/ping", func(ctx *server.Context) {}) })
		})
		t.Run("Extending a controller that does not exist should fail", func(t *testing.T) {
			registry := newRegistry()
			assert.Panics(t, func() { registry.ExtendController(http.MethodGet, "/nonexistent", func(ctx *server.Context) {}) })
		})
		t.Run("Overriding a controller that does not exist should fail", func(t *testing.T) {
			registry := newRegistry()
			assert.Panics(t, func() { registry.OverrideController(http.MethodGet, "/nonexistent", func(ctx *server.Context) {}) })
		})
		t.Run("Boostrap should not panic", func(t *testing.T) {
			assert.NotPanics(t, BootStrap)
		})
	})
}
