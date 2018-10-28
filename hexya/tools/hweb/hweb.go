// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package hweb provides utilities for the HWeb templating system
// such as transpilation to Pongo2 templates.
package hweb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/beevik/etree"
	"github.com/flosch/pongo2"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
)

// TemplateSet allows you to create your own group of templates with their own
// global context (which is shared among all members of the set) and their own
// configuration.
// It's useful for a separation of different kind of templates
// (e. g. web templates vs. mail templates).
type TemplateSet = pongo2.TemplateSet

// A Template to be rendered with server's c.HTML()
type Template = pongo2.Template

// The Context with which to render a HWeb template
type Context = pongo2.Context

// NewSet can be used to create sets with different kind of templates
// (e. g. web from mail templates), with different globals or
// other configurations.
func NewSet(name string, loader pongo2.TemplateLoader) *TemplateSet {
	return pongo2.NewSet(name, loader)
}

// Must panics, if a Template couldn't successfully parsed. This is how you
// would use it:
//     var baseTemplate = pongo2.Must(pongo2.FromFile("templates/base.html"))
func Must(tpl *Template, err error) *Template {
	return pongo2.Must(tpl, err)
}

// ToPongo transpiles the HWeb src template to Pongo2 template
func ToPongo(src []byte) ([]byte, error) {
	doc, err := xmlutils.XMLToDocument(string(src))
	if err != nil {
		return nil, err
	}
	doc.InsertChild(doc.Child[0], &etree.CharData{Data: "{% set _1 = _0 %}"})
	transpileOutput(doc.ChildElements())
	if err = transpileAttributes(doc.ChildElements()); err != nil {
		return nil, err
	}
	if err = transpileConditionals(doc.ChildElements()); err != nil {
		return nil, err
	}
	if err = transpileLoops(doc.ChildElements()); err != nil {
		return nil, err
	}
	if err = transpileCalls(doc.ChildElements()); err != nil {
		return nil, err
	}
	if err = transpileVariables(doc.ChildElements()); err != nil {
		return nil, err
	}
	transpileSmartFields(doc.ChildElements())
	doc.WriteSettings.CanonicalText = true
	res, err := doc.WriteToBytes()
	if err != nil {
		return nil, err
	}
	res = unescapeXMLEntities(res)
	return res, nil
}

// escapeXMLEntities mark xml entities as __[entity]__ so that they can be
// unescaped after XML rendering to valid pongo with unescapeXMLEntities.
//
// This function should be called on text in Pongo2 tags
func escapeXMLEntities(src string) string {
	res := strings.Replace(src, ">", "&gt;", -1)
	res = strings.Replace(res, "<", "&lt;", -1)
	res = strings.Replace(res, "\"", "&quot;", -1)
	res = strings.Replace(res, "'", "&apos;", -1)
	res = strings.Replace(res, "&gt;", "__&gt;__", -1)
	res = strings.Replace(res, "&lt;", "__&lt;__", -1)
	res = strings.Replace(res, "&quot;", "__&quot;__", -1)
	res = strings.Replace(res, "&apos;", "__&apos;__", -1)
	return res
}

// unescapeXMLEntities takes marked entities (as __[entity]__) and
// replace them with the real character.
func unescapeXMLEntities(src []byte) []byte {
	res := bytes.Replace(src, []byte("__&amp;gt;__"), []byte(">"), -1)
	res = bytes.Replace(res, []byte("__&amp;lt;__"), []byte("<"), -1)
	res = bytes.Replace(res, []byte("__&amp;quot;__"), []byte("\""), -1)
	res = bytes.Replace(res, []byte("__&amp;apos;__"), []byte("'"), -1)
	return res
}

// replaceTag replaces the given xml tag in its parents children by the given
// django directive dir. If the xml tag is not t, it removes attributes given
// by attrs
func replaceTag(el *etree.Element, attrs []string, dir string) {
	cd := &etree.CharData{Data: dir}
	switch el.Tag {
	case "t":
		el.Parent().InsertChild(el, cd)
		for len(el.Child) != 0 {
			extractToParent(el.Child[0])
		}
		el.Parent().RemoveChild(el)
	default:
		el.Parent().InsertChild(el, cd)
		for _, key := range attrs {
			el.RemoveAttr(key)
		}
	}
}

// extractToParent extracts this token and
// attach it to its grand parent, just before its parent.
// If tok has no grand parent, just leave it where it is.
func extractToParent(tok etree.Token) {
	if tok.Parent().Parent() == nil {
		return
	}
	tok.Parent().Parent().InsertChild(tok.Parent(), tok)
}

