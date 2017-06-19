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

package menus

import (
	"sort"
	"strconv"
	"sync"

	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/tools/etree"
)

// Registry is the menu Collection of the application
var (
	Registry     *Collection
	bootstrapMap map[string]*Menu
)

// A Collection is a hierarchical and sortable Collection of menus
type Collection struct {
	sync.RWMutex
	Menus    []*Menu
	menusMap map[string]*Menu
}

func (mc *Collection) Len() int {
	return len(mc.Menus)
}

func (mc *Collection) Swap(i, j int) {
	mc.Menus[i], mc.Menus[j] = mc.Menus[j], mc.Menus[i]
}

func (mc *Collection) Less(i, j int) bool {
	return mc.Menus[i].Sequence < mc.Menus[j].Sequence
}

// Add adds a menu to the menu Collection
func (mc *Collection) Add(m *Menu) {
	if m.Action != nil {
		m.HasAction = true
	}
	var targetCollection *Collection
	if m.Parent != nil {
		if m.Parent.Children == nil {
			m.Parent.Children = NewCollection()
		}
		targetCollection = m.Parent.Children
		m.Parent.HasChildren = true
	} else {
		targetCollection = mc
	}
	m.ParentCollection = targetCollection
	targetCollection.Menus = append(targetCollection.Menus, m)
	sort.Sort(targetCollection)

	// We add the menu to the Registry which is the top collection
	mc.Lock()
	defer mc.Unlock()
	Registry.menusMap[m.ID] = m
}

// GetByID returns the Menu with the given id
func (mc *Collection) GetByID(id string) *Menu {
	return mc.menusMap[id]
}

// NewCollection returns a pointer to a new
// Collection instance
func NewCollection() *Collection {
	res := Collection{
		menusMap: make(map[string]*Menu),
	}
	return &res
}

// A Menu is the representation of a single menu item
type Menu struct {
	ID               string
	Name             string
	ParentID         string
	Parent           *Menu
	ParentCollection *Collection
	Children         *Collection
	Sequence         uint8
	ActionID         string
	Action           *actions.BaseAction
	HasChildren      bool
	HasAction        bool
}

// LoadFromEtree reads the menu given etree.Element, creates or updates the menu
// and adds it to the menu registry if it not already.
func LoadFromEtree(element *etree.Element) {
	menu := new(Menu)
	menu.ID = element.SelectAttrValue("id", "NO_ID")
	menu.ActionID = element.SelectAttrValue("action", "")
	menu.Name = element.SelectAttrValue("name", "")
	menu.ParentID = element.SelectAttrValue("parent", "")
	seq, _ := strconv.Atoi(element.SelectAttrValue("sequence", "10"))
	menu.Sequence = uint8(seq)

	bootstrapMap[menu.ID] = menu
}
