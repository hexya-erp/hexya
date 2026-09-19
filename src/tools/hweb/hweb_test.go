// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package hweb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/tools/xmlutils"
)

var template1 = `
<root t-attf-class="toto_{{ name }}">
	<child1 t-att-tag="id | default:42">
	</child1>
</root>
<root2 t-attf-class="titi_{{ value }}" t-att-attr="hi">
</root2>`
var template11 = `
<tag t-att="(&quot;a&quot;, &quot;b&quot;)"/>
`

func TestTranspileAttributes(t *testing.T) {
	t.Run("Testing attribute transpilation", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template1)
		if err != nil {
			panic(err)
		}
		assert.Nil(t, transpileAttributes(doc.ChildElements()))
		resXML, err := doc.WriteToString()
		assert.Nil(t, err)
		assert.EqualValues(t, string(resXML), `
<root class="toto_{{ name }}">
	<child1 tag="{{ id | default:42 }}">
	</child1>
</root>
<root2 class="titi_{{ value }}" attr="{{ hi }}">
</root2>`)
	})
	t.Run("Invalid values should fail", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template11)
		if err != nil {
			panic(err)
		}
		assert.NotNil(t, transpileAttributes(doc.ChildElements()))
		assert.EqualValues(t, transpileAttributes(doc.ChildElements()).Error(), `hweb does not manage t-att attributes (t-att with value '("a", "b")')`)
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
	t.Run("Testing data output", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template2)
		if err != nil {
			panic(err)
		}
		assert.NotPanics(t, func() { transpileOutput(doc.ChildElements()) })
		resXML, err := doc.WriteToString()
		assert.Nil(t, err)
		assert.EqualValues(t, string(resXML), `
<root>
	<child1>
		<p>{{ my_var }}</p>
		<em>{{ my_raw|safe }}</em>
	</child1>
</root>
<root2>
	{{ 42 }}
</root2>
<h2>{{ _1|safe }}</h2>`)
	})
}

var (
	template3 = `
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
<t t-else=""><p>Bye</p></t>`
	template31 = `
<t t-elif="cond">Foo</t>`
	template32 = `
<t t-else="">Bar</t>`
)

