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

package models

import (
	"reflect"
	"testing"

	"github.com/hexya-erp/hexya/src/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMethods(t *testing.T) {
	Convey("Testing simple methods", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Getting all users and calling `PrefixedUser`", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				res := users.Call("PrefixedUser", "Prefix")
				So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
			})
			Convey("Calling `PrefixedUser` with context", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				res := users.WithContext("use_double_square", true).Call("PrefixedUser", "Prefix")
				So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [[jane.smith@example.com]]")
			})
			Convey("Calling super on subset", func() {
				users := env.Pool("User").SearchAll()
				So(users.Call("SubSetSuper").(string), ShouldEqual, "Jane A. SmithJohn Smith")
			})
			Convey("Calling recursive method", func() {
				users := env.Pool("User")
				So(users.Call("RecursiveMethod", 3, "Start"), ShouldEqual, "> > > > Start <, recursion 3 <, recursion 2 <, recursion 1 <")
			})
			Convey("Direct calls from method object", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				res := users.Model().Methods().MustGet("PrefixedUser").Call(users, "Prefix")
				So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
				resMulti := users.Model().Methods().MustGet("OnChangeName").CallMulti(users)
				res1 := resMulti[0].(*ModelData)
				So(res1.FieldMap, ShouldContainKey, "decorated_name")
				So(res1.FieldMap["decorated_name"], ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			})
		}), ShouldBeNil)
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	Convey("Testing non stored computed fields", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Getting one user (Jane) and checking DisplayName", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				So(users.Get(decoratedName), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			})
			Convey("Getting all users (Jane & Will) and checking DisplayName", func() {
				users := env.Pool("User").OrderBy("Name").Call("Fetch").(RecordSet).Collection()
				So(users.Len(), ShouldEqual, 3)
				userRecs := users.Records()
				So(userRecs[0].Get(decoratedName), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
				So(userRecs[1].Get(decoratedName), ShouldEqual, "User: John Smith [<jsmith2@example.com>]")
				So(userRecs[2].Get(decoratedName), ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
			})
			Convey("Testing built-in DisplayName", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				So(users.Get(displayName).(string), ShouldEqual, "Jane A. Smith")
			})
			Convey("Testing computed field through a related field", func() {
				users := env.Pool("User")
				jane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				So(jane.Get(other), ShouldEqual, "Other information")
				So(jane.Get(resume).(RecordSet).Collection().Get(other), ShouldEqual, "Other information")
			})
		}), ShouldBeNil)
	})
}

