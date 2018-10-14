// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package hweb provides utilities for the HWeb templating system
// such as transpilation to Pongo2 templates.
package hweb

import (
	"errors"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
)

// ToPongo transpiles the HWeb src template to Pongo2 template
func ToPongo(src []byte) ([]byte, error) {
	doc, err := xmlutils.XMLToDocument(string(src))
	if err != nil {
		return nil, err
	}
	transpileOutput(doc.ChildElements())
	transpileAttributes(doc.ChildElements())
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
	return res, nil
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
				val = "_0"
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
func transpileAttributes(elts []*etree.Element) {
	for _, elt := range elts {
		for _, attr := range elt.Attr {
			switch {
			case strings.HasPrefix(attr.Key, "t-att-"):
				newKey := strings.TrimPrefix(attr.Key, "t-att-")
				elt.CreateAttr(newKey, fmt.Sprintf("{{ %s }}", attr.Value))
				elt.RemoveAttr(attr.Key)
			case strings.HasPrefix(attr.Key, "t-attf-"):
				newKey := strings.TrimPrefix(attr.Key, "t-attf-")
				elt.CreateAttr(newKey, attr.Value)
				elt.RemoveAttr(attr.Key)
			case attr.Key == "t-att":
				var pair []string
				switch {
				case strutils.StartsAndEndsWith(attr.Value, "{", "}"):
					pairs := strings.Split(strings.Trim(attr.Value, "{}"), ",")
					elt.RemoveAttr(attr.Key)
					for _, pair := range pairs {
						toks := strings.Split(pair, ":")
						key := strings.Trim(strings.TrimSpace(toks[0]), "'\"")
						val := strings.Trim(strings.TrimSpace(toks[1]), "'\"")
						elt.CreateAttr(key, val)
					}
				case strutils.StartsAndEndsWith(attr.Value, "[", "]"):
					pair = strings.Split(strings.Trim(attr.Value, "[]"), ",")
					fallthrough
				case strutils.StartsAndEndsWith(attr.Value, "(", ")"):
					if pair == nil {
						pair = strings.Split(strings.Trim(attr.Value, "()"), ",")
					}
					elt.RemoveAttr(attr.Key)
					key := strings.Trim(strings.TrimSpace(pair[0]), "'\"")
					val := strings.Trim(strings.TrimSpace(pair[1]), "'\"")

					elt.CreateAttr(key, val)
				}
			}
		}
		transpileAttributes(elt.ChildElements())
	}
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
				if loopClosed {
					return errors.New("t-elif found after t-else")
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
			var dir string
			switch loop[i].attr.Key {
			case "t-if":
				dir = fmt.Sprintf("{%% if %s %%}", loop[i].attr.Value)
			case "t-elif":
				dir = fmt.Sprintf("{%% elif %s %%}", loop[i].attr.Value)
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
			replaceTag(elt, []string{}, fmt.Sprintf("{%% set %s = %s %%}", fea.Value, val))
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
		endWithCD := etree.NewCharData(fmt.Sprintf("{%% include \"%s\" %%}\n{%% endwith %%}", fea.Value))
		elt.Parent().InsertChild(xmlutils.NextSibling(elt), endWithCD)
		beginWithCD := etree.NewCharData("{% with _0 = null %}")
		elt.Parent().InsertChild(elt, beginWithCD)
		for _, child := range elt.ChildElements() {
			if child.SelectAttr("t-set") != nil {
				extractToParent(child)
			}
		}
		endCD := etree.NewCharData("{% endmacro %}")
		elt.Parent().InsertChild(xmlutils.NextSibling(elt), endCD)
		replaceTag(elt, []string{}, "{% macro _0() %}")
	}
	return nil
}

// transpileSmartFields handles t-field attributes
func transpileSmartFields(elts []*etree.Element) {

}