func TestTranspileConditionals(t *testing.T) {
	t.Run("Testing conditionals", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template3)
		if err != nil {
			panic(err)
		}
		assert.Nil(t, transpileConditionals(doc.ChildElements()))
		resXML, err := doc.WriteToString()
		assert.Nil(t, err)
		assert.EqualValues(t, string(resXML), `
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
	t.Run("Wrong if/elif/else order should fail", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template31)
		if err != nil {
			panic(err)
		}
		assert.NotNil(t, transpileConditionals(doc.ChildElements()))
		assert.EqualValues(t, transpileConditionals(doc.ChildElements()).Error(), "t-elif found without t-if")
		doc, err = xmlutils.XMLToDocument(template32)
		if err != nil {
			panic(err)
		}
		assert.NotNil(t, transpileConditionals(doc.ChildElements()))
		assert.EqualValues(t, transpileConditionals(doc.ChildElements()).Error(), "t-else found without t-if")
	})
}

var (
	template4 = `
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
	template41 = `
<t t-foreach="lines">
Foo
</t>
`
)

func TestTranspileLoops(t *testing.T) {
	t.Run("Testing loops", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template4)
		if err != nil {
			panic(err)
		}
		assert.Nil(t, transpileLoops(doc.ChildElements()))
		resXML, err := doc.WriteToString()
		assert.Nil(t, err)
		assert.EqualValues(t, string(resXML), `
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
	t.Run("t-foreach without t-as should fail", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template41)
		if err != nil {
			panic(err)
		}
		assert.NotNil(t, transpileLoops(doc.ChildElements()))
		assert.EqualValues(t, transpileLoops(doc.ChildElements()).Error(), "t-foreach without t-as")
	})
}

var (
	template5 = `
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
	template51 = `
<p t-set="foo" t-value="booh"/>`
	template52 = `
<t t-set="bar"/>`
)

func TestTranspileVariables(t *testing.T) {
	t.Run("Testing setting variables", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template5)
		if err != nil {
			panic(err)
		}
		assert.Nil(t, transpileVariables(doc.ChildElements()))
		resXML, err := doc.WriteToString()
		assert.Nil(t, err)
		assert.EqualValues(t, string(resXML), `
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
	t.Run("Wrong t-set tags should fail", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template51)
		if err != nil {
			panic(err)
		}
		assert.NotNil(t, transpileVariables(doc.ChildElements()))
		assert.EqualValues(t, transpileVariables(doc.ChildElements()).Error(), "t-set attribute set on non 't' XML tag")

		doc, err = xmlutils.XMLToDocument(template52)
		if err != nil {
			panic(err)
		}
		assert.NotNil(t, transpileVariables(doc.ChildElements()))
		assert.EqualValues(t, transpileVariables(doc.ChildElements()).Error(), "t-set without t-value nor body")
	})
}

var (
	template6 = `
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
	template61 = `
<p t-call="foo"/>`
)

func TestTranspileCalls(t *testing.T) {
	t.Run("Testing subtemplate calls", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template6)
		if err != nil {
			panic(err)
		}
		assert.Nil(t, transpileCalls(doc.ChildElements()))
		doc.WriteSettings.CanonicalText = true
		resXML, err := doc.WriteToString()
		assert.Nil(t, err)
		assert.EqualValues(t, string(resXML), `
<t t-set="var1" t-value="valueOuter"/>
{% with _0 = null %}<t t-set="var2">
		<h1>Baz</h1>
		<t t-set="var4" t-value="value4"/>
	</t>{% macro _0() %}
	<div>foo</div>
	
	<span>Bar</span>
	
{% endmacro %}{% set __hexya_template_name = "subtemplate" %}{% include __hexya_template_name with var1 = valueInner %}
{% endwith %}
`)
		assert.Nil(t, transpileVariables(doc.ChildElements()))
		resXML, err = doc.WriteToString()
		assert.Nil(t, err)
		assert.EqualValues(t, string(resXML), `
{% set var1 = valueOuter %}
{% with _0 = null %}{% macro var2() %}
		<h1>Baz</h1>
		{% set var4 = value4 %}
	{% endmacro %}{% macro _0() %}
	<div>foo</div>
	
	<span>Bar</span>
	
{% endmacro %}{% set __hexya_template_name = "subtemplate" %}{% include __hexya_template_name with var1 = valueInner %}
{% endwith %}
`)

	})
	t.Run("t-call on non t tag should fail", func(t *testing.T) {
		doc, err := xmlutils.XMLToDocument(template61)
		if err != nil {
			panic(err)
		}
		assert.NotNil(t, transpileCalls(doc.ChildElements()))
		assert.EqualValues(t, transpileCalls(doc.ChildElements()).Error(), "t-call attribute set on non 't' XML tag")
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
	t.Run("Global ToPongo test", func(t *testing.T) {
		res, err := ToPongo([]byte(template7))
		assert.Nil(t, err)
		assert.EqualValues(t, string(res), `{% set _1 = _0 %}
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
					{% if not menu.children %}{% with _0 = null %}{% macro _0() %}{% endmacro %}{% set __hexya_template_name = "web.menu_link" %}{% include __hexya_template_name  %}
{% endwith %}{% endif %}
				</div>
				{% with _0 = null %}{% macro _0() %}{% endmacro %}{% set __hexya_template_name = "web.menu_secondary_submenu" %}{% include __hexya_template_name  %}
{% endwith %}
			{% endfor %}
		</div>
	{% endfor %}
</div>
`)
	})
	t.Run("Malformed templates should fail", func(t *testing.T) {
		_, err := ToPongo([]byte("<a"))
		assert.NotNil(t, err)
		assert.EqualValues(t, err.Error(), "unable to parse XML: XML syntax error on line 1: unexpected EOF")

		_, err = ToPongo([]byte(template31))
		assert.NotNil(t, err)
		assert.EqualValues(t, err.Error(), "t-elif found without t-if")

		_, err = ToPongo([]byte(template41))
		assert.NotNil(t, err)
		assert.EqualValues(t, err.Error(), "t-foreach without t-as")

		_, err = ToPongo([]byte(template51))
		assert.NotNil(t, err)
		assert.EqualValues(t, err.Error(), "t-set attribute set on non 't' XML tag")

		_, err = ToPongo([]byte(template61))
		assert.NotNil(t, err)
		assert.EqualValues(t, err.Error(), "t-call attribute set on non 't' XML tag")
	})
}
