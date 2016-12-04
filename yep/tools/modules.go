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

package tools

import (
	"fmt"
	"io/ioutil"
)

// ListStaticFiles get all file names of the static files that are in
// the "server/static/*/<subDir>" directories.
// Returned
// If diskPath is true, returned file names are relative to the yep directory
// (e.g. yep/server/static/src/js/foo.js) otherwise file names are relative
// to the http root (e.g. /static/src/js/foo.js)
func ListStaticFiles(subDir string, modules []string, diskPath bool) []string {
	var res []string
	for _, module := range modules {
		dirName := fmt.Sprintf("yep/server/static/%s/%s", module, subDir)
		fileInfos, _ := ioutil.ReadDir(dirName)
		for _, fi := range fileInfos {
			if !fi.IsDir() {
				path := fmt.Sprintf("/static/%s/%s/%s", module, subDir, fi.Name())
				if diskPath {
					path = fmt.Sprintf("yep/server/%s", path)
				}
				res = append(res, path)
			}
		}
	}
	return res
}
