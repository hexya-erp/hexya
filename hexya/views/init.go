// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package views

import (
	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
)

const maxInheritanceDepth = 100

var log *logging.Logger

// BootStrap makes the necessary updates to view definitions. In particular:
// - sets the type of the view from the arch root.
// - extracts embedded views
// - populates the fields map from the views arch.
func BootStrap() {
	// Inherit/Extend views
	for loop := 0; loop < maxInheritanceDepth; loop++ {
		// First step: we extend all we can with pure extension views (no ID)
		for i, xmlView := range Registry.rawInheritedViews {
			if xmlView == nil {
				continue
			}
			if xmlView.ID != "" {
				continue
			}
			baseView := Registry.GetByID(xmlView.InheritID)
			if baseView == nil {
				continue
			}
			baseView.updateViewFromXML(xmlView)
			Registry.rawInheritedViews[i] = nil
		}
		// Second step: we create all named extensions we can
		for i, xmlView := range Registry.rawInheritedViews {
			if xmlView == nil {
				continue
			}
			if xmlView.ID == "" {
				continue
			}
			baseView := Registry.GetByID(xmlView.InheritID)
			if baseView == nil {
				continue
			}
			model := baseView.Model
			if xmlView.Model != "" {
				model = xmlView.Model
			}
			priority := baseView.Priority
			if xmlView.Priority != 0 {
				priority = xmlView.Priority
			}
			newView := View{
				ID:          xmlView.ID,
				Priority:    priority,
				Model:       model,
				SubViews:    make(map[string]SubViews),
				arch:        baseView.arch,
				Name:        baseView.Name,
				Type:        baseView.Type,
				arches:      make(map[string]*etree.Element),
				FieldParent: baseView.FieldParent,
			}
			newView.updateViewFromXML(xmlView)
			Registry.Add(&newView)
			Registry.rawInheritedViews[i] = nil
		}
	}
	// Post-process all views
	for _, v := range Registry.views {
		v.postProcess()
	}
}

func init() {
	log = logging.GetLogger("views")
	Registry = NewCollection()
}
