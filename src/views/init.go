// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package views

import (
	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/tools/logging"
	"github.com/hexya-erp/hexya/src/tools/xmlutils"
)

const maxInheritanceDepth = 100

var log logging.Logger

// BootStrap makes the necessary updates to view definitions. In particular:
// - sets the type of the view from the arch root.
// - extracts embedded views
// - populates the fields map from the views arch.
func BootStrap() {
	if !models.BootStrapped() {
		log.Panic("Models must be bootstrapped before bootstrapping views")
	}
	loadModelViews()
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
				arches:      make(map[string]*etree.Document),
				FieldParent: baseView.FieldParent,
			}
			newView.updateViewFromXML(xmlView)
			Registry.Add(&newView)
			Registry.rawInheritedViews[i] = nil
		}
	}
	// Post-process all views
	for _, v := range Registry.views {
		log.Debug("Postprocessing view", "viewID", v.ID, "model", v.Model, "Type", v.Type)
		v.postProcess()
	}
}

// loadModelViews load views that have been defined in the models package during bootstrap
func loadModelViews() {
	for _, views := range models.Views {
		for _, view := range views {
			elt, err := xmlutils.XMLToElement(view)
			if err != nil {
				log.Panic("error while loading view", "error", err, "view", view)
			}
			LoadFromEtree(elt)
		}
	}
}

func init() {
	log = logging.GetLogger("views")
	Registry = NewCollection()
}
