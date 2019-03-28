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
				So(func() { testMap.MustGet("Name", Registry.MustGet("User")) }, ShouldNotPanic)
				So(func() { testMap.MustGet("NoField", Registry.MustGet("User")) }, ShouldPanic)
				So(func() { testMap.MustGet("Profile", Registry.MustGet("User")) }, ShouldPanic)
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
				keys := testMap.FieldNames()
				So(keys, ShouldHaveLength, 4)
				So(keys, ShouldContain, FieldName("Email"))
				So(keys, ShouldContain, FieldName("IsStaff"))
				So(keys, ShouldContain, FieldName("Name"))
				So(keys, ShouldContain, FieldName("Nums"))
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
			johnValues := NewModelData(Registry.MustGet("User")).
				Set("Email", "jsmith2@example.com").
				Set("Nums", 13).
				Set("IsStaff", false)
			num, ok := johnValues.Get("Nums")
			So(num, ShouldEqual, 13)
			So(ok, ShouldBeTrue)
			jv2 := johnValues.Copy()
			johnValues.Unset("Nums")
			num2, ok2 := johnValues.Get("Nums")
			So(num2, ShouldEqual, nil)
			So(ok2, ShouldBeFalse)
			num3, ok3 := jv2.Get("Nums")
			So(num3, ShouldEqual, 13)
			So(ok3, ShouldBeTrue)
		})
		Convey("Checking JSON marshalling of a ModelData", func() {
			johnValues := NewModelData(Registry.MustGet("User")).
				Set("Email", "jsmith2@example.com").
				Set("Nums", 13).
				Set("IsStaff", false)
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
			num, ok := johnValues.Get("Nums")
			So(num, ShouldEqual, 13)
			So(ok, ShouldBeTrue)
			profile, ok := johnValues.Get("Profile")
			So(profile, ShouldEqual, 0)
			So(ok, ShouldBeTrue)
			lastPost, ok := johnValues.Get("LastPost")
			So(lastPost, ShouldEqual, nil)
			So(ok, ShouldBeTrue)
			password, ok := johnValues.Get("Password")
			So(password, ShouldEqual, "")
			So(ok, ShouldBeTrue)
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
				})
			}), ShouldBeNil)
			num, ok := johnValues.Get("Nums")
			So(num, ShouldEqual, 13)
			So(ok, ShouldBeTrue)
			profile, ok := johnValues.Get("Profile")
			So(ok, ShouldBeTrue)
			prof, ok := profile.(RecordSet)
			So(prof.IsEmpty(), ShouldBeTrue)
			lastPost, ok := johnValues.Get("LastPost")
			So(ok, ShouldBeTrue)
			lastP, ok := lastPost.(RecordSet)
			So(lastP.IsEmpty(), ShouldBeTrue)
			password, ok := johnValues.Get("Password")
			So(password, ShouldEqual, "")
			So(ok, ShouldBeTrue)
		})
	})
	Convey("Testing helper functions", t, func() {
		names := []string{"Name", "Email"}
		fields := ConvertToFieldNameSlice(names)
		So(fields, ShouldHaveLength, 2)
		So(fields, ShouldContain, FieldName("Name"))
		So(fields, ShouldContain, FieldName("Email"))
	})
}
