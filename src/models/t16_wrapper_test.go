// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"testing"

	"github.com/hexya-erp/hexya/src/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

type UserSet struct {
	*RecordCollection
}

type DummyStruct struct {
	value string
}

type UserData struct {
	*ModelData
}

func TestWrappers(t *testing.T) {
	Convey("Testing wrappers for RecordSets", t, func() {
		So(func() { RegisterRecordSetWrapper("User", UserSet{}) }, ShouldNotPanic)
		So(func() { RegisterRecordSetWrapper("Profile", int(8)) }, ShouldPanic)
		So(func() { RegisterRecordSetWrapper("Post", DummyStruct{}) }, ShouldPanic)
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			user := env.Pool("User")
			post := env.Pool("Post")
			Convey("Wrapping a User should work", func() {
				wUser := user.Wrap()
				So(wUser, ShouldHaveSameTypeAs, UserSet{})
			})
			Convey("Wrapping a Post should fail", func() {
				So(func() { post.Wrap() }, ShouldPanic)
			})
			Convey("Wrapping a Post as a user should work", func() {
				wUser := post.Wrap("User")
				So(wUser, ShouldHaveSameTypeAs, UserSet{})
			})
		}), ShouldBeNil)
	})
	Convey("Testing wrappers for ModelData", t, func() {
		So(func() { RegisterModelDataWrapper("User", UserData{}) }, ShouldNotPanic)
		So(func() { RegisterModelDataWrapper("Profile", int(8)) }, ShouldPanic)
		So(func() { RegisterModelDataWrapper("Post", DummyStruct{}) }, ShouldPanic)
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userData := NewModelData(Registry.MustGet("User"), FieldMap{"Email": "myuser@example.com"})
			postData := NewModelData(Registry.MustGet("Post"), FieldMap{"Title": "My Post"})
			Convey("Wrapping a user data should work", func() {
				wUserData := userData.Wrap()
				So(wUserData, ShouldHaveSameTypeAs, new(UserData))
			})
			Convey("Wrapping a Post should fail", func() {
				pUserData := postData.Wrap()
				So(pUserData, ShouldHaveSameTypeAs, new(ModelData))
			})
		}), ShouldBeNil)
	})
}
