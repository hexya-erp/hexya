// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package templates

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
	. "github.com/smartystreets/goconvey/convey"
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
	<xpath expr="//t[@t-esc='Email']" position="after">
		<t t-esc="Phone"/>
	</xpath>
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
<template inherit_id="my_other_id" id="new_base_view">
	<xpath expr="//t[@t-esc='Email']" position="after">
		<t t-raw="Fax"/>
	</xpath>
</template>
`

var tmplDef9 = `
<template inherit_id="new_base_view">
	<xpath expr="//t[@t-raw='Fax']" position="before">
		<t t-raw="Mobile"/>
	</xpath>
</template>
`

func loadView(xml string) {
	elt, err := xmlutils.XMLToElement(xml)
	if err != nil {
		panic(err)
	}
	LoadFromEtree(elt)
}

func TestViews(t *testing.T) {
	Convey("Creating Template 1", t, func() {
		Registry.collection = newCollection()
		loadView(tmplDef1)
		BootStrap()
		So(len(Registry.collection.templates), ShouldEqual, 1)
		So(Registry.collection.GetByID("my_id"), ShouldNotBeNil)
		template := Registry.collection.GetByID("my_id")
		So(template.ID, ShouldEqual, "my_id")
		So(template.Page, ShouldEqual, true)
		So(template.Optional, ShouldEqual, false)
		So(template.OptionalDefault, ShouldEqual, false)
		So(template.Priority, ShouldEqual, 16)
		So(string(template.p2Content), ShouldEqual,
			`{% set _1 = _0 %}
	<div>
		{% for line in lines %}<span>
			<h1>{{ line.UserName }}</h1>
			<label for="Age"/>
			<p>Hello World</p>
		</span>{% endfor %}
	</div>