// transpileOutput modifies given elements with t-esc or t-raw tags
func transpileOutput(elts []*etree.Element) {
	for _, elt := range elts {
		transpileOutput(elt.ChildElements())
		for _, attr := range elt.Attr {
			var format string
			switch attr.Key {
			case "t-esc":
				format = "{{ %s }}"
			case "t-raw":
				format = "{{ %s|safe }}"
			default:
				continue
			}
			val := attr.Value
			if val == "0" {
				val = "_1"
			}
			text := fmt.Sprintf(format, val)
			switch elt.Tag {
			case "t":
				cd := &etree.CharData{Data: text}
				elt.Parent().InsertChild(elt, cd)
				elt.Parent().RemoveChild(elt)
			default:
				elt.RemoveAttr(attr.Key)
				elt.SetText(text)
			}
		}
	}
}

// transpileAttributes modifies dynamic attributes of all given elements,
// i.e. t-att, t-att-xxx, t-attf-yyy
func transpileAttributes(elts []*etree.Element) error {
	for _, elt := range elts {
		attrs := make([]etree.Attr, len(elt.Attr))
		copy(attrs, elt.Attr)
		for _, attr := range attrs {
			switch {
			case strings.HasPrefix(attr.Key, "t-att-"):
				newKey := strings.TrimPrefix(attr.Key, "t-att-")
				var attrValue string
				if attr.Value != "" {
					attrValue = fmt.Sprintf("{{ %s }}", escapeXMLEntities(attr.Value))
				}
				elt.CreateAttr(newKey, attrValue)
				elt.RemoveAttr(attr.Key)
			case strings.HasPrefix(attr.Key, "t-attf-"):
				newKey := strings.TrimPrefix(attr.Key, "t-attf-")
				elt.CreateAttr(newKey, escapeXMLEntities(attr.Value))
				elt.RemoveAttr(attr.Key)
			case attr.Key == "t-att":
				var data interface{}
				err := json.Unmarshal([]byte(attr.Value), &data)
				if err != nil {
					return fmt.Errorf("unable to unmarshal %s: %s", attr.Value, err.Error())
				}
				switch d := data.(type) {
				case []interface{}:
					if len(d)%2 != 0 {
						return fmt.Errorf("attribute list %s should have an even number of values", attr.Value)
					}
					for i := 0; i < len(d); i += 2 {
						elt.CreateAttr(fmt.Sprintf("%s", d[i]), fmt.Sprintf("%v", d[i+1]))
					}
				case map[string]interface{}:
					// We sort keys to have be deterministic
					var keys []string
					for k := range d {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						elt.CreateAttr(k, fmt.Sprintf("%v", d[k]))
					}
				default:
					return fmt.Errorf("unable to manage attribute %s with value %s", attr.Key, attr.Value)
				}
				elt.RemoveAttr(attr.Key)
			}
		}
		if err := transpileAttributes(elt.ChildElements()); err != nil {
			return err
		}
	}
	return nil
}

// transpileConditionals extracts all elements of elts with t-if, t-elif, t-else
// attributes to set Pongo conditionals instead
func transpileConditionals(elts []*etree.Element) error {
	type loopElem struct {
		attr etree.Attr
		elem *etree.Element
	}
	var (
		loops      [][]*loopElem
		lastLoop   []*loopElem
		loopClosed bool
		isLoopTag  bool
	)
	for _, elt := range elts {
		transpileConditionals(elt.ChildElements())
		isLoopTag = false
		for _, attr := range elt.Attr {
			switch attr.Key {
			case "t-if":
				if len(lastLoop) > 0 {
					loops = append(loops, lastLoop)
				}
				loopClosed = false
				isLoopTag = true
				lastLoop = []*loopElem{{
					attr: attr,
					elem: elt,
				}}
			case "t-elif":
				if len(lastLoop) == 0 {
					return errors.New("t-elif found without t-if")
				}
				isLoopTag = true
				lastLoop = append(lastLoop, &loopElem{
					attr: attr,
					elem: elt,
				})
			case "t-else":
				if len(lastLoop) == 0 {
					return errors.New("t-else found without t-if")
				}
				loopClosed = true
				isLoopTag = true
				lastLoop = append(lastLoop, &loopElem{
					attr: attr,
					elem: elt,
				})
			}
		}
		if loopClosed || !isLoopTag {
			if len(lastLoop) > 0 {
				loops = append(loops, lastLoop)
				lastLoop = nil
			}
			loopClosed = true
		}
	}
	if len(lastLoop) > 0 {
		loops = append(loops, lastLoop)
	}

	for _, loop := range loops {
		lastTag := loop[len(loop)-1]
		endCD := etree.NewCharData("{% endif %}")
		lastTag.elem.Parent().InsertChild(xmlutils.NextSibling(lastTag.elem), endCD)

		for i := 0; i < len(loop); i++ {
			cond := escapeXMLEntities(loop[i].attr.Value)
			var dir string
			switch loop[i].attr.Key {
			case "t-if":
				dir = fmt.Sprintf("{%% if %s %%}", cond)
			case "t-elif":
				dir = fmt.Sprintf("{%% elif %s %%}", cond)
			case "t-else":
				dir = "{% else %}"
			}
			replaceTag(loop[i].elem, []string{loop[i].attr.Key}, dir)
		}
	}
	return nil
}

