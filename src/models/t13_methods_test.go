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

package models

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/security"
)

func TestMethods(t *testing.T) {
	t.Run("Testing simple methods", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			t.Run("Getting all users and calling `PrefixedUser`", func(t *testing.T) {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				res := users.Call("PrefixedUser", "Prefix")
				assert.EqualValues(t, res.([]string)[0], "Prefix: Jane A. Smith [<jane.smith@example.com>]")
			})
			t.Run("Calling `PrefixedUser` with context", func(t *testing.T) {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				res := users.WithContext("use_double_square", true).Call("PrefixedUser", "Prefix")
				assert.EqualValues(t, res.([]string)[0], "Prefix: Jane A. Smith [[jane.smith@example.com]]")
			})
			t.Run("Calling super on subset", func(t *testing.T) {
				users := env.Pool("User").SearchAll()
				assert.EqualValues(t, users.Call("SubSetSuper").(string), "Jane A. SmithJohn Smith")
			})
			t.Run("Calling recursive method", func(t *testing.T) {
				users := env.Pool("User")
				assert.EqualValues(t, users.Call("RecursiveMethod", 3, "Start"), "> > > > Start <, recursion 3 <, recursion 2 <, recursion 1 <")
			})
			t.Run("Direct calls from method object", func(t *testing.T) {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				res := users.Model().Methods().MustGet("PrefixedUser").Call(users, "Prefix")
				assert.EqualValues(t, res.([]string)[0], "Prefix: Jane A. Smith [<jane.smith@example.com>]")
				resMulti := users.Model().Methods().MustGet("OnChangeName").CallMulti(users)
				res1 := resMulti[0].(*ModelData)
				assert.Contains(t, res1.FieldMap, "decorated_name")
				assert.EqualValues(t, res1.FieldMap["decorated_name"], "User: Jane A. Smith [<jane.smith@example.com>]")
			})
		}))
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	t.Run("Testing non stored computed fields", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			t.Run("Getting one user (Jane) and checking DisplayName", func(t *testing.T) {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, users.Get(decoratedName), "User: Jane A. Smith [<jane.smith@example.com>]")
			})
			t.Run("Getting all users (Jane & Will) and checking DisplayName", func(t *testing.T) {
				users := env.Pool("User").OrderBy("Name").Call("Fetch").(RecordSet).Collection()
				assert.EqualValues(t, users.Len(), 3)
				userRecs := users.Records()
				assert.EqualValues(t, userRecs[0].Get(decoratedName), "User: Jane A. Smith [<jane.smith@example.com>]")
				assert.EqualValues(t, userRecs[1].Get(decoratedName), "User: John Smith [<jsmith2@example.com>]")
				assert.EqualValues(t, userRecs[2].Get(decoratedName), "User: Will Smith [<will.smith@example.com>]")
			})
			t.Run("Testing built-in DisplayName", func(t *testing.T) {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, users.Get(displayName).(string), "Jane A. Smith")
			})
			t.Run("Testing computed field through a related field", func(t *testing.T) {
				users := env.Pool("User")
				jane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, jane.Get(other), "Other information")
				assert.EqualValues(t, jane.Get(resume).(RecordSet).Collection().Get(other), "Other information")
			})
		}))
	})
}

