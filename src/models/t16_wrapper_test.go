// Copyright 2019 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/security"
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
	t.Run("Testing wrappers for RecordSets", func(t *testing.T) {
		assert.NotPanics(t, func() { RegisterRecordSetWrapper("User", UserSet{}) })
		assert.Panics(t, func() { RegisterRecordSetWrapper("Profile", int(8)) })
		assert.Panics(t, func() { RegisterRecordSetWrapper("Post", DummyStruct{}) })
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			user := env.Pool("User")
			post := env.Pool("Post")
			t.Run("Wrapping a User should work", func(t *testing.T) {
				wUser := user.Wrap()
				assert.IsType(t, UserSet{}, wUser)
			})
			t.Run("Wrapping a Post should fail", func(t *testing.T) {
				assert.Panics(t, func() { post.Wrap() })
			})
			t.Run("Wrapping a Post as a user should work", func(t *testing.T) {
				wUser := post.Wrap("User")
				assert.IsType(t, UserSet{}, wUser)
			})
		}))
	})
	t.Run("Testing wrappers for ModelData", func(t *testing.T) {
		assert.NotPanics(t, func() { RegisterModelDataWrapper("User", UserData{}) })
		assert.Panics(t, func() { RegisterModelDataWrapper("Profile", int(8)) })
		assert.Panics(t, func() { RegisterModelDataWrapper("Post", DummyStruct{}) })
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userData := NewModelData(Registry.MustGet("User"), FieldMap{"Email": "myuser@example.com"})
			postData := NewModelData(Registry.MustGet("Post"), FieldMap{"Title": "My Post"})
			t.Run("Wrapping a user data should work", func(t *testing.T) {
				wUserData := userData.Wrap()
				assert.IsType(t, new(UserData), wUserData)
			})
			t.Run("Wrapping a Post should fail", func(t *testing.T) {
				pUserData := postData.Wrap()
				assert.IsType(t, new(ModelData), pUserData)
			})
		}))
	})
}
