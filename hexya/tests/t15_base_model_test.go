// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBaseModelMethods(t *testing.T) {
	Convey("Testing base model methods", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
			Convey("Copy", func() {
				userJane.Write(&h.UserData{Password: "Jane's Password"})
				userJaneCopy := userJane.Copy(&h.UserData{Name: "Jane's Copy", Email2: "js@example.com"})
				So(userJaneCopy.Name(), ShouldEqual, "Jane's Copy")
				So(userJaneCopy.Email(), ShouldEqual, "jane.smith@example.com")
				So(userJaneCopy.Email2(), ShouldEqual, "js@example.com")
				So(userJaneCopy.Password(), ShouldBeBlank)
				So(userJaneCopy.Age(), ShouldEqual, 24)
				So(userJaneCopy.Nums(), ShouldEqual, 2)
				So(userJaneCopy.Posts().Len(), ShouldEqual, 0)
			})
		})
	})
}
