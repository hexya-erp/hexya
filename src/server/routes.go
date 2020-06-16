// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package server

import "github.com/gin-gonic/gin"

// A HandlerFunc is a function that can be used for handling a given request or as a middleware
type HandlerFunc func(*Context)

// RouterGroup is used internally to configure router, a RouterGroup is associated with a prefix
// and an array of handlers (middleware)
type RouterGroup struct {
	gin.RouterGroup
}

// wrapContextFuncs returns a slice of gin.HandlerFunc from a slice of HandlerFunc
func wrapContextFuncs(handlers ...HandlerFunc) []gin.HandlerFunc {
	wrappedHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, hf := range handlers {
		// We use here a closure inside a closure to freeze hf
		wrappedHandlers[i] = func(f HandlerFunc) gin.HandlerFunc {
			return func(ctx *gin.Context) {
				f(&Context{Context: ctx})
			}
		}(hf)
	}
	return wrappedHandlers
}

// Group creates a new router group. You should add all the routes that have common middlwares or the same path prefix.
// For example, all the routes that use a common middlware for authorization could be grouped.
func (rg *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		RouterGroup: *rg.RouterGroup.Group(relativePath, wrapContextFuncs(handlers...)...),
	}
}

// Use adds middleware to the group.
func (rg *RouterGroup) Use(middleware ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.Use(wrapContextFuncs(middleware...)...)
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
func (rg *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.Handle(httpMethod, relativePath, wrapContextFuncs(handlers...)...)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (rg *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.POST(relativePath, wrapContextFuncs(handlers...)...)
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (rg *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.GET(relativePath, wrapContextFuncs(handlers...)...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (rg *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.DELETE(relativePath, wrapContextFuncs(handlers...)...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (rg *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.PATCH(relativePath, wrapContextFuncs(handlers...)...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (rg *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.PUT(relativePath, wrapContextFuncs(handlers...)...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (rg *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.OPTIONS(relativePath, wrapContextFuncs(handlers...)...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (rg *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.HEAD(relativePath, wrapContextFuncs(handlers...)...)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE
func (rg *RouterGroup) Any(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return rg.RouterGroup.Any(relativePath, wrapContextFuncs(handlers...)...)
}
