// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package menus

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/tools/xmlutils"
)

var (
	rootMenuDef = `
<menuitem id="root_menu" name="Root Menu" sequence="20" web_icon="base,static/description/icon.png"/>
`
	childMenuDef = `
<menuitem id="child_menu" name="Child Menu" parent="root_menu" sequence="5"/>
`
	actionMenuDef = `
<menuitem id="action_menu" parent="root_menu" action="my_action" sequence="10"/>
`
	unknownParentMenuDef = `
<menuitem id="orphan_menu" name="Orphan Menu" parent="unknown_menu"/>
`
)

// menuFromXML returns a new Menu from the given XML definition
func menuFromXML(t *testing.T, xmlDef string) *Menu {
	t.Helper()
	element, err := xmlutils.XMLToElement(xmlDef)
	assert.Nil(t, err)
	mMap := AddMenuToMapFromEtree(element, make(map[string]*Menu))
	assert.Len(t, mMap, 1)
	for _, menu := range mMap {
		return menu
	}
	return nil
}

// resetRegistry empties the menus registry and the bootstrap map
func resetRegistry() {
	Registry = NewCollection()
	bootstrapMap = make(map[string]*Menu)
}

func TestAddMenuToMapFromEtree(t *testing.T) {
	t.Run("All attributes given", func(t *testing.T) {
		menu := menuFromXML(t, rootMenuDef)
		assert.Equal(t, int64(1), menu.ID)
		assert.Equal(t, "root_menu", menu.XMLID)
		assert.Equal(t, "Root Menu", menu.Name)
		assert.Equal(t, "", menu.ParentID)
		assert.Equal(t, "", menu.ActionID)
		assert.Equal(t, "base,static/description/icon.png", menu.WebIcon)
		assert.EqualValues(t, 20, menu.Sequence)
	})
	t.Run("Default values", func(t *testing.T) {
		element, err := xmlutils.XMLToElement(`<menuitem/>`)
		assert.Nil(t, err)
		mMap := AddMenuToMapFromEtree(element, make(map[string]*Menu))
		menu := mMap["NO_ID"]
		assert.NotNil(t, menu)
		assert.Equal(t, "", menu.Name)
		assert.EqualValues(t, 10, menu.Sequence)
	})
	t.Run("IDs are incremented in the given map", func(t *testing.T) {
		mMap := make(map[string]*Menu)
		root, err := xmlutils.XMLToElement(rootMenuDef)
		assert.Nil(t, err)
		child, err := xmlutils.XMLToElement(childMenuDef)
		assert.Nil(t, err)
		mMap = AddMenuToMapFromEtree(root, mMap)
		mMap = AddMenuToMapFromEtree(child, mMap)
		assert.Len(t, mMap, 2)
		assert.Equal(t, int64(1), mMap["root_menu"].ID)
		assert.Equal(t, int64(2), mMap["child_menu"].ID)
		assert.Equal(t, "root_menu", mMap["child_menu"].ParentID)
	})
}

func TestLoadFromEtree(t *testing.T) {
	defer resetRegistry()
	resetRegistry()
	element, err := xmlutils.XMLToElement(rootMenuDef)
	assert.Nil(t, err)
	LoadFromEtree(element)
	assert.Len(t, bootstrapMap, 1)
	assert.Contains(t, bootstrapMap, "root_menu")
}

