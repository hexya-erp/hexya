// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package templates

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flosch/pongo2"
	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/tools/xmlutils"
)

var tmplDef1 = `
<template id="my_id" page="True">
	<div>
		<span t-foreach="lines" t-as="line">
			<h1 t-esc="line.UserName"/>
			<label for="Age"/>
			<p>Hello World</p>
		</span>
	</div>
</template>
`

var tmplDef2 = `
<template id="my_other_id" priority="12" optional="enabled">
	<div>
		<h1>Name</h1>
		<div name="position_info">
			<t t-esc="Function"/>
		</div>
		<div name="contact_data">
			<t t-esc="Email"/>
		</div>
	</div>
</template>
`

var tmplDef3 = `
<template inherit_id="my_other_id">
	<div name="position_info" position="inside">
	<t t-esc="CompanyName"/>
		</div>
	<xpath expr="//t[@t-esc='Email']" position="after"><t t-esc="Phone"/></xpath>
</template>
`

var tmplDef4 = `
<template inherit_id="my_other_id">
	<div name="contact_data" position="before">
<div>
			<t t-esc="Address"/>
		</div>
		<hr/>
		</div>
	<h1 position="replace">
<h2><t t-esc="Name"/></h2>
	</h1>
</template>
`

var tmplDef5 = `
<template inherit_id="my_other_id">
	<xpath expr="//t[@t-esc='Address']/.." position="attributes">
		<attribute name="name">address</attribute>
		<attribute name="string">Address</attribute>
	</xpath>
</template>
`

var tmplDef8 = `
<template inherit_id="my_other_id" id="new_base_view" priority="13" optional="disabled" page="True">
	<xpath expr="//t[@t-esc='Email']" position="after"><t t-raw="Fax"/></xpath>
</template>
`

var tmplDef9 = `
<template inherit_id="new_base_view">
	<xpath expr="//t[@t-raw='Fax']" position="before"><t t-raw="Mobile"/></xpath>
</template>
`

var tmplDef10 = `
<template inherit_id="my_other_id">
	<xpath expr="." position="inside">
	<div>1</div>
	<div>2</div>
	</xpath>
</template>
`

func loadTemplate(xml string) {
	elt, err := xmlutils.XMLToElement(xml)
	if err != nil {
		panic(err)
	}
	LoadFromEtree(elt)
}

