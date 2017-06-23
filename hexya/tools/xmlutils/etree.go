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

import "github.com/hexya-erp/hexya/hexya/tools/etree"

// ElementToXML returns the XML string of the given element and
// all its children.
func ElementToXML(element *etree.Element) string {
	doc := etree.NewDocument()
	doc.SetRoot(element)
	doc.IndentTabs()
	xmlStr, err := doc.WriteToString()
	if err != nil {
		log.Panic("Unable to marshal element", "error", err, "element", element)
	}
	return xmlStr
}

// XMLToElement parses the given xml string and returns the root node
func XMLToElement(xmlStr string) *etree.Element {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlStr); err != nil {
		log.Panic("Unable to parse XML", "error", err, "xml", xmlStr)
	}
	return doc.Root()
}

// FindNextSibling returns the next sibling of the given element
func FindNextSibling(element *etree.Element) *etree.Element {
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

// HasParentTag returns true if this element has at least
// one parent node with the given parent tag name
func HasParentTag(element *etree.Element, parent string) bool {
	for e := element.Parent(); e != nil; e = e.Parent() {
		if e.Tag == parent {
			return true
		}
	}
	return false
}
