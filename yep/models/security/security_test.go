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

package security

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAccessControlList(t *testing.T) {
	Convey("Testing Access Control Lists", t, func() {
		acl := NewAccessControlList()
		group1 := NewGroup("Group1")
		group1Inherit := NewGroup("Group1 Inherited", group1)
		group2 := NewGroup("Group2")
		acl.AddPermission(group1, Read)
		acl.AddPermission(group2, All)
		So(len(acl.perms), ShouldEqual, 2)
		So(acl.perms[group1], ShouldEqual, Read)
		So(acl.perms[group2], ShouldEqual, All)

		Convey("Adding permissions to groups", func() {
			acl.AddPermission(group1, Create)
			So(acl.perms[group1], ShouldEqual, Read|Create)
		})

		Convey("Removing permissions from groups", func() {
			acl.RemovePermission(group2, Read)
			So(acl.perms[group2], ShouldEqual, Write|Create|Unlink)
			acl.RemovePermission(group1, Write|Create)
			So(acl.perms[group1], ShouldEqual, Read)
		})

		Convey("Replacing permissions in groups", func() {
			acl.ReplacePermission(group2, Read)
			So(acl.perms[group2], ShouldEqual, Read)
			acl.ReplacePermission(group1, Read|Write)
			So(acl.perms[group1], ShouldEqual, Read|Write)
		})

		Convey("Checking permissions", func() {
			So(acl.CheckPermission(group1, Read), ShouldBeTrue)
			So(acl.CheckPermission(group1Inherit, Read), ShouldBeTrue)
			So(acl.CheckPermission(group2, Read|Write), ShouldBeTrue)
			So(acl.CheckPermission(group2, All), ShouldBeTrue)
			So(acl.CheckPermission(group1, Read|Write), ShouldBeFalse)
			So(acl.CheckPermission(group1Inherit, Read|Create|Unlink), ShouldBeFalse)
		})

	})

}
