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
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

type basicXML struct {
	XMLName xml.Name
	Data    string `xml:",innerxml"`
}

/*
ConcatXML concatenates the XML content of the files given by fileNames
into a valid XML by importing all children of the root node into the
root node of the first file.
*/
func ConcatXML(fileNames []string) ([]byte, [sha1.Size]byte) {
	var reStruct basicXML
	for _, fileName := range fileNames {
		var content basicXML
		cnt, _ := ioutil.ReadFile(fileName)
		err := xml.Unmarshal(cnt, &content)
		if err != nil {
			panic(fmt.Errorf("Unable to parse %s", fileName))
		}
		if reStruct.XMLName.Local == "" {
			reStruct.XMLName = content.XMLName
		}
		reStruct.Data += content.Data
	}
	res, err := xml.Marshal(reStruct)
	if err != nil {
		panic(fmt.Errorf("Unable to convert back to XML: %s", err))
	}
	return res, sha1.Sum(res)
}
