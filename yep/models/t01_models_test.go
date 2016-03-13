/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package models

import (
	"fmt"
	"github.com/npiganeau/yep/yep/orm"
	"testing"
)

func TestSyncDb(t *testing.T) {
	CreateModel("User")
	ExtendModel(new(User), new(User_Extension))
	CreateModel("Profile")
	ExtendModel(new(Profile), new(Profile_Extension))
	CreateModel("Post")
	ExtendModel(new(Post))
	CreateModel("Tag")
	ExtendModel(new(Tag), new(Tag_Extension))

	DeclareMethod("User", "PrefixedUser", func(rs RecordSet, prefix string) []string {
		var res []string
		type User_Simple struct {
			UserName string
		}
		var users []*User_Simple
		rs.ReadAll(&users)
		for _, u := range users {
			res = append(res, fmt.Sprintf("%s: %s", prefix, u.UserName))
		}
		return res
	})

	DeclareMethod("User", "PrefixedUser", func(rs RecordSet, prefix string) []string {
		res := rs.Super(prefix).([]string)
		type User_Email struct {
			Email string
		}
		var users []*User_Email
		rs.ReadAll(&users)
		for i, u := range users {
			res[i] = fmt.Sprintf("%s <%s>", res[i], u.Email)
		}
		return res
	})

	err := orm.RunSyncdb("default", true, orm.Debug)
	throwFail(t, err)

	dORM = orm.NewOrm()
}

type User struct {
	UserName     string `orm:"size(30);unique;string(Name);help(The user's username)"`
	Email        string `orm:"size(100)"`
	Password     string `orm:"size(100)"`
	Status       int16  `orm:"column(Status)"`
	IsStaff      bool
	IsActive     bool     `orm:"default(true)"`
	Profile      *Profile `orm:"null;rel(one);on_delete(set_null)"`
	Posts        []*Post  `orm:"reverse(many)" json:"-"`
	ShouldSkip   string   `orm:"-"`
	Nums         int
	unexport     bool `orm:"-"`
	unexportBool bool
}

func (u *User) TableIndex() [][]string {
	return [][]string{
		{"Id", "UserName"},
		{"Id", "Created"},
	}
}

func (u *User) TableUnique() [][]string {
	return [][]string{
		{"UserName", "Email"},
	}
}

type User_PartialWithPosts struct {
	Email     string
	Email2    string
	IsPremium bool
	Profile   *Profile_PartialWithBestPost
	Posts     []*Post
}

type Profile struct {
	Age      int16
	Money    float64
	User     *User `orm:"reverse(one)" json:"-"`
	BestPost *Post `orm:"rel(one);null"`
}

type Profile_PartialWithBestPost struct {
	Age      int16
	Country  string
	BestPost *Post
}

type Post struct {
	User    *User  `orm:"rel(fk)"`
	Title   string `orm:"size(60)"`
	Content string `orm:"type(text)"`
	Tags    []*Tag `orm:"rel(m2m)"`
}

func (u *Post) TableIndex() [][]string {
	return [][]string{
		{"Id", "Created"},
	}
}

type Tag struct {
	Name     string  `orm:"size(30)"`
	BestPost *Post   `orm:"rel(one);null"`
	Posts    []*Post `orm:"reverse(many)" json:"-"`
}

type User_Extension struct {
	Email2    string `orm:"size(100)"`
	IsPremium bool
}

type Profile_Extension struct {
	City    string
	Country string
}

type Tag_Extension struct {
	Description string
}
