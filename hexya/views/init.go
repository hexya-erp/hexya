// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package views

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
)

var log *logging.Logger

//BootStrap makes the necessary updates to view definitions. In particular:
//- sets the type of the view from the arch root.
//- populates the fields map from the views arch.
func BootStrap() {
	for _, v := range Registry.views {
		archElem := xmlutils.XMLToElement(v.Arch)

		// Set view type
		v.Type = ViewType(archElem.Tag)

		// Populate fields map
		fieldElems := archElem.FindElements("//field")
		for _, f := range fieldElems {
			v.Fields = append(v.Fields, models.FieldName(f.SelectAttr("name").Value))
		}
	}
}

func init() {
	log = logging.GetLogger("views")
	Registry = NewCollection()
}
