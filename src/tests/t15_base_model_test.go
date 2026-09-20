// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
	"github.com/hexya-erp/pool/q"
)

func TestBaseModelMethods(t *testing.T) {
	t.Run("Testing base model methods", func(t *testing.T) {
		t.Run("New", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				dummyUser := h.User().NewSet(env).New(h.User().NewData().
					SetName("DummyUser").
					SetEmail("du@example.com"))
				assert.EqualValues(t, dummyUser.Name(), "DummyUser")
				assert.EqualValues(t, dummyUser.Email(), "du@example.com")
				assert.Empty(t, dummyUser.Email2())
				assert.Less(t, dummyUser.Ids()[0], int64(0))
				assert.Panics(t, func() { dummyUser.ForceLoad() })
				assert.NotPanics(t, func() { dummyUser.SetEmail2("du2@example.com") })
				assert.EqualValues(t, dummyUser.Email2(), "du2@example.com")
				assert.EqualValues(t, dummyUser.DecoratedName(), "User: DummyUser [<du@example.com>]")
				assert.NotPanics(t, func() { dummyUser.Unlink() })
			}))
		})
		t.Run("Copy", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				newProfile := userJane.Profile().Copy(nil)
				userJane.Write(h.User().NewData().SetPassword("Jane's Password"))
				userJaneCopy := userJane.Copy(h.User().NewData().
					SetName("Jane's Copy").
					SetEmail2("js@example.com").
					SetProfile(newProfile))
				assert.EqualValues(t, userJaneCopy.Name(), "Jane's Copy")
				assert.EqualValues(t, userJaneCopy.Email(), "jane.smith@example.com")
				assert.EqualValues(t, userJaneCopy.Email2(), "js@example.com")
				assert.Empty(t, userJaneCopy.Password())
				assert.EqualValues(t, userJaneCopy.Age(), 24)
				assert.EqualValues(t, userJaneCopy.Nums(), 2)
				assert.EqualValues(t, userJaneCopy.Posts().Len(), 2)

				assert.NotPanics(t, func() { userJane.Profile().Copy(nil) })
			}))
		})
		t.Run("Sorted", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				for i := range 20 {
					h.Post().Create(env, h.Post().NewData().
						SetTitle(fmt.Sprintf("Post no %02d", (24-i)%20)).
						SetUser(userJane))
				}
				posts := h.Post().Search(env, q.Post().Title().Contains("Post no")).OrderBy("ID")
				for i, post := range posts.Records() {
					assert.EqualValues(t, post.Title(), fmt.Sprintf("Post no %02d", (24-i)%20))
				}

				sortedPosts := posts.Sorted(func(rs1, rs2 m.PostSet) bool {
					return rs1.Title() < rs2.Title()
				}).Records()
				assert.Len(t, sortedPosts, 20)
				for i, post := range sortedPosts {
					assert.EqualValues(t, post.Title(), fmt.Sprintf("Post no %02d", i))
				}
			}))
		})
		t.Run("Filtered", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				for i := range 20 {
					h.Post().Create(env, h.Post().NewData().
						SetTitle(fmt.Sprintf("Post no %02d", i)).
						SetUser(userJane))
				}
				posts := h.Post().Search(env, q.Post().Title().Contains("Post no"))

				evenPosts := posts.Filtered(func(rs m.PostSet) bool {
					var num int
					fmt.Sscanf(rs.Title(), "Post no %02d", &num)
					if num%2 == 0 {
						return true
					}
					return false
				}).Records()
				assert.Len(t, evenPosts, 10)
				for i := range 10 {
					assert.EqualValues(t, evenPosts[i].Title(), fmt.Sprintf("Post no %02d", 2*i))
				}
			}))
		})
	})
}