`)
	})
	Convey("Creating Template 2", t, func() {
		Registry.collection = newCollection()
		loadView(tmplDef1)
		loadView(tmplDef2)
		BootStrap()
		So(len(Registry.collection.templates), ShouldEqual, 2)
		So(Registry.collection.GetByID("my_other_id"), ShouldNotBeNil)
		template := Registry.collection.GetByID("my_other_id")
		So(template.ID, ShouldEqual, "my_other_id")
		So(template.Page, ShouldEqual, false)
		So(template.Optional, ShouldEqual, true)
		So(template.OptionalDefault, ShouldEqual, true)
		So(template.Priority, ShouldEqual, 12)
		So(string(template.Content("")), ShouldEqual,
			`{% set _1 = _0 %}
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
	Convey("Inheriting Template 2", t, func() {
		Registry.collection = newCollection()
		loadView(tmplDef1)
		loadView(tmplDef2)
		loadView(tmplDef3)
		BootStrap()
		So(len(Registry.collection.templates), ShouldEqual, 2)
		So(Registry.collection.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.collection.GetByID("my_other_id"), ShouldNotBeNil)
		template1 := Registry.collection.GetByID("my_id")
		So(string(template1.Content("")), ShouldEqual,
			`{% set _1 = _0 %}
	<div>
		{% for line in lines %}<span>
			<h1>{{ line.UserName }}</h1>
			<label for="Age"/>
			<p>Hello World</p>
		</span>{% endfor %}
	</div>
`)
		template2 := Registry.collection.GetByID("my_other_id")
		So(string(template2.Content("")), ShouldEqual,
			`{% set _1 = _0 %}<div>
		<h1>Name</h1>
		<div name="position_info">
			{{ Function }}
		{{ CompanyName }}</div>
		<div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>`)
	})

	Convey("More inheritance on Template 2", t, func() {
		Registry.collection = newCollection()
		loadView(tmplDef1)
		loadView(tmplDef2)
		loadView(tmplDef3)
		loadView(tmplDef4)
		BootStrap()
		So(len(Registry.collection.templates), ShouldEqual, 2)
		So(Registry.collection.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.collection.GetByID("my_other_id"), ShouldNotBeNil)
		template2 := Registry.collection.GetByID("my_other_id")
		So(string(template2.Content("")), ShouldEqual,
			`{% set _1 = _0 %}<div>
		<h2>{{ Name }}</h2>
		<div name="position_info">
			{{ Function }}
		{{ CompanyName }}</div>
		<div>
			{{ Address }}
		</div><hr/><div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>`)

	})
	Convey("Modifying inherited modifications on Template 2", t, func() {
		Registry.collection = newCollection()
		loadView(tmplDef1)
		loadView(tmplDef2)
		loadView(tmplDef3)
		loadView(tmplDef4)
		loadView(tmplDef5)
		BootStrap()
		So(len(Registry.collection.templates), ShouldEqual, 2)
		So(Registry.collection.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.collection.GetByID("my_other_id"), ShouldNotBeNil)
		template2 := Registry.collection.GetByID("my_other_id")
		So(string(template2.Content("")), ShouldEqual,
			`{% set _1 = _0 %}<div>
		<h2>{{ Name }}</h2>
		<div name="position_info">
			{{ Function }}
		{{ CompanyName }}</div>
		<div name="address" string="Address">
			{{ Address }}
		</div><hr/><div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>`)
	})
	Convey("Create new base template from inheritance", t, func() {
		Registry.collection = newCollection()
		loadView(tmplDef1)
		loadView(tmplDef2)
		loadView(tmplDef3)
		loadView(tmplDef4)
		loadView(tmplDef5)
		loadView(tmplDef8)
		BootStrap()
		So(Registry.collection.GetByID("my_other_id"), ShouldNotBeNil)
		So(Registry.collection.GetByID("new_base_view"), ShouldNotBeNil)
		template2 := Registry.collection.GetByID("my_other_id")
		newTemplate := Registry.collection.GetByID("new_base_view")
		So(string(template2.Content("")), ShouldEqual,
			`{% set _1 = _0 %}<div>
		<h2>{{ Name }}</h2>
		<div name="position_info">
			{{ Function }}
		{{ CompanyName }}</div>
		<div name="address" string="Address">
			{{ Address }}
		</div><hr/><div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>`)
		So(string(newTemplate.Content("")), ShouldEqual,
			`{% set _1 = _0 %}<div>
		<h2>{{ Name }}</h2>
		<div name="position_info">
			{{ Function }}
		{{ CompanyName }}</div>
		<div name="address" string="Address">
			{{ Address }}
		</div><hr/><div name="contact_data">
			{{ Email }}{{ Fax|safe }}{{ Phone }}
		</div>
	</div>`)
	})
	Convey("Inheriting new base template from inheritance", t, func() {
		Registry.collection = newCollection()
		loadView(tmplDef1)
		loadView(tmplDef2)
		loadView(tmplDef3)
		loadView(tmplDef4)
		loadView(tmplDef5)
		loadView(tmplDef8)
		loadView(tmplDef9)
		BootStrap()
		So(Registry.collection.GetByID("my_other_id"), ShouldNotBeNil)
		So(Registry.collection.GetByID("new_base_view"), ShouldNotBeNil)
		template2 := Registry.collection.GetByID("my_other_id")
		newTemplate := Registry.collection.GetByID("new_base_view")
		So(string(template2.Content("")), ShouldEqual,
			`{% set _1 = _0 %}<div>
		<h2>{{ Name }}</h2>
		<div name="position_info">
			{{ Function }}
		{{ CompanyName }}</div>
		<div name="address" string="Address">
			{{ Address }}
		</div><hr/><div name="contact_data">
			{{ Email }}{{ Phone }}
		</div>
	</div>`)
		So(string(newTemplate.Content("")), ShouldEqual,
			`{% set _1 = _0 %}<div>
		<h2>{{ Name }}</h2>
		<div name="position_info">
			{{ Function }}
		{{ CompanyName }}</div>
		<div name="address" string="Address">
			{{ Address }}
		</div><hr/><div name="contact_data">
			{{ Email }}{{ Mobile|safe }}{{ Fax|safe }}{{ Phone }}
		</div>
	</div>`)
	})
}
