// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package reports

import "github.com/hexya-erp/hexya/src/tools/logging"

var (
	log          logging.Logger
	bootstrapped bool
)

// BootStrap check and initialize the reports
func BootStrap() {
	if bootstrapped {
		log.Panic("Reports are already bootstrapped.")
	}
	for _, report := range Registry.reports {
		if report.Model() == nil {
			log.Panic("Missing model in report", "name", report, "ID", report.ID())
		}
		if err := report.Init(); err != nil {
			log.Panic("Error while initializing a report", "name", report, "ID", report.ID(), "error", err)
		}
	}
	bootstrapped = true
}

func init() {
	log = logging.GetLogger("reports")
	Registry = NewCollection()
}
