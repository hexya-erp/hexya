// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package server

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/menus"
	"github.com/hexya-erp/hexya/src/templates"
	"github.com/hexya-erp/hexya/src/views"
)

// withModules sets the given modules as the module list of the application
// for the duration of the test and restores the previous list afterwards.
func withModules(t *testing.T, modules ...*Module) {
	t.Helper()
	origModules := Modules
	Modules = modules
	t.Cleanup(func() { Modules = origModules })
}

func TestModulesList(t *testing.T) {
	withModules(t)
	t.Run("Registering modules", func(t *testing.T) {
		assert.Len(t, Modules, 0)
		RegisterModule(&Module{Name: "base"})
		RegisterModule(&Module{Name: "web"})
		assert.Len(t, Modules, 2)
	})
	t.Run("Names should return all module names", func(t *testing.T) {
		assert.Equal(t, []string{"base", "web"}, Modules.Names())
	})
	t.Run("Names of an empty list", func(t *testing.T) {
		var ml ModulesList
		assert.Equal(t, []string{}, ml.Names())
	})
}

func TestInitModules(t *testing.T) {
	var preInit, postInit int
	withModules(t,
		&Module{Name: "base", PreInit: func() { preInit++ }, PostInit: func() { postInit++ }},
		// Module without PreInit nor PostInit should be skipped silently
		&Module{Name: "web"})
	t.Run("PreInit should call all modules PreInit funcs", func(t *testing.T) {
		PreInit()
		assert.Equal(t, 1, preInit)
		assert.Equal(t, 0, postInit)
	})
	t.Run("PostInit should call all modules PostInit funcs", func(t *testing.T) {
		PostInit()
		assert.Equal(t, 1, preInit)
		assert.Equal(t, 1, postInit)
	})
}

func TestLoadInternalResources(t *testing.T) {
	t.Run("Loading views, actions, menus and templates", func(t *testing.T) {
		withModules(t, &Module{Name: "testmodule"}, &Module{Name: "unknownmodule"})
		LoadInternalResources("testdata")
		view := views.Registry.GetByID("test_view_id")
		assert.NotNil(t, view)
		assert.Equal(t, "Test View", view.Name)
		action, ok := actions.Registry.GetByXMLID("test_action_id")
		assert.True(t, ok)
		assert.Equal(t, "Test Action", action.Name)
		tmpl, err := templates.Registry.FromCache("test_template_id")
		assert.Nil(t, err)
		assert.NotNil(t, tmpl)
		menus.BootStrap()
		menu := menus.Registry.GetByXMLID("test_menu_id")
		assert.NotNil(t, menu)
		assert.Equal(t, "Test Menu", menu.Name)
	})
	t.Run("Unknown XML tags should panic", func(t *testing.T) {
		withModules(t, &Module{Name: "badmodule"})
		assert.Panics(t, func() { LoadInternalResources("testdata") })
	})
	t.Run("Unparseable XML files should panic", func(t *testing.T) {
		assert.Panics(t, func() { loadXMLResourceFile(filepath.Join("testdata", "unknown.xml")) })
	})
}

func TestLoadRecords(t *testing.T) {
	// Modules without data nor demo directory should be skipped without
	// trying to reach the database.
	withModules(t, &Module{Name: "testmodule"})
	assert.NotPanics(t, func() { LoadDataRecords("testdata") })
	assert.NotPanics(t, func() { LoadDemoRecords("testdata") })
}

func TestLoadTranslations(t *testing.T) {
	withModules(t, &Module{Name: "testmodule"}, &Module{Name: "unknownmodule"})
	t.Run("Loading the translations of all modules", func(t *testing.T) {
		LoadTranslations("testdata", []string{"fr", "de"})
		assert.Equal(t, "Menu de test", i18n.TranslateResourceItem("fr", "test_menu_id", "Test Menu"))
		assert.Equal(t, "Test Menu", i18n.TranslateResourceItem("de", "test_menu_id", "Test Menu"))
	})
	t.Run("ALL should be replaced by the list of all languages", func(t *testing.T) {
		assert.NotPanics(t, func() { LoadTranslations("testdata", []string{"all"}) })
		assert.Equal(t, "Menu de test", i18n.TranslateResourceItem("fr", "test_menu_id", "Test Menu"))
	})
	t.Run("Unknown i18n directories should be ignored", func(t *testing.T) {
		assert.NotPanics(t, func() { LoadModuleTranslations(filepath.Join("testdata", "unknown"), []string{"fr"}) })
	})
}
