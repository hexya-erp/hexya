// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package views

import "github.com/hexya-erp/hexya/hexya/tools/logging"

var log *logging.Logger

//BootStrap makes the necessary updates to view definitions. In particular:
//- sets the type of the view from the arch root.
//- extracts embedded views
//- populates the fields map from the views arch.
func BootStrap() {
	for _, v := range Registry.views {
		v.postProcess()
	}
}

func init() {
	log = logging.GetLogger("views")
	Registry = NewCollection()
}
