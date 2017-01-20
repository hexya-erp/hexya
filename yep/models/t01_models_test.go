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
)

func TestCreateDB(t *testing.T) {
	Convey("Creating DataBase...", t, func() {
		user := NewModel("User", new(struct {
			ID            int64
			UserName      string `yep:"unique;string(Name);help(The user's username)"`
			DecoratedName string `yep:"compute(computeDecoratedName)"`
			Email         string `yep:"size(100);help(The user's email address);index"`
			Password      string
			Status        int16 `yep:"json(status_json)"`
			IsStaff       bool
			IsActive      bool
			Profile       RecordCollection `yep:"type(many2one);comodel(Profile)"` //;on_delete(set_null)"`
			Age           int16            `yep:"compute(computeAge);store;depends(Profile.Age,Profile)"`
			Posts         RecordCollection `yep:"type(one2many);fk(User);comodel(Post)"`
			Nums          int
			unexportBool  bool
			PMoney        float64          `yep:"related(Profile.Money)"`
			LastPost      RecordCollection `yep:"embed;type(many2one);comodel(Post)"`
		}))

		user.Extend(new(struct {
			Email2    string
			IsPremium bool
		}))

		addressMI := NewMixinModel("AddressMixIn", new(struct {
			Street string
			Zip    string
			City   string
		}))

		profile := NewModel("Profile", new(struct {
			Age      int16
			Money    float64
			User     RecordCollection `yep:"type(many2one);comodel(User)"`
			BestPost RecordCollection `yep:"type(one2one);comodel(Post)"`
		}))

		profile.MixInModel(addressMI)
		profile.Extend(new(struct {
			City    string
			Country string
		}))

		NewModel("Post", new(struct {
			User    RecordCollection `yep:"type(many2one);comodel(User)"`
			Title   string
			Content string           `yep:"type(text)"`
			Tags    RecordCollection `yep:"type(many2many);comodel(Tag)"`
		}))

		activeMI := NewMixinModel("ActiveMixIn", new(struct {
			Active bool
		}))
		MixInAllModels("ActiveMixIn")

		tag := NewModel("Tag", new(struct {
			Name     string
			BestPost RecordCollection `yep:"type(many2one);comodel(Post)"`
			Posts    RecordCollection `yep:"type(many2many);comodel(Post)"`
		}))

		tag.Extend(new(struct {
			Description string
		}))

		user.CreateMethod("PrefixedUser", "",
			func(rc RecordCollection, prefix string) []string {
				var res []string
				for _, u := range rc.Records() {
					res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get("UserName")))
				}
				return res
			})

		user.ExtendMethod("PrefixedUser", "",
			func(rc RecordCollection, prefix string) []string {
				res := rc.Super(prefix).([]string)
				for i, u := range rc.Records() {
					email := u.Get("Email").(string)
					res[i] = fmt.Sprintf("%s %s", res[i], rc.Call("DecorateEmail", email))
				}
				return res
			})

		user.CreateMethod("DecorateEmail", "",
			func(rc RecordCollection, email string) string {
				return fmt.Sprintf("<%s>", email)
			})

		user.ExtendMethod("DecorateEmail", "",
			func(rc RecordCollection, email string) string {
				res := rc.Super(email).(string)
				return fmt.Sprintf("[%s]", res)
			})

		user.CreateMethod("computeDecoratedName", "",
			func(rc RecordCollection) FieldMap {
				res := make(FieldMap)
				res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
				return res
			})

		user.CreateMethod("computeAge", "",
			func(rc RecordCollection) (FieldMap, []FieldName) {
				res := make(FieldMap)
				res["Age"] = rc.Get("Profile").(RecordCollection).Get("Age").(int16)
				return res, []FieldName{}
			})

		activeMI.CreateMethod("IsActivated", "",
			func(rc RecordCollection) bool {
				return rc.Get("Active").(bool)
			})

		addressMI.CreateMethod("SayHello", "",
			func(rc RecordCollection) string {
				return "Hello !"
			})

		addressMI.CreateMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				return fmt.Sprintf("%s, %s %s", rc.Get("Street"), rc.Get("Zip"), rc.Get("City"))
			})

		profile.CreateMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				res := rc.Super().(string)
				return fmt.Sprintf("%s, %s", res, rc.Get("Country"))
			})

		addressMI.ExtendMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				res := rc.Super().(string)
				return fmt.Sprintf("<%s>", res)
			})

		profile.ExtendMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				res := rc.Super().(string)
				return fmt.Sprintf("[%s]", res)
			})

		// Creating a dummy table to check that it is correctly removed by Bootstrap
		db.MustExec("CREATE TABLE IF NOT EXISTS shouldbedeleted (id serial NOT NULL PRIMARY KEY)")
	})

	Convey("Database creation should run fine", t, func() {
		Convey("Dummy table should exist", func() {
			So(testAdapter.tables(), ShouldContainKey, "shouldbedeleted")
		})
		Convey("Bootstrap should not panic", func() {
			So(BootStrap, ShouldNotPanic)
		})
		Convey("All models should have a DB table", func() {
			dbTables := testAdapter.tables()
			for tableName, mi := range Registry.registryByTableName {
				if mi.isMixin() {
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
			if mi.isMixin() {
				continue
			}
			dbExecuteNoTx(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, tn))
		}
	})
}
