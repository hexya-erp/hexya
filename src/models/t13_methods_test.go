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
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				res := users.Call("PrefixedUser", "Prefix")
				So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
			})
			Convey("Calling `PrefixedUser` with context", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
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
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				res := users.Model().Methods().MustGet("PrefixedUser").Call(users, "Prefix")
				So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
				resMulti := users.Model().Methods().MustGet("OnChangeName").CallMulti(users)
				res1 := resMulti[0].(FieldMap)
				So(res1, ShouldContainKey, "DecoratedName")
				So(res1["DecoratedName"], ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			})
		}), ShouldBeNil)
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	Convey("Testing non stored computed fields", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Getting one user (Jane) and checking DisplayName", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(users.Get("DecoratedName"), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			})
			Convey("Getting all users (Jane & Will) and checking DisplayName", func() {
				users := env.Pool("User").OrderBy("Name").Call("Fetch").(RecordSet).Collection()
				So(users.Len(), ShouldEqual, 3)
				userRecs := users.Records()
				So(userRecs[0].Get("DecoratedName"), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
				So(userRecs[1].Get("DecoratedName"), ShouldEqual, "User: John Smith [<jsmith2@example.com>]")
				So(userRecs[2].Get("DecoratedName"), ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
			})
			Convey("Testing built-in DisplayName", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(users.Get("DisplayName").(string), ShouldEqual, "Jane A. Smith")
			})
			Convey("Testing computed field through a related field", func() {
				users := env.Pool("User")
				jane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(jane.Get("Other"), ShouldEqual, "Other information")
				So(jane.Get("Resume").(RecordSet).Collection().Get("Other"), ShouldEqual, "Other information")
			})
		}), ShouldBeNil)
	})
}

func TestComputedStoredFields(t *testing.T) {
	Convey("Testing stored computed fields", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			Convey("Checking that user Jane is 23", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Get("Age"), ShouldEqual, 23)
			})
			Convey("Checking that user Will has no age since no profile", func() {
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				So(userWill.Get("Age"), ShouldEqual, 0)
			})
			Convey("It's Jane's birthday, change her age, commit and check", func() {
				jane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(jane.Get("Name"), ShouldEqual, "Jane A. Smith")
				So(jane.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 12345)
				jane.Get("Profile").(RecordSet).Collection().Set("Age", 24)

				jane.Load()
				jane.Get("Profile").(RecordSet).Collection().Load()
				So(jane.Get("Age"), ShouldEqual, 24)
			})
			Convey("Adding a Profile to Will, writing to DB and checking Will's age", func() {
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				userWill.Load()
				So(userWill.Get("Name"), ShouldEqual, "Will Smith")
				willProfileData := FieldMap{
					"Age":   36,
					"Money": 5100,
				}
				willProfile := env.Pool("Profile").Call("Create", willProfileData)
				userWill.Set("Profile", willProfile)

				userWill.Load()
				So(userWill.Get("Age"), ShouldEqual, 36)
			})
			Convey("Checking inverse method by changing will's age", func() {
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				userWill.Load()
				So(userWill.Get("Age"), ShouldEqual, 36)
				userWill.Set("Age", int16(34))
				So(userWill.Get("Age"), ShouldEqual, 34)
				userWill.Load()
				So(userWill.Get("Age"), ShouldEqual, 34)
			})
			Convey("Checking that setting a computed field with no inverse panics", func() {
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				So(func() { userWill.Set("DecoratedName", "FooBar") }, ShouldPanic)
			})
		}), ShouldBeNil)
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	Convey("Testing non stored related fields", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			Convey("Checking that users PMoney is correct", func() {
				userJohn := users.Search(users.Model().Field("Name").Equals("John Smith"))
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn.Get("PMoney"), ShouldEqual, 0)
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Get("PMoney"), ShouldEqual, 12345)
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				So(userWill.Get("PMoney"), ShouldEqual, 5100)
			})
			Convey("Checking that PMoney is correct after update of Profile", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Get("PMoney"), ShouldEqual, 12345)
				userJane.Get("Profile").(RecordSet).Collection().Set("Money", 54321)
				So(userJane.Get("PMoney"), ShouldEqual, 54321)
			})
			Convey("Checking that we can update PMoney directly", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Get("PMoney"), ShouldEqual, 12345)
				userJane.Set("PMoney", 67890)
				So(userJane.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 67890)
				So(userJane.Get("PMoney"), ShouldEqual, 67890)
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				So(userWill.Get("PMoney"), ShouldEqual, 5100)

				userJane.Union(userWill).Set("PMoney", 100)
				So(userJane.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 100)
				So(userJane.Get("PMoney"), ShouldEqual, 100)
				So(userWill.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 100)
				So(userWill.Get("PMoney"), ShouldEqual, 100)
			})
			Convey("Checking that we can search PMoney directly", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				pmoneyUser := users.Search(users.Model().Field("PMoney").Equals(12345))
				So(pmoneyUser.Len(), ShouldEqual, 1)
				So(pmoneyUser.Ids()[0], ShouldEqual, userJane.Ids()[0])
				pUsers := users.Search(users.Model().Field("PMoney").Equals(12345).Or().Field("PMoney").Equals(5100))
				So(pUsers.Len(), ShouldEqual, 2)
				So(pUsers.Ids(), ShouldContain, userJane.Ids()[0])
				So(pUsers.Ids(), ShouldContain, userWill.Ids()[0])
			})
			Convey("Checking that we can order by PMoney", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				userJane.Set("PMoney", 64)
				pUsers := users.SearchAll().OrderBy("PMoney DESC")
				So(pUsers.Len(), ShouldEqual, 3)
				pUsersRecs := pUsers.Records()
				// pUsersRecs[0] is userJohn because its pMoney is Null.
				So(pUsersRecs[1].Equals(userWill), ShouldBeTrue)
				So(pUsersRecs[2].Equals(userJane), ShouldBeTrue)
			})
		}), ShouldBeNil)
	})
}

