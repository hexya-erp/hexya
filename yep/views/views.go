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
	"fmt"
	"strings"
	"sync"

	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/tools/etree"
	"github.com/npiganeau/yep/yep/tools/xmlutils"
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

var _ driver.Valuer = &ViewRef{}
var _ sql.Scanner = &ViewRef{}
var _ json.Marshaler = &ViewRef{}
var _ json.Unmarshaler = &ViewRef{}
var _ xml.UnmarshalerAttr = &ViewRef{}

// ViewTuple is an array of two strings representing a view:
// - The first one is the ID of the view
// - The second one is the view type corresponding to the view ID
type ViewTuple struct {
	ID   string
	Type ViewType
}

// MarshalJSON is the JSON marshalling method of ViewTuple.
// It marshals ViewTuple into a list [id, type].
func (vt ViewTuple) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]string{vt.ID, string(vt.Type)})
}

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
		if view.Priority >= v.Priority {
			break
		}
	}
	defer vc.Unlock()
	vc.views[v.ID] = v
	endElems := make([]*View, len(vc.orderedViews[v.Model][index:]))
	copy(endElems, vc.orderedViews[v.Model][index:])
	vc.orderedViews[v.Model] = append(append(vc.orderedViews[v.Model][:index], v), endElems...)
}

// GetByID returns the View with the given id
func (vc *Collection) GetByID(id string) *View {
	return vc.views[id]
}

// GetFirstViewForModel returns the first view of type viewType for the given model
func (vc *Collection) GetFirstViewForModel(model string, viewType ViewType) *View {
	for _, view := range vc.orderedViews[model] {
		if view.Type == viewType {
			return view
		}
	}
	log.Panic("No view of this type in model", "type", viewType, "model", model)
	return nil
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

// View is the internal definition of a view in the application
type View struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Model       string   `json:"model"`
	Type        ViewType `json:"type"`
	Priority    uint8    `json:"priority"`
	Arch        string   `json:"arch"`
	FieldParent string   `json:"field_parent"`
	//Toolbar     actions.Toolbar `json:"toolbar"`
	Fields []models.FieldName
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
	xmlBytes := []byte(xmlutils.ElementToXML(element))
	var viewXML ViewXML
	if err := xml.Unmarshal(xmlBytes, &viewXML); err != nil {
		log.Panic("Unable to unmarshal element", "error", err, "bytes", string(xmlBytes))
	}
	updateViewRegistry(viewXML)
}

// updateViewRegistry creates or updates the view in the Registry
// that is defined by the given ViewXML.
func updateViewRegistry(viewXML ViewXML) {
	if viewXML.InheritID != "" {
		// Update an existing view
		updateExistingViewFromXML(viewXML)
	} else {
		// Create a new view
		createNewViewFromXML(viewXML)
	}
}

// createNewViewFromXML creates and register a new view with the given XML
func createNewViewFromXML(viewXML ViewXML) {
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
		Arch:        arch,
		FieldParent: viewXML.FieldParent,
	}
	Registry.Add(&view)
}

// updateExistingViewFromXML updates an existing view with the given XML
// viewXML must have an InheritID
func updateExistingViewFromXML(viewXML ViewXML) {
	baseView := Registry.GetByID(viewXML.InheritID)
	baseElem := xmlutils.XMLToElement(baseView.Arch)
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
	baseView.Arch = xmlutils.ElementToXML(baseElem)
}

// getInheritXPathFromSpec returns an XPath string that is suitable for
// searching the base view and find the node to modify.
func getInheritXPathFromSpec(spec *etree.Element) string {
	var xpath string
	if spec.Tag == "xpath" {
		// We have an xpath expression, we take it
		xpath = spec.SelectAttr("expr").Value
	} else {
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
		xpath = fmt.Sprintf("//%s%s", spec.Tag, attrStr)
	}
	return xpath
}
