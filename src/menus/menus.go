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

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/src/actions"
)

// Registry is the menu Collection of the application
var (
	Registry     *Collection
	bootstrapMap map[string]*Menu
)

// A Collection is a hierarchical and sortable Collection of menus
type Collection struct {
	sync.RWMutex
	Menus        []*Menu
	menusMap     map[string]*Menu
	menusMapByID map[int64]*Menu
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
	Registry.menusMap[m.XMLID] = m
	Registry.menusMapByID[m.ID] = m
}

// GetByID returns the Menu with the given id
func (mc *Collection) GetByID(id int64) *Menu {
	mc.RLock()
	defer mc.RUnlock()
	return mc.menusMapByID[id]
}

// GetByXMLID returns the Menu with the given xmlid
func (mc *Collection) GetByXMLID(xmlid string) *Menu {
	mc.RLock()
	defer mc.RUnlock()
	return mc.menusMap[xmlid]
}

// All returns all menus recursively
func (mc *Collection) All() []*Menu {
	mc.RLock()
	defer mc.RUnlock()
	var res []*Menu
	for _, menu := range mc.menusMap {
		res = append(res, menu)
	}
	return res
}

// NewCollection returns a pointer to a new
// Collection instance
func NewCollection() *Collection {
	res := Collection{
		menusMap:     make(map[string]*Menu),
		menusMapByID: make(map[int64]*Menu),
	}
	return &res
}

// A Menu is the representation of a single menu item
type Menu struct {
	ID               int64
	XMLID            string
	Name             string
	ParentID         string
	Parent           *Menu
	ParentCollection *Collection
	Children         *Collection
	Sequence         uint8
	ActionID         string
	Action           *actions.Action
	HasChildren      bool
	HasAction        bool
	WebIcon          string
	names            map[string]string
}

// TranslatedName returns the translated name of this menu
// in the given language
func (m Menu) TranslatedName(lang string) string {
	res, ok := m.names[lang]
	if !ok {
		res = m.Name
	}
	return res
}

// LoadFromEtree reads the menu given etree.Element, creates or updates the menu
// and adds it to the menu registry if it not already.
func LoadFromEtree(element *etree.Element) {
	AddMenuToMapFromEtree(element, bootstrapMap)
}

// AddMenuToMapFromEtree reads the menu from the given element
// and adds it to the given map.
func AddMenuToMapFromEtree(element *etree.Element, mMap map[string]*Menu) map[string]*Menu {
	seq, _ := strconv.Atoi(element.SelectAttrValue("sequence", "10"))
	menu := Menu{
		XMLID:    element.SelectAttrValue("id", "NO_ID"),
		ActionID: element.SelectAttrValue("action", ""),
		Name:     element.SelectAttrValue("name", ""),
		ParentID: element.SelectAttrValue("parent", ""),
		WebIcon:  element.SelectAttrValue("web_icon", ""),
		Sequence: uint8(seq),
	}
	mMap[menu.XMLID] = &menu
	return mMap
}
