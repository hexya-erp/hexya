// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import "github.com/hexya-erp/hexya/hexya/server"

// Registry is the central collection of all the application controllers
var Registry *Group

// A Controller is a server function that is called through
// an http route.
type Controller struct {
	route    Route
	handlers []server.HandlerFunc
}

// A Group is used to group routes with common prefix, in order
// to apply it specific middlewares.
type Group struct {
	relativePath string
	controllers  map[Route]*Controller
	groups       map[string]*Group
	static       map[string]string
	middleWares  []server.HandlerFunc
}

// newGroup returns a pointer to a new empty Group
func newGroup(relativePath string) *Group {
	res := Group{
		relativePath: relativePath,
		controllers:  make(map[Route]*Controller),
		groups:       make(map[string]*Group),
		static:       make(map[string]string),
	}
	return &res
}

// AddGroup adds a sub-group to this group with the given
// relativePath and returns the newly created group.
// It panics if the group already exists.
func (g *Group) AddGroup(relativePath string) *Group {
	if _, exists := g.groups[relativePath]; exists {
		log.Panic("Group already exists in this group", "path", relativePath, "group", g.relativePath)
	}
	newGrp := newGroup(relativePath)
	g.groups[relativePath] = newGrp
	return newGrp
}

// HasController returns true if the given method and path already has a controller function
func (g *Group) HasController(method, relativePath string) bool {
	route := Route{
		Method: method,
		Path:   relativePath,
	}
	if _, exists := g.controllers[route]; exists {
		return true
	}
	return false
}

// AddController creates a controller for the given method and path and sets
// fnct as the base handler function for this controller.
// It panics if such a controller already exists.
func (g *Group) AddController(method, relativePath string, fnct server.HandlerFunc) {
	route := Route{
		Method: method,
		Path:   relativePath,
	}
	if _, exists := g.controllers[route]; exists {
		log.Panic("Trying to add a controller that already exists", "method", method, "path", relativePath)
	}
	controller := &Controller{
		route:    route,
		handlers: []server.HandlerFunc{fnct},
	}
	g.controllers[route] = controller
}

// ExtendController extends the controller for the given method and path
// with the given fnct handler function.
//
// The handler fnct should call its context Super() method to call the
// original controller implementation.
// Note that if Super() is not called explicitly, the original implementation
// is automatically called at the end of the handler fnct. If this is not
// the wanted behaviour, use OverrideController instead.
//
// ExtendController panics if such a controller does not exist
func (g *Group) ExtendController(method, relativePath string, fnct server.HandlerFunc) {
	route := Route{
		Method: method,
		Path:   relativePath,
	}
	if _, exists := g.controllers[route]; !exists {
		log.Panic("Trying to extend a non-existent controller",
			"method", method, "path", relativePath)
	}
	g.controllers[route].handlers = append([]server.HandlerFunc{fnct}, g.controllers[route].handlers...)
}

// OverrideController overrides the controller for the given method and path
// with the given fnct handler function.
//
// Call to the handler fnct context Super() method has no effect, since all previous
// implementations are discarded.
// If this is not the wanted behaviour, use ExtendController instead.
//
// OverrideController panics if such a controller does not exist
func (g *Group) OverrideController(method, relativePath string, fnct server.HandlerFunc) {
	route := Route{
		Method: method,
		Path:   relativePath,
	}
	if _, exists := g.controllers[route]; !exists {
		log.Panic("Trying to override a non-existent controller",
			"method", method, "path", relativePath)
	}
	g.controllers[route].handlers = append([]server.HandlerFunc{fnct})
}

// AddStatic creates a new route at relativePath that will serve
// the static files found at fsPath on the file system.
func (g *Group) AddStatic(relativePath, fsPath string) {
	if _, exists := g.static[relativePath]; exists {
		log.Panic("Static path already exists in this group", "path", relativePath, "group", g.relativePath)
	}
	g.static[relativePath] = fsPath
}

// AddMiddleWare adds the given fnct as a new middleware for this group
// fnct will be executed before any other middleware of this group.
//
// Call the Next() method of fnct's context to call the next middleware.
// If Next is not called explicitly, the next middleware is called automatically
// at the end of fnct.
func (g *Group) AddMiddleWare(fnct server.HandlerFunc) {
	g.middleWares = append([]server.HandlerFunc{fnct}, g.middleWares...)
}

// MustGetGroup returns the sub group of this group for the given relativePath
// It panics if this group does not exist
func (g *Group) MustGetGroup(relativePath string) *Group {
	group, exists := g.groups[relativePath]
	if !exists {
		log.Panic("Group not found", "group", relativePath, "base", g.relativePath)
	}
	return group
}

// GetGroup returns the sub group of this group for the given relativePath
// the returned boolean is 'false' if the sub group does not exists for this group
func (g *Group) GetGroup(relativePath string) (*Group, bool) {
	group, exists := g.groups[relativePath]
	return group, exists
}

// createRoutes creates the router groups and routes defined in this Group
// in the given underlying server.RouterGroup recursively.
func (g *Group) createRoutes(base *server.RouterGroup) {
	for _, mw := range g.middleWares {
		base.Use(mw)
	}
	for path, grp := range g.groups {
		newRtGrp := base.Group(path)
		grp.createRoutes(newRtGrp)
	}
	for route, ctlr := range g.controllers {
		base.Handle(route.Method, route.Path, ctlr.handlers...)
	}
	for path, fsPath := range g.static {
		base.Static(path, fsPath)
	}
}

// A Route is the combination of a URI (Path) and an HTTP Method
type Route struct {
	Path   string
	Method string
}