func TestEmbeddedModels(t *testing.T) {
	Convey("Testing embedded models", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
			Convey("Checking that Jane's resume exists", func() {
				So(userJane.Get("Resume").(RecordSet).IsEmpty(), ShouldBeFalse)
			})
			Convey("Adding a proper resume to Jane", func() {
				userJane.Get("Resume").(RecordSet).Collection().Set("Experience", "Hexya developer for 10 years")
				userJane.Get("Resume").(RecordSet).Collection().Set("Leisure", "Music, Sports")
				userJane.Get("Resume").(RecordSet).Collection().Set("Education", "MIT")
				userJane.Set("Education", "Berkeley")
			})
			Convey("Checking that we can access jane's resume directly", func() {
				So(userJane.Get("Experience"), ShouldEqual, "Hexya developer for 10 years")
				So(userJane.Get("Leisure"), ShouldEqual, "Music, Sports")
				So(userJane.Get("Education"), ShouldEqual, "Berkeley")
				So(userJane.Get("Resume").(RecordSet).Collection().Get("Experience"), ShouldEqual, "Hexya developer for 10 years")
				So(userJane.Get("Resume").(RecordSet).Collection().Get("Leisure"), ShouldEqual, "Music, Sports")
				So(userJane.Get("Resume").(RecordSet).Collection().Get("Education"), ShouldEqual, "MIT")
			})
		}), ShouldBeNil)
	})
}

