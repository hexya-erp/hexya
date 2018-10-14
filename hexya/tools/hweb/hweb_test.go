// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package hweb

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
	. "github.com/smartystreets/goconvey/convey"
)

var template1 = `
<root t-attf-class="toto_{{ name }}">
	<child1 t-att-tag="id or 42">
		<child2 t-att="{'a': &quot;aVal&quot;, 'b'': 234}"/>
		<child3 t-att="['c', 'dVal']"/>
	</child1>
</root>
<root2 t-attf-class="titi_{{ value }}">
</root2>`

func TestTranspileAttributes(t *testing.T) {
	Convey("Testing attribute transpilation", t, func() {
		doc, err := xmlutils.XMLToDocument(template1)
		if err != nil {
			panic(err)
		}
		So(func() { transpileAttributes(doc.ChildElements()) }, ShouldNotPanic)
		resXML, err := doc.WriteToString()
		So(err, ShouldBeNil)
		So(string(resXML), ShouldEqual, `
<root class="toto_{{ name }}">
	<child1 tag="{{ id or 42 }}">
		<child2 a="aVal" b="234"/>
		<child3 c="dVal"/>
	</child1>
</root>
<root2 class="titi_{{ value }}">
</root2>`)
	})
}

var template2 = `
<root>
	<child1>
		<p><t t-esc="my_var"/></p>
		<em t-raw="my_raw"/>
	</child1>
</root>
<root2>
	<t t-esc="42"/>
</root2>
<h2 t-raw="0"/>`

func TestTranspileOutput(t *testing.T) {
	Convey("Testing data output", t, func() {
		doc, err := xmlutils.XMLToDocument(template2)
		if err != nil {
			panic(err)
		}
		So(func() { transpileOutput(doc.ChildElements()) }, ShouldNotPanic)
		resXML, err := doc.WriteToString()
		So(err, ShouldBeNil)
		So(string(resXML), ShouldEqual, `
<root>
	<child1>
		<p>{{ my_var }}</p>
		<em>{{ my_raw|safe }}</em>
	</child1>
</root>
<root2>
	{{ 42 }}
</root2>
<h2>{{ _0|safe }}</h2>`)
	})
}

var template3 = `
<root>
	<child1>
		<t t-if="cond1 or cond2">
			<p t-if="cond4">Foo</p>
			<t t-else="">Dum</t>
		</t>
		<p t-elif="cond3">Bar</p>
		<h1 t-else="">Baz</h1>
		<span t-if="cond5">Hello World</span>
		<tag no="condition"/>
		<t t-if="cond6">
			<a href="somewhere">Hi</a>
		</t>
	</child1>
</root>
<r t-if="cond7" otherAttr="sth">Bonjour</r>
<t t-else=""><p>Bye</p>`

func TestTranspileConditionals(t *testing.T) {
	Convey("Testing conditionals", t, func() {
		doc, err := xmlutils.XMLToDocument(template3)
		if err != nil {
			panic(err)
		}
		So(transpileConditionals(doc.ChildElements()), ShouldBeNil)
		resXML, err := doc.WriteToString()
		So(err, ShouldBeNil)
		So(string(resXML), ShouldEqual, `
<root>
	<child1>
		{% if cond1 or cond2 %}
			{% if cond4 %}<p>Foo</p>
			{% else %}Dum{% endif %}
		
		{% elif cond3 %}<p>Bar</p>
		{% else %}<h1>Baz</h1>{% endif %}
		{% if cond5 %}<span>Hello World</span>{% endif %}
		<tag no="condition"/>
		{% if cond6 %}
			<a href="somewhere">Hi</a>
		{% endif %}
	</child1>
</root>
{% if cond7 %}<r otherAttr="sth">Bonjour</r>
{% else %}<p>Bye</p>{% endif %}`)
	})
}

var template4 = `
<root>
	<child1>
		<t t-foreach="[1, 2, 3]" t-as="i">
			<span>Hello World!</span>
			<p t-foreach="buzz" t-as="baz"><t t-esc="baz"/></p>
		</t>
		<h1 t-foreach="list" t-as="item" otherTag="sth">
			Waouh <t t-raw="item"/>
		</h1>
		<nofortag attr="foo"/>
	</child1>
</root>
<t t-foreach="list_again" t-as="item-again" otherAttr="I'm out!">
	<p>Bye</p>
</t>
`

func TestTranspileLoops(t *testing.T) {
	Convey("Testing loops", t, func() {
		doc, err := xmlutils.XMLToDocument(template4)
		if err != nil {
			panic(err)
		}
		So(transpileLoops(doc.ChildElements()), ShouldBeNil)
		resXML, err := doc.WriteToString()
		So(err, ShouldBeNil)
		So(string(resXML), ShouldEqual, `
<root>
	<child1>
		{% for i in [1, 2, 3] %}
			<span>Hello World!</span>
			{% for baz in buzz %}<p><t t-esc="baz"/></p>{% endfor %}
		{% endfor %}
		{% for item in list %}<h1 otherTag="sth">
			Waouh <t t-raw="item"/>
		</h1>{% endfor %}
		<nofortag attr="foo"/>
	</child1>
</root>
{% for item-again in list_again %}
	<p>Bye</p>
{% endfor %}
`)
	})
}