func TestComputedStoredFields(t *testing.T) {
	Convey("Testing stored computed fields", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			profileModel := Registry.MustGet("Profile")
			Convey("Checking that user Jane is 23", func() {
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				So(userJane.Get(age), ShouldEqual, 23)
			})
			Convey("Checking that user Will has no age since no profile", func() {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				So(userWill.Get(age), ShouldEqual, 0)
			})
			Convey("It's Jane's birthday, change her age, commit and check", func() {
				jane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				So(jane.Get(Name), ShouldEqual, "Jane A. Smith")
				So(jane.Get(profile).(RecordSet).Collection().Get(money), ShouldEqual, 12345)
				jane.Get(profile).(RecordSet).Collection().Set(age, 24)

				jane.Load()
				jane.Get(profile).(RecordSet).Collection().Load()
				So(jane.Get(age), ShouldEqual, 24)
			})
			Convey("Adding a Profile to Will, writing to DB and checking Will's age", func() {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				userWill.Load()
				So(userWill.Get(Name), ShouldEqual, "Will Smith")
				willProfileData := NewModelData(profileModel).
					Set(age, 36).
					Set(money, 5100)
				willProfile := env.Pool("Profile").Call("Create", willProfileData)
				userWill.Set(profile, willProfile)

				userWill.Load()
				So(userWill.Get(age), ShouldEqual, 36)
			})
			Convey("Checking inverse method by changing will's age", func() {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				userWill.Load()
				So(userWill.Get(age), ShouldEqual, 36)
				userWill.Set(age, int16(34))
				So(userWill.Get(age), ShouldEqual, 34)
				userWill.Load()
				So(userWill.Get(age), ShouldEqual, 34)
			})
			Convey("Checking that unlinking a record recomputes their dependencies", func() {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				userWill.Get(profile).(RecordSet).Collection().Call("Unlink")
				So(userWill.Get(age), ShouldEqual, 0)
			})
			Convey("Recreating a profile for userWill", func() {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				willProfileData := NewModelData(profileModel).
					Set(age, 36).
					Set(money, 5100)
				willProfile := env.Pool("Profile").Call("Create", willProfileData)
				userWill.Set(profile, willProfile)
				So(userWill.Get(age), ShouldEqual, 36)
			})
			Convey("Checking that setting a computed field with no inverse panics", func() {
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				So(func() { userWill.Set(decoratedName, "FooBar") }, ShouldPanic)
			})
			Convey("Checking that a computed field can trigger another one", func() {
				jane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				post := jane.Get(posts).(RecordSet).Collection().Records()[0]
				So(jane.Get(Name), ShouldEqual, "Jane A. Smith")
				So(post.Get(writerAge), ShouldEqual, 24)
				jane.Get(profile).(RecordSet).Collection().Set(age, 25)
				So(post.Get(writerAge), ShouldEqual, 25)
				jane.Set(age, int16(24))
				So(post.Get(writerAge), ShouldEqual, 24)
			})
		}), ShouldBeNil)
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	Convey("Testing non stored related fields", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			Convey("Checking that users PMoney is correct", func() {
				userJohn := users.Search(users.Model().Field(Name).Equals("John Smith"))
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn.Get(pMoney), ShouldEqual, 0)
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				So(userJane.Get(pMoney), ShouldEqual, 12345)
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				So(userWill.Get(pMoney), ShouldEqual, 5100)
			})
			Convey("Checking that PMoney is correct after update of Profile", func() {
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				So(userJane.Get(pMoney), ShouldEqual, 12345)
				userJane.Get(profile).(RecordSet).Collection().Set(money, 54321)
				So(userJane.Get(pMoney), ShouldEqual, 54321)
			})
			Convey("Checking that we can update PMoney directly", func() {
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				So(userJane.Get(pMoney), ShouldEqual, 12345)
				userJane.Set(pMoney, 67890)
				So(userJane.Get(profile).(RecordSet).Collection().Get(money), ShouldEqual, 67890)
				So(userJane.Get(pMoney), ShouldEqual, 67890)
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				So(userWill.Get(pMoney), ShouldEqual, 5100)

				userJane.Union(userWill).Set(pMoney, 100)
				So(userJane.Get(profile).(RecordSet).Collection().Get(money), ShouldEqual, 100)
				So(userJane.Get(pMoney), ShouldEqual, 100)
				So(userWill.Get(profile).(RecordSet).Collection().Get(money), ShouldEqual, 100)
				So(userWill.Get(pMoney), ShouldEqual, 100)
			})
			Convey("Checking that we can search PMoney directly", func() {
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				pmoneyUser := users.Search(users.Model().Field(pMoney).Equals(12345))
				So(pmoneyUser.Len(), ShouldEqual, 1)
				So(pmoneyUser.Ids()[0], ShouldEqual, userJane.Ids()[0])
				pUsers := users.Search(users.Model().Field(pMoney).Equals(12345).Or().Field(pMoney).Equals(5100))
				So(pUsers.Len(), ShouldEqual, 2)
				So(pUsers.Ids(), ShouldContain, userJane.Ids()[0])
				So(pUsers.Ids(), ShouldContain, userWill.Ids()[0])
			})
			Convey("Checking that we can order by PMoney", func() {
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userWill := users.Search(users.Model().Field(email).Equals("will.smith@example.com"))
				userJane.Set(pMoney, 64)
				pUsers := users.SearchAll().OrderBy("PMoney DESC")
				So(pUsers.Len(), ShouldEqual, 3)
				pUsersRecs := pUsers.Records()
				// pUsersRecs[0] is userJohn because its pMoney is Null.
				So(pUsersRecs[1].Equals(userWill), ShouldBeTrue)
				So(pUsersRecs[2].Equals(userJane), ShouldBeTrue)
			})
			Convey("Checking that we can chain related fields", func() {
				emptyPosts := env.Pool("Post")
				post := emptyPosts.Search(emptyPosts.Model().Field(title).Equals("1st Post"))
				So(post.Len(), ShouldEqual, 1)
				So(post.Get(writerMoney), ShouldEqual, 12345)
			})
			Convey("Checking that we can chain on a related M2O", func() {
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				emptyComments := env.Pool("Comment")
				comment := emptyComments.Search(emptyComments.Model().Field(text).Equals("First Comment"))
				So(comment.Len(), ShouldEqual, 1)
				So(comment.Get(writerMoney), ShouldEqual, 12345)
				So(comment.Get(postWriter).(RecordSet).Collection().Equals(userJane), ShouldBeTrue)
			})
		}), ShouldBeNil)
	})
}

