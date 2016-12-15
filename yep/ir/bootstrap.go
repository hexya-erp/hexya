// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ir

import (
	"strings"
)

/*
BootStrap computes all views, actions and menus after they have been
added by the modules.
*/
func BootStrap() {
	computeViews()
	computeActions()
}

/*
ComputeViews makes the necessary updates to view definitions. In particular:
- sets the type of the view from the arch root.
- populates the fields map from the views arch.
*/
func computeViews() {
	for _, v := range ViewsRegistry.views {
		archElem := xmlToElement(v.Arch)

		// Set view type
		v.Type = ViewType(archElem.Tag)

		// Populate fields map
		fieldElems := archElem.FindElements("//field")
		for _, f := range fieldElems {
			v.Fields = append(v.Fields, f.SelectAttr("name").Value)
		}
	}
}

/*
ComputeActions makes the necessary updates to action definitions. In particular:
- Add a few default values
- Add View to Views if not already present
- Add all views that are not specified
*/
func computeActions() {
	for _, a := range ActionsRegistry.actions {
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
			vType := ViewsRegistry.GetViewById(a.View[0]).Type
			newRef := ViewTuple{
				ID:   a.View[0],
				Type: vType,
			}
			a.Views = append(a.Views, newRef)
		}

		// Add views of ViewMode that are not specified
		modesStr := strings.Split(a.ViewMode, ",")
		modes := make([]ViewType, len(modesStr))
		for i, v := range modesStr {
			modes[i] = ViewType(strings.TrimSpace(v))
		}
	modeLoop:
		for _, mode := range modes {
			for _, vRef := range a.Views {
				if vRef.Type == mode {
					continue modeLoop
				}
			}
			// No view defined for mode, we need to find it.
			view := ViewsRegistry.GetFirstViewForModel(a.Model, ViewType(mode))
			newRef := ViewTuple{
				ID:   view.ID,
				Type: view.Type,
			}
			a.Views = append(a.Views, newRef)
		}

		// Fixes
		fixViewModes(a)
	}

}

/*
For OpenERP historical reasons, tree views are called 'list' when
in ActionViewType 'form' and 'tree' when in ActionViewType 'tree'.
fixViewModes makes the necessary changes to the given action.
*/
func fixViewModes(a *BaseAction) {
	if a.ActViewType == ActionViewTypeForm {
		for i, v := range a.Views {
			if v.Type == VIEW_TYPE_TREE {
				v.Type = VIEW_TYPE_LIST
			}
			a.Views[i].Type = v.Type
		}
	}
}
