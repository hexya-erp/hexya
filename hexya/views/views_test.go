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
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
	. "github.com/smartystreets/goconvey/convey"
)

var viewDef1 = `
<view id="my_id" name="My View" model="User">
	<form>
		<group>
			<field name="UserName"/>
			<field name="Age"/>
		</group>
	</form>
</view>
`

var viewDef2 = `
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

var viewDef3 = `
<view inherit_id="my_other_id">
	<group name="position_info" position="inside">
		<field name="CompanyName"/>
	</group>
	<xpath expr="//field[@name='Email']" position="after">
		<field name="Phone"/>
	</xpath>
</view>
`

var viewDef4 = `
<view inherit_id="my_other_id">
	<group name="contact_data" position="before">
		<group>
			<field name="Address"/>
		</group>
		<hr/>
	</group>
	<h1 position="replace">
		<h2><field name="Name"/></h2>
	</h1>
</view>
`

var viewDef5 = `
<view inherit_id="my_other_id">
	<xpath expr="//field[@name='Address']/.." position="attributes">
		<attribute name="name">address</attribute>
		<attribute name="string">Address</attribute>
	</xpath>
</view>
`

var viewDef6 = `
<view id="my_tree_id" model="User">
	<tree>
		<field name="UserName"/>
		<field name="Age"/>
	</tree>
</view>
`

var viewDef7 = `
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

var viewDef8 = `
<view inherit_id="my_other_id" id="new_base_view">
	<xpath expr="//field[@name='Email']" position="after">
		<field name="Fax"/>
	</xpath>
</view>
`

var viewDef9 = `
<view inherit_id="new_base_view">
	<xpath expr="//field[@name='Fax']" position="attributes">
		<attribute name="widget">phone</attribute>
	</xpath>
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
		So(xmlutils.ElementToXML(view.Arch("")), ShouldEqual,
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
		So(xmlutils.ElementToXML(view.Arch("")), ShouldEqual,
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
		BootStrap()
		So(len(Registry.views), ShouldEqual, 2)
		So(Registry.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		view1 := Registry.GetByID("my_id")
		So(xmlutils.ElementToXML(view1.Arch("")), ShouldEqual,
			`<form>
	<group>
		<field name="UserName"/>
		<field name="Age"/>
	</group>
</form>
`)
		view2 := Registry.GetByID("my_other_id")
		So(xmlutils.ElementToXML(view2.Arch("")), ShouldEqual,
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
		BootStrap()
		So(len(Registry.views), ShouldEqual, 2)
		So(Registry.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		view2 := Registry.GetByID("my_other_id")
		So(xmlutils.ElementToXML(view2.Arch("")), ShouldEqual,
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
		BootStrap()
		So(len(Registry.views), ShouldEqual, 2)
		So(Registry.GetByID("my_id"), ShouldNotBeNil)
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		view2 := Registry.GetByID("my_other_id")
		So(xmlutils.ElementToXML(view2.Arch("")), ShouldEqual,
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
		So(view1.Type, ShouldEqual, ViewTypeForm)
		So(view2.Type, ShouldEqual, ViewTypeForm)
		So(view3.Type, ShouldEqual, ViewTypeTree)
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
		So(xmlutils.ElementToXML(view.Arch("")), ShouldEqual,
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
		viewCategories := view.SubViews["Categories"]
		So(viewCategories, ShouldHaveLength, 2)
		viewCategoriesForm := viewCategories[ViewTypeForm]
		So(viewCategoriesForm.ID, ShouldEqual, "embedded_form_childview_Categories_1")
		So(xmlutils.ElementToXML(viewCategoriesForm.Arch("")), ShouldEqual, `<form>
	<h1>This is my form</h1>
	<field name="Name"/>
	<field name="Color"/>
	<field name="Sequence"/>
</form>
`)
		viewCategoriesTree := viewCategories[ViewTypeTree]
		So(viewCategoriesTree.ID, ShouldEqual, "embedded_form_childview_Categories_0")
		So(xmlutils.ElementToXML(viewCategoriesTree.Arch("")), ShouldEqual, `<tree>
	<field name="Name"/>
	<field name="Color"/>
</tree>
`)

		viewGroups := view.SubViews["Groups"]
		So(viewGroups, ShouldHaveLength, 1)
		viewGroupsTree := viewGroups[ViewTypeTree]
		So(viewGroupsTree.ID, ShouldEqual, "embedded_form_childview_Groups_0")
		So(xmlutils.ElementToXML(viewGroupsTree.Arch("")), ShouldEqual, `<tree>
	<field name="Name"/>
	<field name="Active"/>
</tree>
`)
	})
	Convey("Testing GetViews functions", t, func() {
		allViews := Registry.GetAll()
		So(allViews, ShouldHaveLength, 4)
		userViews := Registry.GetAllViewsForModel("User")
		So(userViews, ShouldHaveLength, 3)
		userFirstView := Registry.GetFirstViewForModel("User", ViewTypeForm)
		So(userFirstView.ID, ShouldEqual, "my_id")
	})
	Convey("Testing default views", t, func() {
		soModel := models.NewModel("SaleOrder")
		soModel.AddFields(map[string]models.FieldDefinition{
			"Name": models.CharField{},
		})
		soSearch := Registry.GetFirstViewForModel("SaleOrder", ViewTypeSearch)
		So(xmlutils.ElementToXML(soSearch.arch), ShouldEqual, `<search>
	<field name="Name"/>
