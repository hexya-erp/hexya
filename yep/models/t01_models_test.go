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
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateDB(t *testing.T) {
	Convey("Creating DataBase...", t, func() {
		user := NewModel("User")
		user.AddCharField("Name", StringFieldParams{String: "Name", Help: "The user's username", Unique: true})
		user.AddCharField("DecoratedName", StringFieldParams{Compute: "computeDecoratedName"})
		user.AddCharField("Email", StringFieldParams{Help: "The user's email address", Size: 100, Index: true})
		user.AddCharField("Password", StringFieldParams{})
		user.AddIntegerField("Status", SimpleFieldParams{JSON: "status_json", GoType: new(int16)})
		user.AddBooleanField("IsStaff", SimpleFieldParams{})
		user.AddBooleanField("IsActive", SimpleFieldParams{})
		user.AddMany2OneField("Profile", ForeignKeyFieldParams{RelationModel: "Profile"})
		user.AddIntegerField("Age", SimpleFieldParams{Compute: "computeAge", Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)})
		user.AddOne2ManyField("Posts", ReverseFieldParams{RelationModel: "Post", ReverseFK: "User"})
		user.AddFloatField("PMoney", FloatFieldParams{Related: "Profile.Money"})
		user.AddMany2OneField("LastPost", ForeignKeyFieldParams{RelationModel: "Post", Embed: true})
		user.AddCharField("Email2", StringFieldParams{})
		user.AddBooleanField("IsPremium", SimpleFieldParams{})
		user.AddIntegerField("Nums", SimpleFieldParams{GoType: new(int)})
		user.AddFloatField("Size", FloatFieldParams{})

		profile := NewModel("Profile")
		profile.AddIntegerField("Age", SimpleFieldParams{GoType: new(int16)})
		profile.AddFloatField("Money", FloatFieldParams{})
		profile.AddMany2OneField("User", ForeignKeyFieldParams{RelationModel: "User"})
		profile.AddOne2OneField("BestPost", ForeignKeyFieldParams{RelationModel: "Post"})
		profile.AddCharField("City", StringFieldParams{})
		profile.AddCharField("Country", StringFieldParams{})

		post := NewModel("Post")
		post.AddMany2OneField("User", ForeignKeyFieldParams{RelationModel: "User"})
		post.AddCharField("Title", StringFieldParams{})
		post.AddTextField("Content", StringFieldParams{})
		post.AddMany2ManyField("Tags", Many2ManyFieldParams{RelationModel: "Tag"})

		tag := NewModel("Tag")
		tag.AddCharField("Name", StringFieldParams{})
		tag.AddMany2OneField("BestPost", ForeignKeyFieldParams{RelationModel: "Post"})
		tag.AddMany2ManyField("Posts", Many2ManyFieldParams{RelationModel: "Post"})
		tag.AddCharField("Description", StringFieldParams{})

		addressMI := NewMixinModel("AddressMixIn")
		addressMI.AddCharField("Street", StringFieldParams{})
		addressMI.AddCharField("Zip", StringFieldParams{})
		addressMI.AddCharField("City", StringFieldParams{})
		profile.MixInModel(addressMI)

		activeMI := NewMixinModel("ActiveMixIn")
		activeMI.AddBooleanField("Active", SimpleFieldParams{})
		MixInAllModels(activeMI)

		viewModel := NewManualModel("UserView")
		viewModel.AddCharField("Name", StringFieldParams{})
		viewModel.AddCharField("City", StringFieldParams{})

		user.AddMethod("PrefixedUser", "",
			func(rc RecordCollection, prefix string) []string {
				var res []string
				for _, u := range rc.Records() {
					res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get("Name")))
				}
				return res
			})

		user.ExtendMethod("PrefixedUser", "",
			func(rc RecordCollection, prefix string) []string {
				res := rc.Super().Call("PrefixedUser", prefix).([]string)
				for i, u := range rc.Records() {
					email := u.Get("Email").(string)
					res[i] = fmt.Sprintf("%s %s", res[i], rc.Call("DecorateEmail", email))
				}
				return res
			})

		user.AddMethod("DecorateEmail", "",
			func(rc RecordCollection, email string) string {
				return fmt.Sprintf("<%s>", email)
			})

		user.ExtendMethod("DecorateEmail", "",
			func(rc RecordCollection, email string) string {
				res := rc.Super().Call("DecorateEmail", email).(string)
				return fmt.Sprintf("[%s]", res)
			})

		user.AddMethod("computeDecoratedName", "",
			func(rc RecordCollection) FieldMap {
				res := make(FieldMap)
				res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
				return res
			})

		user.AddMethod("computeAge", "",
			func(rc RecordCollection) (FieldMap, []FieldNamer) {
				res := make(FieldMap)
				res["Age"] = rc.Get("Profile").(RecordCollection).Get("Age").(int16)
				return res, []FieldNamer{}
			})

		user.AddMethod("UpdateCity", "",
			func(rc RecordCollection, value string) {
				rc.Get("Profile").(RecordCollection).Set("City", value)
			})

		activeMI.AddMethod("IsActivated", "",
			func(rc RecordCollection) bool {
				return rc.Get("Active").(bool)
			})

		addressMI.AddMethod("SayHello", "",
			func(rc RecordCollection) string {
				return "Hello !"
			})

		addressMI.AddMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				return fmt.Sprintf("%s, %s %s", rc.Get("Street"), rc.Get("Zip"), rc.Get("City"))
			})

		profile.AddMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("%s, %s", res, rc.Get("Country"))
			})

		addressMI.ExtendMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("<%s>", res)
			})

		profile.ExtendMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("[%s]", res)
			})

		// Creating a dummy table to check that it is correctly removed by Bootstrap
		dbExecuteNoTx("CREATE TABLE IF NOT EXISTS shouldbedeleted (id serial NOT NULL PRIMARY KEY)")
	})

	Convey("Database creation should run fine", t, func() {
		Convey("Dummy table should exist", func() {
			So(testAdapter.tables(), ShouldContainKey, "shouldbedeleted")
		})
		Convey("Bootstrap should not panic", func() {
			So(BootStrap, ShouldNotPanic)
			So(SyncDatabase, ShouldNotPanic)
		})
		Convey("Creating SQL view should run fine", func() {
			So(func() {
				dbExecuteNoTx(`DROP VIEW IF EXISTS user_view;
					CREATE VIEW user_view AS (
						SELECT u.id, u.name, p.city, u.active
						FROM "user" u
							LEFT JOIN "profile" p ON p.id = u.profile_id
					)`)
			}, ShouldNotPanic)
		})
		Convey("All models should have a DB table", func() {
			dbTables := testAdapter.tables()
			for tableName, mi := range Registry.registryByTableName {
				if mi.isMixin() || mi.isManual() {
					continue
				}
				So(dbTables[tableName], ShouldBeTrue)
			}
		})
		Convey("All DB tables should have a model", func() {
			for dbTable := range testAdapter.tables() {
				So(Registry.registryByTableName, ShouldContainKey, dbTable)
			}
		})
	})
	Convey("Truncating all tables...", t, func() {
		for tn, mi := range Registry.registryByTableName {
			if mi.isMixin() || mi.isManual() {
				continue
			}
			dbExecuteNoTx(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, tn))
		}
	})
}
