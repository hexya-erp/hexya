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

package ir

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var actionDef1 string = `
<action id="my_action">
	<name>My Action</name>
	<type>form</type>
	<model>ResPartner</model>
	<view_mode>tree,form</view_mode>
</action>
`

func TestActions(t *testing.T) {
	Convey("Creating Action 1", t, func() {
		LoadActionFromEtree(xmlToElement(actionDef1))
		So(len(ActionsRegistry.actions), ShouldEqual, 1)
		So(ActionsRegistry.GetActionById("my_action"), ShouldNotBeNil)
		action := ActionsRegistry.GetActionById("my_action")
		So(action.ID, ShouldEqual, "my_action")
		So(action.Name, ShouldEqual, "My Action")
		So(action.Model, ShouldEqual, "ResPartner")
		So(action.ViewMode, ShouldEqual, "tree,form")
	})
}
