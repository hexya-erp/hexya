// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package menus

import (
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
)

var log *logging.Logger

// BootStrap the menus by linking parents and children
// and populates the Registry
func BootStrap() {
	for _, menu := range bootstrapMap {
		if menu.ParentID != "" {
			parentMenu := bootstrapMap[menu.ParentID]
			if parentMenu == nil {
				log.Panic("Unknown parent menu ID", "parentID", menu.ParentID)
			}
			menu.Parent = parentMenu
		}
		if menu.ActionID != "" {
			menu.Action = actions.Registry.MustGetById(menu.ActionID)
			if menu.Name == "" {
				menu.Name = menu.Action.Name
			}
		}
		Registry.Add(menu)
	}
}

func init() {
	Registry = NewCollection()
	bootstrapMap = make(map[string]*Menu)
	log = logging.GetLogger("menus")
}
