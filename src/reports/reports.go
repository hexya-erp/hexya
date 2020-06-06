// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package reports

import (
	"errors"
	"fmt"
	"sync"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/types"
)

// Registry is the report collection of the application
var Registry *Collection

// A Collection is a collection of reports
type Collection struct {
	sync.RWMutex
	reports map[string]Report
}

// NewCollection returns a pointer to a new
// Collection instance
func NewCollection() *Collection {
	res := Collection{
		reports: make(map[string]Report),
	}
	return &res
}

// Add adds the given report to our Collection
func (rr *Collection) Add(r Report) {
	if err := rr.add(r, false); err != nil {
		log.Panic("error while registering a report", "name", r, "ID", r.ID(), "error", err.Error())
	}
}

// Replace the given report in our Collection
func (rr *Collection) Replace(r Report) {
	if err := rr.add(r, true); err != nil {
		log.Panic("error while replacing a report", "name", r, "ID", r.ID(), "error", err.Error())
	}
}

// add or replace a report in the registry
func (rr *Collection) add(r Report, replace bool) error {
	if bootstrapped {
		return errors.New("you cannot add a report after bootstrap")
	}
	rr.Lock()
	defer rr.Unlock()
	_, ok := rr.reports[r.ID()]
	switch {
	case ok && !replace:
		return errors.New("trying to register a report that already exists")
	case !ok && replace:
		return errors.New("trying to replace a report that doesn't exist")
	}
	rr.reports[r.ID()] = r
	return nil
}

// Get returns the Report with the given id
func (rr *Collection) Get(id string) (Report, bool) {
	r, ok := rr.reports[id]
	return r, ok
}

// MustGet returns the Report with the given id.
// It panics if the record doesn't exist.
func (rr *Collection) MustGet(id string) Report {
	r, ok := rr.Get(id)
	if !ok {
		log.Panic("Report doesn't exist", "name", r, "ID", r.ID())
	}
	return r
}

// Data holds the item specific data used to render a report.
type Data map[string]interface{}

// A Document is the result of rendering a report
type Document struct {
	// Content is the binary content of the rendered report
	Content []byte
	// MimeType of the content
	MimeType string
	// Filename for the document
	Filename string
}

// A Report is bound to a model and can render a report given an ID of the model.
//
// A report typically contains:
// - A template
// - A rendering mechanism to inject data into the template and provide the output to the user.
type Report interface {
	fmt.Stringer
	// ID returns the unique identifying code of this report
	ID() string
	// Model that this report is bound to.
	Model() models.Modeler
	// Render this report.
	Render(id int64, additionalData Data) (*Document, error)
	// Init initializes the report. Init is called at bootstrap.
	Init() error
	// Type returns the type of the record
	Type() string
}

// Register add the given report to the registry
func Register(r Report) {
	Registry.Add(r)
}

// GetAction returns an action for the report with the given reportID for the record
// with the given id and optional additionalData.
func GetAction(reportID string, id int64, additionalData Data) *actions.Action {
	report := Registry.MustGet(reportID)
	return &actions.Action{
		Type:       actions.ActionReport,
		Name:       report.String(),
		Model:      report.Model().Underlying().Name(),
		Context:    types.NewContext().WithKey("active_ids", []int64{id}),
		Flags:      nil,
		Data:       additionalData,
		ReportName: reportID,
		ReportType: report.Type(),
		ReportFile: reportID,
	}
}
