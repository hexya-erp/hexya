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

package xmlutils

import (
	"crypto/sha1"
	"encoding/xml"
	"io/ioutil"

	"github.com/hexya-erp/hexya/hexya/tools/logging"
)

var log *logging.Logger

type basicXML struct {
	XMLName xml.Name
	Data    string `xml:",innerxml"`
}

// ConcatXML concatenates the XML content of the files given by fileNames
// into a valid XML by importing all children of the root node into the
// root node of the first file. This function also returns the sha1sum of
// the result.
func ConcatXML(fileNames []string) ([]byte, [sha1.Size]byte) {
	var reStruct basicXML
	for _, fileName := range fileNames {
		var content basicXML
		cnt, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Panic("Unable to open file", "file", fileName, "error", err)
		}
		err = xml.Unmarshal(cnt, &content)
		if err != nil {
			log.Panic("Unable to parse file", "file", fileName, "error", err)
		}
		if reStruct.XMLName.Local == "" {
			reStruct.XMLName = content.XMLName
		}
		reStruct.Data += content.Data
	}
	res, err := xml.Marshal(reStruct)
	if err != nil {
		log.Panic("Unable to convert back to XML", "error", err)
	}
	return res, sha1.Sum(res)
}

func init() {
	log = logging.GetLogger("xmlutils")
}
