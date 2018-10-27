// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package templates

import (
	"strconv"

	"github.com/hexya-erp/hexya/hexya/tools/logging"
)

const maxInheritanceDepth = 100

var log logging.Logger

// BootStrap actions.
// This function must be called prior to any access to the templates Registry.
func BootStrap() {
	// Inherit/Extend templates
	for loop := 0; loop < maxInheritanceDepth; loop++ {
		// First step: we extend all we can with pure extension templates (no ID)
		for i, xmlTmpl := range Registry.collection.rawInheritedTemplates {
			if xmlTmpl == nil {
				continue
			}
			if xmlTmpl.ID != "" {
				continue
			}
			baseTmpl := Registry.collection.GetByID(xmlTmpl.InheritID)
			if baseTmpl == nil {
				continue
			}
			baseTmpl.updateFromXML(xmlTmpl)
			Registry.collection.rawInheritedTemplates[i] = nil
		}
		// Second step: we create all named extensions we can
		for i, xmlTmpl := range Registry.collection.rawInheritedTemplates {
			if xmlTmpl == nil {
				continue
			}
			if xmlTmpl.ID == "" {
				continue
			}
			baseTmpl := Registry.collection.GetByID(xmlTmpl.InheritID)
			if baseTmpl == nil {
				continue
			}
			priority := baseTmpl.Priority
			if xmlTmpl.Priority != 0 {
				priority = xmlTmpl.Priority
			}
			page := baseTmpl.Page
			if newPage, err := strconv.ParseBool(xmlTmpl.Page); err == nil {
				page = newPage
			}
			optional := baseTmpl.Optional
			if xmlTmpl.Optional != "" {
				optional = true
			}
			optionalDefault := baseTmpl.OptionalDefault
			if xmlTmpl.Optional != "" {
				optionalDefault = xmlTmpl.Optional == "enabled"
			}
			newTmpl := Template{
				ID:              xmlTmpl.ID,
				Priority:        priority,
				Page:            page,
				Optional:        optional,
				OptionalDefault: optionalDefault,
				hWebContent:     baseTmpl.hWebContent,
				p2Contents:      make(map[string][]byte),
			}
			newTmpl.updateFromXML(xmlTmpl)
			Registry.collection.Add(&newTmpl)
			Registry.collection.rawInheritedTemplates[i] = nil
		}
	}
	// Post-process all templates
	for _, t := range Registry.collection.templates {
		log.Debug("Postprocessing template", "ID", t.ID)
		t.postProcess()
	}
}

func init() {
	log = logging.GetLogger("templates")
	Registry = NewTemplateSet()
}
