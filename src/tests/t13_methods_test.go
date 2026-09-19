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

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
	"github.com/hexya-erp/pool/q"
)

func TestMethods(t *testing.T) {
	t.Run("Testing simple methods", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			t.Run("Getting all users and calling `PrefixedUser`", func(t *testing.T) {
				users := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				res := users.PrefixedUser("Prefix")
				assert.EqualValues(t, res[0], "Prefix: Jane A. Smith [<jane.smith@example.com>]")
			})
			t.Run("Calling super on subset", func(t *testing.T) {
				assert.EqualValues(t, h.User().NewSet(env).SearchAll().SubSetSuper(), "Jane A. SmithJohn Smith")
			})
			t.Run("Calling recursive method", func(t *testing.T) {
				assert.EqualValues(t, h.User().NewSet(env).RecursiveMethod(3, "Start"), "> > > > Start <, recursion 3 <, recursion 2 <, recursion 1 <")
			})
		}))
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	t.Run("Testing non stored computed fields", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			t.Run("Getting one user (Jane) and checking DisplayName", func(t *testing.T) {
				users := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				assert.EqualValues(t, users.DecoratedName(), "User: Jane A. Smith [<jane.smith@example.com>]")
			})
			t.Run("Getting all users (Jane & Will) and checking DisplayName", func(t *testing.T) {
				users := h.User().NewSet(env).OrderBy("Name")
				assert.EqualValues(t, users.Len(), 3)
				userRecs := users.Records()
				assert.EqualValues(t, userRecs[0].DecoratedName(), "User: Jane A. Smith [<jane.smith@example.com>]")
				assert.EqualValues(t, userRecs[1].DecoratedName(), "User: John Smith [<jsmith2@example.com>]")
				assert.EqualValues(t, userRecs[2].DecoratedName(), "User: Will Smith [<will.smith@example.com>]")
			})
			t.Run("Testing built-in DisplayName", func(t *testing.T) {
				users := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.DisplayName(), "Jane A. Smith")
			})
		}))
	})
}

func TestComputedStoredFields(t *testing.T) {
	t.Run("Testing stored computed fields", func(t *testing.T) {
		assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			t.Run("Checking that user Jane is 23", func(t *testing.T) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.Age(), 23)
			})
			t.Run("Checking that user Will has no age since no profile", func(t *testing.T) {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				assert.EqualValues(t, userWill.Age(), 0)
			})
			t.Run("It's Jane's birthday, change her age, commit and check", func(t *testing.T) {
				jane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				assert.EqualValues(t, jane.Name(), "Jane A. Smith")
				assert.EqualValues(t, jane.Profile().Money(), 12345)
				jane.Profile().SetAge(24)

				jane.Load()
				jane.Profile().Load()
				assert.EqualValues(t, jane.Age(), 24)
			})
			t.Run("Adding a Profile to Will, writing to DB and checking Will's age", func(t *testing.T) {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				userWill.Load()
				assert.EqualValues(t, userWill.Name(), "Will Smith")
				willProfileData := h.Profile().NewData().
					SetAge(36).
					SetMoney(5100)
				willProfile := h.Profile().Create(env, willProfileData)
				userWill.SetProfile(willProfile)

				userWill.Load()
				assert.EqualValues(t, userWill.Age(), 36)
			})
			t.Run("Checking inverse method by changing will's age", func(t *testing.T) {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				userWill.Load()
				assert.EqualValues(t, userWill.Age(), 36)
				userWill.SetAge(34)
				assert.EqualValues(t, userWill.Age(), 34)
				userWill.Load()
				assert.EqualValues(t, userWill.Age(), 34)
			})
			t.Run("Checking that unlinking a record recomputes their dependencies", func(t *testing.T) {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				userWill.Profile().Unlink()
				assert.EqualValues(t, userWill.Age(), 0)
			})
			t.Run("Recreating a profile for userWill", func(t *testing.T) {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				willProfileData := h.Profile().NewData().
					SetAge(36).
					SetMoney(5100)
				willProfile := h.Profile().Create(env, willProfileData)
				userWill.SetProfile(willProfile)
				assert.EqualValues(t, userWill.Age(), 36)
			})
			t.Run("Checking that setting a computed field with no inverse panics", func(t *testing.T) {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				assert.Panics(t, func() { userWill.SetDecoratedName("FooBar") })
			})
		}))
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	t.Run("Testing non stored related fields", func(t *testing.T) {
		t.Run("Checking that users PMoney is correct", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJohn := h.User().Search(env, q.User().Name().Equals("John Smith"))
				assert.EqualValues(t, userJohn.Len(), 1)
				assert.EqualValues(t, userJohn.PMoney(), 0)
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.PMoney(), 12345)
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				assert.EqualValues(t, userWill.PMoney(), 5100)
			}))
		})
		t.Run("Checking that PMoney is correct after update of Profile", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.PMoney(), 12345)
				userJane.Profile().SetMoney(54321)
				assert.EqualValues(t, userJane.PMoney(), 54321)
			}))
		})
		t.Run("Checking that we can update PMoney directly", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.PMoney(), 12345)
				userJane.SetPMoney(67890)
				assert.EqualValues(t, userJane.Profile().Money(), 67890)
				assert.EqualValues(t, userJane.PMoney(), 67890)
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				assert.EqualValues(t, userWill.PMoney(), 5100)

				userJane.Union(userWill).SetPMoney(100)
				assert.EqualValues(t, userJane.Profile().Money(), 100)
				assert.EqualValues(t, userJane.PMoney(), 100)
				assert.EqualValues(t, userWill.Profile().Money(), 100)
				assert.EqualValues(t, userWill.PMoney(), 100)
			}))
		})
		t.Run("Checking that we can search PMoney directly", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				pmoneyUser := h.User().Search(env, q.User().PMoney().Equals(12345))
				assert.EqualValues(t, pmoneyUser.Len(), 1)
				assert.EqualValues(t, pmoneyUser.Ids()[0], userJane.Ids()[0])
				pUsers := h.User().Search(env, q.User().PMoney().Equals(12345).Or().PMoney().Equals(5100))
				assert.EqualValues(t, pUsers.Len(), 2)
				assert.Contains(t, pUsers.Ids(), userJane.Ids()[0])
				assert.Contains(t, pUsers.Ids(), userWill.Ids()[0])
			}))
		})
		t.Run("Checking that we can order by PMoney", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				userJane.SetPMoney(64)
				pUsers := h.User().NewSet(env).SearchAll().OrderBy("PMoney DESC")
				assert.EqualValues(t, pUsers.Len(), 3)
				pUsersRecs := pUsers.Records()
				// pUsersRecs[0] is userJohn because its pMoney is Null.
				assert.True(t, pUsersRecs[1].Equals(userWill))
				assert.True(t, pUsersRecs[2].Equals(userJane))
			}))
		})
		t.Run("Checking that we can chain related fields", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				post := h.Post().Search(env, q.Post().Title().Equals("1st Post"))
				assert.EqualValues(t, post.Len(), 1)
				assert.EqualValues(t, post.WriterMoney(), 12345)
			}))
		})
		t.Run("Checking that we can chain on a related M2O", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				comment := h.Comment().Search(env, q.Comment().Text().Equals("First Comment"))
				assert.EqualValues(t, comment.Len(), 1)
				assert.EqualValues(t, comment.WriterMoney(), 12345)
				assert.True(t, comment.PostWriter().Equals(userJane))
			}))
		})
	})
}

