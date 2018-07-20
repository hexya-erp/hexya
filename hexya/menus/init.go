// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package menus

import (
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
)

var log logging.Logger

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
		var noName bool
		if menu.ActionID != "" {
			menu.Action = actions.Registry.MustGetById(menu.ActionID)
			if menu.Name == "" {
				noName = true
				menu.Name = menu.Action.Name
			}
		}
		// Add translations
		if menu.names == nil {
			menu.names = make(map[string]string)
		}
		for _, lang := range i18n.Langs {
			nameTrans := i18n.TranslateResourceItem(lang, menu.ID, menu.Name)
			if noName {
				nameTrans = i18n.TranslateResourceItem(lang, menu.Action.ID, menu.Name)
			}
			menu.names[lang] = nameTrans
		}
		Registry.Add(menu)
	}
}

func init() {
	Registry = NewCollection()
	bootstrapMap = make(map[string]*Menu)
	log = logging.GetLogger("menus")
}
