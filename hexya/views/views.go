// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package views

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
)

// A ViewType defines the type of a view
type ViewType string

// View types
const (
	VIEW_TYPE_TREE     ViewType = "tree"
	VIEW_TYPE_LIST     ViewType = "list"
	VIEW_TYPE_FORM     ViewType = "form"
	VIEW_TYPE_GRAPH    ViewType = "graph"
	VIEW_TYPE_CALENDAR ViewType = "calendar"
	VIEW_TYPE_DIAGRAM  ViewType = "diagram"
	VIEW_TYPE_GANTT    ViewType = "gantt"
	VIEW_TYPE_KANBAN   ViewType = "kanban"
	VIEW_TYPE_SEARCH   ViewType = "search"
	VIEW_TYPE_QWEB     ViewType = "qweb"
)

// translatableAttributes is the list of XML attribute names the
// value of which needs to be translated.
var translatableAttributes = []string{"string", "help", "sum", "confirm", "placeholder"}

// Registry is the views collection of the application
var Registry *Collection

// MakeViewRef creates a ViewRef from a view id
func MakeViewRef(id string) ViewRef {
	view := Registry.GetByID(id)
	if view == nil {
		return ViewRef{}
	}
	return ViewRef{id, view.Name}
}

// ViewRef is an array of two strings representing a view:
// - The first one is the ID of the view
// - The second one is the name of the view
type ViewRef [2]string

// MarshalJSON is the JSON marshalling method of ViewRef.
// It marshals empty ViewRef into null instead of ["", ""].
func (vr ViewRef) MarshalJSON() ([]byte, error) {
	if vr[0] == "" {
		return json.Marshal(nil)
	}
	return json.Marshal([2]string{vr[0], vr[1]})
}

// UnmarshalJSON is the JSON unmarshalling method of ViewRef.
// It unmarshals null into an empty ViewRef.
func (vr *ViewRef) UnmarshalJSON(data []byte) error {
	var dst interface{}
	if err := json.Unmarshal(data, &dst); err == nil && dst == nil {
		vr = &ViewRef{"", ""}
		return nil
	}
	var dstArray [2]string
	if err := json.Unmarshal(data, &dstArray); err != nil {
		return err
	}
	*vr = ViewRef(dstArray)
	return nil
}

// UnmarshalXMLAttr is the XML unmarshalling method of ViewRef.
// It unmarshals null into an empty ViewRef.
func (vr *ViewRef) UnmarshalXMLAttr(attr xml.Attr) error {
	*vr = MakeViewRef(attr.Value)
	return nil
}

// Value extracts ID of our ViewRef for storing in the database.
func (vr ViewRef) Value() (driver.Value, error) {
	return driver.Value(vr[0]), nil
}

// Scan fetches the name of our view from the ID
// stored in the database to fill the ViewRef.
func (vr *ViewRef) Scan(src interface{}) error {
	var source string
	switch s := src.(type) {
	case string:
		source = s
	case []byte:
		source = string(s)
	default:
		return fmt.Errorf("Invalid type for ViewRef: %T", src)
	}
	*vr = MakeViewRef(source)
	return nil
}

// ID returns the ID of the current view reference
func (vr ViewRef) ID() string {
	return vr[0]
}

// Name returns the name of the current view reference
func (vr ViewRef) Name() string {
	return vr[1]
}

// IsNull returns true if this ViewRef references no view
func (vr ViewRef) IsNull() bool {
	return vr[0] == "" && vr[1] == ""
}

var _ driver.Valuer = &ViewRef{}
var _ sql.Scanner = &ViewRef{}
var _ json.Marshaler = &ViewRef{}
var _ json.Unmarshaler = &ViewRef{}
var _ xml.UnmarshalerAttr = &ViewRef{}

// ViewTuple is an array of two strings representing a view:
// - The first one is the ID of the view
// - The second one is the view type corresponding to the view ID
type ViewTuple struct {
	ID   string   `xml:"id,attr"`
	Type ViewType `xml:"type,attr"`
}

