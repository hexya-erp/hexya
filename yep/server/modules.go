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
	"path"
	"path/filepath"
	"runtime"

	"github.com/beevik/etree"
	"github.com/npiganeau/yep/yep/ir"
	"github.com/npiganeau/yep/yep/tools"
)

var symlinkDirs = []string{"static", "templates", "data", "views"}

type Module struct {
	Name     string
	PostInit func()
}

var Modules []*Module

/*
RegisterModules registers the given module in the server
This function should be called in the init() function of
all YEP Addons.
*/
func RegisterModule(mod *Module) {
	createModuleSymlinks(mod)
	Modules = append(Modules, mod)
}

/*
createModuleSymlinks create the symlinks of the given module in the
server directory.
*/
func createModuleSymlinks(mod *Module) {
	_, fileName, _, ok := runtime.Caller(2)
	if !ok {
		tools.LogAndPanic(log, "Unable to find caller", "module", mod.Name)
	}
	for _, dir := range symlinkDirs {
		srcPath := fmt.Sprintf(path.Dir(fileName)+"/%s", dir)
		dstPath := fmt.Sprintf("yep/server/%s/%s", dir, mod.Name)
		if _, err := os.Stat(srcPath); err == nil {
			os.Symlink(srcPath, dstPath)
		}
	}
}

/*
cleanModuleSymlinks removes all symlinks in the server symlink directories.
Note that this function actually removes and recreates the symlink directories.
*/
func cleanModuleSymlinks() {
	for _, dir := range symlinkDirs {
		dirPath := fmt.Sprintf("yep/server/%s", dir)
		os.RemoveAll(dirPath)
		os.Mkdir(dirPath, 0775)
	}
}

/*
LoadInternalResources loads all data in the 'views' directory, that are
- views,
- actions,
- menu items,
*/
func LoadInternalResources() {
	for _, mod := range Modules {
		dataDir := fmt.Sprintf("yep/server/views/%s", mod.Name)
		if _, err := os.Stat(dataDir); err != nil {
			// No data dir in this module
			continue
		}
		loadData(dataDir)
	}
}

/*
loadData loads the data defined in the given data directory.
*/
func loadData(dataDir string) {
	dataFiles, err := filepath.Glob(dataDir + "/*")
	if err != nil {
		tools.LogAndPanic(log, "Unable to scan directory for data files", "dir", dataDir)
	}
	for _, dataFile := range dataFiles {
		if path.Ext(dataFile) == ".xml" {
			loadXMLDataFile(dataFile)
		} else if path.Ext(dataFile) != ".csv" {
			loadCSVDataFile(dataFile)
		}
	}
}

/*
loadXMLDataFile loads the data from an XML data file into memory.
*/
func loadXMLDataFile(fileName string) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fileName); err != nil {
		tools.LogAndPanic(log, "Error loading XML data file", "file", fileName, "error", err)
	}
	//var noupdate bool
	for _, dataTag := range doc.FindElements("yep/data") {
		//noupdateStr := dataTag.SelectAttrValue("noupdate", "false")
		//if strings.ToLower(noupdateStr) == "true" {
		//	noupdate = true
		//}
		for _, object := range dataTag.ChildElements() {
			switch object.Tag {
			case "view":
				ir.LoadViewFromEtree(object)
			case "action":
				ir.LoadActionFromEtree(object)
			case "menuitem":
				ir.LoadMenuFromEtree(object)
			case "record":
			default:
				tools.LogAndPanic(log, "Unknown XML tag", "tag", object.Tag)
			}
		}
	}
}

/*
loadCSVDataFile loads the data from a CSV data file into memory.
*/
func loadCSVDataFile(fileName string) {
	// TODO
}
