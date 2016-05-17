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
	"fmt"
	"strings"
	"sync"

	"github.com/beevik/etree"
	"github.com/npiganeau/yep/yep/orm"
)

type ActionType string

const (
	ACTION_ACT_WINDOW ActionType = "ir.actions.act_window"
	ACTION_SERVER     ActionType = "ir.actions.server"
)

type ActionViewType string

const (
	ACTION_VIEW_TYPE_FORM ActionViewType = "form"
	ACTION_VIEW_TYPE_TREE ActionViewType = "tree"
)

var ActionsRegistry *ActionsCollection

func MakeActionRef(id string) ActionRef {
	action := ActionsRegistry.GetActionById(id)
	if action == nil {
		return ActionRef{}
	}
	return ActionRef{id, action.Name}
}

type ActionRef [2]string

func (e *ActionRef) String() string {
	sl := []string{e[0], e[1]}
	return fmt.Sprintf(`[%s]`, strings.Join(sl, ","))
}

func (e *ActionRef) FieldType() int {
	return orm.TypeTextField
}

func (e *ActionRef) SetRaw(value interface{}) error {
	switch d := value.(type) {
	case string:
		dTrimmed := strings.Trim(d, "[]")
		tokens := strings.Split(dTrimmed, ",")
		if len(tokens) > 1 {
			*e = [2]string{tokens[0], tokens[1]}
			return nil
		}
		e = nil
		return fmt.Errorf("<ActionRef.SetRaw>Unable to parse %s", d)
	default:
		return fmt.Errorf("<ActionRef.SetRaw> unknown value `%v`", value)
	}
}

func (e *ActionRef) RawValue() interface{} {
	return e.String()
}

func (e *ActionRef) MarshalJSON() ([]byte, error) {
	if e[0] == "" {
		return json.Marshal(nil)
	}
	sl := []string{e[0], e[1]}
	return json.Marshal(sl)
}

var _ orm.Fielder = new(ActionRef)

type ActionsCollection struct {
	sync.RWMutex
	actions map[string]*BaseAction
}

// NewActionCollection returns a pointer to a new
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

type BaseAction struct {
	ID         string     `json:"id"`
	Type       ActionType `json:"type"`
	Name       string     `json:"name"`
	Model      string     `json:"res_model"`
	ResID      int64      `json:"res_id"`
	Groups     []string   `json:"groups_id"`
	Domain     string     `json:"domain"`
	Help       string     `json:"help"`
	SearchView ViewRef    `json:"search_view_id"`
	SrcModel   string     `json:"src_model"`
	Usage      string     `json:"usage"`
	//Flags interface{}`json:"flags"`
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
	//SearchView  string         `json:"search_view"`
	Filter bool  `json:"filter"`
	Limit  int64 `json:"limit"`
	//Context models.Context `json:"context"`
}

type Toolbar struct {
	Print  []*BaseAction `json:"print"`
	Action []*BaseAction `json:"action"`
	Relate []*BaseAction `json:"relate"`
}

/*
LoadActionFromEtree reads the action given etree.Element, creates or updates the action
and adds it to the action registry if it not already.
*/
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
				panic(fmt.Errorf("Unknown field type `%s` in view `%s`", fieldType, actionHash["id"]))
			}
		} else {
			actionHash[name] = fieldNode.Text()
		}
	}
	// We marshal viewHash in JSON and then unmarshal into a View struct
	bytes, _ := json.Marshal(actionHash)
	var act BaseAction
	if err := json.Unmarshal(bytes, &act); err != nil {
		panic(fmt.Errorf("Unable to unmarshal action: %s", err))
	}
	ActionsRegistry.AddAction(&act)
}
