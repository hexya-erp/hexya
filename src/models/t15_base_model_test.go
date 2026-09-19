// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types/dates"
)

func TestBaseModelMethods(t *testing.T) {
	t.Run("Testing base model methods", func(t *testing.T) {
		t.Run("LastUpdate", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				commentModel := Registry.MustGet("Comment")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				assert.LessOrEqual(t, userJane.Get(lastupdate).(dates.DateTime).Sub(userJane.Get(writeDate).(dates.DateTime)), 1*time.Second)
				newComment := commentModel.Create(env, NewModelData(commentModel).
					Set(text, "MyComment"))
				time.Sleep(1*time.Second + 100*time.Millisecond)
				assert.LessOrEqual(t, newComment.Get(lastupdate).(dates.DateTime).Sub(newComment.Get(createDate).(dates.DateTime)), 1*time.Second)
			}))
		})
		t.Run("Load and Read", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				userJane = userJane.Call("Load", []FieldName{ID, Name, age, posts, profile}).(RecordSet).Collection()
				res := userJane.Call("Read", []FieldName{Name, age, posts, profile})
				assert.Len(t, res, 1)
				fMap := res.([]RecordData)[0].Underlying().FieldMap
				assert.Len(t, fMap, 5)
				assert.Contains(t, fMap, "name")
				assert.EqualValues(t, fMap["name"], "Jane A. Smith")
				assert.Contains(t, fMap, "age")
				assert.EqualValues(t, fMap["age"], 24)
				assert.Contains(t, fMap, "posts_ids")
				assert.Len(t, fMap["posts_ids"].(RecordSet).Collection().Ids(), 2)
				assert.Contains(t, fMap, "profile_id")
				assert.EqualValues(t, fMap["profile_id"].(RecordSet).Collection().Get(ID), userJane.Get(profile).(RecordSet).Collection().Get(ID))
				assert.Contains(t, fMap, "id")
				assert.EqualValues(t, fMap["id"], userJane.Ids()[0])
			}))
		})
		t.Run("Browse and BrowseOne", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				jid := userJane.Ids()[0]
				j2 := userModel.Browse(env, []int64{jid})
				assert.True(t, j2.Equals(userJane))
				j21 := userModel.BrowseOne(env, jid)
				assert.True(t, j21.Equals(userJane))
				j22 := env.Pool("User").Call("Browse", []int64{jid}).(RecordSet).Collection()
				assert.True(t, j22.Equals(userJane))
				j23 := env.Pool("User").Call("BrowseOne", jid).(RecordSet).Collection()
				assert.True(t, j23.Equals(userJane))
			}))
		})
		t.Run("SearchCount", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				countSingle := userJane.Call("SearchCount").(int)
				assert.EqualValues(t, countSingle, 1)
				allCount := env.Pool(userModel.name).Call("SearchCount").(int)
				assert.EqualValues(t, allCount, 3)
				countTags := env.Pool("Tag").WithContext("lang", "fr_FR").
					Search(Registry.MustGet("Tag").Field(description).Contains("Nouvelle ")).Call("SearchCount")
				assert.EqualValues(t, countTags, 2)
			}))
		})
		t.Run("Copy", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				profileModel := Registry.MustGet("Profile")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				newProfile := userJane.Get(profile).(RecordSet).Collection().Call("Copy", NewModelData(profileModel)).(RecordSet).Collection()
				assert.False(t, newProfile.Equals(userJane.Get(profile).(RecordSet).Collection()))
				userJane.Call("Write", NewModelData(userModel).Set(password, "Jane's Password"))
				userJaneCopy := userJane.Call("Copy", NewModelData(userModel).
					Set(Name, "Jane's Copy").
					Set(email2, "js@example.com")).(RecordSet).Collection()
				assert.False(t, userJaneCopy.IsEmpty())
				assert.False(t, userJaneCopy.Equals(userJane))
				assert.EqualValues(t, userJaneCopy.Get(Name), "Jane A. Smith (copy)")
				assert.EqualValues(t, userJaneCopy.Get(email), "jane.smith@example.com")
				assert.EqualValues(t, userJaneCopy.Get(email2), "js@example.com")
				assert.Empty(t, userJaneCopy.Get(password))
				assert.False(t, userJaneCopy.Get(profile).(RecordSet).Collection().Equals(userJane.Get(profile).(RecordSet)))
				assert.EqualValues(t, userJaneCopy.Get(profileAge), 24)
				assert.EqualValues(t, userJaneCopy.Get(age), 24)
				assert.EqualValues(t, userJaneCopy.Get(nums), 2)
				assert.EqualValues(t, userJaneCopy.Get(posts).(RecordSet).Collection().Len(), 2)

				assert.NotPanics(t, func() { userJane.Get(profile).(RecordSet).Collection().Call("Copy", nil) })
			}))
		})
		t.Run("FieldGet and FieldsGet", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				fInfo := userJane.Call("FieldGet", FieldName(Name)).(*FieldInfo)
				assert.EqualValues(t, fInfo.String, "Name")
				assert.EqualValues(t, fInfo.Help, "The user's username")
				assert.EqualValues(t, fInfo.Type, fieldtype.Char)
				fInfos := userJane.Call("FieldsGet", FieldsGetArgs{}).(map[string]*FieldInfo)
				assert.Len(t, fInfos, 35)
			}))
		})
		t.Run("NameGet", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.Get(displayName), "Jane A. Smith")
				janeProfile := userJane.Get(profile).(RecordSet).Collection()
				assert.EqualValues(t, janeProfile.Get(displayName), fmt.Sprintf("Profile(%d)", janeProfile.Get(ID)))
			}))
		})
		t.Run("DefaultGet", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				defaults := userJane.Call("DefaultGet").(*ModelData)
				assert.Len(t, defaults.FieldMap, 14)
				assert.Contains(t, defaults.FieldMap, "status_json")
				assert.EqualValues(t, defaults.FieldMap["status_json"], 12)
				assert.Contains(t, defaults.FieldMap, "hexya_external_id")
				assert.Contains(t, defaults.FieldMap, "is_active")
				assert.EqualValues(t, defaults.FieldMap["is_active"], false)
				assert.Contains(t, defaults.FieldMap, "active")
				assert.EqualValues(t, defaults.FieldMap["active"], true)
				assert.Contains(t, defaults.FieldMap, "is_premium")
				assert.EqualValues(t, defaults.FieldMap["is_premium"], false)
				assert.Contains(t, defaults.FieldMap, "is_staff")
				assert.EqualValues(t, defaults.FieldMap["is_staff"], false)
			}))
		})
		t.Run("New", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				dummyUser := env.Pool("User").Call("New", NewModelData(userModel).
					Set(Name, "DummyUser").
					Set(email, "du@example.com")).(RecordSet).Collection()
				assert.EqualValues(t, dummyUser.Get(Name), "DummyUser")
				assert.EqualValues(t, dummyUser.Get(email), "du@example.com")
				assert.Empty(t, dummyUser.Get(email2))
				assert.Less(t, dummyUser.Ids()[0], int64(0))
				assert.Panics(t, func() { dummyUser.ForceLoad() })
				assert.NotPanics(t, func() { dummyUser.Set(email2, "du2@example.com") })
				assert.EqualValues(t, dummyUser.Get(email2), "du2@example.com")
				assert.EqualValues(t, dummyUser.Get(decoratedName), "User: DummyUser [<du@example.com>]")
				assert.NotPanics(t, func() { dummyUser.unlink() })
			}))
		})
		t.Run("Onchange", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				t.Run("Testing with existing RecordSet", func(t *testing.T) {
					res := userJane.Call("Onchange", OnchangeParams{
						Fields:   []FieldName{Name, coolType, age},
						Onchange: map[string]string{"Name": "1", "CoolType": "1", "Age": "1"},
						Values:   NewModelData(userModel, FieldMap{"Name": "William", "CoolType": "cool", "IsCool": false, "DecoratedName": false, "Profile": false, "Age": int16(24)}),
					}).(OnchangeResult)
					fMap := res.Value.Underlying().FieldMap
					assert.Len(t, fMap, 3)
					assert.Contains(t, fMap, "decorated_name")
					assert.EqualValues(t, fMap["decorated_name"], "User: William [<jane.smith@example.com>]")
					assert.Contains(t, fMap, "is_cool")
					assert.EqualValues(t, fMap["is_cool"], true)
					assert.Contains(t, fMap, "age")
					assert.EqualValues(t, fMap["age"], int16(0))
					assert.Empty(t, res.Warning)
					assert.Len(t, res.Filters, 1)
					var fKey FieldName
					var fValue Conditioner
					for k, v := range res.Filters {
						fKey = k
						fValue = v
					}
					assert.EqualValues(t, fKey.Name(), lastPost.Name())
					assert.EqualValues(t, fKey.JSON(), lastPost.JSON())
					assert.EqualValues(t, fValue.Underlying().String(), `AND Street = addr
`)
				})
				t.Run("Testing onchange warnings", func(t *testing.T) {
					res := userJane.Call("Onchange", OnchangeParams{
						Fields:   []FieldName{Name, coolType, age},
						Onchange: map[string]string{"Name": "1", "CoolType": "1", "age": "1"},
						Values:   NewModelData(userModel, FieldMap{"Name": "Warning User", "CoolType": "cool", "IsCool": false, "DecoratedName": false, "Profile": false, "age": int16(24)}),
					}).(OnchangeResult)
					assert.EqualValues(t, res.Warning, "We have a warning here")
				})
				t.Run("Testing with new RecordSet", func(t *testing.T) {
					res := env.Pool("User").Call("Onchange", OnchangeParams{
						Fields:   []FieldName{Name, email, coolType, nums, nums},
						Onchange: map[string]string{"Name": "1", "CoolType": "1", "Nums": "1"},
						Values:   NewModelData(userModel, FieldMap{"Name": "", "Email": "", "CoolType": "cool", "IsCool": false, "DecoratedName": false}),
					}).(OnchangeResult)
					fMap := res.Value.Underlying().FieldMap
					assert.Len(t, fMap, 2)
					assert.Contains(t, fMap, "decorated_name")
					assert.EqualValues(t, fMap["decorated_name"], "User:  [<>]")
					assert.Contains(t, fMap, "is_cool")
					assert.EqualValues(t, fMap["is_cool"], true)
				})
				t.Run("Testing with new RecordSet and related field", func(t *testing.T) {
					post := env.Pool("Post").SearchAll().Limit(1)
					res := env.Pool("User").Call("Onchange", OnchangeParams{
						Fields:   []FieldName{Name, email, coolType, nums, nums, mana},
						Onchange: map[string]string{"Name": "1", "CoolType": "1", "Nums": "1", "Mana": "1"},
						Values: NewModelData(userModel, FieldMap{"Name": "", "Email": "", "CoolType": "cool",
							"IsCool": false, "DecoratedName": false, "Mana": float32(12.3), "BestProfilePost": false}),
					}).(OnchangeResult)
					fMap := res.Value.Underlying().FieldMap
					assert.Len(t, fMap, 3)
					assert.Contains(t, fMap, "decorated_name")
					assert.EqualValues(t, fMap["decorated_name"], "User:  [<>]")
					assert.Contains(t, fMap, "is_cool")
					assert.EqualValues(t, fMap["is_cool"], true)
					assert.Contains(t, fMap, "best_profile_post_id")
					assert.True(t, fMap["best_profile_post_id"].(RecordSet).Collection().Equals(post))
				})
			}))
		})
		t.Run("CheckRecursion", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				tagModel := Registry.MustGet("Tag")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				assert.True(t, userJane.Call("CheckRecursion").(bool))
				tag1 := env.Pool("Tag").Call("Create", NewModelData(tagModel).
					Set(Name, "Tag1")).(RecordSet).Collection()
				assert.True(t, tag1.Call("CheckRecursion").(bool))
				tag2 := env.Pool("Tag").Call("Create", NewModelData(tagModel).
					Set(Name, "Tag2").
					Set(parent, tag1)).(RecordSet).Collection()
				assert.True(t, tag2.Call("CheckRecursion").(bool))
				tag3 := env.Pool("Tag").Call("Create", NewModelData(tagModel).
					Set(Name, "Tag1").
					Set(parent, tag2)).(RecordSet).Collection()
				assert.True(t, tag3.Call("CheckRecursion").(bool))
				tag1.Set(parent, tag3)
				assert.False(t, tag1.Call("CheckRecursion").(bool))
				assert.False(t, tag2.Call("CheckRecursion").(bool))
				assert.False(t, tag3.Call("CheckRecursion").(bool))
				tagNeg := env.Pool("Tag").Call("New", NewModelData(tagModel).
					Set(Name, "Tag1").
					Set(parent, tag2)).(RecordSet).Collection()
				assert.True(t, tagNeg.Call("CheckRecursion").(bool))
			}))
		})
		t.Run("Browse", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				browsedUser := env.Pool("User").Call("Browse", []int64{userJane.Ids()[0]}).(RecordSet).Collection()
				assert.Len(t, browsedUser.Ids(), 1)
				assert.Contains(t, browsedUser.Ids(), userJane.Ids()[0])
			}))
		})
		t.Run("Equals", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				browsedUser := env.Pool("User").Call("Browse", []int64{userJane.Ids()[0]}).(RecordSet).Collection()
				assert.Equal(t, true, browsedUser.Call("Equals", userJane))
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("John Smith")).(RecordSet).Collection()
				assert.Equal(t, false, userJohn.Call("Equals", userJane))
				johnAndJane := userJohn.Union(userJane)
				usersJ := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Like("J% Smith")).(RecordSet).Collection()
				assert.Len(t, usersJ.Records(), 2)
				assert.True(t, usersJ.Equals(johnAndJane))

				assert.False(t, InvalidRecordCollection("User").Equals(usersJ))
				assert.False(t, env.Pool("Profile").Equals(userJane))
			}))
		})
		t.Run("Union", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				userWill := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("Will Smith")).(RecordSet).Collection()
				johnAndWill := userWill.Union(userJohn)
				assert.EqualValues(t, johnAndJane.Len(), 2)
				assert.EqualValues(t, johnAndWill.Len(), 2)
				all := johnAndJane.Union(johnAndWill)
				assert.EqualValues(t, all.Len(), 3)
				assert.True(t, all.Intersect(userJane).Equals(userJane))
				assert.True(t, all.Intersect(userJohn).Equals(userJohn))
				assert.True(t, all.Intersect(userWill).Equals(userWill))

				assert.False(t, InvalidRecordCollection("User").Union(userJane).IsValid())
				assert.Panics(t, func() { env.Pool("Profile").Union(userJane) })
			}))
		})
		t.Run("Subtract", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				assert.True(t, johnAndJane.Subtract(userJane).Equals(userJohn))
				assert.True(t, johnAndJane.Call("Subtract", userJohn).(RecordSet).Collection().Equals(userJane))

				assert.False(t, InvalidRecordCollection("User").Subtract(userJane).IsValid())
				assert.Panics(t, func() { env.Pool("Profile").Subtract(userJane) })
			}))
		})
		t.Run("Intersect", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field(Name).Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				assert.True(t, johnAndJane.Intersect(userJane).Equals(userJane))
				assert.True(t, johnAndJane.Call("Intersect", userJohn).(RecordSet).Collection().Equals(userJohn))

				assert.False(t, InvalidRecordCollection("User").Intersect(userJane).IsValid())
				assert.Panics(t, func() { env.Pool("Profile").Intersect(userJane) })
			}))
		})
		t.Run("ConvertLimitToInt", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				assert.EqualValues(t, ConvertLimitToInt(12), 12)
				assert.EqualValues(t, ConvertLimitToInt(false), -1)
				assert.EqualValues(t, ConvertLimitToInt(0), 0)
				assert.EqualValues(t, ConvertLimitToInt(nil), 80)
			}))
		})
		t.Run("CartesianProduct", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				tagModel := Registry.MustGet("Tag")
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

				assert.Len(t, cartesianProductSlices(), 0)

				product0 := tagA.CartesianProduct()
				assert.Len(t, product0, 1)
				assert.True(t, product0[0].Equals(tagA))

				product1 := tagsAB.CartesianProduct(tagsCD)
				assert.Len(t, product1, 4)
				assert.True(t, contains(product1,
					tagA.Union(tagC),
					tagA.Union(tagD),
					tagB.Union(tagC),
					tagB.Union(tagD)))

				product2 := tagsAB.CartesianProduct(tagsEFG)
				assert.Len(t, product2, 6)
				assert.True(t, contains(product2,
					tagA.Union(tagE),
					tagA.Union(tagF),
					tagA.Union(tagG),
					tagB.Union(tagE),
					tagB.Union(tagF),
					tagB.Union(tagG)))

				product3 := tagsAB.CartesianProduct(tagsCD, tagsEFG)
				assert.Len(t, product3, 12)
				assert.True(t, contains(product3,
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
					tagB.Union(tagD).Union(tagG)))
			}))
		})
		t.Run("Sorted", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				postModel := Registry.MustGet("Post")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", NewModelData(postModel).
						Set(title, fmt.Sprintf("Post no %02d", (24-i)%20)).
						Set(user, userJane))
				}
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Contains("Post no")).OrderBy("ID")
				for i, post := range rPosts.Records() {
					assert.EqualValues(t, post.Get(title), fmt.Sprintf("Post no %02d", (24-i)%20))
				}

				sortedPosts := rPosts.Call("Sorted", func(rs1 RecordSet, rs2 RecordSet) bool {
					return rs1.Collection().Get(title).(string) < rs2.Collection().Get(title).(string)
				}).(RecordSet).Collection().Records()
				assert.Len(t, sortedPosts, 20)
				for i, post := range sortedPosts {
					assert.EqualValues(t, post.Get(title), fmt.Sprintf("Post no %02d", i))
				}

				assert.False(t, InvalidRecordCollection("Post").Sorted(func(rs1 RecordSet, rs2 RecordSet) bool {
					return rs1.Collection().Get(title).(string) < rs2.Collection().Get(title).(string)
				}).IsValid())
			}))
		})
		t.Run("SortedDefault", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				tagModel := Registry.MustGet("Tag")
				postModel := Registry.MustGet("Post")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				t.Run("With posts", func(t *testing.T) {
					for i := 0; i < 20; i++ {
						env.Pool("Post").Call("Create", NewModelData(postModel).
							Set(title, fmt.Sprintf("Post no %02d", (24-i)%20)).
							Set(user, userJane))
					}
					rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Contains("Post no")).OrderBy("ID")
					for i, post := range rPosts.Records() {
						assert.EqualValues(t, post.Get(title), fmt.Sprintf("Post no %02d", (24-i)%20))
					}

					sortedPosts := rPosts.Call("SortedDefault").(RecordSet).Collection().Records()
					assert.Len(t, sortedPosts, 20)
					for i, post := range sortedPosts {
						assert.EqualValues(t, post.Get(title), fmt.Sprintf("Post no %02d", i))
					}
				})
				t.Run("With tags", func(t *testing.T) {
					env.Pool("Tag").SearchAll().Call("Unlink")
					for i := 0; i < 20; i++ {
						env.Pool("Tag").Call("Create", NewModelData(tagModel).
							Set(Name, fmt.Sprintf("Tag %02d", i/2)))
					}
					rTags := env.Pool("Tag").SearchAll()
					sortedTags := rTags.Call("SortedDefault").(RecordSet).Collection().Records()
					for i, tag := range sortedTags {
						assert.EqualValues(t, tag.Get(Name), fmt.Sprintf("Tag %02d", 9-(i/2)))
						assert.EqualValues(t, tag.Get(ID), int(sortedTags[0].ids[0])+1-i+i%2-(i+1)%2)
					}
				})
			}))
		})
		t.Run("SortedByField", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				postModel := Registry.MustGet("Post")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", NewModelData(postModel).
						Set(title, fmt.Sprintf("Post no %02d", (24-i)%20)).
						Set(user, userJane))
				}
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Contains("Post no")).OrderBy("ID")
				for i, post := range rPosts.Records() {
					assert.EqualValues(t, post.Get(title), fmt.Sprintf("Post no %02d", (24-i)%20))
				}

				sortedPosts := rPosts.Call("SortedByField", FieldName(title), false).(RecordSet).Collection().Records()
				assert.Len(t, sortedPosts, 20)
				for i, post := range sortedPosts {
					assert.EqualValues(t, post.Get(title), fmt.Sprintf("Post no %02d", i))
				}

				revSortedPosts := rPosts.Call("SortedByField", FieldName(title), true).(RecordSet).Collection().Records()
				assert.Len(t, revSortedPosts, 20)
				for i, post := range revSortedPosts {
					assert.EqualValues(t, post.Get(title), fmt.Sprintf("Post no %02d", 19-i))
				}
			}))
		})
		t.Run("Testing one2many sets keep the default order", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				postModel := Registry.MustGet("Post")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				userJane.Get(posts).(RecordSet).Collection().Call("Unlink")
				for i := 0; i < 20; i++ {
					env.Pool("Post").Call("Create", NewModelData(postModel).
						Set(title, fmt.Sprintf("Post no %02d", 19-i)).
						Set(user, userJane))
				}

				rPosts := userJane.Get(posts).(RecordSet).Collection()
				assert.EqualValues(t, rPosts.Len(), 20)
				for i, post := range rPosts.Records() {
					assert.EqualValues(t, post.Get(title), fmt.Sprintf("Post no %02d", i))
				}
			}))
		})
		t.Run("Filtered", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				postModel := Registry.MustGet("Post")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
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
				assert.Len(t, evenPosts, 10)
				for i := 0; i < 10; i++ {
					assert.EqualValues(t, evenPosts[i].Get(title), fmt.Sprintf("Post no %02d", 2*i))
				}

				assert.False(t, InvalidRecordCollection("Post").Filtered(func(rs RecordSet) bool {
					return true
				}).IsValid())
			}))
		})
		t.Run("CheckExecutionPermissions", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				res := env.Pool("User").Call("CheckExecutionPermission", Registry.MustGet("User").Methods().MustGet("Load"), []bool{true})
				assert.Equal(t, true, res)
			}))
		})
		t.Run("convertTotRecordSet", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				profileID := userJane.Get(profile).(RecordSet).Collection().Ids()[0]
				res := env.Pool("User").convertToRecordSet(profileID, "Profile")
				assert.EqualValues(t, res.Ids()[0], profileID)
				res = env.Pool("User").convertToRecordSet(false, "Profile")
				assert.True(t, res.IsEmpty())
				res = env.Pool("User").convertToRecordSet([]interface{}{float64(profileID)}, "Profile")
				assert.EqualValues(t, res.Ids()[0], profileID)
				res = env.Pool("User").convertToRecordSet(int(profileID), "Profile")
				assert.EqualValues(t, res.Ids()[0], profileID)
				res = env.Pool("User").convertToRecordSet(float64(profileID), "Profile")
				assert.EqualValues(t, res.Ids()[0], profileID)
				res = env.Pool("User").convertToRecordSet([]float64{float64(profileID)}, "Profile")
				assert.EqualValues(t, res.Ids()[0], profileID)
				res = env.Pool("User").convertToRecordSet([]int{int(profileID)}, "Profile")
				assert.EqualValues(t, res.Ids()[0], profileID)
				assert.Panics(t, func() { env.Pool("User").convertToRecordSet("", "Profile") })
				res = env.Pool("User").convertToRecordSet(userJane.Get(profile), "Profile")
				assert.EqualValues(t, res.Ids()[0], profileID)
			}))
		})
		t.Run("EnsureOne", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				assert.NotPanics(t, func() { userJane.EnsureOne() })
				assert.Panics(t, func() { env.Pool("User").EnsureOne() })
				users := env.Pool("User").SearchAll()
				assert.Greater(t, users.Len(), 0)
				assert.Panics(t, func() { users.EnsureOne() })
			}))
		})
		t.Run("GetRecord", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				assert.True(t, env.Pool("User").Call("GetRecord", userJane.Get(hexyaExternalID)).(RecordSet).Collection().Equals(userJane))
			}))
		})
		t.Run("SearchByName", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				userJane := userModel.Search(env, userModel.Field(email).Equals("jane.smith@example.com"))
				j := env.Pool("User").Call("SearchByName", "Jane A. Smith", operator.Operator(""), userModel.Field(isStaff).Equals(false), 10).(RecordSet).Collection()
				assert.True(t, j.Equals(userJane))
			}))
		})
	})
}