func TestComputedStoredFields(t *testing.T) {
	t.Run("Testing stored computed fields", func(t *testing.T) {
		assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			profileModel := Registry.MustGet("Profile")
			t.Run("Checking that user Jane is 23", func(t *testing.T) {
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.Get(age), 23)
			})
			t.Run("Checking that user Will has no age since no profile", func(t *testing.T) {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				assert.EqualValues(t, userWill.Get(age), 0)
			})
			t.Run("It's Jane's birthday, change her age, commit and check", func(t *testing.T) {
				jane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, jane.Get(Name), "Jane A. Smith")
				assert.EqualValues(t, jane.Get(profile).(RecordSet).Collection().Get(money), 12345)
				jane.Get(profile).(RecordSet).Collection().Set(age, 24)

				jane.Load()
				jane.Get(profile).(RecordSet).Collection().Load()
				assert.EqualValues(t, jane.Get(age), 24)
			})
			t.Run("Adding a Profile to Will, writing to DB and checking Will's age", func(t *testing.T) {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				userWill.Load()
				assert.EqualValues(t, userWill.Get(Name), "Will Smith")
				willProfileData := NewModelData(profileModel).
					Set(age, 36).
					Set(money, 5100)
				willProfile := env.Pool("Profile").Call("Create", willProfileData)
				userWill.Set(profile, willProfile)

				userWill.Load()
				assert.EqualValues(t, userWill.Get(age), 36)
			})
			t.Run("Checking inverse method by changing will's age", func(t *testing.T) {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				userWill.Load()
				assert.EqualValues(t, userWill.Get(age), 36)
				userWill.Set(age, int16(34))
				assert.EqualValues(t, userWill.Get(age), 34)
				userWill.Load()
				assert.EqualValues(t, userWill.Get(age), 34)
			})
			t.Run("Checking that unlinking a record recomputes their dependencies", func(t *testing.T) {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				userWill.Get(profile).(RecordSet).Collection().Call("Unlink")
				assert.EqualValues(t, userWill.Get(age), 0)
			})
			t.Run("Recreating a profile for userWill", func(t *testing.T) {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				willProfileData := NewModelData(profileModel).
					Set(age, 36).
					Set(money, 5100)
				willProfile := env.Pool("Profile").Call("Create", willProfileData)
				userWill.Set(profile, willProfile)
				assert.EqualValues(t, userWill.Get(age), 36)
			})
			t.Run("Checking that setting a computed field with no inverse panics", func(t *testing.T) {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				assert.Panics(t, func() { userWill.Set(decoratedName, "FooBar") })
			})
			t.Run("Checking that a computed field can trigger another one", func(t *testing.T) {
				jane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				post := jane.Get(posts).(RecordSet).Collection().Records()[0]
				assert.EqualValues(t, jane.Get(Name), "Jane A. Smith")
				assert.EqualValues(t, post.Get(writerAge), 24)
				jane.Get(profile).(RecordSet).Collection().Set(age, 25)
				assert.EqualValues(t, post.Get(writerAge), 25)
				jane.Set(age, int16(24))
				assert.EqualValues(t, post.Get(writerAge), 24)
			})
		}))
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	t.Run("Testing non stored related fields", func(t *testing.T) {
		t.Run("Checking that users PMoney is correct", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJohn := users.Search(users.Model().Field(Name).Equals("John Smith"))
				assert.EqualValues(t, userJohn.Len(), 1)
				assert.EqualValues(t, userJohn.Get(pMoney), 0)
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.Get(pMoney), 12345)
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				assert.EqualValues(t, userWill.Get(pMoney), 5100)
			}))
		})
		t.Run("Checking that PMoney is correct after update of Profile", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.Get(pMoney), 12345)
				userJane.Get(profile).(RecordSet).Collection().Set(money, 54321)
				assert.EqualValues(t, userJane.Get(pMoney), 54321)
			}))
		})
		t.Run("Checking that we can update PMoney directly", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.Get(pMoney), 12345)
				userJane.Set(pMoney, 67890)
				assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(money), 67890)
				assert.EqualValues(t, userJane.Get(pMoney), 67890)
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				assert.EqualValues(t, userWill.Get(pMoney), 5100)

				userJane.Union(userWill).Set(pMoney, 100)
				assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(money), 100)
				assert.EqualValues(t, userJane.Get(pMoney), 100)
				assert.EqualValues(t, userWill.Get(profile).(RecordSet).Collection().Get(money), 100)
				assert.EqualValues(t, userWill.Get(pMoney), 100)
			}))
		})
		t.Run("Checking that we can search PMoney directly", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				pmoneyUser := users.Search(users.Model().Field(pMoney).Equals(12345))
				assert.EqualValues(t, pmoneyUser.Len(), 1)
				assert.EqualValues(t, pmoneyUser.Ids()[0], userJane.Ids()[0])
				pUsers := users.Search(users.Model().Field(pMoney).Equals(12345).Or().Field(pMoney).Equals(5100))
				assert.EqualValues(t, pUsers.Len(), 2)
				assert.Contains(t, pUsers.Ids(), userJane.Ids()[0])
				assert.Contains(t, pUsers.Ids(), userWill.Ids()[0])
			}))
		})
		t.Run("Checking that we can order by PMoney", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				userJane.Set(pMoney, 64)
				pUsers := users.SearchAll().OrderBy("PMoney DESC")
				assert.EqualValues(t, pUsers.Len(), 3)
				pUsersRecs := pUsers.Records()
				// pUsersRecs[0] is userJohn because its pMoney is Null.
				assert.True(t, pUsersRecs[1].Equals(userWill))
				assert.True(t, pUsersRecs[2].Equals(userJane))
			}))
		})
		t.Run("Checking that we can chain related fields", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				emptyPosts := env.Pool("Post")
				post := emptyPosts.Search(emptyPosts.Model().Field(title).Equals("1st Post"))
				assert.EqualValues(t, post.Len(), 1)
				assert.EqualValues(t, post.Get(writerMoney), 12345)
			}))
		})
		t.Run("Checking that we can chain on a related M2O", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				emptyComments := env.Pool("Comment")
				comment := emptyComments.Search(emptyComments.Model().Field(text).Equals("First Comment"))
				assert.EqualValues(t, comment.Len(), 1)
				assert.EqualValues(t, comment.Get(writerMoney), 12345)
				assert.True(t, comment.Get(postWriter).(RecordSet).Collection().Equals(userJane))
			}))
		})
	})
}