func TestMixedInModels(t *testing.T) {
	Convey("Testing mixed in models", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			Convey("Checking that mixed in functions are correctly inherited", func() {
				janeProfile := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com")).Get("Profile").(RecordSet).Collection()
				So(janeProfile.Call("PrintAddress"), ShouldEqual, "[<165 5th Avenue, 0305 New York>, USA]")
				So(janeProfile.Call("SayHello"), ShouldEqual, "Hello !")
			})
			Convey("Checking mixing in all models", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				userJane.Set("Active", true)
				So(userJane.Get("Active").(bool), ShouldEqual, true)
				So(userJane.Call("IsActivated").(bool), ShouldEqual, true)
				janeProfile := userJane.Get("Profile").(RecordSet).Collection()
				janeProfile.Set("Active", true)
				So(janeProfile.Get("Active").(bool), ShouldEqual, true)
				So(janeProfile.Call("IsActivated").(bool), ShouldEqual, true)
			})
		}), ShouldBeNil)
	})
}

func TestContextedFields(t *testing.T) {
	Convey("Testing contexted fields", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			tags := env.Pool("Tag")
			decs := env.Pool("TagHexyaDescription")
			var tagc *RecordCollection
			Convey("Creating record with a single contexted field", func() {
				tagc = tags.Call("Create", FieldMap{
					"Name":        "Contexted tag",
					"Description": "Translated description",
				}).(RecordSet).Collection()
				So(tagc.Get("DescriptionHexyaContexts").(RecordSet).Len(), ShouldEqual, 1)
				So(tagc.Get("Description"), ShouldEqual, "Translated description")

				tagc.WithContext("lang", "fr_FR").Set("Description", "Description traduite")
				So(tagc.Get("DescriptionHexyaContexts").(RecordSet).Len(), ShouldEqual, 2)
				So(tagc.Get("Description"), ShouldEqual, "Translated description")

				newTag := tags.WithContext("lang", "fr_FR").Search(tags.Model().Field("name").Equals("Contexted tag"))
				newTag.Load("description")
				So(newTag.Get("Description"), ShouldEqual, "Description traduite")

				So(tagc.Get("Description"), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "fr_FR").Get("Description"), ShouldEqual, "Description traduite")
				So(tagc.Get("Description"), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "de_DE").Get("Description"), ShouldEqual, "Translated description")

				tagc.WithContext("lang", "fr_FR").Set("Description", "Nouvelle traduction")
				So(tagc.Get("Description"), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "fr_FR").Get("Description"), ShouldEqual, "Nouvelle traduction")
				So(tagc.WithContext("lang", "de_DE").Get("Description"), ShouldEqual, "Translated description")

				tagc.WithContext("lang", "de_DE").Set("Description", "übersetzte Beschreibung")
				So(tagc.Get("Description"), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "fr_FR").Get("Description"), ShouldEqual, "Nouvelle traduction")
				So(tagc.WithContext("lang", "de_DE").Get("Description"), ShouldEqual, "übersetzte Beschreibung")

				tagc.WithContext("lang", "es_ES").Set("Description", "descripción traducida")
				So(tagc.Get("Description"), ShouldEqual, "Translated description")
				So(tagc.WithContext("lang", "fr_FR").Get("Description"), ShouldEqual, "Nouvelle traduction")
				So(tagc.WithContext("lang", "de_DE").Get("Description"), ShouldEqual, "übersetzte Beschreibung")
				So(tagc.WithContext("lang", "es_ES").Get("Description"), ShouldEqual, "descripción traducida")
				So(tagc.WithContext("lang", "it_IT").Get("Description"), ShouldEqual, "Translated description")
			})
			Convey("Creating a record with a contexted field should also create for default context", func() {
				tags.WithContext("lang", "fr_FR").Call("Create", FieldMap{
					"Name":        "Contexted tag 2",
					"Description": "Description en français",
				}).(RecordSet).Collection()
				tag := tags.Search(tags.Model().Field("name").Equals("Contexted tag 2"))
				So(tag.Get("Description"), ShouldEqual, "Description en français")
				So(tag.WithContext("lang", "en_US").Get("Description"), ShouldEqual, "Description en français")
				So(tag.WithContext("lang", "fr_FR").Get("Description"), ShouldEqual, "Description en français")
				tag.WithContext("lang", "en_US").Set("Description", "Description in English")
				So(tag.WithContext("lang", "en_US").Get("Description"), ShouldEqual, "Description in English")
				So(tag.WithContext("lang", "fr_FR").Get("Description"), ShouldEqual, "Description en français")
				So(tag.WithContext("lang", "de_DE").Get("Description"), ShouldEqual, "Description en français")
				So(tag.Get("Description"), ShouldEqual, "Description en français")
				thc := decs.Search(decs.Model().Field("Record").Equals(tag.Ids()[0]).And().Field("lang").Equals(""))
				So(thc.Len(), ShouldEqual, 1)
			})
			Convey("Updating in another transaction should not recreate a default value", func() {
				tag := tags.WithContext("lang", "fr_FR").Search(tags.Model().Field("name").Equals("Contexted tag 2"))
				So(tag.Get("Description"), ShouldEqual, "Description en français")
				thc := decs.Search(decs.Model().Field("Record").Equals(tag.Ids()[0]).And().Field("lang").Equals(""))
				So(thc.Len(), ShouldEqual, 1)

				tag.Set("Description", "Nouvelle description en français")
				thc = decs.Search(decs.Model().Field("Record").Equals(tag.Ids()[0]).And().Field("lang").Equals(""))
				So(thc.Len(), ShouldEqual, 1)
				So(tag.Get("Description"), ShouldEqual, "Nouvelle description en français")
			})
			Convey("Changing language should recreate a default value (new transaction)", func() {
				tag := tags.WithContext("lang", "es_ES").Search(tags.Model().Field("name").Equals("Contexted tag 2"))
				So(tag.Get("Description"), ShouldEqual, "Description en français")
				thc := decs.Search(decs.Model().Field("Record").Equals(tag.Ids()[0]).And().Field("lang").Equals(""))
				So(thc.Len(), ShouldEqual, 1)

				tag.Set("Description", "descripción traducida")
				thc = decs.Search(decs.Model().Field("Record").Equals(tag.Ids()[0]).And().Field("lang").Equals(""))
				So(thc.Len(), ShouldEqual, 1)
				So(tag.Get("Description"), ShouldEqual, "descripción traducida")
			})
			Convey("Deleting a record with a contexted field should delete all contexts", func() {
				newTag := tags.Call("Create", FieldMap{
					"Name":        "Contexted tag 3",
					"Description": "Description to translate",
				}).(RecordSet).Collection()
				So(newTag.Get("Description"), ShouldEqual, "Description to translate")
				newTag.WithContext("lang", "fr_FR").Set("Description", "Description en français")
				So(newTag.WithContext("lang", "fr_FR").Get("Description"), ShouldEqual, "Description en français")
				newTag.WithContext("lang", "de_DE").Set("Description", "übersetzte Beschreibung")
				So(newTag.WithContext("lang", "de_DE").Get("Description"), ShouldEqual, "übersetzte Beschreibung")
				nID := newTag.Ids()[0]
				newTag.Call("Unlink")
				dec := decs.Search(decs.Model().Field("Record").Equals(nID).Or().Field("Record").IsNull())
				So(dec.IsEmpty(), ShouldBeTrue)
			})
		}), ShouldBeNil)
	})
	Convey("Testing contexted group by queries", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			tags := env.Pool("Tag")
			tags.SearchAll().Call("Unlink")
			Convey("Simple group by query", func() {
				tag1 := tags.Call("Create", FieldMap{
					"Name":        "Contexted tag",
					"Description": "Translated description",
				}).(RecordSet).Collection()
				tag1.WithContext("lang", "fr_FR").Set("Description", "Description traduite")
				tag2 := tags.Call("Create", FieldMap{
					"Name":        "Contexted tag",
					"Description": "Translated description",
				}).(RecordSet).Collection()
				tag2.WithContext("lang", "fr_FR").Set("Description", "Description traduite")
				tags.Call("Create", FieldMap{
					"Name":        "Contexted tag",
					"Description": "Other description",
				}).(RecordSet).Collection()
				gbq := tags.WithContext("lang", "fr_FR").SearchAll().GroupBy(FieldName("Description")).Aggregates(FieldName("Description"))
				So(gbq, ShouldHaveLength, 2)
				So(gbq[0].Values, ShouldContainKey, "Description")
				So(gbq[0].Values["Description"], ShouldBeIn, []string{"Other description", "Description traduite"})
				switch gbq[0].Values["Description"] {
				case "Description traduite":
					So(gbq[0].Count, ShouldEqual, 2)
					So(gbq[1].Count, ShouldEqual, 1)
					So(gbq[1].Values["Description"], ShouldEqual, "Other description")
				case "Other description":
					So(gbq[0].Count, ShouldEqual, 1)
					So(gbq[1].Count, ShouldEqual, 2)
					So(gbq[1].Values["Description"], ShouldEqual, "Description traduite")
				default:
					t.FailNow()
				}
				gbq = tags.SearchAll().GroupBy(FieldName("Description")).Aggregates(FieldName("Description"))
				So(gbq[0].Values, ShouldContainKey, "Description")
				So(gbq[0].Values["Description"], ShouldBeIn, []string{"Other description", "Translated description"})
				switch gbq[0].Values["Description"] {
				case "Translated description":
					So(gbq[0].Count, ShouldEqual, 2)
					So(gbq[1].Count, ShouldEqual, 1)
					So(gbq[1].Values["Description"], ShouldEqual, "Other description")
				case "Other description":
					So(gbq[0].Count, ShouldEqual, 1)
					So(gbq[1].Count, ShouldEqual, 2)
					So(gbq[1].Values["Description"], ShouldEqual, "Translated description")
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
			userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
			Convey("convertFunctionArg", func() {
				So(convertFunctionArg(userJane, reflect.TypeOf(*new(int64)), 126).Interface(), ShouldEqual, 126)
				prof := convertFunctionArg(userJane, reflect.TypeOf(TestProfileSet{}), userJane.Get("Profile"))
				So(prof.Type(), ShouldEqual, reflect.TypeOf(TestProfileSet{}))
				So(prof.Interface().(TestProfileSet).Collection().Equals(userJane.Get("Profile").(RecordSet).Collection()), ShouldBeTrue)
				prof = convertFunctionArg(userJane, reflect.TypeOf(new(RecordCollection)), userJane.Get("Profile"))
				So(prof.Type(), ShouldEqual, reflect.TypeOf(new(RecordCollection)))
				So(prof.Interface().(*RecordCollection).Equals(userJane.Get("Profile").(RecordSet).Collection()), ShouldBeTrue)
				vals := convertFunctionArg(userJane, reflect.TypeOf(FieldMap{}), FieldMap{"key": "value"})
				So(vals.Type(), ShouldEqual, reflect.TypeOf(FieldMap{}))
				So(vals.Interface(), ShouldHaveLength, 1)
				So(vals.Interface(), ShouldContainKey, "key")
				So(vals.Interface().(FieldMap)["key"], ShouldEqual, "value")
				vals = convertFunctionArg(userJane, reflect.TypeOf(new(ModelData)), FieldMap{"IsStaff": true})
				So(vals.Type(), ShouldEqual, reflect.TypeOf(new(ModelData)))
				So(vals.Interface().(*ModelData).FieldMap, ShouldHaveLength, 1)
				So(vals.Interface().(*ModelData).FieldMap, ShouldContainKey, "IsStaff")
				So(vals.Interface().(*ModelData).FieldMap["IsStaff"], ShouldEqual, true)
				cond := users.Model().Field("Name").Equals("Jane Smith")
				c := convertFunctionArg(userJane, reflect.TypeOf(TestUserCondition{}), cond)
				So(c.Type(), ShouldEqual, reflect.TypeOf(TestUserCondition{}))
				So(c.Interface().(TestUserCondition).Underlying().String(), ShouldEqual, cond.String())
			})
		}), ShouldBeNil)
	})
}