func TestTemplates(t *testing.T) {
	t.Run("Setting two languages", func(t *testing.T) {
		i18n.Langs = []string{"fr", "de"}
	})
	t.Run("Creating Template 1", func(t *testing.T) {
		loadTemplate(tmplDef1)
		BootStrap()
		assert.EqualValues(t, len(Registry.collection.templates), 1)
		assert.NotNil(t, Registry.collection.GetByID("my_id"))
		template := Registry.collection.GetByID("my_id")
		assert.EqualValues(t, template.ID, "my_id")
		assert.EqualValues(t, template.Page, true)
		assert.EqualValues(t, template.Optional, false)
		assert.EqualValues(t, template.OptionalDefault, false)
		assert.EqualValues(t, template.Priority, 16)
		assert.EqualValues(t, string(template.p2Content), `{% set _1 = _0 %}
	<div>
		{% for line in lines %}<span>
			<h1>{{ line.UserName }}</h1>
			<label for="Age"/>
			<p>Hello World</p>
		</span>{% endfor %}
	</div>
`)
	})
	t.Run("Creating Template 2", func(t *testing.T) {
		loadTemplate(tmplDef1)
		loadTemplate(tmplDef2)
		BootStrap()
		assert.EqualValues(t, len(Registry.collection.templates), 2)
		assert.NotNil(t, Registry.collection.GetByID("my_other_id"))
		template := Registry.collection.GetByID("my_other_id")
		assert.EqualValues(t, template.ID, "my_other_id")
		assert.EqualValues(t, template.Page, false)
		assert.EqualValues(t, template.Optional, true)
		assert.EqualValues(t, template.OptionalDefault, true)
		assert.EqualValues(t, template.Priority, 12)
		assert.EqualValues(t, string(template.Content("")), `{% set _1 = _0 %}
	<div>
		<h1>Name</h1>
		<div name="position_info">
			{{ Function }}
		</div>
		<div name="contact_data">
			{{ Email }}
		</div>
	</div>
`)
	})
	t.Run("Inheriting Template 2", func(t *testing.T) {
		loadTemplate(tmplDef1)
		loadTemplate(tmplDef2)
		loadTemplate(tmplDef3)
		BootStrap()
		assert.EqualValues(t, len(Registry.collection.templates), 2)
		assert.NotNil(t, Registry.collection.GetByID("my_id"))
		assert.NotNil(t, Registry.collection.GetByID("my_other_id"))
		template1 := Registry.collection.GetByID("my_id")
		assert.EqualValues(t, string(template1.Content("")), `{% set _1 = _0 %}
	<div>
		{% for line in lines %}<span>
			<h1>{{ line.UserName }}</h1>
			<label for="Age"/>
			<p>Hello World</p>
		</span>{% endfor %}
	</div>
`)
		template2 := Registry.collection.GetByID("my_other_id")
		assert.EqualValues(t, string(template2.Content("")), `{% set _1 = _0 %}
	<div>
		<h1>Name</h1>
		<div name="position_info">
			{{ Function }}
			{{ CompanyName }}
		</div>
		<div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>
`)
	})

	t.Run("More inheritance on Template 2", func(t *testing.T) {
		loadTemplate(tmplDef1)
		loadTemplate(tmplDef2)
		loadTemplate(tmplDef3)
		loadTemplate(tmplDef4)
		BootStrap()
		assert.EqualValues(t, len(Registry.collection.templates), 2)
		assert.NotNil(t, Registry.collection.GetByID("my_id"))
		assert.NotNil(t, Registry.collection.GetByID("my_other_id"))
		template2 := Registry.collection.GetByID("my_other_id")
		assert.EqualValues(t, string(template2.Content("")), `{% set _1 = _0 %}
	<div>
		<h2>{{ Name }}</h2>
	
		<div name="position_info">
			{{ Function }}
			{{ CompanyName }}
		</div>
		<div>
			{{ Address }}
		</div>
		<hr/>
		<div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>
`)

	})
	t.Run("Modifying inherited modifications on Template 2", func(t *testing.T) {
		loadTemplate(tmplDef1)
		loadTemplate(tmplDef2)
		loadTemplate(tmplDef3)
		loadTemplate(tmplDef4)
		loadTemplate(tmplDef5)
		BootStrap()
		assert.EqualValues(t, len(Registry.collection.templates), 2)
		assert.NotNil(t, Registry.collection.GetByID("my_id"))
		assert.NotNil(t, Registry.collection.GetByID("my_other_id"))
		template2 := Registry.collection.GetByID("my_other_id")
		assert.EqualValues(t, string(template2.Content("")), `{% set _1 = _0 %}
	<div>
		<h2>{{ Name }}</h2>
	
		<div name="position_info">
			{{ Function }}
			{{ CompanyName }}
		</div>
		<div name="address" string="Address">
			{{ Address }}
		</div>
		<hr/>
		<div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>
`)
	})
	t.Run("Create new base template from inheritance", func(t *testing.T) {
		loadTemplate(tmplDef1)
		loadTemplate(tmplDef2)
		loadTemplate(tmplDef3)
		loadTemplate(tmplDef4)
		loadTemplate(tmplDef5)
		loadTemplate(tmplDef8)
		BootStrap()
		assert.NotNil(t, Registry.collection.GetByID("my_other_id"))
		assert.NotNil(t, Registry.collection.GetByID("new_base_view"))
		template2 := Registry.collection.GetByID("my_other_id")
		newTemplate := Registry.collection.GetByID("new_base_view")
		assert.EqualValues(t, string(template2.Content("")), `{% set _1 = _0 %}
	<div>
		<h2>{{ Name }}</h2>
	
		<div name="position_info">
			{{ Function }}
			{{ CompanyName }}
		</div>
		<div name="address" string="Address">
			{{ Address }}
		</div>
		<hr/>
		<div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>
`)
		assert.EqualValues(t, template2.Priority, 12)
		assert.False(t, template2.Page)
		assert.True(t, template2.Optional)
		assert.True(t, template2.OptionalDefault)
		assert.EqualValues(t, newTemplate.Priority, 13)
		assert.True(t, newTemplate.Page)
		assert.True(t, newTemplate.Optional)
		assert.False(t, newTemplate.OptionalDefault)
		assert.EqualValues(t, string(newTemplate.Content("")), `{% set _1 = _0 %}
	<div>
		<h2>{{ Name }}</h2>
	
		<div name="position_info">
			{{ Function }}
			{{ CompanyName }}
		</div>
		<div name="address" string="Address">
			{{ Address }}
		</div>
		<hr/>
		<div name="contact_data">
			{{ Email }}{{ Fax|safe }}{{ Phone }}
		</div>
	</div>
`)
	})
	t.Run("Inheriting new base template from inheritance", func(t *testing.T) {
		Registry.collection = newCollection()
		loadTemplate(tmplDef1)
		loadTemplate(tmplDef2)
		loadTemplate(tmplDef3)
		loadTemplate(tmplDef4)
		loadTemplate(tmplDef5)
		loadTemplate(tmplDef8)
		loadTemplate(tmplDef9)
		BootStrap()
		assert.NotNil(t, Registry.collection.GetByID("my_other_id"))
		assert.NotNil(t, Registry.collection.GetByID("new_base_view"))
		template2 := Registry.collection.GetByID("my_other_id")
		newTemplate := Registry.collection.GetByID("new_base_view")
		assert.EqualValues(t, string(template2.Content("")), `{% set _1 = _0 %}
	<div>
		<h2>{{ Name }}</h2>
	
		<div name="position_info">
			{{ Function }}
			{{ CompanyName }}
		</div>
		<div name="address" string="Address">
			{{ Address }}
		</div>
		<hr/>
		<div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>
`)
		assert.EqualValues(t, string(newTemplate.Content("")), `{% set _1 = _0 %}
	<div>
		<h2>{{ Name }}</h2>
	
		<div name="position_info">
			{{ Function }}
			{{ CompanyName }}
		</div>
		<div name="address" string="Address">
			{{ Address }}
		</div>
		<hr/>
		<div name="contact_data">
			{{ Email }}{{ Mobile|safe }}{{ Fax|safe }}{{ Phone }}
		</div>
	</div>
`)
	})
	t.Run("Inherited modifications on root element", func(t *testing.T) {
		loadTemplate(tmplDef1)
		loadTemplate(tmplDef2)
		loadTemplate(tmplDef3)
		loadTemplate(tmplDef4)
		loadTemplate(tmplDef5)
		loadTemplate(tmplDef10)
		BootStrap()
		assert.NotNil(t, Registry.collection.GetByID("my_other_id"))
		template2 := Registry.collection.GetByID("my_other_id")
		assert.EqualValues(t, string(template2.Content("")), `{% set _1 = _0 %}
	<div>
		<h2>{{ Name }}</h2>
	
		<div name="position_info">
			{{ Function }}
			{{ CompanyName }}
		</div>
		<div name="address" string="Address">
			{{ Address }}
		</div>
		<hr/>
		<div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>
	<div>1</div>
	<div>2</div>
	`)
	})
	t.Run("Testing gin loading", func(t *testing.T) {
		inst := Registry.Instance("my_id", pongo2.Context{
			"lines": []map[string]string{
				{"UserName": "jsmith"},
				{"UserName": "wsmith"},
			}})
		w := httptest.NewRecorder()
		inst.Render(w)
		body, _ := ioutil.ReadAll(w.Result().Body)
		assert.EqualValues(t, string(body), `
	<div>
		<span>
			<h1>jsmith</h1>
			<label for="Age"/>
			<p>Hello World</p>
		</span><span>
			<h1>wsmith</h1>
			<label for="Age"/>
			<p>Hello World</p>
		</span>
	</div>
`)
	})
}
