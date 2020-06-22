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
	"fmt"

	"github.com/beevik/etree"
)

// DocumentToXML returns the XML bytes of the given document
func DocumentToXML(doc *etree.Document) ([]byte, error) {
	doc.IndentTabs()
	xml, err := doc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal document: %s", err)
	}
	return xml, nil
}

// DocumentToXMLNoIndent returns the XML bytes of the given document
// without indenting the result.
//
// Use this function when the XML is HTML that needs to keep <tag></tag> syntax
func DocumentToXMLNoIndent(doc *etree.Document) ([]byte, error) {
	xml, err := doc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal document: %s", err)
	}
	return xml, nil
}

// ElementToXML returns the XML bytes of the given element and
// all its children.
func ElementToXML(element *etree.Element) ([]byte, error) {
	doc := etree.NewDocument()
	doc.SetRoot(element.Copy())
	doc.IndentTabs()
	xml, err := doc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal element: %s", err)
	}
	return xml, nil
}

// ElementToXMLNoIndent returns the XML bytes of the given element and
// all its children, without indenting the result.
//
// Use this function when the XML is HTML that needs to keep <tag></tag> syntax
func ElementToXMLNoIndent(element *etree.Element) ([]byte, error) {
	doc := etree.NewDocument()
	doc.SetRoot(element.Copy())
	xml, err := doc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal element: %s", err)
	}
	return xml, nil
}

// XMLToDocument parses the given xml string and returns an etree.Document
func XMLToDocument(xmlStr string) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlStr); err != nil {
		return nil, fmt.Errorf("unable to parse XML: %s", err)
	}
	return doc, nil
}

// XMLToElement parses the given xml string and returns the root node
func XMLToElement(xmlStr string) (*etree.Element, error) {
	doc, err := XMLToDocument(xmlStr)
	return doc.Root(), err
}

// NextSibling returns the next sibling of the given token or nil if this
// is the last token of its parent
func NextSibling(token etree.Token) etree.Token {
	var found bool
	for _, el := range token.Parent().Child {
		if found {
			return el
		}
		if el == token {
			found = true
		}
	}
	return nil
}

// HasParentTag returns true if this element has at least
// one ancestor node with the given parent tag name
func HasParentTag(element *etree.Element, parent string) bool {
	for e := element.Parent(); e != nil; e = e.Parent() {
		if e.Tag == parent {
			return true
		}
	}
	return false
}

// CopyElement deep copies the given element, setting it as root to a new document
func CopyElement(element *etree.Element) *etree.Element {
	el := element.Copy()
	doc := etree.NewDocument()
	doc.SetRoot(el)
	return el
}
