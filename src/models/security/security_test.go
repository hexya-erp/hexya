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

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/types"
)

func TestGroupRegistry(t *testing.T) {
	group1 := Registry.NewGroup("group1_test", "Group 1")
	group2 := Registry.NewGroup("group2_test", "Group 2")
	group3 := Registry.NewGroup("group3_test", "Group 3", group1)
	group4 := Registry.NewGroup("group4_test", "Group 4", group3)
	group5 := Registry.NewGroup("group5_test", "Group 5", group1)
	t.Run("Testing Group Registry", func(t *testing.T) {
		t.Run("Testing basic access methods", func(t *testing.T) {
			assert.EqualValues(t, group1.ID(), "group1_test")
			assert.EqualValues(t, group2.Name(), "Group 2")
			assert.EqualValues(t, group3.String(), "Group(group3_test)")
			assert.True(t, group3.Implies(group1))
			assert.False(t, group1.Implies(group3))
			assert.Len(t, group4.ImpliedGroups(), 1)
			assert.Contains(t, group4.ImpliedGroups(), group3)
			assert.EqualValues(t, Registry.GetGroup("group4_test"), group4)
			allGroups := Registry.AllGroups()
			assert.Len(t, allGroups, 7)
			assert.Contains(t, allGroups, GroupAdmin)
			assert.Contains(t, allGroups, GroupEveryone)
			assert.Contains(t, allGroups, group1)
			assert.Contains(t, allGroups, group4)
		})
		t.Run("Registering an existing group should fail", func(t *testing.T) {
			assert.Panics(t, func() { Registry.NewGroup("group1_test", "Group 1 again") })
		})
		t.Run("Testing HasMembership function", func(t *testing.T) {
			Registry.AddMembership(2, group1)
			assert.True(t, Registry.HasMembership(2, group1))
			assert.True(t, Registry.HasMembership(2, GroupEveryone))
			Registry.RemoveAllMembershipsForUser(2)
		})
		t.Run("Members of a group should be member of parent groups", func(t *testing.T) {
			Registry.AddMembership(2, group1)
			assert.Len(t, Registry.UserGroups(2), 2)
			assert.Contains(t, Registry.UserGroups(2), group1)
			assert.Contains(t, Registry.UserGroups(2), GroupEveryone)
			assert.True(t, Registry.HasMembership(2, group1))

			Registry.AddMembership(3, group2)
			assert.Len(t, Registry.UserGroups(3), 2)
			assert.Contains(t, Registry.UserGroups(3), group2)
			assert.Contains(t, Registry.UserGroups(3), GroupEveryone)

			Registry.AddMembership(4, group3)
			assert.Len(t, Registry.UserGroups(4), 3)
			assert.Contains(t, Registry.UserGroups(4), group1)
			assert.Contains(t, Registry.UserGroups(4), group3)
			assert.Contains(t, Registry.UserGroups(4), GroupEveryone)

			Registry.AddMembership(5, group4)
			assert.Len(t, Registry.UserGroups(5), 4)
			assert.Contains(t, Registry.UserGroups(5), group1)
			assert.Contains(t, Registry.UserGroups(5), group3)
			assert.Contains(t, Registry.UserGroups(5), group4)
			assert.Contains(t, Registry.UserGroups(5), GroupEveryone)

			Registry.AddMembership(6, group5)
			assert.Len(t, Registry.UserGroups(6), 3)
			assert.Contains(t, Registry.UserGroups(6), group1)
			assert.Contains(t, Registry.UserGroups(6), group5)
			assert.Contains(t, Registry.UserGroups(6), GroupEveryone)
		})
		t.Run("Removing a group should remove all memberships (incl. inherited)", func(t *testing.T) {
			Registry.UnregisterGroup(group3)

			assert.NotContains(t, Registry.groups, group3.ID())
			assert.Empty(t, group4.ImpliedGroups())

			assert.EqualValues(t, len(Registry.UserGroups(2)), 2)
			assert.Contains(t, Registry.UserGroups(2), group1)
			assert.Contains(t, Registry.UserGroups(2), GroupEveryone)
			assert.EqualValues(t, len(Registry.UserGroups(3)), 2)
			assert.Contains(t, Registry.UserGroups(3), group2)
			assert.Contains(t, Registry.UserGroups(3), GroupEveryone)
			assert.Len(t, Registry.UserGroups(4), 1)
			assert.Contains(t, Registry.UserGroups(4), GroupEveryone)
			assert.EqualValues(t, len(Registry.UserGroups(5)), 2)
			assert.Contains(t, Registry.UserGroups(5), group4)
			assert.Contains(t, Registry.UserGroups(5), GroupEveryone)
			assert.EqualValues(t, len(Registry.UserGroups(6)), 3)
			assert.Contains(t, Registry.UserGroups(6), group1)
			assert.Contains(t, Registry.UserGroups(6), group5)
			assert.Contains(t, Registry.UserGroups(6), GroupEveryone)
		})
		t.Run("Removing a membership should remove inherited too", func(t *testing.T) {
			Registry.RemoveMembership(6, group5)

			assert.Len(t, Registry.UserGroups(2), 2)
			assert.Contains(t, Registry.UserGroups(2), group1)
			assert.Contains(t, Registry.UserGroups(2), GroupEveryone)
			assert.Len(t, Registry.UserGroups(3), 2)
			assert.Contains(t, Registry.UserGroups(3), group2)
			assert.Contains(t, Registry.UserGroups(3), GroupEveryone)
			assert.Len(t, Registry.UserGroups(4), 1)
			assert.Contains(t, Registry.UserGroups(4), GroupEveryone)
			assert.Len(t, Registry.UserGroups(5), 2)
			assert.Contains(t, Registry.UserGroups(5), group4)
			assert.Contains(t, Registry.UserGroups(5), GroupEveryone)
			assert.Len(t, Registry.UserGroups(6), 1)
			assert.Contains(t, Registry.UserGroups(6), GroupEveryone)
		})
		t.Run("Removing inherited membership should not change anything", func(t *testing.T) {
			// Recreating group 3
			group3 = Registry.NewGroup("group3_test", "Group 3", group1)
			group4.inherits[group3] = true

			Registry.AddMembership(6, group5)
			Registry.AddMembership(6, group4)
			assert.Len(t, Registry.UserGroups(6), 5)
			assert.Contains(t, Registry.UserGroups(6), group1)
			assert.EqualValues(t, Registry.UserGroups(6)[group1], InheritedGroup)
			assert.Contains(t, Registry.UserGroups(6), group3)
			assert.EqualValues(t, Registry.UserGroups(6)[group3], InheritedGroup)
			assert.Contains(t, Registry.UserGroups(6), group4)
			assert.EqualValues(t, Registry.UserGroups(6)[group4], NativeGroup)
			assert.Contains(t, Registry.UserGroups(6), group5)
			assert.EqualValues(t, Registry.UserGroups(6)[group5], NativeGroup)
			assert.Contains(t, Registry.UserGroups(6), GroupEveryone)
			assert.EqualValues(t, Registry.UserGroups(6)[group5], NativeGroup)

			Registry.RemoveMembership(6, group3)
			assert.Len(t, Registry.UserGroups(6), 5)
			assert.Contains(t, Registry.UserGroups(6), group1)
			assert.Contains(t, Registry.UserGroups(6), group3)
			assert.EqualValues(t, Registry.UserGroups(6)[group3], InheritedGroup)
			assert.Contains(t, Registry.UserGroups(6), group4)
			assert.Contains(t, Registry.UserGroups(6), group5)
			assert.Contains(t, Registry.UserGroups(6), GroupEveryone)
		})
		t.Run("Removing membership should not impact other inherited fields", func(t *testing.T) {
			Registry.RemoveMembership(6, group4)
			assert.Len(t, Registry.UserGroups(6), 3)
			assert.Contains(t, Registry.UserGroups(6), group1)
			assert.Contains(t, Registry.UserGroups(6), group5)
			assert.Contains(t, Registry.UserGroups(6), GroupEveryone)
		})
	})
}

type simpleAuthBackend struct{}

func (a simpleAuthBackend) Authenticate(login, secret string, _ *types.Context) (int64, error) {
	if login != "admin" {
		return 0, UserNotFoundError("admin")
	}
	if secret != "secret" {
		return 0, InvalidCredentialsError("admin")
	}
	return 1, nil
}

func TestAuthBackend(t *testing.T) {
	t.Run("Testing authentication backend", func(t *testing.T) {
		AuthenticationRegistry.RegisterBackend(simpleAuthBackend{})
		id, err := AuthenticationRegistry.Authenticate("admin", "secret", nil)
		assert.Nil(t, err)
		assert.EqualValues(t, id, 1)
		id, err = AuthenticationRegistry.Authenticate("admin2", "secret", nil)
		assert.EqualValues(t, err, UserNotFoundError("admin2"))
		assert.EqualValues(t, err.Error(), "User not found admin2")
		assert.EqualValues(t, id, 0)
		id, err = AuthenticationRegistry.Authenticate("admin", "wrong", nil)
		assert.EqualValues(t, err, InvalidCredentialsError("admin"))
		assert.EqualValues(t, err.Error(), "Wrong credentials for user admin")
		assert.EqualValues(t, id, 0)
	})
}
