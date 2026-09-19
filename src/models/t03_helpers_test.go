// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/security"
)

type TestProfileSet struct {
	*RecordCollection
}

type TestUserData struct {
	*ModelData
}

type TestUserCondition struct {
	*Condition
}

func TestTypes(t *testing.T) {
	t.Run("Testing models types", func(t *testing.T) {
		t.Run("Testing FieldMap methods", func(t *testing.T) {
			testMap := FieldMap{
				"Name":    "John Smith",
				"Email":   "jsmith2@example.com",
				"Nums":    13,
				"IsStaff": false,
			}
			t.Run("MustGet", func(t *testing.T) {
				assert.NotPanics(t, func() { testMap.MustGet(Registry.MustGet("User").FieldName("Name")) })
				assert.Panics(t, func() { testMap.MustGet(Registry.MustGet("User").FieldName("NoField")) })
				assert.Panics(t, func() { testMap.MustGet(Registry.MustGet("User").FieldName("Profile")) })
			})
			t.Run("RemovePKIfZero", func(t *testing.T) {
				testMap["id"] = int64(12)
				testMap.RemovePKIfZero()
				assert.EqualValues(t, testMap["id"], int64(12))
				testMap["id"] = int64(0)
				testMap.RemovePKIfZero()
				_, ok := testMap["id"]
				assert.False(t, ok)
				testMap["ID"] = int64(0)
				testMap.RemovePKIfZero()
				_, ok = testMap["ID"]
				assert.False(t, ok)
			})
			t.Run("OrderedKeys", func(t *testing.T) {
				keys := testMap.OrderedKeys()
				assert.Len(t, keys, 4)
				assert.EqualValues(t, keys[0], "Email")
				assert.EqualValues(t, keys[1], "IsStaff")
				assert.EqualValues(t, keys[2], "Name")
				assert.EqualValues(t, keys[3], "Nums")
			})
			t.Run("Keys", func(t *testing.T) {
				keys := testMap.Keys()
				assert.Len(t, keys, 4)
				assert.Contains(t, keys, "Email")
				assert.Contains(t, keys, "IsStaff")
				assert.Contains(t, keys, "Name")
				assert.Contains(t, keys, "Nums")
			})
			t.Run("FieldNames", func(t *testing.T) {
				keys := testMap.FieldNames(Registry.MustGet("User"))
				assert.Len(t, keys, 4)
				assert.Contains(t, keys, Registry.MustGet("User").FieldName("Email"))
				assert.Contains(t, keys, Registry.MustGet("User").FieldName("IsStaff"))
				assert.Contains(t, keys, Registry.MustGet("User").FieldName("Name"))
				assert.Contains(t, keys, Registry.MustGet("User").FieldName("Nums"))
			})
			t.Run("Values", func(t *testing.T) {
				keys := testMap.Values()
				assert.Len(t, keys, 4)
				assert.Contains(t, keys, "John Smith")
				assert.Contains(t, keys, "jsmith2@example.com")
				assert.Contains(t, keys, 13)
				assert.Contains(t, keys, false)
			})
		})
		t.Run("Checking ModelData methods", func(t *testing.T) {
			numsField := Registry.MustGet("User").FieldName("Nums")
			johnValues := NewModelData(Registry.MustGet("User")).
				Set(Registry.MustGet("User").FieldName("Email"), "jsmith2@example.com").
				Set(numsField, 13).
				Set(Registry.MustGet("User").FieldName("IsStaff"), false)
			assert.True(t, johnValues.Has(numsField))
			assert.EqualValues(t, johnValues.Get(numsField), 13)
			jv2 := johnValues.Copy()
			johnValues.Unset(numsField)
			assert.False(t, johnValues.Has(numsField))
			assert.EqualValues(t, johnValues.Get(numsField), nil)
			assert.True(t, jv2.Has(numsField))
			assert.EqualValues(t, jv2.Get(numsField), 13)
		})
		t.Run("Checking JSON marshalling of a ModelData", func(t *testing.T) {
			johnValues := NewModelData(Registry.MustGet("User")).
				Set(Registry.MustGet("User").FieldName("Email"), "jsmith2@example.com").
				Set(Registry.MustGet("User").FieldName("Nums"), 13).
				Set(Registry.MustGet("User").FieldName("IsStaff"), false)
			jData, err := json.Marshal(johnValues)
			assert.Nil(t, err)
			var fm FieldMap
			err = json.Unmarshal(jData, &fm)
			assert.Nil(t, err)
			assert.Len(t, fm, 3)
			assert.Contains(t, fm, "email")
			assert.Contains(t, fm, "nums")
			assert.Contains(t, fm, "is_staff")
			assert.EqualValues(t, fm["email"], "jsmith2@example.com")
			assert.IsType(t, float64(0), fm["nums"])
			assert.EqualValues(t, fm["nums"], 13)
			assert.EqualValues(t, fm["is_staff"], false)
			md := NewModelData(Registry.MustGet("User"), fm)
			assert.IsType(t, int(0), md.Get(nums))
			assert.EqualValues(t, md.Get(nums), 13)
		})
		t.Run("Checking NewModelData with FieldMap", func(t *testing.T) {
			johnValues := NewModelData(Registry.MustGet("User"), FieldMap{
				"Email":    "jsmith2@example.com",
				"Nums":     13,
				"IsStaff":  false,
				"Profile":  false,
				"LastPost": nil,
				"Password": false,
			})
			assert.EqualValues(t, johnValues.Get(Registry.MustGet("User").FieldName("Nums")), 13)
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Nums")))
			assert.EqualValues(t, johnValues.Get(Registry.MustGet("User").FieldName("Profile")), 0)
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Profile")))
			assert.EqualValues(t, johnValues.Get(Registry.MustGet("User").FieldName("LastPost")), nil)
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("LastPost")))
			assert.EqualValues(t, johnValues.Get(Registry.MustGet("User").FieldName("Password")), "")
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Password")))
		})
		t.Run("Checking NewModelDataFromRS with FieldMap", func(t *testing.T) {
			var johnValues *ModelData
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				johnValues = NewModelDataFromRS(env.Pool("User"), FieldMap{
					"Email":    "jsmith2@example.com",
					"Nums":     13,
					"IsStaff":  false,
					"Profile":  false,
					"LastPost": nil,
					"Password": false,
					"Size":     []byte("12.34"),
					"Mana":     []byte("234.5"),
				})
			}))
			assert.EqualValues(t, johnValues.Get(Registry.MustGet("User").FieldName("Nums")), 13)
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Nums")))
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Profile")))
			assert.True(t, johnValues.Get(Registry.MustGet("User").FieldName("Profile")).(RecordSet).IsEmpty())
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("LastPost")))
			assert.True(t, johnValues.Get(Registry.MustGet("User").FieldName("LastPost")).(RecordSet).IsEmpty())
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Password")))
			assert.EqualValues(t, johnValues.Get(Registry.MustGet("User").FieldName("Password")), "")
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Size")))
			assert.EqualValues(t, johnValues.Get(Registry.MustGet("User").FieldName("Size")), 12.34)
			assert.IsType(t, *new(float64), johnValues.Get(Registry.MustGet("User").FieldName("Size")))
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Mana")))
			assert.EqualValues(t, johnValues.Get(Registry.MustGet("User").FieldName("Mana")), 234.5)
			assert.IsType(t, *new(float32), johnValues.Get(Registry.MustGet("User").FieldName("Mana")))
		})
		t.Run("Testing Create feature of ModelData", func(t *testing.T) {
			johnValues := NewModelData(Registry.MustGet("User")).
				Set(Registry.MustGet("User").FieldName("Email"), "jsmith2@example.com").
				Set(Registry.MustGet("User").FieldName("Nums"), 13).
				Set(Registry.MustGet("User").FieldName("IsStaff"), false).
				Create(Registry.MustGet("User").FieldName("Profile"), NewModelData(Registry.MustGet("Profile")).
					Set(Registry.MustGet("Profile").FieldName("Age"), 23).
					Set(Registry.MustGet("Profile").FieldName("Money"), 12345).
					Set(Registry.MustGet("Profile").FieldName("Street"), "165 5th Avenue").
					Set(Registry.MustGet("Profile").FieldName("City"), "New York").
					Set(Registry.MustGet("Profile").FieldName("Zip"), "0305").
					Set(Registry.MustGet("Profile").FieldName("Country"), "USA")).
				Create(Registry.MustGet("User").FieldName("Posts"), NewModelData(Registry.MustGet("Post")).
					Set(Registry.MustGet("Post").FieldName("Title"), "1st Post").
					Set(Registry.MustGet("Post").FieldName("Content"), "Content of first post")).
				Create(Registry.MustGet("User").FieldName("Posts"), NewModelData(Registry.MustGet("Post")).
					Set(Registry.MustGet("Post").FieldName("Title"), "2nd Post").
					Set(Registry.MustGet("Post").FieldName("Content"), "Content of second post"))
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Email")))
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Profile")))
			assert.True(t, johnValues.Has(Registry.MustGet("User").FieldName("Posts")))

			assert.Panics(t, func() {
				NewModelData(Registry.MustGet("User")).
					Create(Registry.MustGet("User").FieldName("Profile"), NewModelData(Registry.MustGet("Post")).
						Set(Registry.MustGet("Profile").FieldName("Age"), 23).
						Set(Registry.MustGet("Profile").FieldName("Money"), 12345).
						Set(Registry.MustGet("Profile").FieldName("Street"), "165 5th Avenue").
						Set(Registry.MustGet("Profile").FieldName("City"), "New York").
						Set(Registry.MustGet("Profile").FieldName("Zip"), "0305").
						Set(Registry.MustGet("Profile").FieldName("Country"), "USA"))
			})
		})
		t.Run("Testing ModelData Scanning", func(t *testing.T) {
			md := NewModelData(Registry.MustGet("User"))
			err := md.Scan(nil)
			assert.Nil(t, err)
			assert.Len(t, md.FieldMap, 0)
			err = md.Scan(map[string]interface{}{"Nums": 12})
			assert.Nil(t, err)
			assert.Len(t, md.FieldMap, 1)
			assert.Contains(t, md.FieldMap, "Nums")
			assert.EqualValues(t, md.FieldMap["Nums"], 12)
			err = md.Scan(FieldMap{"Nums": 12})
			assert.Nil(t, err)
			assert.Len(t, md.FieldMap, 1)
			assert.Contains(t, md.FieldMap, "Nums")
			assert.EqualValues(t, md.FieldMap["Nums"], 12)
			err = md.Scan("wrong")
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "unexpected type string to represent RecordData: wrong")
		})
	})
	t.Run("Testing FieldNames", func(t *testing.T) {
		t.Run("Creating a new FieldName", func(t *testing.T) {
			fn := NewFieldName("Name", "json")
			assert.EqualValues(t, fn.Name(), "Name")
			assert.EqualValues(t, fn.JSON(), "json")
		})
		t.Run("Unmarshalling FieldNames", func(t *testing.T) {
			data := []byte(`["name1", "name2"]`)
			var fn FieldNames
			err := json.Unmarshal(data, &fn)
			assert.Nil(t, err)
			assert.Len(t, fn, 2)
			assert.EqualValues(t, fn[0].Name(), "name1")
			assert.EqualValues(t, fn[0].JSON(), "name1")
			assert.EqualValues(t, fn[1].Name(), "name2")
			assert.EqualValues(t, fn[1].JSON(), "name2")
			data = []byte(`{}`)
			err = json.Unmarshal(data, &fn)
			assert.NotNil(t, err)
		})
		t.Run("Listing names and json of FieldNames", func(t *testing.T) {
			fn := FieldNames{
				fieldName{name: "Name", json: "name"},
				fieldName{name: "User", json: "user_id"},
			}
			names := fn.Names()
			assert.Len(t, names, 2)
			assert.EqualValues(t, names[0], "Name")
			assert.EqualValues(t, names[1], "User")
			jsons := fn.JSON()
			assert.Len(t, jsons, 2)
			assert.EqualValues(t, jsons[0], "name")
			assert.EqualValues(t, jsons[1], "user_id")
		})
	})
}
