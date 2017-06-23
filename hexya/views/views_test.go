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

package views

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
	. "github.com/smartystreets/goconvey/convey"
)

var viewDef1 string = `
<view id="my_id" name="My View" model="User">
	<form>
		<group>
			<field name="UserName"/>
			<field name="Age"/>
		</group>
	</form>
</view>
`

var viewDef2 string = `
<view id="my_other_id" model="Partner" priority="12">
	<form>
		<h1><field name="Name"/></h1>
		<group name="position_info">
			<field name="Function"/>
		</group>
		<group name="contact_data">
			<field name="Email"/>
		</group>
	</form>
</view>
`

var viewDef3 string = `
<view inherit_id="my_other_id">
	<group name="position_info" position="inside">
		<field name="CompanyName"/>
	</group>
	<xpath expr="//field[@name='Email']" position="after">
		<field name="Phone"/>
	</group>
</view>
`

var viewDef4 string = `
<view inherit_id="my_other_id">
	<group name="contact_data" position="before">
		<group>
			<field name="Address"/>
		</group>
		<hr/>
	</group>
	<h1 position="replace">
		<h2><field name="Name"/></h2>
	</group>
</view>
`

var viewDef5 string = `
<view inherit_id="my_other_id">
	<xpath expr="//field[@name='Address']/.." position="attributes">
		<attribute name="name">address</attribute>
		<attribute name="string">Address</attribute>
	</xpath>
</view>
`

var viewDef6 string = `
<view id="my_tree_id" model="User">
	<tree>
		<field name="UserName"/>
		<field name="Age"/>
	</tree>
</view>
`

var viewDef7 string = `
<view id="embedded_form" model="User">
	<form>
		<field name="Name"/>
		<field name="Age"/>
		<field name="Categories">
			<tree>
				<field name="Name"/>
				<field name="Color"/>
			</tree>
			<!-- Comment -->
			<form>
				<h1>This is my form</h1>
				<field name="Name"/>
				<field name="Color"/>
				<field name="Sequence"/>
			</form>
		</field>
		<field name="Groups">
			<tree>
				<field name="Name"/>
				<field name="Active"/>
			</tree>
		</field>
	</form>
</view>
`

