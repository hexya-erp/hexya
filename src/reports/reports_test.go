// Copyright 2020 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package reports_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/reports"
)

func TestReports(t *testing.T) {
	t.Run("Creating models", func(t *testing.T) {
		user := models.NewModel("User")
		user.AddFields(map[string]models.FieldDefinition{
			"UserName": fields.Char{},
			"Age":      fields.Integer{},
		})
		models.BootStrap()
	})
	t.Run("Testing TextReport", func(t *testing.T) {
		report := reports.TextReport{
			Id:       "sample_report",
			Name:     "Sample Report",
			Modeler:  models.Registry.MustGet("User"),
			MimeType: "text/plain",
			Filename: "sample.txt",
			Template: `
Welcome to my sample report
===========================
Name: {{ .Name }}
Age: {{ .Age }}
`,
			DataFunc: func(id int64, data reports.Data) reports.Data {
				return reports.Data{
					"Name": "Jane Smith",
					"Age":  24,
				}
			},
		}
		report2 := report
		report2.Id = "sample_html"
		report2.Filename = "sample.html"
		report2.MimeType = "text/html"
		t.Run("Registering a text report", func(t *testing.T) {
			assert.NotPanics(t, func() { reports.Register(&report) })
			assert.NotPanics(t, func() { reports.Register(&report2) })
		})
		t.Run("Registering twice should panic", func(t *testing.T) {
			assert.Panics(t, func() { reports.Register(&report) })
		})
		t.Run("Replacing a text report", func(t *testing.T) {
			rep := report
			rep.Name = "New Sample Report"
			assert.NotPanics(t, func() { reports.Registry.Replace(&rep) })
		})
		t.Run("Replacing a report that doesn't exist should fail", func(t *testing.T) {
			rep := report
			rep.Id = "sample_report_2"
			assert.Panics(t, func() { reports.Registry.Replace(&rep) })
		})
		t.Run("Bootstrapping reports", func(t *testing.T) {
			rep2 := reports.Registry.MustGet("sample_html").(*reports.TextReport)
			rep2.Modeler = nil
			assert.Panics(t, reports.BootStrap)
			rep2.Modeler = models.Registry.MustGet("User")
			rep2.Filename = ""
			assert.Panics(t, reports.BootStrap)
			rep2.Filename = "sample.html"
			assert.NotPanics(t, reports.BootStrap)
		})
		t.Run("Bootstrapping twice should panic", func(t *testing.T) {
			assert.Panics(t, reports.BootStrap)
		})
		t.Run("Registering a report after bootstrap should panic", func(t *testing.T) {
			assert.Panics(t, func() { reports.Registry.Replace(&report) })
		})
		t.Run("Fetching a report from registry", func(t *testing.T) {
			rep, ok := reports.Registry.Get("sample_report")
			assert.True(t, ok)
			assert.NotNil(t, rep)
			assert.NotPanics(t, func() { reports.Registry.MustGet("sample_report") })
			assert.Panics(t, func() { reports.Registry.MustGet("sample_report_2") })
			assert.EqualValues(t, rep.String(), "New Sample Report")
		})
		t.Run("Rendering text report", func(t *testing.T) {
			rep := reports.Registry.MustGet("sample_report")
			doc, err := rep.Render(1, nil)
			assert.Nil(t, err)
			assert.EqualValues(t, doc.MimeType, "text/plain")
			assert.EqualValues(t, doc.Filename, "sample.txt")
			assert.EqualValues(t, string(doc.Content), `
Welcome to my sample report
===========================
Name: Jane Smith
Age: 24
`)
		})
		t.Run("Rendering html report", func(t *testing.T) {
			rep := reports.Registry.MustGet("sample_html")
			doc, err := rep.Render(1, nil)
			assert.Nil(t, err)
			assert.EqualValues(t, doc.MimeType, "text/html")
			assert.EqualValues(t, doc.Filename, "sample.html")
			assert.EqualValues(t, string(doc.Content), `
Welcome to my sample report
===========================
Name: Jane Smith
Age: 24
`)
		})
		t.Run("Testing loading error cases", func(t *testing.T) {
			rep := report
			rep.Filename = ""
			err := rep.Init()
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "incomplete TextReport: Filename is not set")
			rep.Filename = report.Filename
			rep.DataFunc = nil
			err = rep.Init()
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "incomplete TextReport: DataFunc is not set")
			rep.DataFunc = report.DataFunc
			rep.Template = `BEGIN {{ .Name } END`
			err = rep.Init()
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "error while loading TextReport template: template: :1: unexpected \"}\" in operand")
			rep.Template = report.Template
			rep.MimeType = "application/json"
			err = rep.Init()
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "unsupported mime type 'application/json' for TextReport")
		})
		t.Run("Testing rendering error cases", func(t *testing.T) {
			rep := report
			rep.Template = "{{ eq .Age \"something\" }}"
			_, err := rep.Render(1, nil)
			assert.NotNil(t, err)
			assert.Contains(t, []string{
				`template: :1:3: executing "" at <eq .Age "something">: error calling eq: invalid type for comparison`,
				`template: :1:3: executing "" at <eq .Age "something">: error calling eq: incompatible types for comparison: int and string`}, err.Error())
			rep = report2
			rep.Template = "{{ eq .Age \"something\" }}"
			_, err = rep.Render(1, nil)
			assert.NotNil(t, err)
			assert.Contains(t, []string{
				`template: :1:3: executing "" at <eq .Age "something">: error calling eq: invalid type for comparison`,
				`template: :1:3: executing "" at <eq .Age "something">: error calling eq: incompatible types for comparison: int and string`}, err.Error())
		})
		t.Run("Calling GetAction", func(t *testing.T) {
			act := reports.GetAction("sample_html", 3, reports.Data{"foo": "bar"})
			assert.EqualValues(t, act.Type, actions.ActionReport)
			assert.EqualValues(t, act.Name, "Sample Report")
			assert.EqualValues(t, act.Model, "User")
			assert.Equal(t, act.Data, map[string]any{"foo": "bar"})
			assert.EqualValues(t, act.ReportName, "sample_html")
			assert.EqualValues(t, act.ReportFile, "sample_html")
			assert.EqualValues(t, act.ReportType, "TextReport")
			assert.Len(t, act.Context.GetIntegerSlice("active_ids"), 1)
			assert.Contains(t, act.Context.GetIntegerSlice("active_ids"), int64(3))
		})
	})
}
