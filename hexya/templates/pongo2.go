// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package templates

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/flosch/pongo2"
)

// Abs calculates the path to a given template. Whenever a path must be resolved
// due to an import from another template, the base equals the parent template's path.
func (tc *Collection) Abs(base, name string) string {
	// templates paths have a "LANG/template_name" pattern so we stick to the same virtual directory
	return path.Join(path.Dir(base), name)
}

// Get returns an io.Reader where the template's content can be read from.
func (tc *Collection) Get(pth string) (io.Reader, error) {
	lang := path.Dir(pth)
	name := path.Base(pth)
	template := tc.GetByID(name)
	if template == nil {
		return nil, fmt.Errorf("unknown template %s", pth)
	}
	return bytes.NewBuffer(template.Content(lang)), nil
}

var _ pongo2.TemplateLoader = new(Collection)
