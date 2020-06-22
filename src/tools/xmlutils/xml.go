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

// Package xmlutils provides utilities for working with XML in the
// context of the Hexya Framework.
package xmlutils

import (
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/beevik/etree"
)

type basicXML struct {
	XMLName xml.Name
	Data    string `xml:",innerxml"`
}

// ConcatXML concatenates the XML content of the files given by fileNames
// into a valid XML by importing all children of the root node into the
// root node of the first file. This function also returns the sha1sum of
// the result.
func ConcatXML(fileNames []string) ([]byte, [sha1.Size]byte, error) {
	var reStruct basicXML
	for _, fileName := range fileNames {
		var content basicXML
		cnt, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, [sha1.Size]byte{}, fmt.Errorf("unable to open XML file %s: %s", fileName, err)
		}
		err = xml.Unmarshal(cnt, &content)
		if err != nil {
			return nil, [sha1.Size]byte{}, fmt.Errorf("unable to parse XML file %s: %s", fileName, err)
		}
		if reStruct.XMLName.Local == "" {
			reStruct.XMLName = content.XMLName
		}
		reStruct.Data += content.Data
	}
	res, err := xml.Marshal(reStruct)
	if err != nil {
		return nil, [sha1.Size]byte{}, fmt.Errorf("unable to convert back to XML: %s", err)

	}
	return res, sha1.Sum(res), nil
}

// ApplyExtensions returns a copy of base with extension specs applied.
func ApplyExtensions(base *etree.Document, specs *etree.Document) (*etree.Document, error) {
	baseElem := base.Copy()
	for _, spec := range specs.ChildElements() {
		xpath, err := getInheritXPathFromSpec(spec)
		if err != nil {
			specBytes, _ := ElementToXML(spec)
			return nil, fmt.Errorf("error in spec %s: %s", string(specBytes), err)
		}
		nodeToModify := baseElem.FindElement(xpath)
		if nodeToModify == nil {
			return nil, fmt.Errorf("node not found in parent view: %s", xpath)
		}
		modifyAction := spec.SelectAttr("position")
		if modifyAction == nil {
			specBytes, _ := ElementToXML(spec)
			return nil, fmt.Errorf("spec should include 'position' attribute : %s", string(specBytes))
		}
		specChild := spec.Copy().Child
		if len(specChild) > 0 {
			if sp0, ok := specChild[0].(*etree.CharData); ok && strings.HasPrefix(sp0.Data, "\n") {
				// Remove first \n from spec if present
				sp0.Data = strings.TrimPrefix(sp0.Data, "\n")
			}
		}
		switch modifyAction.Value {
		case "before":
			for _ = range specChild {
				nodeToModify.Parent().InsertChild(nodeToModify, specChild[0])
			}
		case "after":
			nextNode := NextSibling(nodeToModify)
			for _ = range specChild {
				nodeToModify.Parent().InsertChild(nextNode, specChild[0])
			}
		case "replace":
			for _ = range specChild {
				nodeToModify.Parent().InsertChild(nodeToModify, specChild[0])
			}
			nodeToModify.Parent().RemoveChild(nodeToModify)
		case "inside":
			for _ = range specChild {
				nodeToModify.AddChild(specChild[0])
			}
		case "attributes":
			for _, node := range spec.FindElements("./attribute") {
				attrName := node.SelectAttr("name").Value
				nodeToModify.RemoveAttr(attrName)
				nodeToModify.CreateAttr(attrName, node.Text())
			}
		}
	}
	return baseElem, nil
}

// getInheritXPathFromSpec returns an XPath string that is suitable for
// searching the base view and find the node to modify.
func getInheritXPathFromSpec(spec *etree.Element) (string, error) {
	if spec.Tag == "xpath" {
		// We have an xpath expression, we take it
		return spec.SelectAttr("expr").Value, nil
	}
	if len(spec.Attr) < 1 || len(spec.Attr) > 2 {
		return "", errors.New("invalid view inherit spec")
	}
	var attrStr string
	for _, attr := range spec.Attr {
		if attr.Key != "position" {
			attrStr = fmt.Sprintf("[@%s='%s']", attr.Key, attr.Value)
			break
		}
	}
	return fmt.Sprintf("//%s%s", spec.Tag, attrStr), nil
}