func TestEmbeddedModels(t *testing.T) {
	t.Run("Testing embedded models", func(t *testing.T) {
		assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
			t.Run("Checking that Jane's resume exists", func(t *testing.T) {
				assert.False(t, userJane.Get(resume).(RecordSet).IsEmpty())
				assert.True(t, userJane.Get(resume).(RecordSet).IsNotEmpty())
			})
			t.Run("Adding a proper resume to Jane", func(t *testing.T) {
				userJane.Get(resume).(RecordSet).Collection().Set(experience, "Hexya developer for 10 years")
				userJane.Set(leisure, "Music, Sports")
				userJane.Get(resume).(RecordSet).Collection().Set(education, "MIT")
				userJane.Set(education, "Berkeley")
			})
			t.Run("Checking that we can access jane's resume directly", func(t *testing.T) {
				assert.EqualValues(t, userJane.Get(experience), "Hexya developer for 10 years")
				assert.EqualValues(t, userJane.Get(leisure), "Music, Sports")
				assert.EqualValues(t, userJane.Get(education), "Berkeley")
				assert.EqualValues(t, userJane.Get(resume).(RecordSet).Collection().Get(experience), "Hexya developer for 10 years")
				assert.EqualValues(t, userJane.Get(resume).(RecordSet).Collection().Get(leisure), "Music, Sports")
				assert.EqualValues(t, userJane.Get(resume).(RecordSet).Collection().Get(education), "MIT")
			})
		}))
	})
}

func TestMixedInModels(t *testing.T) {
	t.Run("Testing mixed in models", func(t *testing.T) {
		t.Run("Checking that mixed in functions are correctly inherited", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				janeProfile := users.Search(users.Model().Field(email).Equals("jane.smith@example.com")).Get(profile).(RecordSet).Collection()
				assert.EqualValues(t, janeProfile.Call("PrintAddress"), "[<165 5th Avenue, 0305 New York>, USA]")
				assert.EqualValues(t, janeProfile.Call("SayHello"), "Hello !")
			}))
		})
		t.Run("Checking mixing in all models", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Set(active, true)
				assert.EqualValues(t, userJane.Get(active).(bool), true)
				assert.EqualValues(t, userJane.Call("IsActivated").(bool), true)
				janeProfile := userJane.Get(profile).(RecordSet).Collection()
				janeProfile.Set(active, true)
				assert.EqualValues(t, janeProfile.Get(active).(bool), true)
				assert.EqualValues(t, janeProfile.Call("IsActivated").(bool), true)
			}))
		})
	})
}

