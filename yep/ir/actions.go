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
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/beevik/etree"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/tools/logging"
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

// ActionsRegistry is the action collection of the application
var ActionsRegistry *ActionsCollection

// MakeActionRef creates an ActionRef from an action id
func MakeActionRef(id string) ActionRef {
	action := ActionsRegistry.GetActionById(id)
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

var _ driver.Valuer = ActionRef{}
var _ sql.Scanner = &ActionRef{}
var _ json.Marshaler = &ActionRef{}

// An ActionsCollection is a collection of actions
type ActionsCollection struct {
	sync.RWMutex
	actions map[string]*BaseAction
}

// NewActionsCollection returns a pointer to a new
// ActionsCollection instance
func NewActionsCollection() *ActionsCollection {
	res := ActionsCollection{
		actions: make(map[string]*BaseAction),
	}
	return &res
}

// AddAction adds the given action to our ActionsCollection
func (ar *ActionsCollection) AddAction(a *BaseAction) {
	ar.Lock()
	defer ar.Unlock()
	ar.actions[a.ID] = a
}

// GetActionById returns the Action with the given id
func (ar *ActionsCollection) GetActionById(id string) *BaseAction {
	return ar.actions[id]
}

// A BaseAction is the definition of an action. Actions define the
// behavior of the system in response to user actions.
type BaseAction struct {
	ID           string         `json:"id"`
	Type         ActionType     `json:"type"`
	Name         string         `json:"name"`
	Model        string         `json:"res_model"`
	ResID        int64          `json:"res_id"`
	Groups       []string       `json:"groups_id"`
	Domain       string         `json:"domain"`
	Help         string         `json:"help"`
	SearchView   ViewRef        `json:"search_view_id"`
	SrcModel     string         `json:"src_model"`
	Usage        string         `json:"usage"`
	Views        []ViewRef      `json:"views"`
	View         ViewRef        `json:"view_id"`
	AutoRefresh  bool           `json:"auto_refresh"`
	ManualSearch bool           `json:"-"`
	ActViewType  ActionViewType `json:"-"`
	ViewMode     string         `json:"view_mode"`
	ViewIds      []string       `json:"view_ids"`
	Multi        bool           `json:"multi"`
	Target       string         `json:"target"`
	AutoSearch   bool           `json:"auto_search"`
	Filter       bool           `json:"filter"`
	Limit        int64          `json:"limit"`
	Context      *types.Context `json:"context"`
	//Flags interface{}`json:"flags"`
	//SearchView  string         `json:"search_view"`
}

// A Toolbar holds the actions in the toolbar of the action manager
type Toolbar struct {
	Print  []*BaseAction `json:"print"`
	Action []*BaseAction `json:"action"`
	Relate []*BaseAction `json:"relate"`
}

// LoadActionFromEtree reads the action given etree.Element, creates or updates the action
// and adds it to the action registry if it not already.
func LoadActionFromEtree(element *etree.Element) {
	// We populate an actionHash from XML data fields
	actionHash := make(map[string]interface{})
	actionHash["id"] = element.SelectAttrValue("id", "NO_ID")
	for _, fieldNode := range element.FindElements("field") {
		name := fieldNode.SelectAttrValue("name", "NO_NAME")
		ref := fieldNode.SelectAttrValue("ref", "")
		if ref != "" && (name == "view_id" || name == "search_view_id") {
			actionHash[name] = MakeViewRef(ref)
		} else if len(fieldNode.ChildElements()) > 0 {
			fieldType := fieldNode.SelectAttrValue("type", "text")
			switch fieldType {
			case "html":
				nodeDoc := etree.NewDocument()
				nodeDoc.SetRoot(fieldNode.ChildElements()[0].Copy())
				value, _ := nodeDoc.WriteToString()
				actionHash[name] = value
			default:
				logging.LogAndPanic(log, "Unknown field type", "type", fieldType, "action", actionHash["id"])
			}
		} else {
			actionHash[name] = fieldNode.Text()
		}
	}
	// We marshal viewHash in JSON and then unmarshal into a View struct
	bytes, _ := json.Marshal(actionHash)
	var act BaseAction
	if err := json.Unmarshal(bytes, &act); err != nil {
		logging.LogAndPanic(log, "Unable to unmarshal action", "actionHash", actionHash, "error", err)
	}
	ActionsRegistry.AddAction(&act)
}
