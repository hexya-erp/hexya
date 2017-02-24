// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"github.com/inconshreveable/log15"
	"github.com/npiganeau/yep/yep/server"
	"github.com/npiganeau/yep/yep/tools/logging"
)

var log log15.Logger

// BootStrap creates the actual controllers from the controllers registry.
// This function must be called before starting the http server.
func BootStrap() {
	Registry.createRoutes(server.GetServer().Group("/"))
}

func init() {
	log = logging.GetLogger("controllers")
	Registry = newGroup("/")
}