func TestContextedFields(t *testing.T) {
	t.Run("Testing contexted fields", func(t *testing.T) {
		assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			mTags := env.Pool("Tag")
			decs := env.Pool("TagHexyaDescription")
			var tagc *RecordCollection
			t.Run("Creating record with a single contexted field", func(t *testing.T) {
				tagc = mTags.Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag").
					Set(description, "Translated description")).(RecordSet).Collection()
				assert.EqualValues(t, tagc.Get(descriptionHexyaContexts).(RecordSet).Len(), 1)
				assert.EqualValues(t, tagc.Get(description), "Translated description")

				tagc.WithContext("lang", "fr_FR").Set(description, "Description traduite")
				assert.EqualValues(t, tagc.Get(descriptionHexyaContexts).(RecordSet).Len(), 2)
				assert.EqualValues(t, tagc.Get(description), "Translated description")

				newTag := mTags.WithContext("lang", "fr_FR").Search(mTags.Model().Field(Name).Equals("Contexted tag"))
				newTag.Load(description)
				assert.EqualValues(t, newTag.Get(description), "Description traduite")

				assert.EqualValues(t, tagc.Get(description), "Translated description")
				assert.EqualValues(t, tagc.WithContext("lang", "fr_FR").Get(description), "Description traduite")
				assert.EqualValues(t, tagc.Get(description), "Translated description")
				assert.EqualValues(t, tagc.WithContext("lang", "de_DE").Get(description), "Translated description")

				tagc.WithContext("lang", "fr_FR").Set(description, "Nouvelle traduction")
				assert.EqualValues(t, tagc.Get(description), "Translated description")
				assert.EqualValues(t, tagc.WithContext("lang", "fr_FR").Get(description), "Nouvelle traduction")
				assert.EqualValues(t, tagc.WithContext("lang", "de_DE").Get(description), "Translated description")

				tagc.WithContext("lang", "de_DE").Set(description, "übersetzte Beschreibung")
				assert.EqualValues(t, tagc.Get(description), "Translated description")
				assert.EqualValues(t, tagc.WithContext("lang", "fr_FR").Get(description), "Nouvelle traduction")
				assert.EqualValues(t, tagc.WithContext("lang", "de_DE").Get(description), "übersetzte Beschreibung")

				tagc.WithContext("lang", "es_ES").Set(description, "descripción traducida")
				assert.EqualValues(t, tagc.Get(description), "Translated description")
				assert.EqualValues(t, tagc.WithContext("lang", "fr_FR").Get(description), "Nouvelle traduction")
				assert.EqualValues(t, tagc.WithContext("lang", "de_DE").Get(description), "übersetzte Beschreibung")
				assert.EqualValues(t, tagc.WithContext("lang", "es_ES").Get(description), "descripción traducida")
				assert.EqualValues(t, tagc.WithContext("lang", "it_IT").Get(description), "Translated description")
			})
			t.Run("Creating a record with a contexted field should also create for default context", func(t *testing.T) {
				mTags.WithContext("lang", "fr_FR").Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag 2").
					Set(description, "Description en français")).(RecordSet).Collection()
				tag := mTags.Search(mTags.Model().Field(Name).Equals("Contexted tag 2"))
				assert.EqualValues(t, tag.Get(description), "Description en français")
				assert.EqualValues(t, tag.WithContext("lang", "en_US").Get(description), "Description en français")
				assert.EqualValues(t, tag.WithContext("lang", "fr_FR").Get(description), "Description en français")
				tag.WithContext("lang", "en_US").Set(description, "Description in English")
				assert.EqualValues(t, tag.WithContext("lang", "en_US").Get(description), "Description in English")
				assert.EqualValues(t, tag.WithContext("lang", "fr_FR").Get(description), "Description en français")
				assert.EqualValues(t, tag.WithContext("lang", "de_DE").Get(description), "Description en français")
				assert.EqualValues(t, tag.Get(description), "Description en français")
				thc := decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				assert.EqualValues(t, thc.Len(), 1)
			})
			t.Run("Updating in another transaction should not recreate a default value", func(t *testing.T) {
				tag := mTags.WithContext("lang", "fr_FR").Search(mTags.Model().Field(Name).Equals("Contexted tag 2"))
				assert.EqualValues(t, tag.Get(description), "Description en français")
				thc := decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				assert.EqualValues(t, thc.Len(), 1)

				tag.Set(description, "Nouvelle description en français")
				thc = decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				assert.EqualValues(t, thc.Len(), 1)
				assert.EqualValues(t, tag.Get(description), "Nouvelle description en français")
			})
			t.Run("Changing language should recreate a default value (new transaction)", func(t *testing.T) {
				tag := mTags.WithContext("lang", "es_ES").Search(mTags.Model().Field(Name).Equals("Contexted tag 2"))
				assert.EqualValues(t, tag.Get(description), "Description en français")
				thc := decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				assert.EqualValues(t, thc.Len(), 1)

				tag.Set(description, "descripción traducida")
				thc = decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				assert.EqualValues(t, thc.Len(), 1)
				assert.EqualValues(t, tag.Get(description), "descripción traducida")
			})
			t.Run("Deleting a record with a contexted field should delete all contexts", func(t *testing.T) {
				newTag := mTags.Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag 3").
					Set(description, "Description to translate")).(RecordSet).Collection()
				assert.EqualValues(t, newTag.Get(description), "Description to translate")
				newTag.WithContext("lang", "fr_FR").Set(description, "Description en français")
				assert.EqualValues(t, newTag.WithContext("lang", "fr_FR").Get(description), "Description en français")
				newTag.WithContext("lang", "de_DE").Set(description, "übersetzte Beschreibung")
				assert.EqualValues(t, newTag.WithContext("lang", "de_DE").Get(description), "übersetzte Beschreibung")
				nID := newTag.Ids()[0]
				newTag.Call("Unlink")
				dec := decs.Search(decs.Model().Field(record).Equals(nID).Or().Field(record).IsNull())
				assert.True(t, dec.IsEmpty())
			})
		}))
	})
	t.Run("Testing contexted group by queries", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			mTags := env.Pool("Tag")
			mTags.SearchAll().Call("Unlink")
			t.Run("Simple group by query", func(t *testing.T) {
				tag1 := mTags.Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag").
					Set(description, "Translated description")).(RecordSet).Collection()
				tag1.WithContext("lang", "fr_FR").Set(description, "Description traduite")
				tag2 := mTags.Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag").
					Set(description, "Translated description")).(RecordSet).Collection()
				tag2.WithContext("lang", "fr_FR").Set(description, "Description traduite")
				mTags.Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag").
					Set(description, "Other description")).(RecordSet).Collection()
				gbq := mTags.WithContext("lang", "fr_FR").SearchAll().GroupBy(FieldName(description)).Aggregates(FieldName(description))
				assert.Len(t, gbq, 2)
				assert.True(t, gbq[0].Values.Has(description))
				des := gbq[0].Values.Get(description)
				assert.Contains(t, []string{"Other description", "Description traduite"}, des)
				switch des {
				case "Description traduite":
					assert.EqualValues(t, gbq[0].Count, 2)
					assert.EqualValues(t, gbq[1].Count, 1)
					des1 := gbq[1].Values.Get(description)
					assert.EqualValues(t, des1, "Other description")
				case "Other description":
					assert.EqualValues(t, gbq[0].Count, 1)
					assert.EqualValues(t, gbq[1].Count, 2)
					des1 := gbq[1].Values.Get(description)
					assert.EqualValues(t, des1, "Description traduite")
				default:
					t.FailNow()
				}
				gbq = mTags.SearchAll().GroupBy(FieldName(description)).Aggregates(FieldName(description))
				assert.True(t, gbq[0].Values.Has(description))
				des = gbq[0].Values.Get(description)
				assert.Contains(t, []string{"Other description", "Translated description"}, des)
				switch des {
				case "Translated description":
					assert.EqualValues(t, gbq[0].Count, 2)
					assert.EqualValues(t, gbq[1].Count, 1)
					des1 := gbq[1].Values.Get(description)
					assert.EqualValues(t, des1, "Other description")
				case "Other description":
					assert.EqualValues(t, gbq[0].Count, 1)
					assert.EqualValues(t, gbq[1].Count, 2)
					des1 := gbq[1].Values.Get(description)
					assert.EqualValues(t, des1, "Translated description")
				default:
					t.FailNow()
				}
			})
		}))
	})
}

