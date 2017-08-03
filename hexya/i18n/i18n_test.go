// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/models/types"
	. "github.com/smartystreets/goconvey/convey"
)

func checkTranslation() {
	So(Registry.fieldSelection, ShouldHaveLength, 2)
	So(Registry.fieldSelection, ShouldContainKey, selectionRef{lang: "fr", model: "Profile", field: "State", source: "Active"})
	So(Registry.fieldSelection[selectionRef{lang: "fr", model: "Profile", field: "State", source: "Active"}], ShouldEqual, "Actif")
	So(Registry.fieldSelection, ShouldContainKey, selectionRef{lang: "fr", model: "Profile", field: "State", source: "Inactive"})
	So(Registry.fieldSelection[selectionRef{lang: "fr", model: "Profile", field: "State", source: "Inactive"}], ShouldEqual, "Inactif")
	So(Registry.fieldDescription, ShouldContainKey, fieldRef{lang: "fr", model: "User", field: "Active"})
	So(Registry.fieldDescription[fieldRef{lang: "fr", model: "User", field: "Active"}], ShouldEqual, "Actif")
	So(Registry.fieldHelp, ShouldContainKey, fieldRef{lang: "fr", model: "User", field: "Active"})
	So(Registry.fieldHelp[fieldRef{lang: "fr", model: "User", field: "Active"}], ShouldEqual, "Lorsqu'il est inactif,\nun utilisateur ne sera pas autorisé à se connecter")
	So(Registry.resource, ShouldContainKey, resourceRef{lang: "fr", viewID: "user_view_id", source: "Profile Data"})
	So(Registry.resource[resourceRef{lang: "fr", viewID: "user_view_id", source: "Profile Data"}], ShouldEqual, "Données du profil")
	So(Registry.code, ShouldContainKey, codeRef{lang: "fr", context: "base", source: "You are not allowed to perform this operation"})
	So(Registry.code[codeRef{lang: "fr", context: "base", source: "You are not allowed to perform this operation"}], ShouldEqual, "Vous n'êtes pas autorisé à faire cette opération")
}

func TestI18N(t *testing.T) {
	Convey("Testing translation framework", t, func() {
		Convey("Loading translation from file", func() {
			So(func() { LoadPOFile("testdata/fr.po") }, ShouldNotPanic)
			checkTranslation()
		})
		Convey("Loading a second time the same file should not change anything", func() {
			LoadPOFile("testdata/fr.po")
			checkTranslation()
		})
		Convey("Translating field description should work", func() {
			trans := TranslateFieldDescription("fr", "User", "Active", "")
			So(trans, ShouldEqual, "Actif")
			trans = TranslateFieldDescription("de", "User", "Active", "Active")
			So(trans, ShouldEqual, "Active")
			trans = TranslateFieldDescription("fr", "User", "Login", "Login")
			So(trans, ShouldEqual, "Login")
		})
		Convey("Translating field help should work", func() {
			trans := TranslateFieldHelp("fr", "User", "Active", "")
			So(trans, ShouldEqual, "Lorsqu'il est inactif,\nun utilisateur ne sera pas autorisé à se connecter")
			trans = TranslateFieldHelp("de", "User", "Active", "defaultValue")
			So(trans, ShouldEqual, "defaultValue")
			trans = TranslateFieldHelp("fr", "User", "Login", "defaultHelp")
			So(trans, ShouldEqual, "defaultHelp")
		})
		Convey("Translating field selection should work", func() {
			trans := TranslateFieldSelection("fr", "Profile", "State", types.Selection{"active": "Active", "inactive": "Inactive"})
			So(trans, ShouldHaveLength, 2)
			So(trans["active"], ShouldEqual, "Actif")
			So(trans["inactive"], ShouldEqual, "Inactif")
			trans = TranslateFieldSelection("de", "Profile", "State", types.Selection{"active": "Active", "inactive": "Inactive"})
			So(trans, ShouldHaveLength, 2)
			So(trans["active"], ShouldEqual, "Active")
			So(trans["inactive"], ShouldEqual, "Inactive")
			trans = TranslateFieldSelection("fr", "Profile", "State", types.Selection{"active": "Active", "inactive": "Unknown"})
			So(trans, ShouldHaveLength, 2)
			So(trans["active"], ShouldEqual, "Actif")
			So(trans["inactive"], ShouldEqual, "Unknown")
		})
		Convey("Translating views should work", func() {
			trans := TranslateResourceItem("fr", "user_view_id", "Profile Data")
			So(trans, ShouldEqual, "Données du profil")
			trans = TranslateResourceItem("de", "user_view_id", "Profile Data")
			So(trans, ShouldEqual, "Profile Data")
			trans = TranslateResourceItem("fr", "user_view2_id", "Profile Data")
			So(trans, ShouldEqual, "Profile Data")
		})
		Convey("Translating code should work", func() {
			trans := TranslateCode("fr", "base", "You are not allowed to perform this operation")
			So(trans, ShouldEqual, "Vous n'êtes pas autorisé à faire cette opération")
			trans = TranslateCode("de", "base", "You are not allowed to perform this operation")
			So(trans, ShouldEqual, "You are not allowed to perform this operation")
			trans = TranslateCode("fr", "stock", "You are not allowed to perform this operation")
			So(trans, ShouldEqual, "You are not allowed to perform this operation")
		})
		Convey("Testing translation overrides", func() {
			LoadPOFile("testdata/fr-override.po")
			trans := TranslateFieldSelection("fr", "Profile", "State", types.Selection{"active": "Active", "inactive": "Inactive"})
			So(trans, ShouldHaveLength, 2)
			So(trans["active"], ShouldEqual, "Activé")
			So(trans["inactive"], ShouldEqual, "Inactif")
			transField := TranslateFieldDescription("fr", "User", "Active", "")
			So(transField, ShouldEqual, "Actif")
			transView := TranslateResourceItem("fr", "user_view_id", "Profile Data")
			So(transView, ShouldEqual, "Données du profil")
		})
		Convey("Testing invalid PO files", func() {
			So(func() { LoadPOFile("testdata/invalid-po.txt") }, ShouldPanic)
			So(func() { LoadPOFile("testdata/no-lang.po") }, ShouldPanic)
			So(func() { LoadPOFile("testdata/invalid-field.po") }, ShouldPanic)
			So(func() { LoadPOFile("testdata/invalid-help.po") }, ShouldPanic)
			So(func() { LoadPOFile("testdata/invalid-selection.po") }, ShouldPanic)
		})
	})
}
