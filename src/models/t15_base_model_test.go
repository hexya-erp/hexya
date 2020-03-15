// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBaseModelMethods(t *testing.T) {
	Convey("Testing base model methods", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			profileModel := Registry.MustGet("Profile")
			tagModel := Registry.MustGet("Tag")
			postModel := Registry.MustGet("Post")
			commentModel := Registry.MustGet("Comment")
			userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
			Convey("LastUpdate", func() {
				So(userJane.Get(lastupdate).(dates.DateTime).Sub(userJane.Get(writeDate).(dates.DateTime)), ShouldBeLessThanOrEqualTo, 1*time.Second)
				newComment := commentModel.Create(env, NewModelData(commentModel).
					Set(text, "MyComment"))
				time.Sleep(1*time.Second + 100*time.Millisecond)
				So(newComment.Get(lastupdate).(dates.DateTime).Sub(newComment.Get(createDate).(dates.DateTime)), ShouldBeLessThanOrEqualTo, 1*time.Second)
			})
			Convey("Load and Read", func() {
				userJane = userJane.Call("Load", []FieldName{ID, Name, age, posts, profile}).(RecordSet).Collection()
				res := userJane.Call("Read", []FieldName{Name, age, posts, profile})
				So(res, ShouldHaveLength, 1)
				fMap := res.([]RecordData)[0].Underlying().FieldMap
				So(fMap, ShouldHaveLength, 5)
				So(fMap, ShouldContainKey, "name")
				So(fMap["name"], ShouldEqual, "Jane A. Smith")
				So(fMap, ShouldContainKey, "age")
				So(fMap["age"], ShouldEqual, 24)
				So(fMap, ShouldContainKey, "posts_ids")
				So(fMap["posts_ids"].(RecordSet).Collection().Ids(), ShouldHaveLength, 2)
				So(fMap, ShouldContainKey, "profile_id")
				So(fMap["profile_id"].(RecordSet).Collection().Get(ID), ShouldEqual, userJane.Get(profile).(RecordSet).Collection().Get(ID))
				So(fMap, ShouldContainKey, "id")
				So(fMap["id"], ShouldEqual, userJane.Ids()[0])
			})
			Convey("Browse and BrowseOne", func() {
				jid := userJane.Ids()[0]
				j2 := userModel.Browse(env, []int64{jid})
				So(j2.Equals(userJane), ShouldBeTrue)
				j21 := userModel.BrowseOne(env, jid)
				So(j21.Equals(userJane), ShouldBeTrue)
				j22 := env.Pool("User").Call("Browse", []int64{jid}).(RecordSet).Collection()
				So(j22.Equals(userJane), ShouldBeTrue)
				j23 := env.Pool("User").Call("BrowseOne", jid).(RecordSet).Collection()
				So(j23.Equals(userJane), ShouldBeTrue)
			})
			Convey("SearchCount", func() {
				countSingle := userJane.Call("SearchCount").(int)
				So(countSingle, ShouldEqual, 1)
				allCount := env.Pool(userModel.name).Call("SearchCount").(int)
				So(allCount, ShouldEqual, 3)
				countTags := env.Pool("Tag").WithContext("lang", "fr_FR").
					Search(Registry.MustGet("Tag").Field(description).Contains("Nouvelle ")).Call("SearchCount")
				So(countTags, ShouldEqual, 2)
			})
			Convey("Copy", func() {
				newProfile := userJane.Get(profile).(RecordSet).Collection().Call("Copy", NewModelData(profileModel)).(RecordSet).Collection()
				So(newProfile.Equals(userJane.Get(profile).(RecordSet).Collection()), ShouldBeFalse)
				userJane.Call("Write", NewModelData(userModel).Set(password, "Jane's Password"))
				userJaneCopy := userJane.Call("Copy", NewModelData(userModel).
					Set(Name, "Jane's Copy").
					Set(email2, "js@example.com")).(RecordSet).Collection()
				So(userJaneCopy.IsEmpty(), ShouldBeFalse)
				So(userJaneCopy.Equals(userJane), ShouldBeFalse)
				So(userJaneCopy.Get(Name), ShouldEqual, "Jane A. Smith (copy)")
				So(userJaneCopy.Get(email), ShouldEqual, "jane.smith@example.com")
				So(userJaneCopy.Get(email2), ShouldEqual, "js@example.com")
				So(userJaneCopy.Get(password), ShouldBeBlank)
				So(userJaneCopy.Get(profile).(RecordSet).Collection().Equals(userJane.Get(profile).(RecordSet)), ShouldBeFalse)
				So(userJaneCopy.Get(profileAge), ShouldEqual, 24)
				So(userJaneCopy.Get(age), ShouldEqual, 24)
				So(userJaneCopy.Get(nums), ShouldEqual, 2)
				So(userJaneCopy.Get(posts).(RecordSet).Collection().Len(), ShouldEqual, 2)

				So(func() { userJane.Get(profile).(RecordSet).Collection().Call("Copy", nil) }, ShouldNotPanic)
			})
			Convey("FieldGet and FieldsGet", func() {
				fInfo := userJane.Call("FieldGet", FieldName(Name)).(*FieldInfo)
				So(fInfo.String, ShouldEqual, "Name")
				So(fInfo.Help, ShouldEqual, "The user's username")
				So(fInfo.Type, ShouldEqual, fieldtype.Char)
				fInfos := userJane.Call("FieldsGet", FieldsGetArgs{}).(map[string]*FieldInfo)
				So(fInfos, ShouldHaveLength, 35)
			})
			Convey("NameGet", func() {
				So(userJane.Get(displayName), ShouldEqual, "Jane A. Smith")
				janeProfile := userJane.Get(profile).(RecordSet).Collection()
				So(janeProfile.Get(displayName), ShouldEqual, fmt.Sprintf("Profile(%d)", janeProfile.Get(ID)))
			})
			Convey("DefaultGet", func() {
				defaults := userJane.Call("DefaultGet").(*ModelData)
				So(defaults.FieldMap, ShouldHaveLength, 14)
				So(defaults.FieldMap, ShouldContainKey, "status_json")
				So(defaults.FieldMap["status_json"], ShouldEqual, 12)
				So(defaults.FieldMap, ShouldContainKey, "hexya_external_id")
				So(defaults.FieldMap, ShouldContainKey, "is_active")
				So(defaults.FieldMap["is_active"], ShouldEqual, false)
				So(defaults.FieldMap, ShouldContainKey, "active")
				So(defaults.FieldMap["active"], ShouldEqual, true)
				So(defaults.FieldMap, ShouldContainKey, "is_premium")
				So(defaults.FieldMap["is_premium"], ShouldEqual, false)
				So(defaults.FieldMap, ShouldContainKey, "is_staff")
				So(defaults.FieldMap["is_staff"], ShouldEqual, false)
			})
			Convey("New", func() {
				dummyUser := env.Pool("User").Call("New", NewModelData(userModel).
					Set(Name, "DummyUser").
					Set(email, "du@example.com")).(RecordSet).Collection()
				So(dummyUser.Get(Name), ShouldEqual, "DummyUser")
				So(dummyUser.Get(email), ShouldEqual, "du@example.com")
				So(dummyUser.Get(email2), ShouldBeEmpty)
				So(dummyUser.Ids()[0], ShouldBeLessThan, 0)
				So(func() { dummyUser.ForceLoad() }, ShouldPanic)
				So(func() { dummyUser.Set(email2, "du2@example.com") }, ShouldNotPanic)
				So(dummyUser.Get(email2), ShouldEqual, "du2@example.com")
				So(dummyUser.Get(decoratedName), ShouldEqual, "User: DummyUser [<du@example.com>]")
				So(func() { dummyUser.unlink() }, ShouldNotPanic)
			})
			Convey("Onchange", func() {
				Convey("Testing with existing RecordSet", func() {
					res := userJane.Call("Onchange", OnchangeParams{
						Fields:   []FieldName{Name, coolType, age},
						Onchange: map[string]string{"Name": "1", "CoolType": "1", "Age": "1"},
						Values:   NewModelData(userModel, FieldMap{"Name": "William", "CoolType": "cool", "IsCool": false, "DecoratedName": false, "Profile": false, "Age": int16(24)}),
					}).(OnchangeResult)
					fMap := res.Value.Underlying().FieldMap
					So(fMap, ShouldHaveLength, 3)
					So(fMap, ShouldContainKey, "decorated_name")
					So(fMap["decorated_name"], ShouldEqual, "User: William [<jane.smith@example.com>]")
					So(fMap, ShouldContainKey, "is_cool")
					So(fMap["is_cool"], ShouldEqual, true)
					So(fMap, ShouldContainKey, "age")
					So(fMap["age"], ShouldEqual, int16(0))
					So(res.Warning, ShouldBeEmpty)
					So(res.Filters, ShouldHaveLength, 1)
					var fKey FieldName
					var fValue Conditioner
					for k, v := range res.Filters {
						fKey = k
						fValue = v
					}
					So(fKey.Name(), ShouldEqual, lastPost.Name())
					So(fKey.JSON(), ShouldEqual, lastPost.JSON())
					So(fValue.Underlying().String(), ShouldEqual, `AND Street = addr
`)
				})
				Convey("Testing onchange warnings", func() {
					res := userJane.Call("Onchange", OnchangeParams{
						Fields:   []FieldName{Name, coolType, age},
						Onchange: map[string]string{"Name": "1", "CoolType": "1", "age": "1"},
						Values:   NewModelData(userModel, FieldMap{"Name": "Warning User", "CoolType": "cool", "IsCool": false, "DecoratedName": false, "Profile": false, "age": int16(24)}),
					}).(OnchangeResult)
					So(res.Warning, ShouldEqual, "We have a warning here")
				})
				Convey("Testing with new RecordSet", func() {
					res := env.Pool("User").Call("Onchange", OnchangeParams{
						Fields:   []FieldName{Name, email, coolType, nums, nums},
						Onchange: map[string]string{"Name": "1", "CoolType": "1", "Nums": "1"},
						Values:   NewModelData(userModel, FieldMap{"Name": "", "Email": "", "CoolType": "cool", "IsCool": false, "DecoratedName": false}),
					}).(OnchangeResult)
					fMap := res.Value.Underlying().FieldMap
					So(fMap, ShouldHaveLength, 2)
					So(fMap, ShouldContainKey, "decorated_name")
					So(fMap["decorated_name"], ShouldEqual, "User:  [<>]")
					So(fMap, ShouldContainKey, "is_cool")
					So(fMap["is_cool"], ShouldEqual, true)
				})
				Convey("Testing with new RecordSet and related field", func() {
					post := env.Pool("Post").SearchAll().Limit(1)
					res := env.Pool("User").Call("Onchange", OnchangeParams{
						Fields:   []FieldName{Name, email, coolType, nums, nums, mana},
						Onchange: map[string]string{"Name": "1", "CoolType": "1", "Nums": "1", "Mana": "1"},
						Values: NewModelData(userModel, FieldMap{"Name": "", "Email": "", "CoolType": "cool",
							"IsCool": false, "DecoratedName": false, "Mana": float32(12.3), "BestProfilePost": false}),
					}).(OnchangeResult)
					fMap := res.Value.Underlying().FieldMap
					So(fMap, ShouldHaveLength, 3)
					So(fMap, ShouldContainKey, "decorated_name")
					So(fMap["decorated_name"], ShouldEqual, "User:  [<>]")
					So(fMap, ShouldContainKey, "is_cool")
					So(fMap["is_cool"], ShouldEqual, true)
					So(fMap, ShouldContainKey, "best_profile_post_id")
					So(fMap["best_profile_post_id"].(RecordSet).Collection().Equals(post), ShouldBeTrue)
				})
			})
			Convey("CheckRecursion", func() {
				So(userJane.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag1 := env.Pool("Tag").Call("Create", NewModelData(tagModel).
					Set(Name, "Tag1")).(RecordSet).Collection()
				So(tag1.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag2 := env.Pool("Tag").Call("Create", NewModelData(tagModel).
					Set(Name, "Tag2").
					Set(parent, tag1)).(RecordSet).Collection()
				So(tag2.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag3 := env.Pool("Tag").Call("Create", NewModelData(tagModel).
					Set(Name, "Tag1").
					Set(parent, tag2)).(RecordSet).Collection()
				So(tag3.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag1.Set(parent, tag3)
				So(tag1.Call("CheckRecursion").(bool), ShouldBeFalse)
				So(tag2.Call("CheckRecursion").(bool), ShouldBeFalse)
				So(tag3.Call("CheckRecursion").(bool), ShouldBeFalse)
				tagNeg := env.Pool("Tag").Call("New", NewModelData(tagModel).
					Set(Name, "Tag1").
					Set(parent, tag2)).(RecordSet).Collection()
				So(tagNeg.Call("CheckRecursion").(bool), ShouldBeTrue)
			})
			Convey("Browse", func() {
				browsedUser := env.Pool("User").Call("Browse", []int64{userJane.Ids()[0]}).(RecordSet).Collection()
				So(browsedUser.Ids(), ShouldHaveLength, 1)
				So(browsedUser.Ids(), ShouldContain, userJane.Ids()[0])
			})
			Convey("Equals", func() {
				browsedUser := env.Pool("User").Call("Browse", []int64{userJane.Ids()[0]}).(RecordSet).Collection()
				So(browsedUser.Call("Equals", userJane), ShouldBeTrue)
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("John Smith")).(RecordSet).Collection()
				So(userJohn.Call("Equals", userJane), ShouldBeFalse)
				johnAndJane := userJohn.Union(userJane)
				usersJ := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Like("J% Smith")).(RecordSet).Collection()
				So(usersJ.Records(), ShouldHaveLength, 2)
				So(usersJ.Equals(johnAndJane), ShouldBeTrue)

				So(InvalidRecordCollection("User").Equals(usersJ), ShouldBeFalse)
				So(env.Pool("Profile").Equals(userJane), ShouldBeFalse)
			})
			Convey("Union", func() {
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				userWill := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("Will Smith")).(RecordSet).Collection()
				johnAndWill := userWill.Union(userJohn)
				So(johnAndJane.Len(), ShouldEqual, 2)
				So(johnAndWill.Len(), ShouldEqual, 2)
				all := johnAndJane.Union(johnAndWill)
				So(all.Len(), ShouldEqual, 3)
				So(all.Intersect(userJane).Equals(userJane), ShouldBeTrue)
				So(all.Intersect(userJohn).Equals(userJohn), ShouldBeTrue)
				So(all.Intersect(userWill).Equals(userWill), ShouldBeTrue)

				So(InvalidRecordCollection("User").Union(userJane).IsValid(), ShouldBeFalse)
				So(func() { env.Pool("Profile").Union(userJane) }, ShouldPanic)
			})
			Convey("Subtract", func() {
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				So(johnAndJane.Subtract(userJane).Equals(userJohn), ShouldBeTrue)
				So(johnAndJane.Call("Subtract", userJohn).(RecordSet).Collection().Equals(userJane), ShouldBeTrue)

				So(InvalidRecordCollection("User").Subtract(userJane).IsValid(), ShouldBeFalse)
				So(func() { env.Pool("Profile").Subtract(userJane) }, ShouldPanic)
			})
			Convey("Intersect", func() {
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				So(johnAndJane.Intersect(userJane).Equals(userJane), ShouldBeTrue)
				So(johnAndJane.Call("Intersect", userJohn).(RecordSet).Collection().Equals(userJohn), ShouldBeTrue)

				So(InvalidRecordCollection("User").Intersect(userJane).IsValid(), ShouldBeFalse)
				So(func() { env.Pool("Profile").Intersect(userJane) }, ShouldPanic)
			})
			Convey("ConvertLimitToInt", func() {
				So(ConvertLimitToInt(12), ShouldEqual, 12)
				So(ConvertLimitToInt(false), ShouldEqual, -1)
				So(ConvertLimitToInt(0), ShouldEqual, 0)
				So(ConvertLimitToInt(nil), ShouldEqual, 80)
			})
			Convey("CartesianProduct", func() {
				tagA := env.Pool("Tag").Call("Create", NewModelData(tagModel).Set(Name, "A")).(RecordSet).Collection()
				tagB := env.Pool("Tag").Call("Create", NewModelData(tagModel).Set(Name, "B")).(RecordSet).Collection()
				tagC := env.Pool("Tag").Call("Create", NewModelData(tagModel).Set(Name, "C")).(RecordSet).Collection()
				tagD := env.Pool("Tag").Call("Create", NewModelData(tagModel).Set(Name, "D")).(RecordSet).Collection()
				tagE := env.Pool("Tag").Call("Create", NewModelData(tagModel).Set(Name, "E")).(RecordSet).Collection()
				tagF := env.Pool("Tag").Call("Create", NewModelData(tagModel).Set(Name, "F")).(RecordSet).Collection()
				tagG := env.Pool("Tag").Call("Create", NewModelData(tagModel).Set(Name, "G")).(RecordSet).Collection()
				tagsAB := tagA.Union(tagB)
				tagsCD := tagC.Union(tagD)
				tagsEFG := tagE.Union(tagF).Union(tagG)

				contains := func(product []*RecordCollection, collections ...*RecordCollection) bool {
				productLoop:
					for _, p := range product {
						for _, c := range collections {
							if c.Equals(p) {
								break productLoop
							}
						}
						return false
					}
					return true
				}

				So(cartesianProductSlices(), ShouldHaveLength, 0)

				product0 := tagA.CartesianProduct()
				So(product0, ShouldHaveLength, 1)
				So(product0[0].Equals(tagA), ShouldBeTrue)

				product1 := tagsAB.CartesianProduct(tagsCD)
				So(product1, ShouldHaveLength, 4)
				So(contains(product1,
					tagA.Union(tagC),
					tagA.Union(tagD),
					tagB.Union(tagC),
					tagB.Union(tagD)), ShouldBeTrue)

				product2 := tagsAB.CartesianProduct(tagsEFG)
				So(product2, ShouldHaveLength, 6)
				So(contains(product2,
					tagA.Union(tagE),
					tagA.Union(tagF),
					tagA.Union(tagG),
					tagB.Union(tagE),
					tagB.Union(tagF),
					tagB.Union(tagG)), ShouldBeTrue)

				product3 := tagsAB.CartesianProduct(tagsCD, tagsEFG)
				So(product3, ShouldHaveLength, 12)
				So(contains(product3,
					tagA.Union(tagC).Union(tagE),
					tagA.Union(tagC).Union(tagF),
					tagA.Union(tagC).Union(tagG),
					tagA.Union(tagD).Union(tagE),
					tagA.Union(tagD).Union(tagF),
					tagA.Union(tagD).Union(tagG),
					tagB.Union(tagC).Union(tagE),
					tagB.Union(tagC).Union(tagF),
					tagB.Union(tagC).Union(tagG),
					tagB.Union(tagD).Union(tagE),
					tagB.Union(tagD).Union(tagF),
					tagB.Union(tagD).Union(tagG)), ShouldBeTrue)
			})
			Convey("Sorted", func() {
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", NewModelData(postModel).
						Set(title, fmt.Sprintf("Post no %02d", (24-i)%20)).
						Set(user, userJane))
				}
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Contains("Post no")).OrderBy("ID")
				for i, post := range rPosts.Records() {
					So(post.Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", (24-i)%20))
				}

				sortedPosts := rPosts.Call("Sorted", func(rs1 RecordSet, rs2 RecordSet) bool {
					return rs1.Collection().Get(title).(string) < rs2.Collection().Get(title).(string)
				}).(RecordSet).Collection().Records()
				So(sortedPosts, ShouldHaveLength, 20)
				for i, post := range sortedPosts {
					So(post.Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", i))
				}

				So(InvalidRecordCollection("Post").Sorted(func(rs1 RecordSet, rs2 RecordSet) bool {
					return rs1.Collection().Get(title).(string) < rs2.Collection().Get(title).(string)
				}).IsValid(), ShouldBeFalse)
			})
			Convey("SortedDefault", func() {
				Convey("With posts", func() {
					for i := 0; i < 20; i++ {
						env.Pool("Post").Call("Create", NewModelData(postModel).
							Set(title, fmt.Sprintf("Post no %02d", (24-i)%20)).
							Set(user, userJane))
					}
					rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Contains("Post no")).OrderBy("ID")
					for i, post := range rPosts.Records() {
						So(post.Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", (24-i)%20))
					}

					sortedPosts := rPosts.Call("SortedDefault").(RecordSet).Collection().Records()
					So(sortedPosts, ShouldHaveLength, 20)
					for i, post := range sortedPosts {
						So(post.Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", i))
					}
				})
				Convey("With tags", func() {
					env.Pool("Tag").SearchAll().Call("Unlink")
					for i := 0; i < 20; i++ {
						env.Pool("Tag").Call("Create", NewModelData(tagModel).
							Set(Name, fmt.Sprintf("Tag %02d", i/2)))
					}
					rTags := env.Pool("Tag").SearchAll()
					sortedTags := rTags.Call("SortedDefault").(RecordSet).Collection().Records()
					for i, tag := range sortedTags {
						So(tag.Get(Name), ShouldEqual, fmt.Sprintf("Tag %02d", 9-(i/2)))
						So(tag.Get(ID), ShouldEqual, int(sortedTags[0].ids[0])+1-i+i%2-(i+1)%2)
					}
				})
			})
			Convey("SortedByField", func() {
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", NewModelData(postModel).
						Set(title, fmt.Sprintf("Post no %02d", (24-i)%20)).
						Set(user, userJane))
				}
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Contains("Post no")).OrderBy("ID")
				for i, post := range rPosts.Records() {
					So(post.Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", (24-i)%20))
				}

				sortedPosts := rPosts.Call("SortedByField", FieldName(title), false).(RecordSet).Collection().Records()
				So(sortedPosts, ShouldHaveLength, 20)
				for i, post := range sortedPosts {
					So(post.Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", i))
				}

				revSortedPosts := rPosts.Call("SortedByField", FieldName(title), true).(RecordSet).Collection().Records()
				So(revSortedPosts, ShouldHaveLength, 20)
				for i, post := range revSortedPosts {
					So(post.Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", 19-i))
				}
			})
			Convey("Testing one2many sets keep the default order", func() {
				userJane.Get(posts).(RecordSet).Collection().Call("Unlink")
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", NewModelData(postModel).
						Set(title, fmt.Sprintf("Post no %02d", 19-i)).
						Set(user, userJane))
				}

				rPosts := userJane.Get(posts).(RecordSet).Collection()
				So(rPosts.Len(), ShouldEqual, 20)
				for i, post := range rPosts.Records() {
					So(post.Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", i))
				}
			})
			Convey("Filtered", func() {
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", NewModelData(postModel).
						Set(title, fmt.Sprintf("Post no %02d", i)).
						Set(user, userJane))
				}
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Contains("Post no"))

				evenPosts := rPosts.Call("Filtered", func(rs RecordSet) bool {
					var num int
					_, err := fmt.Sscanf(rs.Collection().Get(title).(string), "Post no %02d", &num)
					if err != nil {
						t.Error(err)
					}
					if num%2 == 0 {
						return true
					}
					return false
				}).(RecordSet).Collection().Records()
				So(evenPosts, ShouldHaveLength, 10)
				for i := 0; i < 10; i++ {
					So(evenPosts[i].Get(title), ShouldEqual, fmt.Sprintf("Post no %02d", 2*i))
				}

				So(InvalidRecordCollection("Post").Filtered(func(rs RecordSet) bool {
					return true
				}).IsValid(), ShouldBeFalse)
			})
			Convey("CheckExecutionPermissions", func() {
				res := env.Pool("User").Call("CheckExecutionPermission", Registry.MustGet("User").Methods().MustGet("Load"), []bool{true})
				So(res, ShouldBeTrue)
			})
			Convey("convertTotRecordSet", func() {
				profileID := userJane.Get(profile).(RecordSet).Collection().Ids()[0]
				res := env.Pool("User").convertToRecordSet(profileID, "Profile")
				So(res.Ids()[0], ShouldEqual, profileID)
				res = env.Pool("User").convertToRecordSet(false, "Profile")
				So(res.IsEmpty(), ShouldBeTrue)
				res = env.Pool("User").convertToRecordSet([]interface{}{float64(profileID)}, "Profile")
				So(res.Ids()[0], ShouldEqual, profileID)
				res = env.Pool("User").convertToRecordSet(int(profileID), "Profile")
				So(res.Ids()[0], ShouldEqual, profileID)
				res = env.Pool("User").convertToRecordSet(float64(profileID), "Profile")
				So(res.Ids()[0], ShouldEqual, profileID)
				res = env.Pool("User").convertToRecordSet([]float64{float64(profileID)}, "Profile")
				So(res.Ids()[0], ShouldEqual, profileID)
				res = env.Pool("User").convertToRecordSet([]int{int(profileID)}, "Profile")
				So(res.Ids()[0], ShouldEqual, profileID)
				So(func() { env.Pool("User").convertToRecordSet("", "Profile") }, ShouldPanic)
				res = env.Pool("User").convertToRecordSet(userJane.Get(profile), "Profile")
				So(res.Ids()[0], ShouldEqual, profileID)
			})
			Convey("EnsureOne", func() {
				So(func() { userJane.EnsureOne() }, ShouldNotPanic)
				So(func() { env.Pool("User").EnsureOne() }, ShouldPanic)
				users := env.Pool("User").SearchAll()
				So(users.Len(), ShouldBeGreaterThan, 0)
				So(func() { users.EnsureOne() }, ShouldPanic)
			})
			Convey("GetRecord", func() {
				So(env.Pool("User").Call("GetRecord", userJane.Get(hexyaExternalID)).(RecordSet).Collection().Equals(userJane), ShouldBeTrue)
			})
			Convey("SearchByName", func() {
				j := env.Pool("User").Call("SearchByName", "Jane A. Smith", operator.Operator(""), userModel.Field(isStaff).Equals(false), 10).(RecordSet).Collection()
				So(j.Equals(userJane), ShouldBeTrue)
			})
		}), ShouldBeNil)
	})
}

