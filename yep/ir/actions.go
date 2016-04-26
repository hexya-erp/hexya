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

	"github.com/npiganeau/yep/yep/orm"
)

type ActionType string

const (
	ACTION_ACT_WINDOW ActionType = "ir.actions.act_window"
	ACTION_SERVER     ActionType = "ir.actions.server"
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
	// Add View to Views if not already present
	var present bool
	for _, view := range a.Views {
		if len(view) > 0 && len(a.View) > 0 {
			if view[0] == a.View[0] {
				present = true
				break
			}
		}
	}
	if !present && len(a.View) > 0 && a.View[0] != "" {
		vType := ViewsRegistry.GetViewById(a.View[0]).Type
		if vType == VIEW_TYPE_TREE {
			vType = VIEW_TYPE_LIST
		}
		newRef := ViewRef{
			a.View[0],
			string(vType),
		}
		a.Views = append(a.Views, newRef)
	}
	// Set a few default values
	if a.Target == "" {
		a.Target = "current"
	}
	a.AutoSearch = !a.ManualSearch
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
	Views        []ViewRef `json:"views"`
	View         ViewRef   `json:"view_id"`
	AutoRefresh  bool      `json:"auto_refresh"`
	ManualSearch bool      `json:"-"`
	ViewMode     string    `json:"view_mode"`
	ViewIds      []string  `json:"view_ids"`
	Multi        bool      `json:"multi"`
	Target       string    `json:"target"`
	AutoSearch   bool      `json:"auto_search"`
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
