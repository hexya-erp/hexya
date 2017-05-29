// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/hexya/server"
	. "github.com/smartystreets/goconvey/convey"
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

func TestControllers(t *testing.T) {
	Convey("Testing inheritable controllers", t, func() {
		registry := newGroup("/")
		registry.AddGroup("/test")
		Convey("Testing GetGroup", func() {
			grp := registry.GetGroup("/test")
			So(grp, ShouldEqual, registry.groups["/test"])
		})
		Convey("Testing simple addition of controllers", func() {
			grp := registry.GetGroup("/test")
			grp.AddController(http.MethodGet, "/ping", func(ctx *server.Context) {
				ctx.String(http.StatusOK, "pong")
			})
			srv := newServer()
			registry.createRoutes(srv.Group("/"))
			r := performRequest(srv, http.MethodGet, "/test/ping")
			So(r.Code, ShouldEqual, http.StatusOK)
			So(r.Body.String(), ShouldEqual, "pong")
		})
		Convey("Testing inheritance of controller", func() {
			grp := registry.GetGroup("/test")
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
			So(r.Code, ShouldEqual, http.StatusOK)
			So(r.Body.String(), ShouldEqual, "before**afterpong/after2")
		})
		Convey("Testing overriding of controller", func() {
			grp := registry.GetGroup("/test")
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
			So(r.Code, ShouldEqual, http.StatusOK)
			So(r.Body.String(), ShouldEqual, "before**after/after2")
		})
		Convey("Testing group middlewares", func() {
			grp := registry.GetGroup("/test")
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
			So(r.Code, ShouldEqual, http.StatusOK)
			So(r.Body.String(), ShouldEqual, "hexya-middleware-before/pong-middleware")
		})
	})
}
