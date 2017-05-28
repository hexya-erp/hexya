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
	"testing"

	"github.com/npiganeau/yep/yep/tools/xmlutils"
	. "github.com/smartystreets/goconvey/convey"
)

var actionDef1 string = `
<action id="my_action" name="My Action" type="form" model="Partner" view_mode="tree,form"/>
`

func TestActions(t *testing.T) {
	Convey("Creating Action 1", t, func() {
		LoadFromEtree(xmlutils.XMLToElement(actionDef1))
		So(len(Registry.actions), ShouldEqual, 1)
		So(Registry.GetById("my_action"), ShouldNotBeNil)
		action := Registry.GetById("my_action")
		So(action.ID, ShouldEqual, "my_action")
		So(action.Name, ShouldEqual, "My Action")
		So(action.Model, ShouldEqual, "Partner")
		So(action.ViewMode, ShouldEqual, "tree,form")
	})
}