func TestEmbeddedModels(t *testing.T) {
	Convey("Testing embedded models", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
			Convey("Checking that Jane's resume exists", func() {
				So(userJane.Get(resume).(RecordSet).IsEmpty(), ShouldBeFalse)
				So(userJane.Get(resume).(RecordSet).IsNotEmpty(), ShouldBeTrue)
			})
			Convey("Adding a proper resume to Jane", func() {
				userJane.Get(resume).(RecordSet).Collection().Set(experience, "Hexya developer for 10 years")
				userJane.Set(leisure, "Music, Sports")
				userJane.Get(resume).(RecordSet).Collection().Set(education, "MIT")
				userJane.Set(education, "Berkeley")
			})
			Convey("Checking that we can access jane's resume directly", func() {
				So(userJane.Get(experience), ShouldEqual, "Hexya developer for 10 years")
				So(userJane.Get(leisure), ShouldEqual, "Music, Sports")
				So(userJane.Get(education), ShouldEqual, "Berkeley")
				So(userJane.Get(resume).(RecordSet).Collection().Get(experience), ShouldEqual, "Hexya developer for 10 years")
				So(userJane.Get(resume).(RecordSet).Collection().Get(leisure), ShouldEqual, "Music, Sports")
				So(userJane.Get(resume).(RecordSet).Collection().Get(education), ShouldEqual, "MIT")
			})
		}), ShouldBeNil)
	})
}

func TestMixedInModels(t *testing.T) {
	Convey("Testing mixed in models", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			Convey("Checking that mixed in functions are correctly inherited", func() {
				janeProfile := users.Search(users.Model().Field(email).Equals("jane.smith@example.com")).Get(profile).(RecordSet).Collection()
				So(janeProfile.Call("PrintAddress"), ShouldEqual, "[<165 5th Avenue, 0305 New York>, USA]")
				So(janeProfile.Call("SayHello"), ShouldEqual, "Hello !")
			})
			Convey("Checking mixing in all models", func() {
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Set(active, true)
				So(userJane.Get(active).(bool), ShouldEqual, true)
				So(userJane.Call("IsActivated").(bool), ShouldEqual, true)
				janeProfile := userJane.Get(profile).(RecordSet).Collection()
				janeProfile.Set(active, true)
				So(janeProfile.Get(active).(bool), ShouldEqual, true)
				So(janeProfile.Call("IsActivated").(bool), ShouldEqual, true)
			})
		}), ShouldBeNil)
	})
}

