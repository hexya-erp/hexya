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
	"sort"
)

var MenusRegistry *MenuCollection

type MenuCollection struct {
	Menus []*UiMenu
}

func (mr *MenuCollection) Len() int {
	return len(mr.Menus)
}
func (mr *MenuCollection) Swap(i, j int) {
	mr.Menus[i], mr.Menus[j] = mr.Menus[j], mr.Menus[i]
}
func (mr *MenuCollection) Less(i, j int) bool {
	return mr.Menus[i].Sequence < mr.Menus[j].Sequence
}

/*
AddMenu adds a menu to the menu registry
*/
func (mr *MenuCollection) AddMenu(m *UiMenu) {
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
		targetCollection = mr
	}
	targetCollection.Menus = append(targetCollection.Menus, m)
	sort.Sort(targetCollection)
}

type UiMenu struct {
	ID          string
	Name        string
	Parent      *UiMenu
	Children    *MenuCollection
	Sequence    uint8
	Action      *BaseAction
	HasChildren bool
	HasAction   bool
}
