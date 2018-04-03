// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package actions

import (
	"strings"

	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/hexya-erp/hexya/hexya/views"
)

var log *logging.Logger

// BootStrap actions.
// This function must be called prior to any access to the actions Registry.
func BootStrap() {
	for _, a := range Registry.actions {
		switch a.Type {
		case ActionActWindow:
			bootStrapWindowAction(a)
		}
		// Populate translations
		if a.names == nil {
			a.names = make(map[string]string)
		}
		for _, lang := range i18n.Langs {
			nameTrans := i18n.TranslateResourceItem(lang, a.ID, a.Name)
			a.names[lang] = nameTrans
		}
	}
}

// bootStrapWindowAction makes the necessary updates to action definitions. In particular:
// - Add a few default values
// - Add View to Views if not already present
// - Add all views that are not specified
func bootStrapWindowAction(a *Action) {
	// Set a few default values
	if a.Target == "" {
		a.Target = "current"
	}
	a.AutoSearch = !a.ManualSearch
	if a.ActViewType == "" {
		a.ActViewType = ActionViewTypeForm
	}
	a.Help = a.HelpXML.Content

	// Add View to Views if not already present
	var present bool
	// Check if view is present in Views
	for _, view := range a.Views {
		if !a.View.IsNull() {
			if view.ID == a.View.ID() {
				present = true
				break
			}
		}
	}
	// Add View if not present in Views
	if !present && !a.View.IsNull() {
		vType := views.Registry.GetByID(a.View.ID()).Type
		newRef := views.ViewTuple{
			ID:   a.View.ID(),
			Type: vType,
		}
		a.Views = append(a.Views, newRef)
	}

	// Add views of ViewMode that are not specified
	modesStr := strings.Split(a.ViewMode, ",")
	modes := make([]views.ViewType, len(modesStr))
	for i, v := range modesStr {
		modes[i] = views.ViewType(strings.TrimSpace(v))
	}
modeLoop:
	for _, mode := range modes {
		for _, vRef := range a.Views {
			if vRef.Type == mode {
				continue modeLoop
			}
		}
		// No view defined for mode, we need to find it.
		view := views.Registry.GetFirstViewForModel(a.Model, views.ViewType(mode))
		newRef := views.ViewTuple{
			ID:   view.ID,
			Type: view.Type,
		}
		a.Views = append(a.Views, newRef)
	}

	// Fixes
	fixViewModes(a)
}

// fixViewModes makes the necessary changes to the given action.
//
// For OpenERP historical reasons, tree views are called 'list' when
// in ActionViewType 'form' and 'tree' when in ActionViewType 'tree'.
func fixViewModes(a *Action) {
	if a.ActViewType == ActionViewTypeForm {
		for i, v := range a.Views {
			if v.Type == views.ViewTypeTree {
				v.Type = views.ViewTypeList
			}
			a.Views[i].Type = v.Type
		}
	}
}

func init() {
	log = logging.GetLogger("actions")
	Registry = NewCollection()
}