func TestPostBootSequences(t *testing.T) {
	Convey("Testing manual sequences after bootstrap", t, func() {
		testSeq := Registry.MustGetSequence("Test")
		testSeq.Drop()
		seq := CreateSequence("ManualSequence", 1, 1)
		So(seq.JSON, ShouldEqual, "manual_sequence_manseq")
		So(TestAdapter.sequences("%_manseq"), ShouldHaveLength, 1)
		So(TestAdapter.sequences("%_manseq")[0].Name, ShouldEqual, "manual_sequence_manseq")
		So(seq.NextValue(), ShouldEqual, 1)
		So(seq.NextValue(), ShouldEqual, 2)
		seq.Alter(2, 5)
		So(seq.NextValue(), ShouldEqual, 5)
		So(seq.NextValue(), ShouldEqual, 7)
		So(func() { CreateSequence("ManualSequence", 1, 1) }, ShouldPanic)
		seq.Drop()
		So(TestAdapter.sequences("%_manseq"), ShouldHaveLength, 0)
	})
	Convey("Boot sequences cannot be altered or dropped after bootstrap", t, func() {
		bootSeq := Registry.MustGetSequence("TestSequence")
		So(bootSeq.boot, ShouldBeTrue)
		So(func() { bootSeq.Alter(3, 4) }, ShouldPanic)
		So(func() { bootSeq.Drop() }, ShouldPanic)
	})
}

