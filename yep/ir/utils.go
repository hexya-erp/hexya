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
	"github.com/beevik/etree"
	"github.com/npiganeau/yep/yep/tools/logging"
)

// elementToXML returns the XML string of the given element and
// all its children.
func elementToXML(element *etree.Element) string {
	doc := etree.NewDocument()
	doc.SetRoot(element)
	doc.IndentTabs()
	xmlStr, err := doc.WriteToString()
	if err != nil {
		logging.LogAndPanic(log, "Unable to marshal element", "error", err, "element", element)
	}
	return xmlStr
}

// xmlToElement parses the given xml string and returns the root node
func xmlToElement(xmlStr string) *etree.Element {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlStr); err != nil {
		logging.LogAndPanic(log, "Unable to parse XML", "error", "xml", xmlStr)
	}
	return doc.Root()
}

// findNextSibling returns the next sibling of the given element
func findNextSibling(element *etree.Element) *etree.Element {
	var found bool
	for _, el := range element.Parent().ChildElements() {
		if found {
			return el
		}
		if el == element {
			found = true
		}
	}
	return nil
}
