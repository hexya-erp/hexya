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
	"testing"

	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/models/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEnvironment(t *testing.T) {
	Convey("Testing Environment Modifications", t, func() {
		env := NewEnvironment(security.SuperUserID)
		env.context = types.NewContext().WithKey("key", "context value")
		userJane := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
		Convey("Checking WithEnv", func() {
			env2 := NewEnvironment(2)
			userJane1 := userJane.WithEnv(env2)
			So(userJane1.Env().Uid(), ShouldEqual, 2)
			So(userJane.Env().Uid(), ShouldEqual, 1)
			So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
			So(userJane1.Env().Context().IsEmpty(), ShouldBeTrue)
		})
		Convey("Checking WithContext", func() {
			userJane1 := userJane.WithContext("newKey", "This is a different key")
			So(userJane1.Env().Context().HasKey("key"), ShouldBeTrue)
			So(userJane1.Env().Context().HasKey("newKey"), ShouldBeTrue)
			So(userJane1.Env().Context().Get("key"), ShouldEqual, "context value")
			So(userJane1.Env().Context().Get("newKey"), ShouldEqual, "This is a different key")
			So(userJane1.Env().Uid(), ShouldEqual, security.SuperUserID)
			So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
			So(userJane.Env().Context().HasKey("newKey"), ShouldBeFalse)
			So(userJane.Env().Context().Get("key"), ShouldEqual, "context value")
			So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
		})
		Convey("Checking WithNewContext", func() {
			newCtx := types.NewContext().WithKey("newKey", "This is a different key")
			userJane1 := userJane.WithNewContext(newCtx)
			So(userJane1.Env().Context().HasKey("key"), ShouldBeFalse)
			So(userJane1.Env().Context().HasKey("newKey"), ShouldBeTrue)
			So(userJane1.Env().Context().Get("newKey"), ShouldEqual, "This is a different key")
			So(userJane1.Env().Uid(), ShouldEqual, security.SuperUserID)
			So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
			So(userJane.Env().Context().HasKey("newKey"), ShouldBeFalse)
			So(userJane.Env().Context().Get("key"), ShouldEqual, "context value")
			So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
		})
		Convey("Checking Sudo", func() {
			userJane1 := userJane.Sudo(2)
			userJane2 := userJane1.Sudo()
			So(userJane1.Env().Uid(), ShouldEqual, 2)
			So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
			So(userJane2.Env().Uid(), ShouldEqual, security.SuperUserID)
		})
		Convey("Checking combined modifications", func() {
			userJane1 := userJane.Sudo(2)
			userJane2 := userJane1.Sudo()
			userJane = userJane.WithContext("key", "modified value")
			So(userJane.Env().Context().Get("key"), ShouldEqual, "modified value")
			So(userJane1.Env().Context().Get("key"), ShouldEqual, "context value")
			So(userJane1.Env().Uid(), ShouldEqual, 2)
			So(userJane2.Env().Context().Get("key"), ShouldEqual, "context value")
			So(userJane2.Env().Uid(), ShouldEqual, security.SuperUserID)
		})
	})
}
