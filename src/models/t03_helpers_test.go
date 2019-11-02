// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"encoding/json"
	"testing"

	"github.com/hexya-erp/hexya/src/models/security"
	. "github.com/smartystreets/goconvey/convey"
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
	Convey("Testing models types", t, func() {
		Convey("Testing FieldMap methods", func() {
			testMap := FieldMap{
				"Name":    "John Smith",
				"Email":   "jsmith2@example.com",
				"Nums":    13,
				"IsStaff": false,
			}
			Convey("MustGet", func() {
				So(func() { testMap.MustGet(Registry.MustGet("User").FieldName("Name")) }, ShouldNotPanic)
				So(func() { testMap.MustGet(Registry.MustGet("User").FieldName("NoField")) }, ShouldPanic)
				So(func() { testMap.MustGet(Registry.MustGet("User").FieldName("Profile")) }, ShouldPanic)
			})
			Convey("RemovePKIfZero", func() {
				testMap["id"] = int64(12)
				testMap.RemovePKIfZero()
				So(testMap["id"], ShouldEqual, int64(12))
				testMap["id"] = int64(0)
				testMap.RemovePKIfZero()
				_, ok := testMap["id"]
				So(ok, ShouldBeFalse)
				testMap["ID"] = int64(0)
				testMap.RemovePKIfZero()
				_, ok = testMap["ID"]
				So(ok, ShouldBeFalse)
			})
			Convey("OrderedKeys", func() {
				keys := testMap.OrderedKeys()
				So(keys, ShouldHaveLength, 4)
				So(keys[0], ShouldEqual, "Email")
				So(keys[1], ShouldEqual, "IsStaff")
				So(keys[2], ShouldEqual, "Name")
				So(keys[3], ShouldEqual, "Nums")
			})
			Convey("Keys", func() {
				keys := testMap.Keys()
				So(keys, ShouldHaveLength, 4)
				So(keys, ShouldContain, "Email")
				So(keys, ShouldContain, "IsStaff")
				So(keys, ShouldContain, "Name")
				So(keys, ShouldContain, "Nums")
			})
			Convey("FieldNames", func() {
				keys := testMap.FieldNames(Registry.MustGet("User"))
				So(keys, ShouldHaveLength, 4)
				So(keys, ShouldContain, Registry.MustGet("User").FieldName("Email"))
				So(keys, ShouldContain, Registry.MustGet("User").FieldName("IsStaff"))
				So(keys, ShouldContain, Registry.MustGet("User").FieldName("Name"))
				So(keys, ShouldContain, Registry.MustGet("User").FieldName("Nums"))
			})
			Convey("Values", func() {
				keys := testMap.Values()
				So(keys, ShouldHaveLength, 4)
				So(keys, ShouldContain, "John Smith")
				So(keys, ShouldContain, "jsmith2@example.com")
				So(keys, ShouldContain, 13)
				So(keys, ShouldContain, false)
			})
		})
		Convey("Checking ModelData methods", func() {
			numsField := Registry.MustGet("User").FieldName("Nums")
			johnValues := NewModelData(Registry.MustGet("User")).
				Set(Registry.MustGet("User").FieldName("Email"), "jsmith2@example.com").
				Set(numsField, 13).
				Set(Registry.MustGet("User").FieldName("IsStaff"), false)
			So(johnValues.Has(numsField), ShouldBeTrue)
			So(johnValues.Get(numsField), ShouldEqual, 13)
			jv2 := johnValues.Copy()
			johnValues.Unset(numsField)
			So(johnValues.Has(numsField), ShouldBeFalse)
			So(johnValues.Get(numsField), ShouldEqual, nil)
			So(jv2.Has(numsField), ShouldBeTrue)
			So(jv2.Get(numsField), ShouldEqual, 13)
		})
		Convey("Checking JSON marshalling of a ModelData", func() {
			johnValues := NewModelData(Registry.MustGet("User")).
				Set(Registry.MustGet("User").FieldName("Email"), "jsmith2@example.com").
				Set(Registry.MustGet("User").FieldName("Nums"), 13).
				Set(Registry.MustGet("User").FieldName("IsStaff"), false)
			jData, err := json.Marshal(johnValues)
			So(err, ShouldBeNil)
			var fm FieldMap
			err = json.Unmarshal(jData, &fm)
			So(err, ShouldBeNil)
			So(fm, ShouldHaveLength, 3)
			So(fm, ShouldContainKey, "email")
			So(fm, ShouldContainKey, "nums")
			So(fm, ShouldContainKey, "is_staff")
			So(fm["email"], ShouldEqual, "jsmith2@example.com")
			So(fm["nums"], ShouldEqual, 13)
			So(fm["is_staff"], ShouldEqual, false)
		})
		Convey("Checking NewModelData with FieldMap", func() {
			johnValues := NewModelData(Registry.MustGet("User"), FieldMap{
				"Email":    "jsmith2@example.com",
				"Nums":     13,
				"IsStaff":  false,
				"Profile":  false,
				"LastPost": nil,
				"Password": false,
			})
			So(johnValues.Get(Registry.MustGet("User").FieldName("Nums")), ShouldEqual, 13)
			So(johnValues.Has(Registry.MustGet("User").FieldName("Nums")), ShouldBeTrue)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Profile")), ShouldEqual, 0)
			So(johnValues.Has(Registry.MustGet("User").FieldName("Profile")), ShouldBeTrue)
			So(johnValues.Get(Registry.MustGet("User").FieldName("LastPost")), ShouldEqual, nil)
			So(johnValues.Has(Registry.MustGet("User").FieldName("LastPost")), ShouldBeTrue)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Password")), ShouldEqual, "")
			So(johnValues.Has(Registry.MustGet("User").FieldName("Password")), ShouldBeTrue)
		})
		Convey("Checking NewModelDataFromRS with FieldMap", func() {
			var johnValues *ModelData
			So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
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
			}), ShouldBeNil)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Nums")), ShouldEqual, 13)
			So(johnValues.Has(Registry.MustGet("User").FieldName("Nums")), ShouldBeTrue)
			So(johnValues.Has(Registry.MustGet("User").FieldName("Profile")), ShouldBeTrue)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Profile")).(RecordSet).IsEmpty(), ShouldBeTrue)
			So(johnValues.Has(Registry.MustGet("User").FieldName("LastPost")), ShouldBeTrue)
			So(johnValues.Get(Registry.MustGet("User").FieldName("LastPost")).(RecordSet).IsEmpty(), ShouldBeTrue)
			So(johnValues.Has(Registry.MustGet("User").FieldName("Password")), ShouldBeTrue)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Password")), ShouldEqual, "")
			So(johnValues.Has(Registry.MustGet("User").FieldName("Size")), ShouldBeTrue)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Size")), ShouldEqual, 12.34)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Size")), ShouldHaveSameTypeAs, *new(float64))
			So(johnValues.Has(Registry.MustGet("User").FieldName("Mana")), ShouldBeTrue)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Mana")), ShouldEqual, 234.5)
			So(johnValues.Get(Registry.MustGet("User").FieldName("Mana")), ShouldHaveSameTypeAs, *new(float32))
		})
		Convey("Testing Create feature of ModelData", func() {
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
			So(johnValues.Has(Registry.MustGet("User").FieldName("Email")), ShouldBeTrue)
			So(johnValues.Has(Registry.MustGet("User").FieldName("Profile")), ShouldBeTrue)
			So(johnValues.Has(Registry.MustGet("User").FieldName("Posts")), ShouldBeTrue)

			So(func() {
				NewModelData(Registry.MustGet("User")).
					Create(Registry.MustGet("User").FieldName("Profile"), NewModelData(Registry.MustGet("Post")).
						Set(Registry.MustGet("Profile").FieldName("Age"), 23).
						Set(Registry.MustGet("Profile").FieldName("Money"), 12345).
						Set(Registry.MustGet("Profile").FieldName("Street"), "165 5th Avenue").
						Set(Registry.MustGet("Profile").FieldName("City"), "New York").
						Set(Registry.MustGet("Profile").FieldName("Zip"), "0305").
						Set(Registry.MustGet("Profile").FieldName("Country"), "USA"))
			}, ShouldPanic)
		})
		Convey("Testing ModelData Scanning", func() {
			md := NewModelData(Registry.MustGet("User"))
			err := md.Scan(nil)
			So(err, ShouldBeNil)
			So(md.FieldMap, ShouldHaveLength, 0)
			err = md.Scan(map[string]interface{}{"Nums": 12})
			So(err, ShouldBeNil)
			So(md.FieldMap, ShouldHaveLength, 1)
			So(md.FieldMap, ShouldContainKey, "Nums")
			So(md.FieldMap["Nums"], ShouldEqual, 12)
			err = md.Scan(FieldMap{"Nums": 12})
			So(err, ShouldBeNil)
			So(md.FieldMap, ShouldHaveLength, 1)
			So(md.FieldMap, ShouldContainKey, "Nums")
			So(md.FieldMap["Nums"], ShouldEqual, 12)
			err = md.Scan("wrong")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unexpected type string to represent RecordData: wrong")
		})
	})
	Convey("Testing FieldNames", t, func() {
		Convey("Creating a new FieldName", func() {
			fn := NewFieldName("Name", "json")
			So(fn.Name(), ShouldEqual, "Name")
			So(fn.JSON(), ShouldEqual, "json")
		})
		Convey("Unmarshalling FieldNames", func() {
			data := []byte(`["name1", "name2"]`)
			var fn FieldNames
			err := json.Unmarshal(data, &fn)
			So(err, ShouldBeNil)
			So(fn, ShouldHaveLength, 2)
			So(fn[0].Name(), ShouldEqual, "name1")
			So(fn[0].JSON(), ShouldEqual, "name1")
			So(fn[1].Name(), ShouldEqual, "name2")
			So(fn[1].JSON(), ShouldEqual, "name2")
			data = []byte(`{}`)
			err = json.Unmarshal(data, &fn)
			So(err, ShouldNotBeNil)
		})
	})
}