func TestContextedFields(t *testing.T) {
	Convey("Testing contexted fields", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			mTags := env.Pool("Tag")
			decs := env.Pool("TagHexyaDescription")
			var tagc *RecordCollection
			Convey("Creating record with a single contexted field", func() {
				tagc = mTags.Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag").
					Set(description, "Translated description")).(RecordSet).Collection()
				So(tagc.Get(descriptionHexyaContexts).(RecordSet).Len(), ShouldEqual, 1)
				So(tagc.Get(description), ShouldEqual, "Translated description")

				tagc.WithContext("lang", "fr_FR").Set(description, "Description traduite")
				So(tagc.Get(descriptionHexyaContexts).(RecordSet).Len(), ShouldEqual, 2)
				So(tagc.Get(description), ShouldEqual, "Translated description")

				newTag := mTags.WithContext("lang", "fr_FR").Search(mTags.Model().Field(Name).Equals("Contexted tag"))
				newTag.Load(description)
				So(newTag.Get(description), ShouldEqual, "Description traduite")

				So(tagc.Get(description), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "fr_FR").Get(description), ShouldEqual, "Description traduite")
				So(tagc.Get(description), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "de_DE").Get(description), ShouldEqual, "Translated description")

				tagc.WithContext("lang", "fr_FR").Set(description, "Nouvelle traduction")
				So(tagc.Get(description), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "fr_FR").Get(description), ShouldEqual, "Nouvelle traduction")
				So(tagc.WithContext("lang", "de_DE").Get(description), ShouldEqual, "Translated description")

				tagc.WithContext("lang", "de_DE").Set(description, "übersetzte Beschreibung")
				So(tagc.Get(description), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "fr_FR").Get(description), ShouldEqual, "Nouvelle traduction")
				So(tagc.WithContext("lang", "de_DE").Get(description), ShouldEqual, "übersetzte Beschreibung")

				tagc.WithContext("lang", "es_ES").Set(description, "descripción traducida")
				So(tagc.Get(description), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "fr_FR").Get(description), ShouldEqual, "Nouvelle traduction")
				So(tagc.WithContext("lang", "de_DE").Get(description), ShouldEqual, "übersetzte Beschreibung")
				So(tagc.WithContext("lang", "es_ES").Get(description), ShouldEqual, "descripción traducida")
				So(tagc.WithContext("lang", "it_IT").Get(description), ShouldEqual, "Translated description")
			})
			Convey("Creating a record with a contexted field should also create for default context", func() {
				mTags.WithContext("lang", "fr_FR").Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag 2").
					Set(description, "Description en français")).(RecordSet).Collection()
				tag := mTags.Search(mTags.Model().Field(Name).Equals("Contexted tag 2"))
				So(tag.Get(description), ShouldEqual, "Description en français")
				So(tag.WithContext("lang", "en_US").Get(description), ShouldEqual, "Description en français")
				So(tag.WithContext("lang", "fr_FR").Get(description), ShouldEqual, "Description en français")
				tag.WithContext("lang", "en_US").Set(description, "Description in English")
				So(tag.WithContext("lang", "en_US").Get(description), ShouldEqual, "Description in English")
				So(tag.WithContext("lang", "fr_FR").Get(description), ShouldEqual, "Description en français")
				So(tag.WithContext("lang", "de_DE").Get(description), ShouldEqual, "Description en français")
				So(tag.Get(description), ShouldEqual, "Description en français")
				thc := decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				So(thc.Len(), ShouldEqual, 1)
			})
			Convey("Updating in another transaction should not recreate a default value", func() {
				tag := mTags.WithContext("lang", "fr_FR").Search(mTags.Model().Field(Name).Equals("Contexted tag 2"))
				So(tag.Get(description), ShouldEqual, "Description en français")
				thc := decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				So(thc.Len(), ShouldEqual, 1)

				tag.Set(description, "Nouvelle description en français")
				thc = decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				So(thc.Len(), ShouldEqual, 1)
				So(tag.Get(description), ShouldEqual, "Nouvelle description en français")
			})
			Convey("Changing language should recreate a default value (new transaction)", func() {
				tag := mTags.WithContext("lang", "es_ES").Search(mTags.Model().Field(Name).Equals("Contexted tag 2"))
				So(tag.Get(description), ShouldEqual, "Description en français")
				thc := decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				So(thc.Len(), ShouldEqual, 1)

				tag.Set(description, "descripción traducida")
				thc = decs.Search(decs.Model().Field(record).Equals(tag.Ids()[0]).And().Field(lang).IsNull())
				So(thc.Len(), ShouldEqual, 1)
				So(tag.Get(description), ShouldEqual, "descripción traducida")
			})
			Convey("Deleting a record with a contexted field should delete all contexts", func() {
				newTag := mTags.Call("Create", NewModelData(mTags.model).
					Set(Name, "Contexted tag 3").
					Set(description, "Description to translate")).(RecordSet).Collection()
				So(newTag.Get(description), ShouldEqual, "Description to translate")
				newTag.WithContext("lang", "fr_FR").Set(description, "Description en français")
				So(newTag.WithContext("lang", "fr_FR").Get(description), ShouldEqual, "Description en français")
				newTag.WithContext("lang", "de_DE").Set(description, "übersetzte Beschreibung")
				So(newTag.WithContext("lang", "de_DE").Get(description), ShouldEqual, "übersetzte Beschreibung")
				nID := newTag.Ids()[0]
				newTag.Call("Unlink")
				dec := decs.Search(decs.Model().Field(record).Equals(nID).Or().Field(record).IsNull())
				So(dec.IsEmpty(), ShouldBeTrue)
			})
		}), ShouldBeNil)
	})
	Convey("Testing contexted group by queries", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			mTags := env.Pool("Tag")
			mTags.SearchAll().Call("Unlink")
			Convey("Simple group by query", func() {
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
				So(gbq, ShouldHaveLength, 2)
				So(gbq[0].Values.Has(description), ShouldBeTrue)
				des := gbq[0].Values.Get(description)
				So(des, ShouldBeIn, []string{"Other description", "Description traduite"})
				switch des {
				case "Description traduite":
					So(gbq[0].Count, ShouldEqual, 2)
					So(gbq[1].Count, ShouldEqual, 1)
					des1 := gbq[1].Values.Get(description)
					So(des1, ShouldEqual, "Other description")
				case "Other description":
					So(gbq[0].Count, ShouldEqual, 1)
					So(gbq[1].Count, ShouldEqual, 2)
					des1 := gbq[1].Values.Get(description)
					So(des1, ShouldEqual, "Description traduite")
				default:
					t.FailNow()
				}
				gbq = mTags.SearchAll().GroupBy(FieldName(description)).Aggregates(FieldName(description))
				So(gbq[0].Values.Has(description), ShouldBeTrue)
				des = gbq[0].Values.Get(description)
				So(des, ShouldBeIn, []string{"Other description", "Translated description"})
				switch des {
				case "Translated description":
					So(gbq[0].Count, ShouldEqual, 2)
					So(gbq[1].Count, ShouldEqual, 1)
					des1 := gbq[1].Values.Get(description)
					So(des1, ShouldEqual, "Other description")
				case "Other description":
					So(gbq[0].Count, ShouldEqual, 1)
					So(gbq[1].Count, ShouldEqual, 2)
					des1 := gbq[1].Values.Get(description)
					So(des1, ShouldEqual, "Translated description")
				default:
					t.FailNow()
				}
			})
		}), ShouldBeNil)
	})
}