// MarshalJSON is the JSON marshalling method of ViewTuple.
// It marshals ViewTuple into a list [id, type].
func (vt ViewTuple) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]string{vt.ID, string(vt.Type)})
}

// UnmarshalJSON method for ViewTuple
func (vt *ViewTuple) UnmarshalJSON(data []byte) error {
	var src interface{}
	err := json.Unmarshal(data, &src)
	if err != nil {
		return err
	}
	switch s := src.(type) {
	case []interface{}:
		vID, _ := s[0].(string)
		vt.ID = vID
		vt.Type = ViewType(s[1].(string))
	default:
		return errors.New("Unexpected type in ViewTuple Unmarshal")
	}
	return nil
}

var _ json.Marshaler = ViewTuple{}
var _ json.Unmarshaler = &ViewTuple{}

// A Collection is a view collection
type Collection struct {
	sync.RWMutex
	views        map[string]*View
	orderedViews map[string][]*View
}

// NewCollection returns a pointer to a new
// Collection instance
func NewCollection() *Collection {
	res := Collection{
		views:        make(map[string]*View),
		orderedViews: make(map[string][]*View),
	}
	return &res
}

// Add adds the given view to our Collection
func (vc *Collection) Add(v *View) {
	vc.Lock()
	var index int8
	for i, view := range vc.orderedViews[v.Model] {
		index = int8(i)
		if view.Priority > v.Priority {
			break
		}
	}
	defer vc.Unlock()
	vc.views[v.ID] = v
	if index == int8(len(vc.orderedViews)-1) {
		vc.orderedViews[v.Model] = append(vc.orderedViews[v.Model], v)
		return
	}
	endElems := make([]*View, len(vc.orderedViews[v.Model][index:]))
	copy(endElems, vc.orderedViews[v.Model][index:])
	vc.orderedViews[v.Model] = append(append(vc.orderedViews[v.Model][:index], v), endElems...)
}

// GetByID returns the View with the given id
func (vc *Collection) GetByID(id string) *View {
	return vc.views[id]
}

// GetAll returns a list of all views of this Collection.
// Views are returned in an arbitrary order
func (vc *Collection) GetAll() []*View {
	res := make([]*View, len(vc.views))
	var i int
	for _, view := range vc.views {
		res[i] = view
		i++
	}
	return res
}

// GetFirstViewForModel returns the first view of type viewType for the given model
func (vc *Collection) GetFirstViewForModel(model string, viewType ViewType) *View {
	for _, view := range vc.orderedViews[model] {
		if view.Type == viewType {
			return view
		}
	}
	defaultView := vc.defaultViewForModel(model, viewType)
	if defaultView == nil {
		log.Panic("No view of this type in model", "type", viewType, "model", model)
	}
	return defaultView
}

// defaultViewForModel returns a default view for the given model and type
func (vc *Collection) defaultViewForModel(model string, viewType ViewType) *View {
	view := View{
		Model:  model,
		Type:   viewType,
		Fields: []models.FieldName{models.FieldName("Name")},
		arch:   fmt.Sprintf(`<%s><field name="Name"/></%s>`, viewType, viewType),
		arches: make(map[string]string),
	}
	view.translateArch()
	return &view
}

// GetAllViewsForModel returns a list with all views for the given model
func (vc *Collection) GetAllViewsForModel(model string) []*View {
	var res []*View
	for _, view := range vc.views {
		if view.Model == model {
			res = append(res, view)
		}
	}
	return res
}

// LoadFromEtree loads the given view given as Element
// into this collection.
func (vc *Collection) LoadFromEtree(element *etree.Element) {
	xmlBytes := []byte(xmlutils.ElementToXML(element))
	var viewXML ViewXML
	if err := xml.Unmarshal(xmlBytes, &viewXML); err != nil {
		log.Panic("Unable to unmarshal element", "error", err, "bytes", string(xmlBytes))
	}
	if viewXML.InheritID != "" {
		// Update an existing view
		vc.updateExistingViewFromXML(viewXML)
	} else {
		// Create a new view
		vc.createNewViewFromXML(viewXML)
	}
}

