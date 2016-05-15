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
	"runtime"
)

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
RunPostInit runs successively all PostInit() func of all modules.
PostInit() functions are used for actions that need to be done after
bootstrapping the models.
*/
func RunPostInit() {
	PostInit()
	for _, module := range Modules {
		module.PostInit()
	}
}

/*
createModuleSymlinks create the symlinks of the given module in the
server directory.
*/
func createModuleSymlinks(mod *Module) {
	_, fileName, _, ok := runtime.Caller(2)
	if !ok {
		panic(fmt.Errorf("Unable to find caller"))
	}
	modulePathTmpl := path.Dir(fileName) + "/%s"
	serverPathTmpl := "yep/server/%s/%s"
	dirs := []string{"static", "templates", "data"}
	for _, dir := range dirs {
		srcPath := fmt.Sprintf(modulePathTmpl, dir)
		dstPath := fmt.Sprintf(serverPathTmpl, dir, mod.Name)
		if _, err := os.Stat(srcPath); err == nil {
			os.Symlink(srcPath, dstPath)
		}
	}
}
