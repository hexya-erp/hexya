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

package actions

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/tools/xmlutils"
	"github.com/hexya-erp/hexya/src/views"
)

var actionDef1 = `
<action id="my_action" name="My Action" type="ir.actions.act_window" model="Partner" view_mode="tree,form"
        src_model="User" view_id="my_id"/>
`

var actionDef2 = `
<action id="my_action_2" name="My Second Action" model="Partner" type="ir.actions.act_window" view_mode="tree,form">
	<help>
		This is the help message.
		<strong>And this is important!</strong>
	</help>
	<view id="base_view_partner_tree" type="tree"/>
	<view id="base_view_partner_form" type="form"/>
</action>
`

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

func TestActions(t *testing.T) {
	t.Run("Creating models", func(t *testing.T) {
		user := models.NewModel("User")
		partner := models.NewModel("Partner")
		user.AddFields(map[string]models.FieldDefinition{
			"UserName": fields.Char{},
			"Age":      fields.Integer{},
		})
		partner.AddFields(map[string]models.FieldDefinition{
			"Name": fields.Char{},
		})
		models.BootStrap()
	})
	t.Run("Creating Action 1", func(t *testing.T) {
		view1, _ := xmlutils.XMLToElement(viewDef1)
		views.LoadFromEtree(view1)
		views.BootStrap()
		action1, _ := xmlutils.XMLToElement(actionDef1)
		LoadFromEtree(action1)
		assert.EqualValues(t, len(Registry.actions), 1)
		action, ok := Registry.GetByXMLID("my_action")
		assert.NotNil(t, action)
		assert.True(t, ok)
		assert.EqualValues(t, action.XMLID, "my_action")
		assert.EqualValues(t, action.Name, "My Action")
		assert.EqualValues(t, action.Model, "Partner")
		assert.EqualValues(t, action.ViewMode, "tree,form")
	})
	t.Run("Creating Action 2", func(t *testing.T) {
		action2, _ := xmlutils.XMLToElement(actionDef2)
		LoadFromEtree(action2)
		assert.EqualValues(t, len(Registry.actions), 2)
		action, ok := Registry.GetByXMLID("my_action_2")
		assert.NotNil(t, action)
		assert.True(t, ok)
		assert.EqualValues(t, action.XMLID, "my_action_2")
		assert.EqualValues(t, action.Name, "My Second Action")
		assert.EqualValues(t, action.Model, "Partner")
		assert.EqualValues(t, action.ViewMode, "tree,form")
		assert.EqualValues(t, action.View, views.ViewRef{})
		assert.Len(t, action.Views, 2)
		assert.Contains(t, action.Views, views.ViewTuple{ID: "base_view_partner_tree", Type: "tree"})
		assert.Contains(t, action.Views, views.ViewTuple{ID: "base_view_partner_form", Type: "form"})
		assert.EqualValues(t, action.HelpXML.Content, "\n\t\tThis is the help message.\n\t\t\n\t\t<strong>And this is important!</strong>\n\t")
	})
	t.Run("Testing Boostrap and Get functions", func(t *testing.T) {
		BootStrap()
		allActions := Registry.GetAll()
		assert.Len(t, allActions, 2)
		assert.NotPanics(t, func() { Registry.MustGetByXMLID("my_action") })
		assert.Panics(t, func() { Registry.MustGetByXMLID("unknown_id") })
		assert.NotPanics(t, func() { Registry.MustGetById(1) })
		assert.NotPanics(t, func() { Registry.MustGetById(2) })
		assert.Panics(t, func() { Registry.MustGetById(3) })
		act, _ := Registry.GetByID(1)
		assert.NotNil(t, act)
		assert.EqualValues(t, act.XMLID, "my_action")
		userLinkedActions := Registry.GetActionLinksForModel("User")
		assert.Len(t, userLinkedActions, 1)
		tName := userLinkedActions[0].TranslatedName("fr")
		assert.EqualValues(t, tName, "My Action")
		action2 := Registry.MustGetByXMLID("my_action_2")
		assert.EqualValues(t, action2.Help, "\n\t\tThis is the help message.\n\t\t\n\t\t<strong>And this is important!</strong>\n\t")
	})
	t.Run("Testing ActionRef objects", func(t *testing.T) {
		actionRef := MakeActionRef("my_action")
		t.Run("Creating ActionRef instance", func(t *testing.T) {
			assert.EqualValues(t, actionRef.ID(), "my_action")
			assert.EqualValues(t, actionRef.Name(), "My Action")
			data, err := json.Marshal(actionRef)
			assert.Nil(t, err)
			assert.EqualValues(t, string(data), `["my_action","My Action"]`)
			val, err := actionRef.Value()
			assert.Nil(t, err)
			assert.EqualValues(t, val, driver.Value("my_action"))
		})
		t.Run("Creating empty actionRef", func(t *testing.T) {
			emptyAR := MakeActionRef("unknownID")
			assert.EqualValues(t, emptyAR.ID(), "")
			assert.EqualValues(t, emptyAR.Name(), "")
			data, err := json.Marshal(emptyAR)
			assert.Nil(t, err)
			assert.EqualValues(t, string(data), `false`)
			val, err := emptyAR.Value()
			assert.Nil(t, err)
			assert.EqualValues(t, val, driver.Value(""))
		})
		t.Run("Unmarshalling JSON actionRef", func(t *testing.T) {
			data := []byte(`["action_id","Action Name"]`)
			var ar ActionRef
			err := json.Unmarshal(data, &ar)
			assert.Nil(t, err)
			assert.EqualValues(t, ar.ID(), "action_id")
			assert.EqualValues(t, ar.Name(), "Action Name")
		})
		t.Run("Unmarshalling JSON empty actionRef", func(t *testing.T) {
			data := []byte(`null`)
			var ar ActionRef
			err := json.Unmarshal(data, &ar)
			assert.Nil(t, err)
			assert.True(t, ar.IsNull())
		})
		t.Run("Unmarshalling JSON false actionRef", func(t *testing.T) {
			data := []byte(`false`)
			var ar ActionRef
			err := json.Unmarshal(data, &ar)
			assert.Nil(t, err)
			assert.True(t, ar.IsNull())
		})
		t.Run("Scanning actionRefs", func(t *testing.T) {
			var vr ActionRef
			err := vr.Scan("my_action")
			assert.Nil(t, err)
			assert.EqualValues(t, vr.ID(), "my_action")
			assert.EqualValues(t, vr.Name(), "My Action")

			err = vr.Scan([]byte("my_action_2"))
			assert.Nil(t, err)
			assert.EqualValues(t, vr.ID(), "my_action_2")
			assert.EqualValues(t, vr.Name(), "My Second Action")
		})
	})
	t.Run("Testing ActionString objects", func(t *testing.T) {
		act, _ := Registry.GetByXMLID("my_action")
		as := act.ActionString()
		d, err := json.Marshal(as)
		assert.Nil(t, err)
		assert.EqualValues(t, string(d), `"ir.actions.act_window,1"`)
		d, err = json.Marshal(ActionString{})
		assert.Nil(t, err)
		assert.EqualValues(t, string(d), "false")
	})
}
