/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package ir

import (
	"strings"

	"github.com/beevik/etree"
	"github.com/npiganeau/yep/yep/tools"
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
			tools.LogAndPanic(log, "Unable to read view", "view", v.ID, "error", err)
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