func TestEmbeddedModels(t *testing.T) {
	t.Run("Testing embedded models", func(t *testing.T) {
		assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
			t.Run("Checking that Jane's resume exists", func(t *testing.T) {
				assert.False(t, userJane.Resume().IsEmpty())
				assert.True(t, userJane.Resume().IsNotEmpty())
			})
			t.Run("Adding a proper resume to Jane", func(t *testing.T) {
				userJane.Resume().SetExperience("Hexya developer for 10 years")
				userJane.Resume().SetEducation("MIT")
				userJane.Resume().SetLeisure("Music, Sports")
				userJane.SetEducation("Berkeley")
			})
			t.Run("Checking that we can access jane's resume directly", func(t *testing.T) {
				assert.EqualValues(t, userJane.Experience(), "Hexya developer for 10 years")
				assert.EqualValues(t, userJane.Leisure(), "Music, Sports")
				assert.EqualValues(t, userJane.Education(), "Berkeley")
				assert.EqualValues(t, userJane.Resume().Experience(), "Hexya developer for 10 years")
				assert.EqualValues(t, userJane.Resume().Leisure(), "Music, Sports")
				assert.EqualValues(t, userJane.Resume().Education(), "MIT")
			})
		}))
	})
}

func TestMixedInModels(t *testing.T) {
	t.Run("Testing mixed in models", func(t *testing.T) {
		t.Run("Checking that mixed in functions are correctly inherited", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				janeProfile := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com")).Profile()
				assert.EqualValues(t, janeProfile.PrintAddress(), "[<165 5th Avenue, 0305 New York>, USA]")
				assert.EqualValues(t, janeProfile.SayHello(), "Hello !")
			}))
		})
		t.Run("Checking mixing in all models", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userJane.SetActive(true)
				assert.EqualValues(t, userJane.Active(), true)
				assert.EqualValues(t, userJane.IsActivated(), true)
				janeProfile := userJane.Profile()
				janeProfile.SetActive(true)
				assert.EqualValues(t, janeProfile.Active(), true)
				assert.EqualValues(t, janeProfile.IsActivated(), true)
			}))
		})
	})
}

func TestInvalidRecordSets(t *testing.T) {
	t.Run("Testing Invalid Recordsets", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			rc := models.InvalidRecordCollection("User")
			rs := rc.Wrap("User").(m.UserSet)
			t.Run("Getting a field on an invalid RecordSet should return empty value", func(t *testing.T) {
				assert.EqualValues(t, rs.Name(), "")
			})
			t.Run("Getting a relation field on an invalid RecordSet should return invalid recordset", func(t *testing.T) {
				profile := rs.Profile()
				assert.False(t, profile.IsValid())
			})
			t.Run("Calling a method on an invalid RecordSet should panic", func(t *testing.T) {
				assert.Panics(t, func() { rs.PrefixedUser(">>") })
			})
		}))
	})
}
