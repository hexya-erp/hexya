// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package templates

import (
	"net/http"

	"github.com/gin-gonic/gin/render"
	"github.com/hexya-erp/hexya/src/tools/hweb"
)

// Registry is the templates set of the application
var Registry *TemplateSet

// A TemplateSet is a set of pongo2 templates
type TemplateSet struct {
	*hweb.TemplateSet
	collection *Collection
}

// NewTemplateSet returns a pointer to a new empty TemplateSet
func NewTemplateSet() *TemplateSet {
	coll := newCollection()
	res := TemplateSet{
		TemplateSet: hweb.NewSet("-", coll),
		collection:  coll,
	}
	return &res
}

// Instance returns the TemplateRenderer given by its name with the given data
func (ts *TemplateSet) Instance(name string, data interface{}) render.Render {
	template := hweb.Must(ts.FromCache(name))
	return TemplateRenderer{
		Template: template,
		Data:     data.(hweb.Context),
	}
}

var _ render.HTMLRender = new(TemplateSet)

// A TemplateRenderer can render a template with the given data
type TemplateRenderer struct {
	Template *hweb.Template
	Data     hweb.Context
}

// Render this TemplateRenderer to the given ResponseWriter
func (tr TemplateRenderer) Render(w http.ResponseWriter) error {
	tr.WriteContentType(w)
	err := tr.Template.ExecuteWriter(tr.Data, w)
	return err
}

// WriteContentType of this TemplateRenderer
func (tr TemplateRenderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html"}
	}
}

var _ render.Render = TemplateRenderer{}