func TestCollection(t *testing.T) {
	defer resetRegistry()
	resetRegistry()
	rootMenu := menuFromXML(t, rootMenuDef)
	childMenu := menuFromXML(t, childMenuDef)
	childMenu.ID = 2
	childMenu.Parent = rootMenu
	otherRootMenu := &Menu{ID: 3, XMLID: "other_root_menu", Name: "Other Root Menu", Sequence: 5}
	t.Run("Adding menus to the registry", func(t *testing.T) {
		Registry.Add(rootMenu)
		Registry.Add(otherRootMenu)
		assert.Equal(t, 2, Registry.Len())
		// Menus are sorted by sequence
		assert.Equal(t, "other_root_menu", Registry.Menus[0].XMLID)
		assert.Equal(t, "root_menu", Registry.Menus[1].XMLID)
		assert.Equal(t, Registry, rootMenu.ParentCollection)
		assert.False(t, rootMenu.HasChildren)
		assert.False(t, rootMenu.HasAction)
	})
	t.Run("Adding a child menu", func(t *testing.T) {
		Registry.Add(childMenu)
		// Children are not added to the top collection
		assert.Equal(t, 2, Registry.Len())
		assert.True(t, rootMenu.HasChildren)
		assert.NotNil(t, rootMenu.Children)
		assert.Equal(t, 1, rootMenu.Children.Len())
		assert.Equal(t, rootMenu.Children, childMenu.ParentCollection)
	})
	t.Run("Adding a menu with an action", func(t *testing.T) {
		actionMenu := &Menu{ID: 4, XMLID: "action_menu", Name: "Action Menu",
			Action: &actions.Action{XMLID: "my_action", Name: "My Action"}}
		Registry.Add(actionMenu)
		assert.True(t, actionMenu.HasAction)
	})
	t.Run("Getting menus", func(t *testing.T) {
		assert.Equal(t, rootMenu, Registry.GetByID(1))
		assert.Equal(t, childMenu, Registry.GetByID(2))
		assert.Nil(t, Registry.GetByID(100))
		assert.Equal(t, rootMenu, Registry.GetByXMLID("root_menu"))
		assert.Nil(t, Registry.GetByXMLID("unknown_menu"))
		// All returns all menus, including children
		all := Registry.All()
		assert.Len(t, all, 4)
		var xmlIDs []string
		for _, menu := range all {
			xmlIDs = append(xmlIDs, menu.XMLID)
		}
		sort.Strings(xmlIDs)
		assert.Equal(t, []string{"action_menu", "child_menu", "other_root_menu", "root_menu"}, xmlIDs)
	})
	t.Run("Sorting primitives", func(t *testing.T) {
		// Menus are sorted by sequence: action_menu (0), other_root_menu (5), root_menu (20)
		assert.Equal(t, "action_menu", Registry.Menus[0].XMLID)
		assert.True(t, Registry.Less(0, 1))
		Registry.Swap(0, 1)
		assert.Equal(t, "other_root_menu", Registry.Menus[0].XMLID)
		assert.Equal(t, "action_menu", Registry.Menus[1].XMLID)
		assert.False(t, Registry.Less(0, 1))
	})
}

func TestTranslatedName(t *testing.T) {
	t.Run("Menu without translations", func(t *testing.T) {
		menu := Menu{Name: "My Menu"}
		assert.Equal(t, "My Menu", menu.TranslatedName("fr"))
	})
	t.Run("Menu with translations", func(t *testing.T) {
		menu := Menu{Name: "My Menu", names: map[string]string{"fr": "Mon menu"}}
		assert.Equal(t, "Mon menu", menu.TranslatedName("fr"))
		assert.Equal(t, "My Menu", menu.TranslatedName("de"))
	})
}

func TestBootStrap(t *testing.T) {
	defer func() {
		resetRegistry()
		i18n.Langs = nil
	}()
	resetRegistry()
	i18n.Langs = []string{"fr"}
	i18n.LoadPOFile(filepath.Join("testdata", "fr.po"))
	actions.Registry.Add(&actions.Action{XMLID: "my_action", Name: "My Action"})

	t.Run("Bootstrapping menus", func(t *testing.T) {
		for _, def := range []string{rootMenuDef, childMenuDef, actionMenuDef} {
			element, err := xmlutils.XMLToElement(def)
			assert.Nil(t, err)
			LoadFromEtree(element)
		}
		BootStrap()
		assert.Len(t, Registry.All(), 3)
		assert.Equal(t, 1, Registry.Len())
	})
	t.Run("Parents and children should be linked", func(t *testing.T) {
		rootMenu := Registry.GetByXMLID("root_menu")
		assert.NotNil(t, rootMenu)
		assert.Nil(t, rootMenu.Parent)
		assert.True(t, rootMenu.HasChildren)
		assert.Equal(t, 2, rootMenu.Children.Len())
		// Children are sorted by sequence
		assert.Equal(t, "child_menu", rootMenu.Children.Menus[0].XMLID)
		assert.Equal(t, "action_menu", rootMenu.Children.Menus[1].XMLID)
		assert.Equal(t, rootMenu, Registry.GetByXMLID("child_menu").Parent)
	})
	t.Run("Menus with an action should get its name if they have none", func(t *testing.T) {
		actionMenu := Registry.GetByXMLID("action_menu")
		assert.True(t, actionMenu.HasAction)
		assert.Equal(t, "my_action", actionMenu.Action.XMLID)
		assert.Equal(t, "My Action", actionMenu.Name)
	})
	t.Run("Menus should be translated", func(t *testing.T) {
		assert.Equal(t, "Menu racine", Registry.GetByXMLID("root_menu").TranslatedName("fr"))
		assert.Equal(t, "Root Menu", Registry.GetByXMLID("root_menu").TranslatedName("de"))
		// Menus without a name are translated with their action's XMLID
		assert.Equal(t, "Mon action", Registry.GetByXMLID("action_menu").TranslatedName("fr"))
		// Untranslated menus keep their name
		assert.Equal(t, "Child Menu", Registry.GetByXMLID("child_menu").TranslatedName("fr"))
	})
	t.Run("Unknown parent should panic", func(t *testing.T) {
		resetRegistry()
		element, err := xmlutils.XMLToElement(unknownParentMenuDef)
		assert.Nil(t, err)
		LoadFromEtree(element)
		assert.Panics(t, BootStrap)
	})
}
