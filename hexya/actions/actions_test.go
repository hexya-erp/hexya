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

package actions

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
	"github.com/hexya-erp/hexya/hexya/views"
	. "github.com/smartystreets/goconvey/convey"
)

var actionDef1 string = `
<action id="my_action" name="My Action" type="ir.actions.act_window" model="Partner" view_mode="tree,form"
        src_model="User" view_id="my_id"/>
`

var actionDef2 string = `
<action id="my_action_2" name="My Second Action" model="Partner" type="ir.actions.act_window" view_mode="tree,form">
	<view id="base_view_partner_tree" type="tree"/>
	<view id="base_view_partner_form" type="form"/>
</action>
`

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

func TestActions(t *testing.T) {
	Convey("Creating Action 1", t, func() {
		views.LoadFromEtree(xmlutils.XMLToElement(viewDef1))
		views.BootStrap()
		LoadFromEtree(xmlutils.XMLToElement(actionDef1))
		So(len(Registry.actions), ShouldEqual, 1)
		So(Registry.GetById("my_action"), ShouldNotBeNil)
		action := Registry.GetById("my_action")
		So(action.ID, ShouldEqual, "my_action")
		So(action.Name, ShouldEqual, "My Action")
		So(action.Model, ShouldEqual, "Partner")
		So(action.ViewMode, ShouldEqual, "tree,form")
	})
	Convey("Creating Action 2", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(actionDef2))
		So(len(Registry.actions), ShouldEqual, 2)
		So(Registry.GetById("my_action_2"), ShouldNotBeNil)
		action := Registry.GetById("my_action_2")
		So(action.ID, ShouldEqual, "my_action_2")
		So(action.Name, ShouldEqual, "My Second Action")
		So(action.Model, ShouldEqual, "Partner")
		So(action.ViewMode, ShouldEqual, "tree,form")
		So(action.View, ShouldEqual, views.ViewRef{})
		So(action.Views, ShouldHaveLength, 2)
		So(action.Views, ShouldContain, views.ViewTuple{ID: "base_view_partner_tree", Type: "tree"})
		So(action.Views, ShouldContain, views.ViewTuple{ID: "base_view_partner_form", Type: "form"})
	})
	Convey("Testing Boostrap and Get functions", t, func() {
		BootStrap()
		allActions := Registry.GetAll()
		So(allActions, ShouldHaveLength, 2)
		So(func() { Registry.MustGetById("my_action") }, ShouldNotPanic)
		So(func() { Registry.MustGetById("unknown_id") }, ShouldPanic)
		userLinkedActions := Registry.GetActionLinksForModel("User")
		So(userLinkedActions, ShouldHaveLength, 1)
		tName := userLinkedActions[0].TranslatedName("fr")
		So(tName, ShouldEqual, "My Action")
	})
	Convey("Testing ActionRef objects", t, func() {
		actionRef := MakeActionRef("my_action")
		Convey("Creating ActionRef instance", func() {
			So(actionRef.ID(), ShouldEqual, "my_action")
			So(actionRef.Name(), ShouldEqual, "My Action")
			data, err := json.Marshal(actionRef)
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `["my_action","My Action"]`)
			val, err := actionRef.Value()
			So(err, ShouldBeNil)
			So(val, ShouldEqual, driver.Value("my_action"))
		})
		Convey("Creating empty actionRef", func() {
			emptyAR := MakeActionRef("unknownID")
			So(emptyAR.ID(), ShouldEqual, "")
			So(emptyAR.Name(), ShouldEqual, "")
			data, err := json.Marshal(emptyAR)
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `null`)
			val, err := emptyAR.Value()
			So(err, ShouldBeNil)
			So(val, ShouldEqual, driver.Value(""))
		})
		Convey("Unmarshalling JSON actionRef", func() {
			data := []byte(`["action_id","Action Name"]`)
			var ar ActionRef
			err := json.Unmarshal(data, &ar)
			So(err, ShouldBeNil)
			So(ar.ID(), ShouldEqual, "action_id")
			So(ar.Name(), ShouldEqual, "Action Name")
		})
		Convey("Unmarshalling JSON empty actionRef", func() {
			data := []byte(`null`)
			var ar ActionRef
			err := json.Unmarshal(data, &ar)
			So(err, ShouldBeNil)
			So(ar.IsNull(), ShouldBeTrue)
		})
		Convey("Scanning actionRefs", func() {
			var vr ActionRef
			err := vr.Scan("my_action")
			So(err, ShouldBeNil)
			So(vr.ID(), ShouldEqual, "my_action")
			So(vr.Name(), ShouldEqual, "My Action")

			err = vr.Scan([]byte("my_action_2"))
			So(err, ShouldBeNil)
			So(vr.ID(), ShouldEqual, "my_action_2")
			So(vr.Name(), ShouldEqual, "My Second Action")
		})
	})

}
