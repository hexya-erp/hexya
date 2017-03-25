// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package menus

import (
	"github.com/inconshreveable/log15"
	"github.com/npiganeau/yep/yep/tools/logging"
)

var log log15.Logger

// BootStrap the menus by linking parents and children
// and populates the Registry
func BootStrap() {
	for _, menu := range bootstrapMap {
		if menu.ParentID != "" {
			parentMenu := bootstrapMap[menu.ParentID]
			if parentMenu == nil {
				logging.LogAndPanic(log, "Unknown parent menu ID", "parentID", menu.ParentID)
			}
			menu.Parent = parentMenu
		}
		Registry.Add(menu)
	}
}

func init() {
	Registry = NewCollection()
	bootstrapMap = make(map[string]*Menu)
	log = logging.GetLogger("menus")
}
