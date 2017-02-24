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
	"sync"

	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/tools/etree"
	"github.com/npiganeau/yep/yep/tools/logging"
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

// ViewInheritanceMode defines if this is a primary or an extension view
type ViewInheritanceMode string

// View inheritance modes
const (
	VIEW_PRIMARY   ViewInheritanceMode = "primary"
	VIEW_EXTENSION ViewInheritanceMode = "extension"
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

// UnmarshalXML is the XML unmarshalling method of ViewRef.
// It unmarshals null into an empty ViewRef.
func (vr *ViewRef) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var dst string
	if err := decoder.DecodeElement(&dst, &start); err != nil {
		return err
	}
	*vr = MakeViewRef(dst)
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
var _ xml.Unmarshaler = &ViewRef{}

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
	logging.LogAndPanic(log, "No view of this type in model", "type", viewType, "model", model)
	return nil
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
	Fields      []models.FieldName
}

// ViewArch is used to unmarshal the arch node of the XML definition
// of a View.
type ViewArch struct {
	XML string `xml:",innerxml"`
}

// ViewXML is used to unmarshal the XML definition of a View
type ViewXML struct {
	ID              string              `xml:"id,attr"`
	Name            string              `xml:"name"`
	Model           string              `xml:"model"`
	Priority        uint8               `xml:"priority"`
	ArchData        ViewArch            `xml:"arch"`
	InheritID       string              `xml:"inherit_id,attr"`
	FieldParent     string              `xml:"field_parent"`
	InheritanceMode ViewInheritanceMode `xml:"mode"`
}

// LoadFromEtree reads the view given etree.Element, creates or updates the view
// and adds it to the view registry if it not already.
func LoadFromEtree(element *etree.Element) {
	xmlBytes := []byte(xmlutils.ElementToXML(element))
	var viewXML ViewXML
	if err := xml.Unmarshal(xmlBytes, &viewXML); err != nil {
		logging.LogAndPanic(log, "Unable to unmarshal element", "error", err, "bytes", string(xmlBytes))
	}
	updateViewRegistry(viewXML)
}

// updateViewRegistry creates or updates the view in the Registry
// that is defined by the given ViewXML.
func updateViewRegistry(viewXML ViewXML) {
	if viewXML.InheritID != "" {
		// Update an existing view
		baseView := Registry.GetByID(viewXML.InheritID)
		baseElem := xmlutils.XMLToElement(baseView.Arch)
		specDoc := etree.NewDocument()
		if err := specDoc.ReadFromString(viewXML.ArchData.XML); err != nil {
			logging.LogAndPanic(log, "Unable to read inheritance specs", "error", err, "arch", viewXML.ArchData.XML)
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
			//break
		}
		baseView.Arch = xmlutils.ElementToXML(baseElem)
	} else {
		// Create a new view
		priority := uint8(16)
		if viewXML.Priority != 0 {
			priority = viewXML.Priority
		}
		// We check/standardize arch by unmarshalling and marshalling it again
		arch := xmlutils.ElementToXML(xmlutils.XMLToElement(viewXML.ArchData.XML))
		view := View{
			ID:          viewXML.ID,
			Name:        viewXML.Name,
			Model:       viewXML.Model,
			Priority:    priority,
			Arch:        arch,
			FieldParent: viewXML.FieldParent,
		}
		Registry.Add(&view)
	}
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
			logging.LogAndPanic(log, "Invalid view inherit spec", "spec", xmlutils.ElementToXML(spec))
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
