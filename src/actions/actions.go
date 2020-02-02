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
	"strings"
	"sync"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/tools/xmlutils"
	"github.com/hexya-erp/hexya/src/views"
)

// An ActionType defines the type of action
type ActionType string

// Action types
const (
	ActionActWindow   ActionType = "ir.actions.act_window"
	ActionServer      ActionType = "ir.actions.server"
	ActionClient      ActionType = "ir.actions.client"
	ActionCloseWindow ActionType = "ir.actions.act_window_close"
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
	action := Registry.GetByXMLID(id)
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
		return json.Marshal(false)
	}
	return json.Marshal([2]string{ar[0], ar[1]})
}

// UnmarshalJSON for ActionRef.
// Unmarshals false as an empty ActionRef
func (ar *ActionRef) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "null", "false":
		*ar = ActionRef{}
		return nil
	default:
		var aux [2]string
		if err := json.Unmarshal(data, &aux); err != nil {
			return err
		}
		*ar = aux
	}
	return nil
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
		return fmt.Errorf("invalid type for ActionRef: %T", src)
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

// An ActionString is the concatenation of the action type and its ID
// e.g. ir.actions.act_window,76
type ActionString struct {
	Type string
	ID   int64
}

// MarshalJSON for the actionString type. Marshals to false if the action string is empty
func (as ActionString) MarshalJSON() ([]byte, error) {
	if as.ID == 0 {
		return json.Marshal(false)
	}
	return json.Marshal(fmt.Sprintf("%s,%d", as.Type, as.ID))
}

// A Collection is a collection of actions
type Collection struct {
	sync.RWMutex
	actions     map[string]*Action
	actionsByID map[int64]*Action
	links       map[string][]*Action
}

// NewCollection returns a pointer to a new
// Collection instance
func NewCollection() *Collection {
	res := Collection{
		actions:     make(map[string]*Action),
		actionsByID: make(map[int64]*Action),
		links:       make(map[string][]*Action),
	}
	return &res
}

// Add adds the given action to our Collection
func (ar *Collection) Add(a *Action) {
	ar.Lock()
	defer ar.Unlock()
	maxID := int64(1)
	for _, act := range ar.actions {
		if act.ID >= maxID {
			maxID = act.ID + 1
		}
	}
	a.ID = maxID
	ar.actions[a.XMLID] = a
	ar.actionsByID[a.ID] = a
	ar.links[a.SrcModel] = append(ar.links[a.SrcModel], a)
}

// GetByXMLID returns the Action with the given xmlid
func (ar *Collection) GetByXMLID(id string) *Action {
	return ar.actions[id]
}

// GetById returns the Action with the given id
func (ar *Collection) GetById(id int64) *Action {
	return ar.actionsByID[id]
}

// GetAll returns a list of all actions of this Collection.
// Actions are returned in an arbitrary order
func (ar *Collection) GetAll() []*Action {
	res := make([]*Action, len(ar.actions))
	var i int
	for _, action := range ar.actions {
		res[i] = action
		i++
	}
	return res
}

// MustGetByXMLID returns the Action with the given xmlid
// It panics if the id is not found in the action registry
func (ar *Collection) MustGetByXMLID(id string) *Action {
	action, ok := ar.actions[id]
	if !ok {
		log.Panic("Action does not exist", "action_id", id)
	}
	return action
}

// MustGetById returns the Action with the given id
// It panics if the id is not found in the action registry
func (ar *Collection) MustGetById(id int64) *Action {
	action, ok := ar.actionsByID[id]
	if !ok {
		log.Panic("Action does not exist", "action_id", id)
	}
	return action
}

// GetActionLinksForModel returns the list of linked actions
// for the model with the given name
func (ar *Collection) GetActionLinksForModel(modelName string) []*Action {
	return ar.links[modelName]
}

// LoadFromEtree reads the action given etree.Element, creates or updates the action
// and adds it to the given Collection if it not already.
func (ar *Collection) LoadFromEtree(element *etree.Element) {
	xmlBytes, err := xmlutils.ElementToXML(element)
	if err != nil {
		log.Panic("unable to convert element to XML", "error", err)
	}
	var action Action
	if err = xml.Unmarshal(xmlBytes, &action); err != nil {
		log.Panic("Unable to unmarshal element", "error", err, "bytes", string(xmlBytes))
	}
	ar.Add(&action)
}

// actionHelp is a placeholder struct to recover
// help XML from action definition
type actionHelp struct {
	Content string `xml:",innerxml"`
}

