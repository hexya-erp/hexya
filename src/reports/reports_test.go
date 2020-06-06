// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package reports_test

import (
	"testing"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/reports"
	. "github.com/smartystreets/goconvey/convey"
)

func TestReports(t *testing.T) {
	Convey("Creating models", t, func() {
		user := models.NewModel("User")
		user.AddFields(map[string]models.FieldDefinition{
			"UserName": fields.Char{},
			"Age":      fields.Integer{},
		})
		models.BootStrap()
	})
	Convey("Testing TextReport", t, func() {
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
		Convey("Registering a text report", func() {
			So(func() { reports.Register(&report) }, ShouldNotPanic)
			So(func() { reports.Register(&report2) }, ShouldNotPanic)
		})
		Convey("Registering twice should panic", func() {
			So(func() { reports.Register(&report) }, ShouldPanic)
		})
		Convey("Replacing a text report", func() {
			rep := report
			rep.Name = "New Sample Report"
			So(func() { reports.Registry.Replace(&rep) }, ShouldNotPanic)
		})
		Convey("Replacing a report that doesn't exist should fail", func() {
			rep := report
			rep.Id = "sample_report_2"
			So(func() { reports.Registry.Replace(&rep) }, ShouldPanic)
		})
		Convey("Bootstrapping reports", func() {
			rep2 := reports.Registry.MustGet("sample_html").(*reports.TextReport)
			rep2.Modeler = nil
			So(reports.BootStrap, ShouldPanic)
			rep2.Modeler = models.Registry.MustGet("User")
			rep2.Filename = ""
			So(reports.BootStrap, ShouldPanic)
			rep2.Filename = "sample.html"
			So(reports.BootStrap, ShouldNotPanic)
		})
		Convey("Bootstrapping twice should panic", func() {
			So(reports.BootStrap, ShouldPanic)
		})
		Convey("Registering a report after bootstrap should panic", func() {
			So(func() { reports.Registry.Replace(&report) }, ShouldPanic)
		})
		Convey("Fetching a report from registry", func() {
			rep, ok := reports.Registry.Get("sample_report")
			So(ok, ShouldBeTrue)
			So(rep, ShouldNotBeNil)
			So(func() { reports.Registry.MustGet("sample_report") }, ShouldNotPanic)
			So(func() { reports.Registry.MustGet("sample_report_2") }, ShouldPanic)
			So(rep.String(), ShouldEqual, "New Sample Report")
		})
		Convey("Rendering text report", func() {
			rep := reports.Registry.MustGet("sample_report")
			doc, err := rep.Render(1, nil)
			So(err, ShouldBeNil)
			So(doc.MimeType, ShouldEqual, "text/plain")
			So(doc.Filename, ShouldEqual, "sample.txt")
			So(string(doc.Content), ShouldEqual, `
Welcome to my sample report
===========================
Name: Jane Smith
Age: 24
`)
		})
		Convey("Rendering html report", func() {
			rep := reports.Registry.MustGet("sample_html")
			doc, err := rep.Render(1, nil)
			So(err, ShouldBeNil)
			So(doc.MimeType, ShouldEqual, "text/html")
			So(doc.Filename, ShouldEqual, "sample.html")
			So(string(doc.Content), ShouldEqual, `
Welcome to my sample report
===========================
Name: Jane Smith
Age: 24
`)
		})
		Convey("Testing loading error cases", func() {
			rep := report
			rep.Filename = ""
			err := rep.Init()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "incomplete TextReport: Filename is not set")
			rep.Filename = report.Filename
			rep.DataFunc = nil
			err = rep.Init()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "incomplete TextReport: DataFunc is not set")
			rep.DataFunc = report.DataFunc
			rep.Template = `BEGIN {{ .Name } END`
			err = rep.Init()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "error while loading TextReport template: template: :1: unexpected \"}\" in operand")
			rep.Template = report.Template
			rep.MimeType = "application/json"
			err = rep.Init()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unsupported mime type 'application/json' for TextReport")
		})
		Convey("Testing rendering error cases", func() {
			rep := report
			rep.Template = "{{ eq .Unknown \"something\" }}"
			_, err := rep.Render(1, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldBeIn, []string{`template: :1:3: executing "" at <eq .Unknown "something">: error calling eq: invalid type for comparison`, `template: :1:3: executing "" at <eq .Unknown "something">: error calling eq: incompatible types for comparison`})
			rep = report2
			rep.Template = "{{ eq .Unknown \"something\" }}"
			_, err = rep.Render(1, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldBeIn, []string{`template: :1:3: executing "" at <eq .Unknown "something">: error calling eq: invalid type for comparison`, `template: :1:3: executing "" at <eq .Unknown "something">: error calling eq: incompatible types for comparison`})
		})
		Convey("Calling GetAction", func() {
			act := reports.GetAction("sample_html", 3, reports.Data{"foo": "bar"})
			So(act.Type, ShouldEqual, actions.ActionReport)
			So(act.Name, ShouldEqual, "Sample Report")
			So(act.Model, ShouldEqual, "User")
			So(act.Data, ShouldResemble, map[string]interface{}{"foo": "bar"})
			So(act.ReportName, ShouldEqual, "sample_html")
			So(act.ReportFile, ShouldEqual, "sample_html")
			So(act.ReportType, ShouldEqual, "TextReport")
			So(act.Context.GetIntegerSlice("active_ids"), ShouldHaveLength, 1)
			So(act.Context.GetIntegerSlice("active_ids"), ShouldContain, int64(3))
		})
	})
}