func TestPostBootSequences(t *testing.T) {
	t.Run("Testing manual sequences after bootstrap", func(t *testing.T) {
		testSeq := Registry.MustGetSequence("Test")
		testSeq.Drop()
		seq := CreateSequence("ManualSequence", 1, 1)
		assert.EqualValues(t, seq.JSON, "manual_sequence_manseq")
		assert.Len(t, TestAdapter.sequences("%_manseq"), 1)
		assert.EqualValues(t, TestAdapter.sequences("%_manseq")[0].Name, "manual_sequence_manseq")
		assert.EqualValues(t, seq.NextValue(), 1)
		assert.EqualValues(t, seq.NextValue(), 2)
		seq.Alter(2, 5)
		assert.EqualValues(t, seq.NextValue(), 5)
		assert.EqualValues(t, seq.NextValue(), 7)
		assert.Panics(t, func() { CreateSequence("ManualSequence", 1, 1) })
		seq.Drop()
		assert.Len(t, TestAdapter.sequences("%_manseq"), 0)
	})
	t.Run("Boot sequences cannot be altered or dropped after bootstrap", func(t *testing.T) {
		bootSeq := Registry.MustGetSequence("TestSequence")
		assert.True(t, bootSeq.boot)
		assert.Panics(t, func() { bootSeq.Alter(3, 4) })
		assert.Panics(t, func() { bootSeq.Drop() })
	})
}