</search>
`)
		soTree := Registry.GetFirstViewForModel("SaleOrder", ViewTypeTree)
		So(xmlutils.ElementToXML(soTree.arch), ShouldEqual, `<tree>
	<field name="Name"/>
</tree>
`)
	})
	Convey("Create new base view from inheritance", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef8))
		BootStrap()
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		So(Registry.GetByID("new_base_view"), ShouldNotBeNil)
		view2 := Registry.GetByID("my_other_id")
		newView := Registry.GetByID("new_base_view")
		So(xmlutils.ElementToXML(view2.Arch("")), ShouldEqual,
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
		So(xmlutils.ElementToXML(newView.Arch("")), ShouldEqual,
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
		<field name="Fax"/>
		<field name="Phone"/>
	</group>
</form>
`)
	})
	Convey("Inheriting new base view from inheritance", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(viewDef9))
		BootStrap()
		So(Registry.GetByID("my_other_id"), ShouldNotBeNil)
		So(Registry.GetByID("new_base_view"), ShouldNotBeNil)
		view2 := Registry.GetByID("my_other_id")
		newView := Registry.GetByID("new_base_view")
		So(xmlutils.ElementToXML(view2.Arch("")), ShouldEqual,
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
		So(xmlutils.ElementToXML(newView.Arch("")), ShouldEqual,
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
		<field name="Fax" widget="phone"/>
		<field name="Phone"/>
	</group>
</form>
`)
	})

	Convey("Testing ViewRef objects", t, func() {
		userFormRef := MakeViewRef("my_id")
		Convey("Creating ViewRef instance", func() {
			So(userFormRef.ID(), ShouldEqual, "my_id")
			So(userFormRef.Name(), ShouldEqual, "My View")
			data, err := json.Marshal(userFormRef)
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `["my_id","My View"]`)
			val, err := userFormRef.Value()
			So(err, ShouldBeNil)
			So(val, ShouldEqual, driver.Value("my_id"))
		})
		Convey("Creating empty viewRef", func() {
			emptyVR := MakeViewRef("unknownID")
			So(emptyVR.ID(), ShouldEqual, "")
			So(emptyVR.Name(), ShouldEqual, "")
			data, err := json.Marshal(emptyVR)
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `null`)
			val, err := emptyVR.Value()
			So(err, ShouldBeNil)
			So(val, ShouldEqual, driver.Value(""))
		})
		Convey("Unmarshalling JSON viewRef", func() {
			data := []byte(`["view_id","View Name"]`)
			var vr ViewRef
			err := json.Unmarshal(data, &vr)
			So(err, ShouldBeNil)
			So(vr.ID(), ShouldEqual, "view_id")
			So(vr.Name(), ShouldEqual, "View Name")
		})
		Convey("Unmarshalling JSON empty viewRef", func() {
			data := []byte(`null`)
			var vr ViewRef
			err := json.Unmarshal(data, &vr)
			So(err, ShouldBeNil)
			So(vr.IsNull(), ShouldBeTrue)
		})
		Convey("Unmarshalling XML viewRef", func() {
			type stuff struct {
				Ref ViewRef `xml:"ref,attr"`
			}
			data := []byte(`<stuff ref="my_id"/>`)
			var st stuff
			err := xml.Unmarshal(data, &st)
			So(err, ShouldBeNil)
			So(st.Ref.ID(), ShouldEqual, "my_id")
			So(st.Ref.Name(), ShouldEqual, "My View")
		})
		Convey("Scanning viewRefs", func() {
			var vr ViewRef
			err := vr.Scan("my_id")
			So(err, ShouldBeNil)
			So(vr.ID(), ShouldEqual, "my_id")
			So(vr.Name(), ShouldEqual, "My View")

			err = vr.Scan([]byte("my_tree_id"))
			So(err, ShouldBeNil)
			So(vr.ID(), ShouldEqual, "my_tree_id")
			So(vr.Name(), ShouldEqual, "my.tree.id")
		})
	})
	Convey("Testing ViewTuple objects", t, func() {
		Convey("Marshalling a ViewTuple", func() {
			vt := ViewTuple{
				ID:   "my_id",
				Type: ViewTypeForm,
			}
			data, err := json.Marshal(vt)
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `["my_id","form"]`)
		})
		Convey("Unmarshalling ViewTuples", func() {
			data := []byte(`["my_tree_id","tree"]`)
			var vt ViewTuple
			err := json.Unmarshal(data, &vt)
			So(err, ShouldBeNil)
			So(vt.ID, ShouldEqual, "my_tree_id")
			So(vt.Type, ShouldEqual, ViewTypeTree)
		})
	})
}