// transpileLoops extracts all elements of elts with t-foreach
// attributes to set Pongo loops instead
func transpileLoops(elts []*etree.Element) error {
	for _, elt := range elts {
		transpileLoops(elt.ChildElements())
		fea := elt.SelectAttr("t-foreach")
		if fea == nil {
			continue
		}
		aea := elt.SelectAttr("t-as")
		if aea == nil {
			return errors.New("t-foreach without t-as")
		}
		endCD := etree.NewCharData("{% endfor %}")
		elt.Parent().InsertChild(xmlutils.NextSibling(elt), endCD)
		text := fmt.Sprintf("{%% for %s in %s %%}", aea.Value, fea.Value)
		replaceTag(elt, []string{"t-foreach", "t-as"}, text)
	}
	return nil
}

// transpileVariables modifies all children elements of elt with t-set tags
func transpileVariables(elts []*etree.Element) error {
	for _, elt := range elts {
		transpileVariables(elt.ChildElements())
		fea := elt.SelectAttr("t-set")
		if fea == nil {
			continue
		}
		if elt.Tag != "t" {
			return errors.New("t-set attribute set on non 't' XML tag")
		}
		val := elt.SelectAttrValue("t-value", "")
		switch {
		case val != "":
			replaceTag(elt, []string{}, fmt.Sprintf("{%% set %s = %s %%}", fea.Value, escapeXMLEntities(val)))
		case len(elt.Child) > 0:
			endCD := etree.NewCharData("{% endmacro %}")
			elt.Parent().InsertChild(xmlutils.NextSibling(elt), endCD)
			text := fmt.Sprintf("{%% macro %s() %%}", fea.Value)
			replaceTag(elt, []string{}, text)
		default:
			return errors.New("t-set without t-value nor body")
		}
	}
	return nil
}

// transpileCalls modifies all given elements with t-call attributes
// to set Pongo include directive instead
func transpileCalls(elts []*etree.Element) error {
	for _, elt := range elts {
		transpileCalls(elt.ChildElements())
		fea := elt.SelectAttr("t-call")
		if fea == nil {
			continue
		}
		if elt.Tag != "t" {
			return errors.New("t-call attribute set on non 't' XML tag")
		}
		beginWithCD := etree.NewCharData("{% with _0 = null %}")
		elt.Parent().InsertChild(elt, beginWithCD)
		vars := make(map[string]string)
		for _, child := range elt.ChildElements() {
			if child.SelectAttr("t-set") != nil {
				if child.SelectAttr("t-value") != nil {
					vars[child.SelectAttrValue("t-set", "")] = child.SelectAttrValue("t-value", "")
					elt.RemoveChild(child)
					continue
				}
				extractToParent(child)
			}
		}
		var withs []string
		for k, v := range vars {
			withs = append(withs, fmt.Sprintf("%s = %s", k, v))
		}
		with := strings.Join(withs, " ")
		if with != "" {
			with = "with " + with
		}
		// We set a __hexya_template_name variable in order to lazy load the template (avoids infinite recursion)
		endWithCD := etree.NewCharData(fmt.Sprintf("{%% set __hexya_template_name = \"%s\" %%}{%% include __hexya_template_name %s %%}\n{%% endwith %%}", fea.Value, with))
		elt.Parent().InsertChild(xmlutils.NextSibling(elt), endWithCD)
		endCD := etree.NewCharData("{% endmacro %}")
		elt.Parent().InsertChild(xmlutils.NextSibling(elt), endCD)
		replaceTag(elt, []string{}, "{% macro _0() %}")
	}
	return nil
}

// transpileSmartFields handles t-field attributes
func transpileSmartFields(elts []*etree.Element) {

}
