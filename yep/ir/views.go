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

package ir

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"sync"

	"github.com/npiganeau/yep/yep/orm"
)

type ViewType string

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

type ViewInheritanceMode string

const (
	VIEW_PRIMARY   ViewInheritanceMode = "primary"
	VIEW_EXTENSION ViewInheritanceMode = "extension"
)

var ViewsRegistry *ViewsCollection

func MakeViewRef(id string) ViewRef {
	view := ViewsRegistry.GetViewById(id)
	if view == nil {
		return ViewRef{}
	}
	return ViewRef{id, view.Name}
}

type ViewRef [2]string

func (e *ViewRef) String() string {
	sl := []string{e[0], e[1]}
	return fmt.Sprintf(`[%s]`, strings.Join(sl, ","))
}

func (e *ViewRef) FieldType() int {
	return orm.TypeTextField
}

func (e *ViewRef) SetRaw(value interface{}) error {
	switch d := value.(type) {
	case string:
		dTrimmed := strings.Trim(d, "[]")
		tokens := strings.Split(dTrimmed, ",")
		if len(tokens) > 1 {
			*e = [2]string{tokens[0], tokens[1]}
			return nil
		}
		e = nil
		return fmt.Errorf("<ViewRef.SetRaw> Unable to parse %s", d)
	default:
		return fmt.Errorf("<ViewRef.SetRaw> unknown value `%v`", value)
	}
}

func (e *ViewRef) RawValue() interface{} {
	return e.String()
}

func (e *ViewRef) MarshalJSON() ([]byte, error) {
	if e[0] == "" {
		return json.Marshal(nil)
	}
	sl := []string{e[0], e[1]}
	return json.Marshal(sl)
}

var _ orm.Fielder = new(ViewRef)

type ViewsCollection struct {
	sync.RWMutex
	views        map[string]*View
	orderedViews map[string][]*View
}

// NewViewCollection returns a pointer to a new
// ViewsCollection instance
func NewViewsCollection() *ViewsCollection {
	res := ViewsCollection{
		views:        make(map[string]*View),
		orderedViews: make(map[string][]*View),
	}
	return &res
}

// AddView adds the given view to our ViewsCollection
func (vc *ViewsCollection) AddView(v *View) {
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
	endElems := vc.orderedViews[v.Model][index:]
	vc.orderedViews[v.Model] = append(append(vc.orderedViews[v.Model][:index], v), endElems...)
}

// GetViewById returns the View with the given id
func (vc *ViewsCollection) GetViewById(id string) *View {
	return vc.views[id]
}

/*
GetFirstViewForModel returns the first view of type viewType for the given model
*/
func (vc *ViewsCollection) GetFirstViewForModel(model string, viewType ViewType) *View {
	for _, view := range vc.orderedViews[model] {
		if view.Type == viewType {
			return view
		}
	}
	panic(fmt.Errorf("Unable to find view of type `%s` for model `%s`", viewType, model))
}

type View struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Model              string              `json:"model"`
	Type               ViewType            `json:"type"`
	Priority           uint8               `json:"priority"`
	Arch               string              `json:"arch"`
	InheritID          *View               `json:"inherit_id"`
	InheritChildrenIDs []*View             `json:"inherit_children_ids"`
	FieldParent        string              `json:"field_parent"`
	InheritanceMode    ViewInheritanceMode `json:"mode"`
	Fields             []string
	//GroupsID []*Group
}

// Represents a <field> node in a XML view arch
type FieldNode struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
}
