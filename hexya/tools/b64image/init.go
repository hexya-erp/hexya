// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package b64image

import "github.com/hexya-erp/hexya/hexya/tools/logging"

var log *logging.Logger

func init() {
	log = logging.GetLogger("b64image")
}
