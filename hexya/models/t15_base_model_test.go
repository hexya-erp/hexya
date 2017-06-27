// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBaseModelMethods(t *testing.T) {
	Convey("Testing base model methods", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			userJane := userModel.Search(env, userModel.Field("Email").Equals("jane.smith@example.com"))
			Convey("LastUpdate", func() {
				So(userJane.Get("LastUpdate").(types.DateTime).Sub(userJane.Get("WriteDate").(types.DateTime).Time), ShouldBeLessThanOrEqualTo, 1*time.Second)
				newUser := userModel.Create(env, FieldMap{
					"Name":    "Alex Smith",
					"Email":   "jsmith@example.com",
					"IsStaff": true,
					"Nums":    1,
				})
				time.Sleep(1*time.Second + 100*time.Millisecond)
				So(newUser.Get("WriteDate").(types.DateTime).IsZero(), ShouldBeTrue)
				So(newUser.Get("LastUpdate").(types.DateTime).Sub(newUser.Get("CreateDate").(types.DateTime).Time), ShouldBeLessThanOrEqualTo, 1*time.Second)
			})
			Convey("Load and Read", func() {
				userJane = userJane.Call("Load", []string{"ID", "Name", "Age", "Posts", "Profile"}).(RecordCollection)
				res := userJane.Call("Read", []string{"Name", "Age", "Posts", "Profile"})
				So(res, ShouldHaveLength, 1)
				fMap := res.([]FieldMap)[0]
				So(fMap, ShouldHaveLength, 5)
				So(fMap, ShouldContainKey, "Name")
				So(fMap["Name"], ShouldEqual, "Jane A. Smith")
				So(fMap, ShouldContainKey, "Age")
				So(fMap["Age"], ShouldEqual, 24)
				So(fMap, ShouldContainKey, "Posts")
				So(fMap["Posts"].(RecordCollection).Ids(), ShouldHaveLength, 2)
				So(fMap, ShouldContainKey, "Profile")
				So(fMap["Profile"].(RecordCollection).Get("ID"), ShouldEqual, userJane.Get("Profile").(RecordCollection).Get("ID"))
				So(fMap, ShouldContainKey, "id")
				So(fMap["id"], ShouldEqual, userJane.Ids()[0])
			})
			Convey("Copy", func() {
				userJane.Call("Write", FieldMap{"Password": "Jane's Password"})
				userJaneCopy := userJane.Call("Copy").(RecordCollection)
				So(userJaneCopy.Get("Name"), ShouldBeBlank)
				So(userJaneCopy.Get("Email"), ShouldEqual, "jane.smith@example.com")
				So(userJaneCopy.Get("Password"), ShouldBeBlank)
				So(userJaneCopy.Get("Age"), ShouldEqual, 24)
				So(userJaneCopy.Get("Nums"), ShouldEqual, 2)
				So(userJaneCopy.Get("Posts").(RecordCollection).Len(), ShouldEqual, 0)
			})
			Convey("FieldGet and FieldsGet", func() {
				fInfo := userJane.Call("FieldGet", FieldName("Name")).(*FieldInfo)
				So(fInfo.String, ShouldEqual, "Name")
				So(fInfo.Help, ShouldEqual, "The user's username")
				So(fInfo.Type, ShouldEqual, fieldtype.Char)
				fInfos := userJane.Call("FieldsGet", FieldsGetArgs{}).(map[string]*FieldInfo)
				So(fInfos, ShouldHaveLength, 34)
			})
			Convey("NameGet", func() {
				So(userJane.Get("DisplayName"), ShouldEqual, "Jane A. Smith")
				profile := userJane.Get("Profile").(RecordCollection)
				So(profile.Get("DisplayName"), ShouldEqual, fmt.Sprintf("Profile(%d)", profile.Get("ID")))
			})
			Convey("DefaultGet", func() {
				defaults := userJane.Call("DefaultGet").(FieldMap)
				So(defaults, ShouldHaveLength, 2)
				So(defaults, ShouldContainKey, "status_json")
				So(defaults["status_json"], ShouldEqual, 12)
				So(defaults, ShouldContainKey, "hexya_external_id")
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
				}).(RecordCollection)
				So(tag1.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag2 := env.Pool("Tag").Call("Create", FieldMap{
					"Name":   "Tag2",
					"Parent": tag1,
				}).(RecordCollection)
				So(tag2.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag3 := env.Pool("Tag").Call("Create", FieldMap{
					"Name":   "Tag1",
					"Parent": tag2,
				}).(RecordCollection)
				So(tag3.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag1.Set("Parent", tag3)
				So(tag1.Call("CheckRecursion").(bool), ShouldBeFalse)
				So(tag2.Call("CheckRecursion").(bool), ShouldBeFalse)
				So(tag3.Call("CheckRecursion").(bool), ShouldBeFalse)
			})
		})
	})
}