// createNewViewFromXML creates and register a new view with the given XML
func (vc *Collection) createNewViewFromXML(viewXML ViewXML) {
	priority := uint8(16)
	if viewXML.Priority != 0 {
		priority = viewXML.Priority
	}
	name := strings.Replace(viewXML.ID, "_", ".", -1)
	if viewXML.Name != "" {
		name = viewXML.Name
	}
	// We check/standardize arch by unmarshalling and marshalling it again
	arch := xmlutils.ElementToXML(xmlutils.XMLToElement(viewXML.Arch))
	view := View{
		ID:          viewXML.ID,
		Name:        name,
		Model:       viewXML.Model,
		Priority:    priority,
		arch:        arch,
		FieldParent: viewXML.FieldParent,
		SubViews:    make(map[string]SubViews),
		arches:      make(map[string]string),
	}
	vc.Add(&view)
}

// updateExistingViewFromXML updates an existing view with the given XML
// viewXML must have an InheritID
func (vc *Collection) updateExistingViewFromXML(viewXML ViewXML) {
	baseView := vc.GetByID(viewXML.InheritID)
	baseElem := xmlutils.XMLToElement(baseView.arch)
	specDoc := etree.NewDocument()
	if err := specDoc.ReadFromString(viewXML.Arch); err != nil {
		log.Panic("Unable to read inheritance specs", "error", err, "arch", viewXML.Arch)
	}
	for _, spec := range specDoc.ChildElements() {
		xpath := getInheritXPathFromSpec(spec)
		nodeToModify := baseElem.FindElement(xpath)
		nextNode := xmlutils.FindNextSibling(nodeToModify)
		modifyAction := spec.SelectAttr("position")
		switch modifyAction.Value {
		case "before":
			for _, node := range spec.ChildElements() {
				nodeToModify.Parent().InsertChild(nodeToModify, node)
			}
		case "after":
			for _, node := range spec.ChildElements() {
				nodeToModify.Parent().InsertChild(nextNode, node)
			}
		case "replace":
			for _, node := range spec.ChildElements() {
				nodeToModify.Parent().InsertChild(nodeToModify, node)
			}
			nodeToModify.Parent().RemoveChild(nodeToModify)
		case "inside":
			for _, node := range spec.ChildElements() {
				nodeToModify.AddChild(node)
			}
		case "attributes":
			for _, node := range spec.FindElements("./attribute") {
				attrName := node.SelectAttr("name").Value
				nodeToModify.RemoveAttr(attrName)
				nodeToModify.CreateAttr(attrName, node.Text())
			}
		}
	}
	baseView.arch = xmlutils.ElementToXML(baseElem)
}

// View is the internal definition of a view in the application
type View struct {
	ID          string
	Name        string
	Model       string
	Type        ViewType
	Priority    uint8
	arch        string
	FieldParent string
	Fields      []models.FieldName
	SubViews    map[string]SubViews
	arches      map[string]string
}

// A SubViews is a holder for embedded views of a field
type SubViews map[ViewType]*View

// populateFieldsMap scans arch, extract field names and put them in the fields slice
func (v *View) populateFieldsMap() {
	archElem := xmlutils.XMLToElement(v.arch)
	fieldElems := archElem.FindElements("//field")
	for _, f := range fieldElems {
		v.Fields = append(v.Fields, models.FieldName(f.SelectAttr("name").Value))
	}
}

// Arch returns the arch XML string of this view for the given language.
// Call with empty string to get the default language's arch
func (v *View) Arch(lang string) string {
	res, ok := v.arches[lang]
	if !ok || lang == "" {
		res = v.arch
	}
	return res
}

// setViewType sets the Type field with the view type
// scanned from arch
func (v *View) setViewType() {
	archElem := xmlutils.XMLToElement(v.arch)
	v.Type = ViewType(archElem.Tag)
}