func TestFreeTransientModels(t *testing.T) {
	transientModelTimeout = 1500 * time.Millisecond
	// Start workerloop with a very small period
	workerFunctions = []WorkerFunction{NewWorkerFunction(FreeTransientModels, 400*time.Millisecond)}
	RunWorkerLoop()

	var wizID int64
	Convey("Test freeing transient models", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			wizModel := env.Pool("Wizard")
			Convey("Creating a transient record", func() {
				wiz := wizModel.Call("Create", NewModelData(wizModel.model).
					Set(Name, "WizName").
					Set(value, 13)).(RecordSet).Collection()
				wizID = wiz.Ids()[0]
			})
			Convey("Loading the record immediately and it should still be there", func() {
				wiz := wizModel.Call("BrowseOne", wizID).(RecordSet).Collection()
				wiz.Load()
				So(wiz.IsNotEmpty(), ShouldBeTrue)
			})
			Convey("After transient model timeout, it should not be there anymore", func() {
				<-time.After(2 * time.Second)
				wiz := wizModel.Call("BrowseOne", wizID).(RecordSet).Collection()
				wiz.Load()
				So(wiz.IsEmpty(), ShouldBeTrue)
			})
		}), ShouldBeNil)
	})
	StopWorkerLoop()
	if workerStop != nil {
		t.Fail()
	}
}
