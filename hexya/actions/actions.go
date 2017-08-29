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

package actions

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sync"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
	"github.com/hexya-erp/hexya/hexya/views"
)

// An ActionType defines the type of action
type ActionType string

// Action types
const (
	ActionActWindow ActionType = "ir.actions.act_window"
	ActionServer    ActionType = "ir.actions.server"
)

// ActionViewType defines the type of view of an action
type ActionViewType string

// Action view types
const (
	ActionViewTypeForm ActionViewType = "form"
	ActionViewTypeTree ActionViewType = "tree"
)

// Registry is the action collection of the application
var Registry *Collection

// MakeActionRef creates an ActionRef from an action id
func MakeActionRef(id string) ActionRef {
	action := Registry.GetById(id)
	if action == nil {
		return ActionRef{}
	}
	return ActionRef{id, action.Name}
}

// ActionRef is an array of two strings representing an action:
// - The first one is the ID of the action
// - The second one is the name of the action
type ActionRef [2]string

// MarshalJSON is the JSON marshalling method of ActionRef
// It marshals empty ActionRef into null instead of ["", ""].
func (ar ActionRef) MarshalJSON() ([]byte, error) {
	if ar[0] == "" {
		return json.Marshal(nil)
	}
	return json.Marshal([2]string{ar[0], ar[1]})
}

// Value extracts ID of our ActionRef for storing in the database.
func (ar ActionRef) Value() (driver.Value, error) {
	return driver.Value(ar[0]), nil
}

// Scan fetches the name of our action from the ID
// stored in the database to fill the ActionRef.
func (ar *ActionRef) Scan(src interface{}) error {
	switch s := src.(type) {
	case string:
		*ar = MakeActionRef(s)
	case []byte:
		*ar = MakeActionRef(string(s))
	default:
		return fmt.Errorf("Invalid type for ActionRef: %T", src)
	}
	return nil
}

// ID returns the ID of the current action reference
func (ar ActionRef) ID() string {
	return ar[0]
}

// Name returns the name of the current action reference
func (ar ActionRef) Name() string {
	return ar[1]
}

// IsNull returns true if this ActionRef references no action
func (ar ActionRef) IsNull() bool {
	return ar[0] == "" && ar[1] == ""
}

var _ driver.Valuer = ActionRef{}
var _ sql.Scanner = &ActionRef{}
var _ json.Marshaler = &ActionRef{}

// An Collection is a collection of actions
type Collection struct {
	sync.RWMutex
	actions map[string]*BaseAction
	links   map[string][]*BaseAction
}

// NewCollection returns a pointer to a new
// Collection instance
func NewCollection() *Collection {
	res := Collection{
		actions: make(map[string]*BaseAction),
		links:   make(map[string][]*BaseAction),
	}
	return &res
}

// Add adds the given action to our Collection
func (ar *Collection) Add(a *BaseAction) {
	ar.Lock()
	defer ar.Unlock()
	ar.actions[a.ID] = a
	ar.links[a.SrcModel] = append(ar.links[a.SrcModel], a)
}

// GetById returns the Action with the given id
func (ar *Collection) GetById(id string) *BaseAction {
	return ar.actions[id]
}

// GetAll returns a list of all actions of this Collection.
// Actions are returned in an arbitrary order
func (ar *Collection) GetAll() []*BaseAction {
	res := make([]*BaseAction, len(ar.actions))
	var i int
	for _, action := range ar.actions {
		res[i] = action
		i++
	}
	return res
}

// MustGetById returns the Action with the given id
// It panics if the id is not found in the action registry
func (ar *Collection) MustGetById(id string) *BaseAction {
	action, ok := ar.actions[id]
	if !ok {
		log.Panic("Action does not exist", "action_id", id)
	}
	return action
}

// GetActionLinksForModel returns the list of linked actions
// for the model with the given name
func (ar *Collection) GetActionLinksForModel(modelName string) []*BaseAction {
	return ar.links[modelName]
}

// LoadFromEtree reads the action given etree.Element, creates or updates the action
// and adds it to the given Collection if it not already.
func (ar *Collection) LoadFromEtree(element *etree.Element) {
	xmlBytes := []byte(xmlutils.ElementToXML(element))
	var action BaseAction
	if err := xml.Unmarshal(xmlBytes, &action); err != nil {
		log.Panic("Unable to unmarshal element", "error", err, "bytes", string(xmlBytes))
	}
	ar.Add(&action)
}

// A BaseAction is the definition of an action. Actions define the
// behavior of the system in response to user requests.
type BaseAction struct {
	ID           string            `json:"id" xml:"id,attr"`
	Type         ActionType        `json:"type" xml:"type,attr"`
	Name         string            `json:"name" xml:"name,attr"`
	Model        string            `json:"res_model" xml:"model,attr"`
	ResID        int64             `json:"res_id" xml:"res_id,attr"`
	Method       string            `json:"method" xml:"method,attr"`
	Groups       []string          `json:"groups_id" xml:"groups,attr"`
	Domain       string            `json:"domain" xml:"domain,attr"`
	Help         string            `json:"help" xml:"help"`
	SearchView   views.ViewRef     `json:"search_view_id" xml:"search_view_id,attr"`
	SrcModel     string            `json:"src_model" xml:"src_model,attr"`
	Usage        string            `json:"usage" xml:"usage,attr"`
	Views        []views.ViewTuple `json:"views" xml:"view"`
	View         views.ViewRef     `json:"view_id" xml:"view_id,attr"`
	AutoRefresh  bool              `json:"auto_refresh" xml:"auto_refresh,attr"`
	ManualSearch bool              `json:"-" xml:"-"`
	ActViewType  ActionViewType    `json:"-" xml:"view_type"`
	ViewMode     string            `json:"view_mode" xml:"view_mode,attr"`
	Multi        bool              `json:"multi" xml:"multi,attr"`
	Target       string            `json:"target" xml:"target,attr"`
	AutoSearch   bool              `json:"auto_search" xml:"auto_search,attr"`
	Filter       bool              `json:"filter" xml:"filter,attr"`
	Limit        int64             `json:"limit" xml:"limit,attr"`
	Context      *types.Context    `json:"context" xml:"context,attr"`
	names        map[string]string
	//Flags interface{}`json:"flags"`
}

// TranslatedName returns the translated name of this action
// in the given language
func (a BaseAction) TranslatedName(lang string) string {
	res, ok := a.names[lang]
	if !ok {
		res = a.Name
	}
	return res
}

// LoadFromEtree reads the action given etree.Element, creates or updates the action
// and adds it to the action registry if it not already.
func LoadFromEtree(element *etree.Element) {
	Registry.LoadFromEtree(element)
}
