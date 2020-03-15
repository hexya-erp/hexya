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

	"github.com/hexya-erp/hexya/src/models/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGroupRegistry(t *testing.T) {
	group1 := Registry.NewGroup("group1_test", "Group 1")
	group2 := Registry.NewGroup("group2_test", "Group 2")
	group3 := Registry.NewGroup("group3_test", "Group 3", group1)
	group4 := Registry.NewGroup("group4_test", "Group 4", group3)
	group5 := Registry.NewGroup("group5_test", "Group 5", group1)
	Convey("Testing Group Registry", t, func() {
		Convey("Members of a group should be member of parent groups", func() {
			Registry.AddMembership(2, group1)
			So(Registry.UserGroups(2), ShouldHaveLength, 2)
			So(Registry.UserGroups(2), ShouldContainKey, group1)
			So(Registry.UserGroups(2), ShouldContainKey, GroupEveryone)

			Registry.AddMembership(3, group2)
			So(Registry.UserGroups(3), ShouldHaveLength, 2)
			So(Registry.UserGroups(3), ShouldContainKey, group2)
			So(Registry.UserGroups(3), ShouldContainKey, GroupEveryone)

			Registry.AddMembership(4, group3)
			So(Registry.UserGroups(4), ShouldHaveLength, 3)
			So(Registry.UserGroups(4), ShouldContainKey, group1)
			So(Registry.UserGroups(4), ShouldContainKey, group3)
			So(Registry.UserGroups(4), ShouldContainKey, GroupEveryone)

			Registry.AddMembership(5, group4)
			So(Registry.UserGroups(5), ShouldHaveLength, 4)
			So(Registry.UserGroups(5), ShouldContainKey, group1)
			So(Registry.UserGroups(5), ShouldContainKey, group3)
			So(Registry.UserGroups(5), ShouldContainKey, group4)
			So(Registry.UserGroups(5), ShouldContainKey, GroupEveryone)

			Registry.AddMembership(6, group5)
			So(Registry.UserGroups(6), ShouldHaveLength, 3)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
			So(Registry.UserGroups(6), ShouldContainKey, GroupEveryone)
		})
		Convey("Removing a group should remove all memberships (incl. inherited)", func() {
			Registry.UnregisterGroup(group3)

			So(Registry.groups, ShouldNotContainKey, group3.ID)
			So(group4.Inherits, ShouldBeEmpty)

			So(len(Registry.UserGroups(2)), ShouldEqual, 2)
			So(Registry.UserGroups(2), ShouldContainKey, group1)
			So(Registry.UserGroups(2), ShouldContainKey, GroupEveryone)
			So(len(Registry.UserGroups(3)), ShouldEqual, 2)
			So(Registry.UserGroups(3), ShouldContainKey, group2)
			So(Registry.UserGroups(3), ShouldContainKey, GroupEveryone)
			So(Registry.UserGroups(4), ShouldHaveLength, 1)
			So(Registry.UserGroups(4), ShouldContainKey, GroupEveryone)
			So(len(Registry.UserGroups(5)), ShouldEqual, 2)
			So(Registry.UserGroups(5), ShouldContainKey, group4)
			So(Registry.UserGroups(5), ShouldContainKey, GroupEveryone)
			So(len(Registry.UserGroups(6)), ShouldEqual, 3)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
			So(Registry.UserGroups(6), ShouldContainKey, GroupEveryone)
		})
		Convey("Removing a membership should remove inherited too", func() {
			Registry.RemoveMembership(6, group5)

			So(Registry.UserGroups(2), ShouldHaveLength, 2)
			So(Registry.UserGroups(2), ShouldContainKey, group1)
			So(Registry.UserGroups(2), ShouldContainKey, GroupEveryone)
			So(Registry.UserGroups(3), ShouldHaveLength, 2)
			So(Registry.UserGroups(3), ShouldContainKey, group2)
			So(Registry.UserGroups(3), ShouldContainKey, GroupEveryone)
			So(Registry.UserGroups(4), ShouldHaveLength, 1)
			So(Registry.UserGroups(4), ShouldContainKey, GroupEveryone)
			So(Registry.UserGroups(5), ShouldHaveLength, 2)
			So(Registry.UserGroups(5), ShouldContainKey, group4)
			So(Registry.UserGroups(5), ShouldContainKey, GroupEveryone)
			So(Registry.UserGroups(6), ShouldHaveLength, 1)
			So(Registry.UserGroups(6), ShouldContainKey, GroupEveryone)
		})
		Convey("Removing inherited membership should not change anything", func() {
			// Recreating group 3
			group3 = Registry.NewGroup("group3_test", "Group 3", group1)
			group4.Inherits = []*Group{group3}

			Registry.AddMembership(6, group5)
			Registry.AddMembership(6, group4)
			So(Registry.UserGroups(6), ShouldHaveLength, 5)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6)[group1], ShouldEqual, InheritedGroup)
			So(Registry.UserGroups(6), ShouldContainKey, group3)
			So(Registry.UserGroups(6)[group3], ShouldEqual, InheritedGroup)
			So(Registry.UserGroups(6), ShouldContainKey, group4)
			So(Registry.UserGroups(6)[group4], ShouldEqual, NativeGroup)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
			So(Registry.UserGroups(6)[group5], ShouldEqual, NativeGroup)
			So(Registry.UserGroups(6), ShouldContainKey, GroupEveryone)
			So(Registry.UserGroups(6)[group5], ShouldEqual, NativeGroup)

			Registry.RemoveMembership(6, group3)
			So(Registry.UserGroups(6), ShouldHaveLength, 5)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group3)
			So(Registry.UserGroups(6)[group3], ShouldEqual, InheritedGroup)
			So(Registry.UserGroups(6), ShouldContainKey, group4)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
			So(Registry.UserGroups(6), ShouldContainKey, GroupEveryone)
		})
		Convey("Removing membership should not impact other inherited fields", func() {
			Registry.RemoveMembership(6, group4)
			So(Registry.UserGroups(6), ShouldHaveLength, 3)
			So(Registry.UserGroups(6), ShouldContainKey, group1)
			So(Registry.UserGroups(6), ShouldContainKey, group5)
			So(Registry.UserGroups(6), ShouldContainKey, GroupEveryone)
		})
	})
}

type simpleAuthBackend struct{}

func (a simpleAuthBackend) Authenticate(login, secret string, ctx *types.Context) (int64, error) {
	if login != "admin" {
		return 0, UserNotFoundError("admin")
	}
	if secret != "secret" {
		return 0, InvalidCredentialsError("admin")
	}
	return 1, nil
}

func TestAuthBackend(t *testing.T) {
	Convey("Testing authentication backend", t, func() {
		AuthenticationRegistry.RegisterBackend(simpleAuthBackend{})
		id, err := AuthenticationRegistry.Authenticate("admin", "secret", nil)
		So(err, ShouldBeNil)
		So(id, ShouldEqual, 1)
		id, err = AuthenticationRegistry.Authenticate("admin2", "secret", nil)
		So(err, ShouldEqual, UserNotFoundError("admin2"))
		So(err.Error(), ShouldEqual, "User not found admin2")
		So(id, ShouldEqual, 0)
		id, err = AuthenticationRegistry.Authenticate("admin", "wrong", nil)
		So(err, ShouldEqual, InvalidCredentialsError("admin"))
		So(err.Error(), ShouldEqual, "Wrong credentials for user admin")
		So(id, ShouldEqual, 0)
	})
}
