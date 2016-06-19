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

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func PrefixUser(rs RecordSet, prefix string) []string {
	var res []string
	type User_Simple struct {
		ID       int64
		UserName string
	}
	var users []*User_Simple
	rs.ReadAll(&users)
	for _, u := range users {
		res = append(res, fmt.Sprintf("%s: %s", prefix, u.UserName))
	}
	return res
}

func PrefixUserEmailExtension(rs RecordSet, prefix string) []string {

	res := rs.Super(prefix).([]string)
	type User_Email struct {
		ID    int64
		Email string
	}
	var users []*User_Email
	rs.ReadAll(&users)
	for i, u := range users {
		res[i] = fmt.Sprintf("%s %s", res[i], rs.Call("DecorateEmail", u.Email))
	}
	return res
}

func DecorateEmail(rs RecordSet, email string) string {
	return fmt.Sprintf("<%s>", email)
}

func DecorateEmailExtension(rs RecordSet, email string) string {
	res := rs.Super(email).(string)
	return fmt.Sprintf("[%s]", res)
}

func computeDecoratedName(rs RecordSet) FieldMap {
	res := make(FieldMap)
	res["DecoratedName"] = rs.Call("PrefixedUser", "User").([]string)[0]
	return res
}

func computeAge(rs RecordSet) FieldMap {
	res := make(FieldMap)
	type Profile_Simple struct {
		ID  int64
		Age int16
	}
	type User_Simple struct {
		ID      int64
		Profile *Profile_Simple
	}
	user := new(User_Simple)
	//rs.RelatedSel("Profile").ReadOne(user)
	if user.Profile != nil {
		res["Age"] = user.Profile.Age
	}
	return res
}

func TestCreateDB(t *testing.T) {
	Convey("Creating DataBase...", t, func() {
		CreateModel("User")
		ExtendModel("User", new(User), new(User_Extension))
		CreateModel("Profile")
		ExtendModel("Profile", new(Profile), new(Profile_Extension))
		CreateModel("Post")
		ExtendModel("Post", new(Post))
		CreateModel("Tag")
		ExtendModel("Tag", new(Tag), new(Tag_Extension))

		DeclareMethod("User", "PrefixedUser", PrefixUser)
		DeclareMethod("User", "PrefixedUser", PrefixUserEmailExtension)
		DeclareMethod("User", "DecorateEmail", DecorateEmail)
		DeclareMethod("User", "DecorateEmail", DecorateEmailExtension)
		DeclareMethod("User", "computeDecoratedName", computeDecoratedName)
		DeclareMethod("User", "computeAge", computeAge)

		Convey("Dummy table should exist", func() {
			So(testAdapter.tables(), ShouldContainKey, "shouldbedeleted")
		})
		Convey("Bootstrap should not panic", func() {
			So(BootStrap, ShouldNotPanic)
		})
		Convey("All models should have a DB table", func() {
			dbTables := testAdapter.tables()
			for tableName, _ := range modelRegistry.registryByTableName {
				So(dbTables[tableName], ShouldBeTrue)
			}
		})
		Convey("All DB tables should have a model", func() {
			for dbTable, _ := range testAdapter.tables() {
				So(modelRegistry.registryByTableName, ShouldContainKey, dbTable)
			}
		})
	})
}

type User struct {
	ID            int64
	UserName      string `yep:"unique;string(Name);help(The user's username)"`
	DecoratedName string `yep:"compute(computeDecoratedName)"`
	Email         string `yep:"size(100);help(The user's email address)"`
	Password      string
	Status        int16 `yep:"json(status_json)"`
	IsStaff       bool
	IsActive      bool
	Profile       *Profile `yep:"null;type(many2one)"` //;on_delete(set_null)"`
	Age           int16    `yep:"compute(computeAge);store;depends(Profile.Age,Profile)"`
	Posts         []*Post  `yep:"type(one2many)"`
	Nums          int
	unexportBool  bool
}

func (u *User) TableIndex() [][]string {
	return [][]string{
		{"Id", "UserName"},
		{"Id", "Email"},
	}
}

func (u *User) TableUnique() [][]string {
	return [][]string{
		{"UserName", "Email"},
	}
}

type User_PartialWithPosts struct {
	ID        int64
	Email     string
	Email2    string
	IsPremium bool
	Profile   *Profile_PartialWithBestPost
	Posts     []*Post
}

type Profile struct {
	Age      int16
	Money    float64
	User     *User
	BestPost *Post `yep:"type(one2one)"`
}

type Profile_PartialWithBestPost struct {
	ID       int64
	Age      int16
	Country  string
	BestPost *Post
}

type Post struct {
	User    *User
	Title   string
	Content string `yep:"type(text)"`
	Tags    []*Tag `yep:"type(many2many)"`
}

func (u *Post) TableIndex() [][]string {
	return [][]string{
		{"Id", "Title"},
	}
}

type Tag struct {
	Name     string
	BestPost *Post
	Posts    []*Post `yep:"type(many2many)"`
}

type User_Extension struct {
	Email2    string
	IsPremium bool
}

type Profile_Extension struct {
	City    string
	Country string
}

type Tag_Extension struct {
	Description string
}
