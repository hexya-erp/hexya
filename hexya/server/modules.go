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

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/menus"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/views"
)

var symlinkDirs = []string{"static", "templates", "data", "resources", "i18n"}

// A Module is a go package that implements business features.
// This struct is used to register modules.
type Module struct {
	Name     string
	PostInit func()
}

// A ModulesList is a list of Module objects
type ModulesList []*Module

// Names returns a list of all module names in this ModuleList.
func (ml *ModulesList) Names() []string {
	res := make([]string, len(*ml))
	for i, module := range *ml {
		res[i] = module.Name
	}
	return res
}

// Modules is the list of activated modules in the application
var Modules ModulesList

// RegisterModule registers the given module in the server
// This function should be called in the init() function of
// all Hexya Addons.
func RegisterModule(mod *Module) {
	createModuleSymlinks(mod)
	Modules = append(Modules, mod)
}

// createModuleSymlinks create the symlinks of the given module in the
// server directory.
func createModuleSymlinks(mod *Module) {
	_, fileName, _, ok := runtime.Caller(2)
	if !ok {
		log.Panic("Unable to find caller", "module", mod.Name)
	}
	for _, dir := range symlinkDirs {
		srcPath := filepath.Join(filepath.Dir(fileName), dir)
		dstPath := filepath.Join(generate.HexyaDir, "hexya", "server", dir, mod.Name)
		if _, err := os.Stat(srcPath); err == nil {
			os.Symlink(srcPath, dstPath)
		}
	}
}

// cleanModuleSymlinks removes all symlinks in the server symlink directories.
// Note that this function actually removes and recreates the symlink directories.
func cleanModuleSymlinks() {
	for _, dir := range symlinkDirs {
		dirPath := filepath.Join(generate.HexyaDir, "hexya", "server", dir)
		os.RemoveAll(dirPath)
		os.Mkdir(dirPath, 0775)
	}
}

// LoadInternalResources loads all data in the 'resources' directory, that are
// - views,
// - actions,
// - menu items
// Internal resources are defined in XML files.
func LoadInternalResources() {
	loadData("resources", "xml", loadXMLResourceFile)
}

// LoadDataRecords loads all the data records in the 'data' directory into the database.
// Data records are defined in CSV files.
func LoadDataRecords() {
	loadData("data", "csv", models.LoadCSVDataFile)
}

// LoadTranslations loads all translation data from the PO files in the 'i18n' directory
// into the translations registry.
func LoadTranslations(langs []string) {
	for _, mod := range Modules {
		dataDir := filepath.Join(generate.HexyaDir, "hexya", "server", "i18n", mod.Name)
		if _, err := os.Stat(dataDir); err != nil {
			// No resources dir in this module
			return
		}
		LoadModuleTranslations(dataDir, langs)
	}
}

// LoadModuleTranslations loads the PO files in the given directory for the given languages
func LoadModuleTranslations(i18nDir string, langs []string) {
	var poFiles []string
	for _, lang := range langs {
		dataFiles, err := filepath.Glob(fmt.Sprintf("%s/%s.po", i18nDir, lang))
		if err != nil {
			log.Panic("Unable to scan directory for data files", "dir", i18nDir, "type", "po", "error", err)
		}
		poFiles = append(poFiles, dataFiles...)
	}
	dataFilesSorted := sort.StringSlice(poFiles)
	dataFilesSorted.Sort()
	for _, dataFile := range dataFilesSorted {
		i18n.LoadPOFile(dataFile)
	}
}

// loadData loads the files in the given dir with the given extension (without .)
// using the loader function.
func loadData(dir, ext string, loader func(string)) {
	for _, mod := range Modules {
		dataDir := filepath.Join(generate.HexyaDir, "hexya", "server", dir, mod.Name)
		if _, err := os.Stat(dataDir); err != nil {
			// No resources dir in this module
			continue
		}
		dataFiles, err := filepath.Glob(fmt.Sprintf("%s/*.%s", dataDir, ext))
		if err != nil {
			log.Panic("Unable to scan directory for data files", "dir", dataDir, "type", ext, "error", err)
		}
		dataFilesSorted := sort.StringSlice(dataFiles)
		dataFilesSorted.Sort()
		for _, dataFile := range dataFilesSorted {
			loader(dataFile)
		}
	}
}

// loadXMLResourceFile loads the data from an XML data file into memory.
func loadXMLResourceFile(fileName string) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fileName); err != nil {
		log.Panic("Error loading XML data file", "file", fileName, "error", err)
	}
	for _, dataTag := range doc.FindElements("hexya/data") {
		for _, object := range dataTag.ChildElements() {
			switch object.Tag {
			case "view":
				views.LoadFromEtree(object)
			case "action":
				actions.LoadFromEtree(object)
			case "menuitem":
				menus.LoadFromEtree(object)
			default:
				log.Panic("Unknown XML tag", "tag", object.Tag)
			}
		}
	}
}
