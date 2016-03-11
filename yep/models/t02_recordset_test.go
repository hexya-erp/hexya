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
	"github.com/npiganeau/yep/yep/orm"
	"testing"
	"time"
)

var env Environment

func TestCreateRecordSet(t *testing.T) {
	env = NewEnvironment(dORM, 1)
	userJohn := User{
		UserName: "John Smith",
		Email:    "jsmith@example.com",
	}
	userJane := User{
		UserName: "Jane Smith",
		Email:    "jane.smith@example.com",
	}
	users := env.Pool("User").Create(&userJohn)
	throwFail(t, AssertIs(users.Ids(), []int64{1}))
	users2 := env.Pool("User").Create(&userJane)
	throwFail(t, AssertIs(users2.Ids(), []int64{2}))
}

func TestSearchRecordSet(t *testing.T) {
	users := env.Pool(new(User)).Filter("UserName", "Jane Smith").Search()
	throwFail(t, AssertIs(len(users.Ids()), 1))
	var userJane User
	users.ReadOne(&userJane)
	throwFail(t, AssertIs(userJane.UserName, "Jane Smith"))
	throwFail(t, AssertIs(userJane.Email, "jane.smith@example.com"))

	usersAll := env.Pool(new(User)).Search()
	var userStructs []*User_PartialWithPosts
	num := usersAll.ReadAll(&userStructs)
	throwFail(t, AssertIs(num, 2))
	throwFail(t, AssertIs(userStructs[0].Email, "jsmith@example.com"))
	throwFail(t, AssertIs(userStructs[1].Email, "jane.smith@example.com"))
}

func TestUpdateRecordSet(t *testing.T) {
	// Update multi
	users := env.Pool(new(User)).Filter("UserName", "Jane Smith").Search()
	throwFail(t, AssertIs(len(users.Ids()), 1))
	num := users.Write(orm.Params{"UserName": "Jane A. Smith"})
	throwFail(t, AssertIs(num, 1))
	var userJane User
	users.ReadOne(&userJane)
	throwFail(t, AssertIs(userJane.UserName, "Jane A. Smith"))
	throwFail(t, AssertIs(userJane.Email, "jane.smith@example.com"))

	// Update single (from different RecordSet
	type User_WithID struct {
		ID         int64
		UserName   string
		Email      string
		CreateDate time.Time //FIXME: Shouldn't be necessary. To be fixed by computed fields
	}
	var userJohn User_WithID
	env.Pool("User").Filter("UserName", "John Smith").ReadOne(&userJohn)
	userJohn.Email = "jsmith2@example.com"
	users.Write(&userJohn)
	var userJane2 User_WithID
	users.ReadOne(&userJane2)
	throwFail(t, AssertIs(userJane2.UserName, "Jane A. Smith"))
	throwFail(t, AssertIs(userJane2.Email, "jane.smith@example.com"))
	var userJohn2 User
	env.Pool("User").Filter("UserName", "John Smith").ReadOne(&userJohn2)
	throwFail(t, AssertIs(userJohn2.UserName, "John Smith"))
	throwFail(t, AssertIs(userJohn2.Email, "jsmith2@example.com"))
}

func TestDeleteRecordSet(t *testing.T) {
	users := env.Pool("User").Filter("UserName", "John Smith")
	num := users.Unlink()
	throwFail(t, AssertIs(num, 1))
}
