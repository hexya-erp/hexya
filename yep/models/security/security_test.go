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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccessControlList(t *testing.T) {
	Convey("Testing Access Control Lists", t, func() {
		gr := NewGroupCollection()
		acl := NewAccessControlList()
		group1 := gr.NewGroup("group1_test", "Group 1")
		group1Inherit := gr.NewGroup("group1_inherited", "Group 1 Inherited", group1)
		group2 := gr.NewGroup("group2_test", "Group 2")
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

func TestGroupRegistry(t *testing.T) {
	group1 := Registry.NewGroup("group1_test", "Group 1")
	group2 := Registry.NewGroup("group2_test", "Group 2")
	group3 := Registry.NewGroup("group3_test", "Group 3", group1)
	group4 := Registry.NewGroup("group4_test", "Group 4", group3)
	group5 := Registry.NewGroup("group5_test", "Group 5", group1)
	Convey("Testing Group Registry", t, func() {
		Convey("Members of a group should be member of parent groups", func() {
			Registry.AddMembership(2, group1)
			So(len(Registry.UserGroups(2)), ShouldEqual, 1)
			So(Registry.UserGroups(2), ShouldContainKey, group1)

			Registry.AddMembership(3, group2)
			So(len(Registry.UserGroups(3)), ShouldEqual, 1)
			So(Registry.UserGroups(3), ShouldContainKey, group2)

			Registry.AddMembership(4, group3)
			So(len(Registry.UserGroups(4)), ShouldEqual, 2)
			So(Registry.UserGroups(4), ShouldContainKey, group1)
			So(Registry.UserGroups(4), ShouldContainKey, group3)

			Registry.AddMembership(5, group4)
			So(len(Registry.UserGroups(5)), ShouldEqual, 3)
			So(Registry.UserGroups(5), ShouldContainKey, group1)
			So(Registry.UserGroups(5), ShouldContainKey, group3)
			So(Registry.UserGroups(5), ShouldContainKey, group4)

			Registry.AddMembership(6, group5)
			So(len(Registry.UserGroups(6)), ShouldEqual, 2)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
		})
		Convey("Removing a group should remove all memberships (incl. inherited)", func() {
			Registry.UnregisterGroup(group3)

			So(Registry.groups, ShouldNotContainKey, group3.ID)
			So(group4.Inherits, ShouldBeEmpty)

			So(len(Registry.UserGroups(2)), ShouldEqual, 1)
			So(len(Registry.UserGroups(3)), ShouldEqual, 1)
			So(Registry.UserGroups(3), ShouldContainKey, group2)
			So(Registry.UserGroups(2), ShouldContainKey, group1)
			So(Registry.UserGroups(4), ShouldBeEmpty)
			So(len(Registry.UserGroups(5)), ShouldEqual, 1)
			So(Registry.UserGroups(5), ShouldContainKey, group4)
			So(len(Registry.UserGroups(6)), ShouldEqual, 2)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
		})
		Convey("Removing a membership should remove inherited too", func() {
			Registry.RemoveMembership(6, group5)

			So(len(Registry.UserGroups(2)), ShouldEqual, 1)
			So(len(Registry.UserGroups(3)), ShouldEqual, 1)
			So(Registry.UserGroups(3), ShouldContainKey, group2)
			So(Registry.UserGroups(2), ShouldContainKey, group1)
			So(Registry.UserGroups(4), ShouldBeEmpty)
			So(len(Registry.UserGroups(5)), ShouldEqual, 1)
			So(Registry.UserGroups(5), ShouldContainKey, group4)
			So(Registry.UserGroups(6), ShouldBeEmpty)
		})
		Convey("Removing inherited membership should not change anything", func() {
			group4.Inherits = []*Group{group3}

			Registry.AddMembership(6, group5)
			Registry.AddMembership(6, group4)
			So(len(Registry.UserGroups(6)), ShouldEqual, 4)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group3)
			So(Registry.UserGroups(6), ShouldContainKey, group4)
			So(Registry.UserGroups(6), ShouldContainKey, group5)

			Registry.RemoveMembership(6, group3)
			So(len(Registry.UserGroups(6)), ShouldEqual, 4)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group3)
			So(Registry.UserGroups(6), ShouldContainKey, group4)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
		})
		Convey("Removing membership should not impact other inherited fields", func() {
			Registry.RemoveMembership(6, group4)
			So(len(Registry.UserGroups(6)), ShouldEqual, 2)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
		})
	})
}
