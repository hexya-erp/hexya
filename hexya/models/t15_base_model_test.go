// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBaseModelMethods(t *testing.T) {
	Convey("Testing base model methods", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			userJane := userModel.Search(env, userModel.Field("Email").Equals("jane.smith@example.com"))
			Convey("LastUpdate", func() {
				So(userJane.Get("LastUpdate").(dates.DateTime).Sub(userJane.Get("WriteDate").(dates.DateTime).Time), ShouldBeLessThanOrEqualTo, 1*time.Second)
				newUser := userModel.Create(env, FieldMap{
					"Name":    "Alex Smith",
					"Email":   "jsmith@example.com",
					"IsStaff": true,
					"Nums":    1,
				})
				time.Sleep(1*time.Second + 100*time.Millisecond)
				So(newUser.Get("WriteDate").(dates.DateTime).IsZero(), ShouldBeTrue)
				So(newUser.Get("LastUpdate").(dates.DateTime).Sub(newUser.Get("CreateDate").(dates.DateTime).Time), ShouldBeLessThanOrEqualTo, 1*time.Second)
			})
			Convey("Load and Read", func() {
				userJane = userJane.Call("Load", []string{"ID", "Name", "Age", "Posts", "Profile"}).(RecordSet).Collection()
				res := userJane.Call("Read", []string{"Name", "Age", "Posts", "Profile"})
				So(res, ShouldHaveLength, 1)
				fMap := res.([]FieldMap)[0]
				So(fMap, ShouldHaveLength, 5)
				So(fMap, ShouldContainKey, "Name")
				So(fMap["Name"], ShouldEqual, "Jane A. Smith")
				So(fMap, ShouldContainKey, "Age")
				So(fMap["Age"], ShouldEqual, 24)
				So(fMap, ShouldContainKey, "Posts")
				So(fMap["Posts"].(RecordSet).Collection().Ids(), ShouldHaveLength, 2)
				So(fMap, ShouldContainKey, "Profile")
				So(fMap["Profile"].(RecordSet).Collection().Get("ID"), ShouldEqual, userJane.Get("Profile").(RecordSet).Collection().Get("ID"))
				So(fMap, ShouldContainKey, "id")
				So(fMap["id"], ShouldEqual, userJane.Ids()[0])
			})
			Convey("SearchCount", func() {
				countSingle := userJane.Call("SearchCount").(int)
				So(countSingle, ShouldEqual, 1)
				allCount := env.Pool(userModel.name).Call("SearchCount").(int)
				So(allCount, ShouldEqual, 3)
			})
			Convey("Copy", func() {
				newProfile := userJane.Get("Profile").(RecordSet).Collection().Call("Copy", FieldMap{})
				userJane.Call("Write", FieldMap{"Password": "Jane's Password"})
				userJaneCopy := userJane.Call("Copy", FieldMap{
					"Name":    "Jane's Copy",
					"Email2":  "js@example.com",
					"Profile": newProfile,
				}).(RecordSet).Collection()
				So(userJaneCopy.Get("Name"), ShouldEqual, "Jane's Copy")
				So(userJaneCopy.Get("Email"), ShouldEqual, "jane.smith@example.com")
				So(userJaneCopy.Get("Email2"), ShouldEqual, "js@example.com")
				So(userJaneCopy.Get("Password"), ShouldBeBlank)
				So(userJaneCopy.Get("Age"), ShouldEqual, 24)
				So(userJaneCopy.Get("Nums"), ShouldEqual, 2)
				So(userJaneCopy.Get("Posts").(RecordSet).Collection().Len(), ShouldEqual, 0)
			})
			Convey("FieldGet and FieldsGet", func() {
				fInfo := userJane.Call("FieldGet", FieldName("Name")).(*FieldInfo)
				So(fInfo.String, ShouldEqual, "Name")
				So(fInfo.Help, ShouldEqual, "The user's username")
				So(fInfo.Type, ShouldEqual, fieldtype.Char)
				fInfos := userJane.Call("FieldsGet", FieldsGetArgs{}).(map[string]*FieldInfo)
				So(fInfos, ShouldHaveLength, 30)
			})
			Convey("NameGet", func() {
				So(userJane.Get("DisplayName"), ShouldEqual, "Jane A. Smith")
				profile := userJane.Get("Profile").(RecordSet).Collection()
				So(profile.Get("DisplayName"), ShouldEqual, fmt.Sprintf("Profile(%d)", profile.Get("ID")))
			})
			Convey("DefaultGet", func() {
				defaults := userJane.Call("DefaultGet").(FieldMap)
				So(defaults, ShouldHaveLength, 6)
				So(defaults, ShouldContainKey, "status_json")
				So(defaults["status_json"], ShouldEqual, 12)
				So(defaults, ShouldContainKey, "hexya_external_id")
				So(defaults, ShouldContainKey, "is_active")
				So(defaults["is_active"], ShouldEqual, false)
				So(defaults, ShouldContainKey, "active")
				So(defaults["active"], ShouldEqual, true)
				So(defaults, ShouldContainKey, "is_premium")
				So(defaults["is_premium"], ShouldEqual, false)
				So(defaults, ShouldContainKey, "is_staff")
				So(defaults["is_staff"], ShouldEqual, false)
			})
			Convey("Onchange", func() {
				res := userJane.Call("Onchange", OnchangeParams{
					Fields:   []string{"Name"},
					Onchange: map[string]string{"Name": "1"},
					Values:   FieldMap{"Name": "William", "Email": "will@example.com"},
				}).(OnchangeResult)
				fMap := res.Value.FieldMap()
				So(fMap, ShouldHaveLength, 1)
				So(fMap, ShouldContainKey, "decorated_name")
				So(fMap["decorated_name"], ShouldEqual, "User: William [<will@example.com>]")
			})
			Convey("CheckRecursion", func() {
				So(userJane.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag1 := env.Pool("Tag").Call("Create", FieldMap{
					"Name": "Tag1",
				}).(RecordSet).Collection()
				So(tag1.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag2 := env.Pool("Tag").Call("Create", FieldMap{
					"Name":   "Tag2",
					"Parent": tag1,
				}).(RecordSet).Collection()
				So(tag2.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag3 := env.Pool("Tag").Call("Create", FieldMap{
					"Name":   "Tag1",
					"Parent": tag2,
				}).(RecordSet).Collection()
				So(tag3.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag1.Set("Parent", tag3)
				So(tag1.Call("CheckRecursion").(bool), ShouldBeFalse)
				So(tag2.Call("CheckRecursion").(bool), ShouldBeFalse)
				So(tag3.Call("CheckRecursion").(bool), ShouldBeFalse)
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
					Field("Name").Equals("John Smith")).(RecordSet).Collection()
				So(userJohn.Call("Equals", userJane), ShouldBeFalse)
				johnAndJane := userJohn.Union(userJane)
				usersJ := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Like("J% Smith")).(RecordSet).Collection()
				So(usersJ.Records(), ShouldHaveLength, 2)
				So(usersJ.Equals(johnAndJane), ShouldBeTrue)
			})
			Convey("Union", func() {
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				userWill := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Equals("Will Smith")).(RecordSet).Collection()
				johnAndWill := userWill.Union(userJohn)
				So(johnAndJane.Len(), ShouldEqual, 2)
				So(johnAndWill.Len(), ShouldEqual, 2)
				all := johnAndJane.Union(johnAndWill)
				So(all.Len(), ShouldEqual, 3)
				So(all.Intersect(userJane).Equals(userJane), ShouldBeTrue)
				So(all.Intersect(userJohn).Equals(userJohn), ShouldBeTrue)
				So(all.Intersect(userWill).Equals(userWill), ShouldBeTrue)
			})
			Convey("Subtract", func() {
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				So(johnAndJane.Subtract(userJane).Equals(userJohn), ShouldBeTrue)
				So(johnAndJane.Subtract(userJohn).Equals(userJane), ShouldBeTrue)
			})
			Convey("Intersect", func() {
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				So(johnAndJane.Intersect(userJane).Equals(userJane), ShouldBeTrue)
				So(johnAndJane.Call("Intersect", userJohn).(RecordSet).Collection().Equals(userJohn), ShouldBeTrue)
			})
			Convey("ConvertLimitToInt", func() {
				So(ConvertLimitToInt(12), ShouldEqual, 12)
				So(ConvertLimitToInt(false), ShouldEqual, -1)
				So(ConvertLimitToInt(0), ShouldEqual, 0)
				So(ConvertLimitToInt(nil), ShouldEqual, 80)
			})
			Convey("CartesianProduct", func() {
				tagA := env.Pool("Tag").Call("Create", FieldMap{"Name": "A"}).(RecordSet).Collection()
				tagB := env.Pool("Tag").Call("Create", FieldMap{"Name": "B"}).(RecordSet).Collection()
				tagC := env.Pool("Tag").Call("Create", FieldMap{"Name": "C"}).(RecordSet).Collection()
				tagD := env.Pool("Tag").Call("Create", FieldMap{"Name": "D"}).(RecordSet).Collection()
				tagE := env.Pool("Tag").Call("Create", FieldMap{"Name": "E"}).(RecordSet).Collection()
				tagF := env.Pool("Tag").Call("Create", FieldMap{"Name": "F"}).(RecordSet).Collection()
				tagG := env.Pool("Tag").Call("Create", FieldMap{"Name": "G"}).(RecordSet).Collection()
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
					env.Pool("Post").Call("Create", FieldMap{
						"Title": fmt.Sprintf("Post no %02d", (24-i)%20),
						"User":  userJane,
					})
				}
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Title").Contains("Post no")).OrderBy("ID")
				for i, post := range posts.Records() {
					So(post.Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", (24-i)%20))
				}

				sortedPosts := posts.Call("Sorted", func(rs1 RecordSet, rs2 RecordSet) bool {
					return rs1.Collection().Get("Title").(string) < rs2.Collection().Get("Title").(string)
				}).(RecordSet).Collection().Records()
				So(sortedPosts, ShouldHaveLength, 20)
				for i, post := range sortedPosts {
					So(post.Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", i))
				}
			})
			Convey("SortedDefault", func() {
				Convey("With posts", func() {
					for i := 0; i < 20; i++ {
						env.Pool("Post").Call("Create", FieldMap{
							"Title": fmt.Sprintf("Post no %02d", (24-i)%20),
							"User":  userJane,
						})
					}
					posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Title").Contains("Post no")).OrderBy("ID")
					for i, post := range posts.Records() {
						So(post.Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", (24-i)%20))
					}

					sortedPosts := posts.Call("SortedDefault").(RecordSet).Collection().Records()
					So(sortedPosts, ShouldHaveLength, 20)
					for i, post := range sortedPosts {
						So(post.Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", i))
					}
				})
				Convey("With tags", func() {
					env.Pool("Tag").SearchAll().Call("Unlink")
					for i := 0; i < 20; i++ {
						env.Pool("Tag").Call("Create", FieldMap{
							"Name": fmt.Sprintf("Tag %02d", i/2),
						})
					}
					tags := env.Pool("Tag").SearchAll()
					sortedTags := tags.Call("SortedDefault").(RecordSet).Collection().Records()
					for i, tag := range sortedTags {
						So(tag.Get("Name"), ShouldEqual, fmt.Sprintf("Tag %02d", 9-(i/2)))
						So(tag.Get("ID"), ShouldEqual, int(sortedTags[0].ids[0])+1-i+i%2-(i+1)%2)
					}
				})
			})
			Convey("SortedByField", func() {
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", FieldMap{
						"Title": fmt.Sprintf("Post no %02d", (24-i)%20),
						"User":  userJane,
					})
				}
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Title").Contains("Post no")).OrderBy("ID")
				for i, post := range posts.Records() {
					So(post.Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", (24-i)%20))
				}

				sortedPosts := posts.Call("SortedByField", FieldName("Title"), false).(RecordSet).Collection().Records()
				So(sortedPosts, ShouldHaveLength, 20)
				for i, post := range sortedPosts {
					So(post.Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", i))
				}

				revSortedPosts := posts.Call("SortedByField", FieldName("Title"), true).(RecordSet).Collection().Records()
				So(revSortedPosts, ShouldHaveLength, 20)
				for i, post := range revSortedPosts {
					So(post.Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", 19-i))
				}
			})
			Convey("Testing one2many sets keep the default order", func() {
				userJane.Get("Posts").(RecordSet).Collection().Call("Unlink")
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", FieldMap{
						"Title": fmt.Sprintf("Post no %02d", 19-i),
						"User":  userJane,
					})
				}

				posts := userJane.Get("Posts").(RecordSet).Collection()
				So(posts.Len(), ShouldEqual, 20)
				for i, post := range posts.Records() {
					So(post.Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", i))
				}
			})
			Convey("Filtered", func() {
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", FieldMap{
						"Title": fmt.Sprintf("Post no %02d", i),
						"User":  userJane,
					})
				}
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Title").Contains("Post no"))

				evenPosts := posts.Call("Filtered", func(rs RecordSet) bool {
					var num int
					_, err := fmt.Sscanf(rs.Collection().Get("Title").(string), "Post no %02d", &num)
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
					So(evenPosts[i].Get("Title"), ShouldEqual, fmt.Sprintf("Post no %02d", 2*i))
				}
			})
		}), ShouldBeNil)
	})
}