// extractSubViews recursively scans arch for embedded views,
// extract them from arch and add them to SubViews.
func (v *View) extractSubViews() {
	archElem := xmlutils.XMLToElement(v.arch)
	fieldElems := archElem.FindElements("//field")
	for _, f := range fieldElems {
		if xmlutils.HasParentTag(f, "field") {
			// Discard fields of embedded views
			continue
		}
		fieldName := f.SelectAttr("name").Value
		for i, childElement := range f.ChildElements() {
			if _, exists := v.SubViews[fieldName]; !exists {
				v.SubViews[fieldName] = make(SubViews)
			}
			childView := View{
				ID:       fmt.Sprintf("%s_childview_%s_%d", v.ID, fieldName, i),
				arch:     xmlutils.ElementToXML(childElement),
				SubViews: make(map[string]SubViews),
				arches:   make(map[string]string),
			}
			childView.postProcess()
			v.SubViews[fieldName][childView.Type] = &childView
		}
		// Remove all children elements.
		// We do it in a separate loop on tokens to remove text and comments too.
		for _, childToken := range f.Child {
			f.RemoveChild(childToken)
		}
	}
	v.arch = xmlutils.ElementToXML(archElem)
}

// postProcess executes all actions that are needed the view for bootstrapping
func (v *View) postProcess() {
	v.setViewType()
	v.extractSubViews()
	v.populateFieldsMap()
	v.translateArch()
}

// translateArch populates the arches map with all the translations
func (v *View) translateArch() {
	labels := v.TranslatableStrings()
	archElt := xmlutils.XMLToElement(v.arch)
	for _, lang := range i18n.Langs {
		tArchElt := archElt.Copy()
		for _, label := range labels {
			attrElts := tArchElt.FindElements(fmt.Sprintf("//[@%s]", label.Attribute))
			for i, attrElt := range attrElts {
				if attrElt.SelectAttrValue(label.Attribute, "") != label.Value {
					continue
				}
				transLabel := i18n.TranslateResourceItem(lang, v.ID, label.Value)
				attrElts[i].RemoveAttr(label.Attribute)
				attrElts[i].CreateAttr(label.Attribute, transLabel)
			}
		}
		v.arches[lang] = xmlutils.ElementToXML(tArchElt)
	}
}

// A TranslatableAttribute is a reference to an attribute in a
// XML view definition that can be translated.
type TranslatableAttribute struct {
	Attribute string
	Value     string
}

// TranslatableStrings returns the list of all the strings in the
// view arch that must be translated.
func (v *View) TranslatableStrings() []TranslatableAttribute {
	archElt := xmlutils.XMLToElement(v.arch)
	var labels []TranslatableAttribute
	for _, tagName := range translatableAttributes {
		elts := archElt.FindElements(fmt.Sprintf("[@%s]", tagName))
		for _, elt := range elts {
			label := elt.SelectAttrValue(tagName, "")
			if label == "" {
				continue
			}
			labels = append(labels, TranslatableAttribute{Attribute: tagName, Value: label})
		}
	}
	return labels
}

// ViewXML is used to unmarshal the XML definition of a View
type ViewXML struct {
	ID          string `xml:"id,attr"`
	Name        string `xml:"name,attr"`
	Model       string `xml:"model,attr"`
	Priority    uint8  `xml:"priority,attr"`
	Arch        string `xml:",innerxml"`
	InheritID   string `xml:"inherit_id,attr"`
	FieldParent string `xml:"field_parent,attr"`
}

// LoadFromEtree reads the view given etree.Element, creates or updates the view
// and adds it to the view registry if it not already.
func LoadFromEtree(element *etree.Element) {
	Registry.LoadFromEtree(element)
}

// getInheritXPathFromSpec returns an XPath string that is suitable for
// searching the base view and find the node to modify.
func getInheritXPathFromSpec(spec *etree.Element) string {
	if spec.Tag == "xpath" {
		// We have an xpath expression, we take it
		return spec.SelectAttr("expr").Value
	}
	if len(spec.Attr) < 1 || len(spec.Attr) > 2 {
		log.Panic("Invalid view inherit spec", "spec", xmlutils.ElementToXML(spec))
	}
	var attrStr string
	for _, attr := range spec.Attr {
		if attr.Key != "position" {
			attrStr = fmt.Sprintf("[@%s='%s']", attr.Key, attr.Value)
			break
		}
	}
	return fmt.Sprintf("//%s%s", spec.Tag, attrStr)
}
