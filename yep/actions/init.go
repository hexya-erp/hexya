// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package actions

import (
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/npiganeau/yep/yep/tools/logging"
	"github.com/npiganeau/yep/yep/views"
)

var log log15.Logger

//BootStrap makes the necessary updates to action definitions. In particular:
//- Add a few default values
//- Add View to Views if not already present
//- Add all views that are not specified
func BootStrap() {
	for _, a := range Registry.actions {
		// Set a few default values
		if a.Target == "" {
			a.Target = "current"
		}
		a.AutoSearch = !a.ManualSearch
		if a.ActViewType == "" {
			a.ActViewType = ActionViewTypeForm
		}

		// Add View to Views if not already present
		var present bool
		// Check if view is present in Views
		for _, view := range a.Views {
			if view.ID != "" && len(a.View) > 0 {
				if view.ID == a.View[0] {
					present = true
					break
				}
			}
		}
		// Add View if not present in Views
		if !present && len(a.View) > 0 && a.View[0] != "" {
			vType := views.ViewsRegistry.GetViewById(a.View[0]).Type
			newRef := views.ViewTuple{
				ID:   a.View[0],
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
			view := views.ViewsRegistry.GetFirstViewForModel(a.Model, views.ViewType(mode))
			newRef := views.ViewTuple{
				ID:   view.ID,
				Type: view.Type,
			}
			a.Views = append(a.Views, newRef)
		}

		// Fixes
		fixViewModes(a)
	}

}

//For OpenERP historical reasons, tree views are called 'list' when
//in ActionViewType 'form' and 'tree' when in ActionViewType 'tree'.
//fixViewModes makes the necessary changes to the given action.
func fixViewModes(a *BaseAction) {
	if a.ActViewType == ActionViewTypeForm {
		for i, v := range a.Views {
			if v.Type == views.VIEW_TYPE_TREE {
				v.Type = views.VIEW_TYPE_LIST
			}
			a.Views[i].Type = v.Type
		}
	}
}

func init() {
	log = logging.GetLogger("actions")
	Registry = NewActionsCollection()
}
