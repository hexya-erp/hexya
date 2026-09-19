// Copyright 2016 Nicolas Piganeau. All Rights Reserved.
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

	"github.com/stretchr/testify/assert"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/tools/xmlutils"
)

var viewDef1 = `
<view id="my_id" name="My View" model="User">
	<form>
		<group>
			<field name="UserName"/>
			<label for="Age"/>
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
		<field name="UserName"/>
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
var viewDef71 = `
<view inherit_id="embedded_form">
	<field name="UserName" position="attributes">
		<attribute name="required">1</attribute>
	</field>
	<xpath expr="//field[@name='Categories']/form/field[@name='Name']" position="attributes">
		<attribute name="readonly">1</attribute>
	</xpath>
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

var viewDef10 = `
<view id="search_view" model="User">
	<search>
		<field name="UserName"/>
	</search>
</view>
`

func documentToXMLString(elt *etree.Document) string {
	xmlData, err := xmlutils.DocumentToXML(elt)
	if err != nil {
		panic(err)
	}
	return string(xmlData)
}

func loadView(xml string) {
	elt, err := xmlutils.XMLToElement(xml)
	if err != nil {
		panic(err)
	}
	LoadFromEtree(elt)
}

func TestViews(t *testing.T) {
	t.Run("Creating View 1", func(t *testing.T) {
		loadView(viewDef1)
		assert.EqualValues(t, len(Registry.views), 1)
		assert.NotNil(t, Registry.GetByID("my_id"))
		view := Registry.GetByID("my_id")
		assert.EqualValues(t, view.ID, "my_id")
		assert.EqualValues(t, view.Name, "My View")
		assert.EqualValues(t, view.Model, "User")
		assert.EqualValues(t, view.Priority, 16)
		assert.EqualValues(t, documentToXMLString(view.Arch("")), `<form>
	<group>
		<field name="UserName"/>
		<label for="Age"/>
		<field name="Age"/>
	</group>
</form>
`)
	})
	t.Run("Creating View 2", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		assert.EqualValues(t, len(Registry.views), 2)
		assert.NotNil(t, Registry.GetByID("my_other_id"))
		view := Registry.GetByID("my_other_id")
		assert.EqualValues(t, view.ID, "my_other_id")
		assert.EqualValues(t, view.Name, "my.other.id")
		assert.EqualValues(t, view.Model, "Partner")
		assert.EqualValues(t, view.Priority, 12)
		assert.EqualValues(t, documentToXMLString(view.Arch("")), `<form>
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
	t.Run("Bootstrapping views before models should panic", func(t *testing.T) {
		assert.Panics(t, BootStrap)
	})
	t.Run("Creating models and boostrap them", func(t *testing.T) {
		group := models.NewModel("Group")
		category := models.NewModel("Category")
		user := models.NewModel("User")
		partner := models.NewModel("Partner")
		user.NewMethod("OnChangeAge", func(rc *models.RecordCollection) *models.ModelData {
			return models.NewModelData(rc.Model())
		})
		group.AddFields(map[string]models.FieldDefinition{
			"Name":   fields.Char{},
			"Active": fields.Boolean{},
		})
		category.AddFields(map[string]models.FieldDefinition{
			"Name":     fields.Char{},
			"Color":    fields.Integer{},
			"Sequence": fields.Integer{},
		})
		user.AddFields(map[string]models.FieldDefinition{
			"UserName": fields.Char{},
			"Age":      fields.Integer{OnChange: models.Registry.MustGet("User").Methods().MustGet("OnChangeAge")},
			"Groups":   fields.Many2Many{RelationModel: models.Registry.MustGet("Group")},
			"Categories": fields.Many2Many{RelationModel: models.Registry.MustGet("Category"),
				JSON: "category_ids"},
		})
		partner.AddFields(map[string]models.FieldDefinition{
			"Name":        fields.Char{},
			"Function":    fields.Char{},
			"CompanyName": fields.Char{},
			"Email":       fields.Char{},
			"Phone":       fields.Char{},
			"Fax":         fields.Char{},
			"Address":     fields.Char{},
		})
		models.BootStrap()
		models.Views[partner] = []string{`<view id="test_view" model="Partner"><tree><field name="Name"/></tree></view>`}
	})
	t.Run("Setting two languages", func(t *testing.T) {
		i18n.Langs = []string{"fr", "de"}
	})
	t.Run("Inheriting View 2", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		loadView(viewDef3)
		BootStrap()
		assert.EqualValues(t, len(Registry.views), 3)
		assert.NotNil(t, Registry.GetByID("my_id"))
		assert.NotNil(t, Registry.GetByID("my_other_id"))
		view1 := Registry.GetByID("my_id")
		assert.EqualValues(t, documentToXMLString(view1.Arch("")), `<form>
	<group>
		<field name="user_name"/>
		<label for="age"/>
		<field name="age" on_change="1"/>
	</group>
</form>
`)
		view2 := Registry.GetByID("my_other_id")
		assert.EqualValues(t, documentToXMLString(view2.Arch("")), `<form>
	<h1>
		<field name="name"/>
	</h1>
	<group name="position_info">
		<field name="function"/>
		<field name="company_name"/>
	</group>
	<group name="contact_data">
		<field name="email"/>
		<field name="phone"/>
	</group>
</form>
`)
	})
	t.Run("More inheritance on View 2", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		loadView(viewDef3)
		loadView(viewDef4)
		BootStrap()
		assert.EqualValues(t, len(Registry.views), 3)
		assert.NotNil(t, Registry.GetByID("my_id"))
		assert.NotNil(t, Registry.GetByID("my_other_id"))
		view2 := Registry.GetByID("my_other_id")
		assert.EqualValues(t, documentToXMLString(view2.Arch("")), `<form>
	<h2>
		<field name="name"/>
	</h2>
	<group name="position_info">
		<field name="function"/>
		<field name="company_name"/>
	</group>
	<group>
		<field name="address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="email"/>
		<field name="phone"/>
	</group>
</form>
`)
	})
	t.Run("Modifying inherited modifications on View 2", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		loadView(viewDef3)
		loadView(viewDef4)
		loadView(viewDef5)
		BootStrap()
		assert.EqualValues(t, len(Registry.views), 3)
		assert.NotNil(t, Registry.GetByID("my_id"))
		assert.NotNil(t, Registry.GetByID("my_other_id"))
		view2 := Registry.GetByID("my_other_id")
		assert.EqualValues(t, documentToXMLString(view2.Arch("")), `<form>
	<h2>
		<field name="name"/>
	</h2>
	<group name="position_info">
		<field name="function"/>
		<field name="company_name"/>
	</group>
	<group name="address" string="Address">
		<field name="address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="email"/>
		<field name="phone"/>
	</group>
</form>
`)
	})
	t.Run("Bootstrapping views", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		loadView(viewDef3)
		loadView(viewDef4)
		loadView(viewDef5)
		loadView(viewDef6)
		BootStrap()
		view1 := Registry.GetByID("my_id")
		view2 := Registry.GetByID("my_other_id")
		view3 := Registry.GetByID("my_tree_id")
		assert.NotNil(t, view1)
		assert.NotNil(t, view2)
		assert.NotNil(t, view3)
		assert.EqualValues(t, view1.Type, ViewTypeForm)
		assert.EqualValues(t, view2.Type, ViewTypeForm)
		assert.EqualValues(t, view3.Type, ViewTypeTree)
	})
	t.Run("Testing embedded views", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		loadView(viewDef3)
		loadView(viewDef4)
		loadView(viewDef5)
		loadView(viewDef6)
		loadView(viewDef7)
		BootStrap()
		assert.EqualValues(t, len(Registry.views), 5)
		assert.NotNil(t, Registry.GetByID("embedded_form"))
		assert.Nil(t, Registry.GetByID("embedded_form_childview_1"))
		assert.Nil(t, Registry.GetByID("embedded_form_childview_2"))
		view := Registry.GetByID("embedded_form")
		assert.EqualValues(t, view.ID, "embedded_form")
		assert.EqualValues(t, documentToXMLString(view.Arch("")), `<form>
	<field name="user_name"/>
	<field name="age" on_change="1"/>
	<field name="category_ids"/>
	<field name="groups_ids"/>
</form>
`)
		assert.Len(t, view.SubViews, 2)
		assert.Contains(t, view.SubViews, "Categories")
		assert.Contains(t, view.SubViews, "Groups")
		viewCategories := view.SubViews["Categories"]
		assert.Len(t, viewCategories, 2)
		viewCategoriesForm := viewCategories[ViewTypeForm]
		assert.EqualValues(t, viewCategoriesForm.ID, "embedded_form_childview_Categories_1")
		assert.EqualValues(t, documentToXMLString(viewCategoriesForm.Arch("")), `<form>
	<h1>This is my form</h1>
	<field name="name"/>
	<field name="color"/>
	<field name="sequence"/>
</form>
`)
		viewCategoriesTree := viewCategories[ViewTypeTree]
		assert.EqualValues(t, viewCategoriesTree.ID, "embedded_form_childview_Categories_0")
		assert.EqualValues(t, documentToXMLString(viewCategoriesTree.Arch("")), `<tree>
	<field name="name"/>
	<field name="color"/>
</tree>
`)

		viewGroups := view.SubViews["Groups"]
		assert.Len(t, viewGroups, 1)
		viewGroupsTree := viewGroups[ViewTypeTree]
		assert.EqualValues(t, viewGroupsTree.ID, "embedded_form_childview_Groups_0")
		assert.EqualValues(t, documentToXMLString(viewGroupsTree.Arch("")), `<tree>
	<field name="name"/>
	<field name="active"/>
</tree>
`)
	})
	t.Run("Inheriting embedded views", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		loadView(viewDef3)
		loadView(viewDef4)
		loadView(viewDef5)
		loadView(viewDef6)
		loadView(viewDef7)
		loadView(viewDef71)
		BootStrap()
		assert.EqualValues(t, len(Registry.views), 5)
		assert.NotNil(t, Registry.GetByID("embedded_form"))
		assert.Nil(t, Registry.GetByID("embedded_form_childview_1"))
		assert.Nil(t, Registry.GetByID("embedded_form_childview_2"))
		view := Registry.GetByID("embedded_form")
		assert.EqualValues(t, view.ID, "embedded_form")
		assert.EqualValues(t, documentToXMLString(view.Arch("")), `<form>
	<field required="1" name="user_name"/>
	<field name="age" on_change="1"/>
	<field name="category_ids"/>
	<field name="groups_ids"/>
</form>
`)
		assert.Len(t, view.SubViews, 2)
		assert.Contains(t, view.SubViews, "Categories")
		assert.Contains(t, view.SubViews, "Groups")
		viewCategories := view.SubViews["Categories"]
		assert.Len(t, viewCategories, 2)
		viewCategoriesForm := viewCategories[ViewTypeForm]
		assert.EqualValues(t, viewCategoriesForm.ID, "embedded_form_childview_Categories_1")
		assert.EqualValues(t, documentToXMLString(viewCategoriesForm.Arch("")), `<form>
	<h1>This is my form</h1>
	<field readonly="1" name="name"/>
	<field name="color"/>
	<field name="sequence"/>
</form>
`)
		viewCategoriesTree := viewCategories[ViewTypeTree]
		assert.EqualValues(t, viewCategoriesTree.ID, "embedded_form_childview_Categories_0")
		assert.EqualValues(t, documentToXMLString(viewCategoriesTree.Arch("")), `<tree>
	<field name="name"/>
	<field name="color"/>
</tree>
`)

		viewGroups := view.SubViews["Groups"]
		assert.Len(t, viewGroups, 1)
		viewGroupsTree := viewGroups[ViewTypeTree]
		assert.EqualValues(t, viewGroupsTree.ID, "embedded_form_childview_Groups_0")
		assert.EqualValues(t, documentToXMLString(viewGroupsTree.Arch("")), `<tree>
	<field name="name"/>
	<field name="active"/>
</tree>
`)
	})
	t.Run("Testing GetViews functions", func(t *testing.T) {
		allViews := Registry.GetAll()
		assert.Len(t, allViews, 5)
		userViews := Registry.GetAllViewsForModel("User")
		assert.Len(t, userViews, 3)
		userFirstView := Registry.GetFirstViewForModel("User", ViewTypeForm)
		assert.EqualValues(t, userFirstView.ID, "my_id")
	})
	t.Run("Testing default views", func(t *testing.T) {
		soModel := models.NewModel("SaleOrder")
		soModel.AddFields(map[string]models.FieldDefinition{
			"Name": fields.Char{},
		})
		soSearch := Registry.GetFirstViewForModel("SaleOrder", ViewTypeSearch)
		assert.EqualValues(t, documentToXMLString(soSearch.arch), `<search>
	<field name="name"/>
</search>
`)
		soTree := Registry.GetFirstViewForModel("SaleOrder", ViewTypeTree)
		assert.EqualValues(t, documentToXMLString(soTree.arch), `<tree>
	<field name="name"/>
</tree>
`)
	})
	t.Run("Create new base view from inheritance", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		loadView(viewDef3)
		loadView(viewDef4)
		loadView(viewDef5)
		loadView(viewDef6)
		loadView(viewDef7)
		loadView(viewDef8)
		BootStrap()
		assert.NotNil(t, Registry.GetByID("my_other_id"))
		assert.NotNil(t, Registry.GetByID("new_base_view"))
		view2 := Registry.GetByID("my_other_id")
		newView := Registry.GetByID("new_base_view")
		assert.EqualValues(t, documentToXMLString(view2.Arch("")), `<form>
	<h2>
		<field name="name"/>
	</h2>
	<group name="position_info">
		<field name="function"/>
		<field name="company_name"/>
	</group>
	<group name="address" string="Address">
		<field name="address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="email"/>
		<field name="phone"/>
	</group>
</form>
`)
		assert.EqualValues(t, documentToXMLString(newView.Arch("")), `<form>
	<h2>
		<field name="name"/>
	</h2>
	<group name="position_info">
		<field name="function"/>
		<field name="company_name"/>
	</group>
	<group name="address" string="Address">
		<field name="address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="email"/>
		<field name="fax"/>
		<field name="phone"/>
	</group>
</form>
`)
	})
	t.Run("Inheriting new base view from inheritance", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef1)
		loadView(viewDef2)
		loadView(viewDef3)
		loadView(viewDef4)
		loadView(viewDef5)
		loadView(viewDef6)
		loadView(viewDef7)
		loadView(viewDef8)
		loadView(viewDef9)
		BootStrap()
		assert.NotNil(t, Registry.GetByID("my_other_id"))
		assert.NotNil(t, Registry.GetByID("new_base_view"))
		view2 := Registry.GetByID("my_other_id")
		newView := Registry.GetByID("new_base_view")
		assert.EqualValues(t, documentToXMLString(view2.Arch("")), `<form>
	<h2>
		<field name="name"/>
	</h2>
	<group name="position_info">
		<field name="function"/>
		<field name="company_name"/>
	</group>
	<group name="address" string="Address">
		<field name="address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="email"/>
		<field name="phone"/>
	</group>
</form>
`)
		assert.EqualValues(t, documentToXMLString(newView.Arch("")), `<form>
	<h2>
		<field name="name"/>
	</h2>
	<group name="position_info">
		<field name="function"/>
		<field name="company_name"/>
	</group>
	<group name="address" string="Address">
		<field name="address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="email"/>
		<field widget="phone" name="fax"/>
		<field name="phone"/>
	</group>
</form>
`)
	})

	t.Run("Testing ViewRef objects", func(t *testing.T) {
		userFormRef := MakeViewRef("my_id")
		t.Run("Creating ViewRef instance", func(t *testing.T) {
			assert.EqualValues(t, userFormRef.ID(), "my_id")
			assert.EqualValues(t, userFormRef.Name(), "My View")
			data, err := json.Marshal(userFormRef)
			assert.Nil(t, err)
			assert.EqualValues(t, string(data), `["my_id","My View"]`)
			val, err := userFormRef.Value()
			assert.Nil(t, err)
			assert.EqualValues(t, val, driver.Value("my_id"))
		})
		t.Run("Creating empty viewRef", func(t *testing.T) {
			emptyVR := MakeViewRef("unknownID")
			assert.EqualValues(t, emptyVR.ID(), "")
			assert.EqualValues(t, emptyVR.Name(), "")
			data, err := json.Marshal(emptyVR)
			assert.Nil(t, err)
			assert.EqualValues(t, string(data), `null`)
			val, err := emptyVR.Value()
			assert.Nil(t, err)
			assert.EqualValues(t, val, driver.Value(""))
		})
		t.Run("Unmarshalling JSON viewRef", func(t *testing.T) {
			data := []byte(`["view_id","View Name"]`)
			var vr ViewRef
			err := json.Unmarshal(data, &vr)
			assert.Nil(t, err)
			assert.EqualValues(t, vr.ID(), "view_id")
			assert.EqualValues(t, vr.Name(), "View Name")
		})
		t.Run("Unmarshalling JSON empty viewRef", func(t *testing.T) {
			data := []byte(`null`)
			var vr ViewRef
			err := json.Unmarshal(data, &vr)
			assert.Nil(t, err)
			assert.True(t, vr.IsNull())
		})
		t.Run("Unmarshalling XML viewRef", func(t *testing.T) {
			type stuff struct {
				Ref ViewRef `xml:"ref,attr"`
			}
			data := []byte(`<stuff ref="my_id"/>`)
			var st stuff
			err := xml.Unmarshal(data, &st)
			assert.Nil(t, err)
			assert.EqualValues(t, st.Ref.ID(), "my_id")
			assert.EqualValues(t, st.Ref.Name(), "My View")
		})
		t.Run("Scanning viewRefs", func(t *testing.T) {
			var vr ViewRef
			err := vr.Scan("my_id")
			assert.Nil(t, err)
			assert.EqualValues(t, vr.ID(), "my_id")
			assert.EqualValues(t, vr.Name(), "My View")

			err = vr.Scan([]byte("my_tree_id"))
			assert.Nil(t, err)
			assert.EqualValues(t, vr.ID(), "my_tree_id")
			assert.EqualValues(t, vr.Name(), "my.tree.id")
		})
	})
	t.Run("Testing ViewTuple objects", func(t *testing.T) {
		t.Run("Marshalling a ViewTuple", func(t *testing.T) {
			vt := ViewTuple{
				ID:   "my_id",
				Type: ViewTypeForm,
			}
			data, err := json.Marshal(vt)
			assert.Nil(t, err)
			assert.EqualValues(t, string(data), `["my_id","form"]`)
		})
		t.Run("Unmarshalling ViewTuples", func(t *testing.T) {
			data := []byte(`["my_tree_id","tree"]`)
			var vt ViewTuple
			err := json.Unmarshal(data, &vt)
			assert.Nil(t, err)
			assert.EqualValues(t, vt.ID, "my_tree_id")
			assert.EqualValues(t, vt.Type, ViewTypeTree)
		})
	})
	t.Run("Testing search view sanitizing", func(t *testing.T) {
		Registry = NewCollection()
		loadView(viewDef10)
		BootStrap()
		assert.NotNil(t, Registry.GetByID("search_view"))
		searchView := Registry.GetByID("search_view")
		assert.EqualValues(t, documentToXMLString(searchView.Arch("")), `<search>
	<field name="user_name" domain="[]"/>
</search>
`)
	})

}