// A Action is the definition of an action. Actions define the
// behavior of the system in response to user requests.
type Action struct {
	ID           int64                  `json:"id" xml:"-"`
	XMLID        string                 `json:"xmlid" xml:"id,attr"`
	Type         ActionType             `json:"type" xml:"type,attr"`
	Name         string                 `json:"name" xml:"name,attr"`
	Model        string                 `json:"res_model" xml:"model,attr"`
	ResID        int64                  `json:"res_id" xml:"res_id,attr"`
	Method       string                 `json:"method" xml:"method,attr"`
	Groups       []string               `json:"groups_id" xml:"groups,attr"`
	Domain       string                 `json:"domain" xml:"domain,attr"`
	HelpXML      actionHelp             `json:"-" xml:"help"`
	Help         string                 `json:"help" xml:"-"`
	SearchView   views.ViewRef          `json:"search_view_id" xml:"search_view_id,attr"`
	SrcModel     string                 `json:"src_model" xml:"src_model,attr"`
	Usage        string                 `json:"usage" xml:"usage,attr"`
	Views        []views.ViewTuple      `json:"views" xml:"view"`
	View         views.ViewRef          `json:"view_id" xml:"view_id,attr"`
	AutoRefresh  bool                   `json:"auto_refresh" xml:"auto_refresh,attr"`
	ManualSearch bool                   `json:"-" xml:"-"`
	ActViewType  ActionViewType         `json:"-" xml:"view_type"`
	ViewMode     string                 `json:"view_mode" xml:"view_mode,attr"`
	Multi        bool                   `json:"multi" xml:"multi,attr"`
	Target       string                 `json:"target" xml:"target,attr"`
	AutoSearch   bool                   `json:"auto_search" xml:"auto_search,attr"`
	Filter       bool                   `json:"filter" xml:"filter,attr"`
	Limit        int64                  `json:"limit" xml:"limit,attr"`
	Context      *types.Context         `json:"context" xml:"context,attr"`
	Flags        map[string]interface{} `json:"flags"`
	Tag          string                 `json:"tag"`
	names        map[string]string
}

// TranslatedName returns the translated name of this action
// in the given language
func (a Action) TranslatedName(lang string) string {
	res, ok := a.names[lang]
	if !ok {
		res = a.Name
	}
	return res
}

// Sanitize makes the necessary updates to action definitions.
// It is good practice to call Sanitize before sending an action to the client.
func (a *Action) Sanitize() {
	switch a.Type {
	case ActionActWindow:
		a.sanitizeActWindow()
	}
}

// sanitizeActWindow makes the necessary updates to action definitions. In particular:
// - Add a few default values
// - Add View to Views if not already present
// - Add all views that are not specified
func (a *Action) sanitizeActWindow() {
	// Set a few default values
	if a.Target == "" {
		a.Target = "current"
	}
	a.AutoSearch = !a.ManualSearch
	if a.ActViewType == "" {
		a.ActViewType = ActionViewTypeForm
	}
	a.Help = a.HelpXML.Content

	// Add View to Views if not already present
	var present bool
	// Check if view is present in Views
	for _, view := range a.Views {
		if !a.View.IsNull() {
			if view.ID == a.View.ID() {
				present = true
				break
			}
		}
	}
	// Add View if not present in Views
	if !present && !a.View.IsNull() {
		vType := views.Registry.GetByID(a.View.ID()).Type
		newRef := views.ViewTuple{
			ID:   a.View.ID(),
			Type: vType,
		}
		a.Views = append(a.Views, newRef)
	}

	// Add views of ViewMode that are not specified
	modesStr := strings.Split(a.ViewMode, ",")
	modes := make([]views.ViewType, len(modesStr))
	for i, v := range modesStr {
		modes[i] = views.ViewType(strings.TrimSpace(v))
	}
modeLoop:
	for _, mode := range modes {
		for _, vRef := range a.Views {
			if vRef.Type == mode {
				continue modeLoop
			}
		}
		// No view defined for mode, we need to find it.
		view := views.Registry.GetFirstViewForModel(a.Model, mode)
		newRef := views.ViewTuple{
			ID:   view.ID,
			Type: view.Type,
		}
		a.Views = append(a.Views, newRef)
	}

	// Fixes
	a.fixViewModes()
}

// fixViewModes makes the necessary changes to the given action.
//
// For OpenERP historical reasons, tree views are called 'list' when
// in ActionViewType 'form' and 'tree' when in ActionViewType 'tree'.
func (a *Action) fixViewModes() {
	if a.ActViewType == ActionViewTypeForm {
		for i, v := range a.Views {
			if v.Type == views.ViewTypeTree {
				v.Type = views.ViewTypeList
			}
			a.Views[i].Type = v.Type
		}
	}
}

// ActionString returns the ActionString associated with this action.
func (a Action) ActionString() ActionString {
	return ActionString{ID: a.ID, Type: string(a.Type)}
}

// LoadFromEtree reads the action given etree.Element, creates or updates the action
// and adds it to the action registry if it not already.
func LoadFromEtree(element *etree.Element) {
	Registry.LoadFromEtree(element)
}
