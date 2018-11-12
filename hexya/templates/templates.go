// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package templates

import (
	"encoding/xml"
	"strconv"
	"sync"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/tools/hweb"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
)

// A Collection of templates
type Collection struct {
	sync.RWMutex
	templates             map[string]*Template
	rawInheritedTemplates []*TemplateXML
}

// newCollection returns a pointer to a new
// Collection instance
func newCollection() *Collection {
	res := Collection{
		templates: make(map[string]*Template),
	}
	return &res
}

// Add the given template to the Collection
func (tc *Collection) Add(t *Template) {
	tc.Lock()
	defer tc.Unlock()
	tc.templates[t.ID] = t
}

// GetByID returns the Template with the given id
func (tc *Collection) GetByID(id string) *Template {
	return tc.templates[id]
}

// LoadFromEtree loads the given template given as Element
// into this collection.
func (tc *Collection) LoadFromEtree(element *etree.Element) {
	xmlBytes, err := xmlutils.ElementToXMLNoIndent(element)
	if err != nil {
		log.Panic("Unable to convert element to XML", "error", err)
	}
	var templateXML TemplateXML
	if err = xml.Unmarshal(xmlBytes, &templateXML); err != nil {
		log.Panic("Unable to unmarshal element", "error", err, "bytes", string(xmlBytes))
	}
	if templateXML.InheritID != "" {
		// Update an existing template.
		// Put in raw inherited templates for now, as the base template may not exist yet.
		tc.rawInheritedTemplates = append(tc.rawInheritedTemplates, &templateXML)
		return
	}
	// Create a new template
	tc.createNewTemplateFromXML(&templateXML)
}

// createNewViewFromXML creates and register a new template with the given XML
func (tc *Collection) createNewTemplateFromXML(templateXML *TemplateXML) {
	priority := uint8(16)
	if templateXML.Priority != 0 {
		priority = templateXML.Priority
	}
	optional := templateXML.Optional != ""
	optionalDefault := templateXML.Optional == "enabled"
	page, _ := strconv.ParseBool(templateXML.Page)
	tmpl := Template{
		ID:              templateXML.ID,
		Priority:        priority,
		Page:            page,
		Optional:        optional,
		OptionalDefault: optionalDefault,
		hWebContent:     templateXML.Content,
		p2Contents:      make(map[string][]byte),
	}
	tc.Add(&tmpl)
}

// A Template holds information of a HWeb template
type Template struct {
	*hweb.Template
	ID              string
	Priority        uint8
	Page            bool
	Optional        bool
	OptionalDefault bool
	hWebContent     []byte
	p2Content       []byte
	p2Contents      map[string][]byte
}

// Content returns the template data for the given language.
// Call with empty string to get the default language's data.
func (t *Template) Content(lang string) []byte {
	res, ok := t.p2Contents[lang]
	if !ok || lang == "" {
		res = t.p2Content
	}
	return res
}

// updateFromXML updates this template with the given XML
// templateXML must have an InheritID
func (t *Template) updateFromXML(templateXML *TemplateXML) {
	specDoc, err := xmlutils.XMLToDocument(string(templateXML.Content))
	if err != nil {
		log.Panic("Unable to read inheritance specs", "error", err, "arch", string(templateXML.Content))
	}
	content, err := xmlutils.XMLToElement(string(t.hWebContent))
	if err != nil {
		log.Panic("Error while reading base template content", "error", err, "template", t.ID, "content", string(t.hWebContent))
	}
	newContent, err := xmlutils.ApplyExtensions(content, specDoc)
	if err != nil {
		log.Panic("Error while applying template extension specs", "error", err, "specTmpl", templateXML.ID,
			"specs", string(templateXML.Content), "template", t.ID, "content", string(t.hWebContent))
	}
	ncs, err := xmlutils.ElementToXMLNoIndent(newContent)
	if err != nil {
		log.Panic("Error while converting back to XML", "error", err, "content", newContent,
			"specTmpl", templateXML.ID, "template", t.ID)
	}
	t.hWebContent = ncs
}

// postProcess the template by
//
// - populating the p2Content (Pongo2)
// - populating the p2Contents map with all the translations
func (t *Template) postProcess() {
	p2Content, err := hweb.ToPongo(t.hWebContent)
	if err != nil {
		log.Panic("Error while transpiling to Pongo2", "error", err, "template", t.ID)
	}
	t.p2Content = p2Content
	for _, lang := range i18n.Langs {
		p2Content, err = hweb.ToPongo([]byte(i18n.TranslateResourceItem(lang, t.ID, string(t.hWebContent))))
		if err != nil {
			log.Panic("Error while transpiling translation to Pongo2", "error", err, "template", t.ID)
		}
		t.p2Contents[lang] = p2Content
	}
}

// TemplateXML is used to unmarshal the XML definition of a template
type TemplateXML struct {
	ID        string `xml:"id,attr"`
	InheritID string `xml:"inherit_id,attr"`
	Content   []byte `xml:",innerxml"`
	Priority  uint8  `xml:"priority,attr"`
	Page      string `xml:"page,attr"`
	Optional  string `xml:"optional,attr"`
}

// LoadFromEtree reads the view given etree.Element, creates or updates the template
// and adds it to the template registry if it not already.
func LoadFromEtree(element *etree.Element) {
	Registry.collection.LoadFromEtree(element)
}
