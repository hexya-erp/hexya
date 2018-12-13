// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
	})
}
