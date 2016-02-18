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

package orm

import (
	"testing"
	"time"
)

func TestFullUser(t *testing.T) {
	// Full user struct
	type User struct {
		ID           int
		UserName     string
		Email        string
		Password     string
		Status       int16
		IsStaff      bool
		IsActive     bool
		Created      time.Time
		Updated      time.Time
		Profile      *Profile
		Posts        []*Post
		ShouldSkip   string
		Nums         int
		Langs        SliceStringField
		Extra        JSONField
		unexport     bool
		unexportBool bool
		Email2       string
		IsPremium    bool
	}

	// Insert full users
	user1 := User{
		UserName: "John Smith",
		Email:    "jsmith@example.com",
		Status:   132,
		IsActive: true,
	}
	id1, err := dORM.Insert(&user1)
	throwFail(t, err)

	user2 := User{
		UserName: "Jane Smith",
		Email:    "j2smith@example.com",
		Email2:   "jane.smith@example.net",
		Status:   12,
		IsActive: true,
	}
	id2, err := dORM.Insert(&user2)
	throwFail(t, err)

	// Query full user
	user := User{ID: int(id1)}
	err = dORM.Read(&user)
	throwFail(t, err)
	throwFail(t, AssertIs(user.ID, id1))

	// Query through QuerySet with full user
	var user3 User
	qs := dORM.QueryTable("user")
	err = qs.Filter("email2", "jane.smith@example.net").One(&user3)
	throwFail(t, err)
	throwFail(t, AssertIs(user3.ID, id2))
}