func TestRecursionProtection(t *testing.T) {
	t.Run("Testing protection against recursion", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			t.Run("Endless recursive method calls should panic", func(t *testing.T) {
				assert.Panics(t, func() { env.Pool("User").Call("EndlessRecursion") })
			})
			t.Run("Loop calls should not trigger recursion protection", func(t *testing.T) {
				assert.NotPanics(t, func() {
					for range int(maxRecursionDepth) + 10 {
						env.Pool("Profile").Call("SayHello")
					}
				})
			})
			t.Run("Recursion should be triggered exactly at the max recursion depth", func(t *testing.T) {
				assert.NotPanics(t, func() { env.Pool("User").Call("RecursiveMethod", int(maxRecursionDepth)/2-1, "Hi!") })
				assert.Panics(t, func() { env.Pool("User").Call("RecursiveMethod", int(maxRecursionDepth)/2, "Hi!") })
			})
		}))
	})
}

func TestInternalMethodFunctions(t *testing.T) {
	t.Run("Testing internal method functions", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
			RegisterRecordSetWrapper("Profile", TestProfileSet{})
			t.Run("convertFunctionArg", func(t *testing.T) {
				assert.EqualValues(t, convertFunctionArg(reflect.TypeFor[int64](), 126).Interface(), 126)
				prof := convertFunctionArg(reflect.TypeFor[TestProfileSet](), userJane.Get(profile))
				assert.EqualValues(t, prof.Type(), reflect.TypeFor[TestProfileSet]())
				assert.True(t, prof.Interface().(TestProfileSet).Collection().Equals(userJane.Get(profile).(RecordSet).Collection()))
				prof = convertFunctionArg(reflect.TypeFor[*RecordCollection](), userJane.Get(profile))
				assert.EqualValues(t, prof.Type(), reflect.TypeFor[*RecordCollection]())
				assert.True(t, prof.Interface().(*RecordCollection).Equals(userJane.Get(profile).(RecordSet).Collection()))
				vals := convertFunctionArg(reflect.TypeFor[*ModelData](), NewModelData(users.model, FieldMap{"name": "Mike"}))
				assert.EqualValues(t, vals.Type(), reflect.TypeFor[*ModelData]())
				assert.Len(t, vals.Interface().(*ModelData).FieldMap, 1)
				assert.Contains(t, vals.Interface().(*ModelData).FieldMap, "name")
				assert.EqualValues(t, vals.Interface().(*ModelData).FieldMap["name"], "Mike")
				vals = convertFunctionArg(reflect.TypeFor[*TestUserData](), NewModelData(users.model, FieldMap{"IsStaff": true}))
				assert.EqualValues(t, vals.Type(), reflect.TypeFor[*ModelData]())
				assert.Len(t, vals.Interface().(*ModelData).FieldMap, 1)
				assert.Contains(t, vals.Interface().(*ModelData).FieldMap, "is_staff")
				assert.EqualValues(t, vals.Interface().(*ModelData).FieldMap["is_staff"], true)
				cond := users.Model().Field(Name).Equals("Jane Smith")
				c := convertFunctionArg(reflect.TypeFor[TestUserCondition](), cond)
				assert.EqualValues(t, c.Type(), reflect.TypeFor[TestUserCondition]())
				assert.EqualValues(t, c.Interface().(TestUserCondition).Underlying().String(), cond.String())
			})
			t.Run("MethodType", func(t *testing.T) {
				meth := users.model.methods.MustGet("OnChangeMana")
				assert.EqualValues(t, meth.MethodType(), reflect.TypeFor[func(*RecordCollection) *ModelData]())
			})
			t.Run("Name", func(t *testing.T) {
				meth := users.model.methods.MustGet("ComputeCoolType")
				assert.EqualValues(t, meth.Name(), "ComputeCoolType")
			})
		}))
	})
}

func TestInvalidRecordSets(t *testing.T) {
	t.Run("Testing Invalid Recordsets", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			rc := InvalidRecordCollection("User")
			t.Run("Getting a field on an invalid RecordSet should return empty value", func(t *testing.T) {
				assert.EqualValues(t, rc.Get(Name), "")
			})
			t.Run("Getting a relation field on an invalid RecordSet should return invalid recordset", func(t *testing.T) {
				rel, ok := rc.Get(profile).(*RecordCollection)
				assert.True(t, ok)
				assert.False(t, rel.IsValid())
				assert.EqualValues(t, rel.model.name, "Profile")
			})
			t.Run("Calling a method on an invalid RecordSet should panic", func(t *testing.T) {
				assert.Panics(t, func() { rc.Call("PrefixedUser", ">>") })
			})
		}))
	})
}