var template5 = `
<root>
	<child1>
		<t t-set="var1" t-value="my_value"/>
		<t t-set="var2">
	Hello world, with <mytag data="foo">bar</mytag>
		</t>
	</child1>
</root>
<t t-set="var3" t-value="other_value"/>
`

func TestTranspileVariables(t *testing.T) {
	Convey("Testing setting variables", t, func() {
		doc, err := xmlutils.XMLToDocument(template5)
		if err != nil {
			panic(err)
		}
		So(transpileVariables(doc.ChildElements()), ShouldBeNil)
		resXML, err := doc.WriteToString()
		So(err, ShouldBeNil)
		So(string(resXML), ShouldEqual, `
<root>
	<child1>
		{% set var1 = my_value %}
		{% macro var2() %}
	Hello world, with <mytag data="foo">bar</mytag>
		{% endmacro %}
	</child1>
</root>
{% set var3 = other_value %}
`)
	})
}

var template6 = `
<t t-set="var1" t-value="valueOuter"/>
<t t-call="subtemplate">
	<div>foo</div>
	<t t-set="var1" t-value="valueInner"/>
	<span>Bar</span>
	<t t-set="var2">
		<h1>Baz</h1>
		<t t-set="var4" t-value="value4"/>
	</t>
</t>
`

func TestTranspileCalls(t *testing.T) {
	Convey("Testing subtemplate calls", t, func() {
		doc, err := xmlutils.XMLToDocument(template6)
		if err != nil {
			panic(err)
		}
		So(transpileCalls(doc.ChildElements()), ShouldBeNil)
		doc.WriteSettings.CanonicalText = true
		resXML, err := doc.WriteToString()
		So(err, ShouldBeNil)
		So(string(resXML), ShouldEqual, `
<t t-set="var1" t-value="valueOuter"/>
{% with _0 = null %}<t t-set="var1" t-value="valueInner"/><t t-set="var2">
		<h1>Baz</h1>
		<t t-set="var4" t-value="value4"/>
	</t>{% macro _0() %}
	<div>foo</div>
	
	<span>Bar</span>
	
{% endmacro %}{% include "subtemplate" %}
{% endwith %}
`)
		So(transpileVariables(doc.ChildElements()), ShouldBeNil)
		resXML, err = doc.WriteToString()
		So(err, ShouldBeNil)
		So(string(resXML), ShouldEqual, `
{% set var1 = valueOuter %}
{% with _0 = null %}{% set var1 = valueInner %}{% macro var2() %}
		<h1>Baz</h1>
		{% set var4 = value4 %}
	{% endmacro %}{% macro _0() %}
	<div>foo</div>
	
	<span>Bar</span>
	
{% endmacro %}{% include "subtemplate" %}
{% endwith %}
`)

	})
}

var template7 = `
<a class="o_sub_menu_logo" t-attf-href="/web{% if debug %}?debug{ %endif %}">
	<span class="oe_logo_edit">Edit Company data</span>
	<img src='/web/binary/company_logo'/>
</a>
<div class="o_sub_menu_content">
	<t t-foreach="menu_data.children" t-as="menu">
		<div style="display: none" class="oe_secondary_menu" t-att-data-menu-parent="menu.id">
			<t t-foreach="menu.children" t-as="menu">
				<div class="oe_secondary_menu_section" t-att-data-menu-xmlid="menu.xmlid">
					<t t-if="menu.children"><t t-esc="menu.name"/></t>
					<t t-if="not menu.children"><t t-call="web.menu_link"/></t>
				</div>
				<t t-call="web.menu_secondary_submenu"/>
			</t>
		</div>
	</t>
</div>
`

func TestToPongo(t *testing.T) {
	Convey("Global ToPongo test", t, func() {
		res, err := ToPongo([]byte(template7))
		So(err, ShouldBeNil)
		So(string(res), ShouldEqual, `
<a class="o_sub_menu_logo" href="/web{% if debug %}?debug{ %endif %}">
	<span class="oe_logo_edit">Edit Company data</span>
	<img src="/web/binary/company_logo"/>
</a>
<div class="o_sub_menu_content">
	{% for menu in menu_data.children %}
		<div style="display: none" class="oe_secondary_menu" data-menu-parent="{{ menu.id }}">
			{% for menu in menu.children %}
				<div class="oe_secondary_menu_section" data-menu-xmlid="{{ menu.xmlid }}">
					{% if menu.children %}{{ menu.name }}{% endif %}
					{% if not menu.children %}{% with _0 = null %}{% macro _0() %}{% endmacro %}{% include "web.menu_link" %}
{% endwith %}{% endif %}
				</div>
				{% with _0 = null %}{% macro _0() %}{% endmacro %}{% include "web.menu_secondary_submenu" %}
{% endwith %}
			{% endfor %}
		</div>
	{% endfor %}
</div>
`)
	})
}