func TestRecursionProtection(t *testing.T) {
	Convey("Testing protection against recursion", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Endless recursive method calls should panic", func() {
				So(func() { env.Pool("User").Call("EndlessRecursion") }, ShouldPanic)
			})
			Convey("Loop calls should not trigger recursion protection", func() {
				So(func() {
					for i := 0; i < int(maxRecursionDepth)+10; i++ {
						env.Pool("Profile").Call("SayHello")
					}
				}, ShouldNotPanic)
			})
			Convey("Recursion should be triggered exactly at the max recursion depth", func() {
				So(func() { env.Pool("User").Call("RecursiveMethod", int(maxRecursionDepth)/2-1, "Hi!") }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("RecursiveMethod", int(maxRecursionDepth)/2, "Hi!") }, ShouldPanic)
			})
		}), ShouldBeNil)
	})
}

func TestTypeConversionInMethodCall(t *testing.T) {
	Convey("Testing type conversion in method call", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
			RegisterRecordSetWrapper("Profile", TestProfileSet{})
			Convey("convertFunctionArg", func() {
				So(convertFunctionArg(reflect.TypeOf(*new(int64)), 126).Interface(), ShouldEqual, 126)
				prof := convertFunctionArg(reflect.TypeOf(TestProfileSet{}), userJane.Get(profile))
				So(prof.Type(), ShouldEqual, reflect.TypeOf(TestProfileSet{}))
				So(prof.Interface().(TestProfileSet).Collection().Equals(userJane.Get(profile).(RecordSet).Collection()), ShouldBeTrue)
				prof = convertFunctionArg(reflect.TypeOf(new(RecordCollection)), userJane.Get(profile))
				So(prof.Type(), ShouldEqual, reflect.TypeOf(new(RecordCollection)))
				So(prof.Interface().(*RecordCollection).Equals(userJane.Get(profile).(RecordSet).Collection()), ShouldBeTrue)
				vals := convertFunctionArg(reflect.TypeOf(new(ModelData)), NewModelData(users.model, FieldMap{"name": "Mike"}))
				So(vals.Type(), ShouldEqual, reflect.TypeOf(new(ModelData)))
				So(vals.Interface().(*ModelData).FieldMap, ShouldHaveLength, 1)
				So(vals.Interface().(*ModelData).FieldMap, ShouldContainKey, "name")
				So(vals.Interface().(*ModelData).FieldMap["name"], ShouldEqual, "Mike")
				vals = convertFunctionArg(reflect.TypeOf(new(TestUserData)), NewModelData(users.model, FieldMap{"IsStaff": true}))
				So(vals.Type(), ShouldEqual, reflect.TypeOf(new(ModelData)))
				So(vals.Interface().(*ModelData).FieldMap, ShouldHaveLength, 1)
				So(vals.Interface().(*ModelData).FieldMap, ShouldContainKey, "is_staff")
				So(vals.Interface().(*ModelData).FieldMap["is_staff"], ShouldEqual, true)
				cond := users.Model().Field(Name).Equals("Jane Smith")
				c := convertFunctionArg(reflect.TypeOf(TestUserCondition{}), cond)
				So(c.Type(), ShouldEqual, reflect.TypeOf(TestUserCondition{}))
				So(c.Interface().(TestUserCondition).Underlying().String(), ShouldEqual, cond.String())
			})
		}), ShouldBeNil)
	})
}

func TestInvalidRecordSets(t *testing.T) {
	Convey("Testing Invalid Recordsets", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			rc := InvalidRecordCollection("User")
			Convey("Getting a field on an invalid RecordSet should return empty value", func() {
				So(rc.Get(Name), ShouldEqual, "")
			})
			Convey("Getting a relation field on an invalid RecordSet should return invalid recordset", func() {
				rel, ok := rc.Get(profile).(*RecordCollection)
				So(ok, ShouldBeTrue)
				So(rel.IsValid(), ShouldBeFalse)
				So(rel.model.name, ShouldEqual, "Profile")
			})
			Convey("Calling a method on an invalid RecordSet should panic", func() {
				So(func() { rc.Call("PrefixedUser", ">>") }, ShouldPanic)
			})
		}), ShouldBeNil)
	})
}