func TestViews(t *testing.T) {
	Convey("Creating View 1", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef1))
		So(len(Registry.views), ShouldEqual, 1)
		So(Registry.GetByID("my_id"), ShouldNotBeNil)
		view := Registry.GetByID("my_id")
		So(view.ID, ShouldEqual, "my_id")
		So(view.Name, ShouldEqual, "My View")
		So(view.Model, ShouldEqual, "User")
		So(view.Priority, ShouldEqual, 16)
		So(view.Arch, ShouldEqual,
			`<form>
	<group>
		<field name="UserName"/>
		<field name="Age"/>
	</group>
</form>
`)
	})
	Convey("Creating View 2", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef2))
		So(len(Registry.views), ShouldEqual, 2)
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		view := Registry.GetByID("my_other_id")
		So(view.ID, ShouldEqual, "my_other_id")
		So(view.Name, ShouldEqual, "my.other.id")
		So(view.Model, ShouldEqual, "Partner")
		So(view.Priority, ShouldEqual, 12)
		So(view.Arch, ShouldEqual,
			`<form>
	<h1>
		<field name="Name"/>
	</h1>
	<group name="position_info">
		<field name="Function"/>
	</group>
	<group name="contact_data">
		<field name="Email"/>
	</group>
</form>
`)
	})
	Convey("Inheriting View 2", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef3))
		So(len(Registry.views), ShouldEqual, 2)
		So(Registry.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		view1 := Registry.GetByID("my_id")
		So(view1.Arch, ShouldEqual,
			`<form>
	<group>
		<field name="UserName"/>
		<field name="Age"/>
	</group>
</form>
`)
		view2 := Registry.GetByID("my_other_id")
		So(view2.Arch, ShouldEqual,
			`<form>
	<h1>
		<field name="Name"/>
	</h1>
	<group name="position_info">
		<field name="Function"/>
		<field name="CompanyName"/>
	</group>
	<group name="contact_data">
		<field name="Email"/>
		<field name="Phone"/>
	</group>
</form>
`)
	})
	Convey("More inheritance on View 2", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef4))
		So(len(Registry.views), ShouldEqual, 2)
		So(Registry.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		view2 := Registry.GetByID("my_other_id")
		So(view2.Arch, ShouldEqual,
			`<form>
	<h2>
		<field name="Name"/>
	</h2>
	<group name="position_info">
		<field name="Function"/>
		<field name="CompanyName"/>
	</group>
	<group>
		<field name="Address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="Email"/>
		<field name="Phone"/>
	</group>
</form>
`)
	})
	Convey("Modifying inherited modifications on View 2", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef5))
		So(len(Registry.views), ShouldEqual, 2)
		So(Registry.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		view2 := Registry.GetByID("my_other_id")
		So(view2.Arch, ShouldEqual,
			`<form>
	<h2>
		<field name="Name"/>
	</h2>
	<group name="position_info">
		<field name="Function"/>
		<field name="CompanyName"/>
	</group>
	<group name="address" string="Address">
		<field name="Address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="Email"/>
		<field name="Phone"/>
	</group>
</form>
`)
	})
	Convey("Bootstrapping views", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef6))
		BootStrap()
		view1 := Registry.GetByID("my_id")
		view2 := Registry.GetByID("my_other_id")
		view3 := Registry.GetByID("my_tree_id")
		So(view1, ShouldNotBeNil)
		So(view2, ShouldNotBeNil)
		So(view3, ShouldNotBeNil)
		So(view1.Type, ShouldEqual, VIEW_TYPE_FORM)
		So(view2.Type, ShouldEqual, VIEW_TYPE_FORM)
		So(view3.Type, ShouldEqual, VIEW_TYPE_TREE)
	})
	Convey("Testing embedded views", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef7))
		BootStrap()
		So(len(Registry.views), ShouldEqual, 4)
		So(Registry.GetByID("embedded_form"), ShouldNotBeNil)
		So(Registry.GetByID("embedded_form_childview_1"), ShouldBeNil)
		So(Registry.GetByID("embedded_form_childview_2"), ShouldBeNil)
		view := Registry.GetByID("embedded_form")
		So(view.ID, ShouldEqual, "embedded_form")
		So(view.Arch, ShouldEqual,
			`<form>
	<field name="Name"/>
	<field name="Age"/>
	<field name="Categories"/>
	<field name="Groups"/>
</form>
`)
		So(view.SubViews, ShouldHaveLength, 2)
		So(view.SubViews, ShouldContainKey, "Categories")
		So(view.SubViews, ShouldContainKey, "Groups")
		viewsCategories := view.SubViews["Categories"]
		So(viewsCategories, ShouldHaveLength, 2)
		viewCategoriesForm := viewsCategories[VIEW_TYPE_FORM]
		So(viewCategoriesForm.ID, ShouldEqual, "embedded_form_childview_Categories_1")
		So(viewCategoriesForm.Arch, ShouldEqual, `<form>
	<h1>This is my form</h1>
	<field name="Name"/>
	<field name="Color"/>
	<field name="Sequence"/>
</form>
`)
		viewCategoriesTree := viewsCategories[VIEW_TYPE_TREE]
		So(viewCategoriesTree.ID, ShouldEqual, "embedded_form_childview_Categories_0")
		So(viewCategoriesTree.Arch, ShouldEqual, `<tree>
	<field name="Name"/>
	<field name="Color"/>
</tree>
`)

		viewsGroups := view.SubViews["Groups"]
		So(viewsGroups, ShouldHaveLength, 1)
		viewGroupsTree := viewsGroups[VIEW_TYPE_TREE]
		So(viewGroupsTree.ID, ShouldEqual, "embedded_form_childview_Groups_0")
		So(viewGroupsTree.Arch, ShouldEqual, `<tree>
	<field name="Name"/>
	<field name="Active"/>
</tree>
`)
	})
}
