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

	"github.com/beevik/etree"
	"github.com/npiganeau/yep/yep/tools/logging"
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
		doc := etree.NewDocument()
		if err := doc.ReadFromString(v.Arch); err != nil {
			logging.LogAndPanic(log, "Unable to read view", "view", v.ID, "error", err)
		}

		// Set view type
		v.Type = ViewType(doc.ChildElements()[0].Tag)

		// Populate fields map
		fieldElems := doc.FindElements("//field")
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
			a.ActViewType = ACTION_VIEW_TYPE_FORM
		}

		// Add View to Views if not already present
		var present bool
		// Check if view is present in Views
		for _, view := range a.Views {
			if len(view) > 0 && len(a.View) > 0 {
				if view[0] == a.View[0] {
					present = true
					break
				}
			}
		}
		// Add View if not present in Views
		if !present && len(a.View) > 0 && a.View[0] != "" {
			vType := ViewsRegistry.GetViewById(a.View[0]).Type
			newRef := ViewRef{
				0: a.View[0],
				1: string(vType),
			}
			a.Views = append(a.Views, newRef)
		}

		// Add views of ViewMode that are not specified
		modes := strings.Split(a.ViewMode, ",")
		for i, v := range modes {
			modes[i] = strings.TrimSpace(v)
		}
	modeLoop:
		for _, mode := range modes {
			for _, vRef := range a.Views {
				if vRef[1] == mode {
					continue modeLoop
				}
			}
			// No view defined for mode, we need to find it.
			view := ViewsRegistry.GetFirstViewForModel(a.Model, ViewType(mode))
			newRef := ViewRef{
				0: view.ID,
				1: string(view.Type),
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
	if a.ActViewType == ACTION_VIEW_TYPE_FORM {
		for i, v := range a.Views {
			vType := ViewType(v[1])
			if vType == VIEW_TYPE_TREE {
				vType = VIEW_TYPE_LIST
			}
			a.Views[i][1] = string(vType)
		}
	}
}
