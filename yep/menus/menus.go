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

	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/tools/etree"
)

// Registry is the menu collection of the application
var Registry *MenuCollection

// A MenuCollection is a hierarchical and sortable collection of menus
type MenuCollection struct {
	sync.RWMutex
	Menus    []*Menu
	menusMap map[string]*Menu
}

func (mc *MenuCollection) Len() int {
	return len(mc.Menus)
}
func (mc *MenuCollection) Swap(i, j int) {
	mc.Menus[i], mc.Menus[j] = mc.Menus[j], mc.Menus[i]
}
func (mc *MenuCollection) Less(i, j int) bool {
	return mc.Menus[i].Sequence < mc.Menus[j].Sequence
}

// AddMenu adds a menu to the menu collection
func (mc *MenuCollection) AddMenu(m *Menu) {
	if m.Action != nil {
		m.HasAction = true
	}
	var targetCollection *MenuCollection
	if m.Parent != nil {
		if m.Parent.Children == nil {
			m.Parent.Children = new(MenuCollection)
		}
		targetCollection = m.Parent.Children
		m.Parent.HasChildren = true
	} else {
		targetCollection = mc
	}
	m.ParentCollection = targetCollection
	targetCollection.Menus = append(targetCollection.Menus, m)
	sort.Sort(targetCollection)

	// We add the menu to the top parent menu map
	topParent := m
	for topParent.Parent != nil {
		topParent = topParent.Parent
	}
	mc.Lock()
	defer mc.Unlock()
	topParent.ParentCollection.menusMap[m.ID] = m
}

// GetMenuById returns the Menu with the given id
func (mc *MenuCollection) GetMenuById(id string) *Menu {
	return mc.menusMap[id]
}

// NewMenuCollection returns a pointer to a new
// MenuCollection instance
func NewMenuCollection() *MenuCollection {
	res := MenuCollection{
		menusMap: make(map[string]*Menu),
	}
	return &res
}

// A Menu is the representation of a single menu item
type Menu struct {
	ID               string
	Name             string
	Parent           *Menu
	ParentCollection *MenuCollection
	Children         *MenuCollection
	Sequence         uint8
	Action           *actions.BaseAction
	HasChildren      bool
	HasAction        bool
}

// LoadFromEtree reads the menu given etree.Element, creates or updates the menu
// and adds it to the menu registry if it not already.
func LoadFromEtree(element *etree.Element) {
	menu := new(Menu)
	menu.ID = element.SelectAttrValue("id", "NO_ID")
	actionID := element.SelectAttrValue("action", "")
	defaultName := "No name"
	if actionID != "" {
		menu.Action = actions.Registry.GetActionById(actionID)
		defaultName = menu.Action.Name
	}
	menu.Name = element.SelectAttrValue("name", defaultName)
	parentID := element.SelectAttrValue("parent", "")
	if parentID != "" {
		menu.Parent = Registry.GetMenuById(parentID)
	}
	seq, _ := strconv.Atoi(element.SelectAttrValue("sequence", "10"))
	menu.Sequence = uint8(seq)

	Registry.AddMenu(menu)
}

func init() {
	Registry = NewMenuCollection()
}
