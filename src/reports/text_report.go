// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package reports

import (
	"bytes"
	"errors"
	"fmt"
	html "html/template"
	text "text/template"

	"github.com/hexya-erp/hexya/src/models"
)

// A TextReport is a simple text or html report.
//
// It uses the html/template package if the report"s mime type is text/html
// and text/template if the mimetype is text/plain.
type TextReport struct {
	Id       string
	Name     string
	Modeler  models.Modeler
	MimeType string
	Filename string
	Template string
	DataFunc func(int64, Data) Data
}

// Render this report. Returns the rendered report, the report mimetype,the report filename and an error.
func (r *TextReport) Render(id int64, additionalData Data) (*Document, error) {
	var rawText bool
	if r.MimeType == "text/plain" {
		rawText = true
	}
	data := r.DataFunc(id, additionalData)
	var res bytes.Buffer
	switch rawText {
	case true:
		template := text.Must(text.New("").Parse(r.Template))
		err := template.Execute(&res, data)
		if err != nil {
			return nil, err
		}
	case false:
		template := html.Must(html.New("").Parse(r.Template))
		err := template.Execute(&res, data)
		if err != nil {
			return nil, err
		}
	}
	return &Document{
		Content:  res.Bytes(),
		MimeType: r.MimeType,
		Filename: r.Filename,
	}, nil
}

// Init initializes the report. Init is called at bootstrap.
func (r *TextReport) Init() error {
	if r.MimeType != "text/html" && r.MimeType != "text/plain" {
		return fmt.Errorf("unsupported mime type '%s' for TextReport", r.MimeType)
	}
	if r.DataFunc == nil {
		return errors.New("incomplete TextReport: DataFunc is not set")
	}
	if r.Filename == "" {
		return errors.New("incomplete TextReport: Filename is not set")
	}
	_, err := text.New("").Parse(r.Template)
	if err != nil {
		return fmt.Errorf("error while loading TextReport template: %s", err)
	}
	return nil
}

func (r *TextReport) String() string {
	return r.Name
}

// ID returns the unique identifying code of this report
func (r *TextReport) ID() string {
	return r.Id
}

// Model returns the name of the model that this report is bound to.
func (r *TextReport) Model() models.Modeler {
	return r.Modeler
}

// Type of the report: TextReport
func (r *TextReport) Type() string {
	return "TextReport"
}