func TestFreeTransientModels(t *testing.T) {
	transientModelTimeout = 1500 * time.Millisecond
	// Start workerloop with a very small period
	workerFunctions = []WorkerFunction{NewWorkerFunction(FreeTransientModels, 400*time.Millisecond)}
	RunWorkerLoop()

	var wizID int64
	t.Run("Test freeing transient models", func(t *testing.T) {
		t.Run("Creating a transient record", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				wizModel := env.Pool("Wizard")
				wiz := wizModel.Call("Create", NewModelData(wizModel.model).
					Set(Name, "WizName").
					Set(value, 13)).(RecordSet).Collection()
				wizID = wiz.Ids()[0]
			}))
		})
		t.Run("Loading the record immediately and it should still be there", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				wizModel := env.Pool("Wizard")
				wiz := wizModel.Call("BrowseOne", wizID).(RecordSet).Collection()
				wiz.Load()
				assert.True(t, wiz.IsNotEmpty())
			}))
		})
		t.Run("After transient model timeout, it should not be there anymore", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				wizModel := env.Pool("Wizard")
				<-time.After(2 * time.Second)
				wiz := wizModel.Call("BrowseOne", wizID).(RecordSet).Collection()
				wiz.Load()
				assert.True(t, wiz.IsEmpty())
			}))
		})
	})
	StopWorkerLoop()
	if workerStop != nil {
		t.Fail()
	}
}
