// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/types"
)

func checkTranslation(t *testing.T) {
	assert.Len(t, Registry.fieldSelection, 2)
	assert.Contains(t, Registry.fieldSelection, selectionRef{lang: "fr", model: "Profile", field: "State", source: "Active"})
	assert.EqualValues(t, Registry.fieldSelection[selectionRef{lang: "fr", model: "Profile", field: "State", source: "Active"}], "Actif")
	assert.Contains(t, Registry.fieldSelection, selectionRef{lang: "fr", model: "Profile", field: "State", source: "Inactive"})
	assert.EqualValues(t, Registry.fieldSelection[selectionRef{lang: "fr", model: "Profile", field: "State", source: "Inactive"}], "Inactif")
	assert.Contains(t, Registry.fieldDescription, fieldRef{lang: "fr", model: "User", field: "Active"})
	assert.EqualValues(t, Registry.fieldDescription[fieldRef{lang: "fr", model: "User", field: "Active"}], "Actif")
	assert.Contains(t, Registry.fieldHelp, fieldRef{lang: "fr", model: "User", field: "Active"})
	assert.EqualValues(t, Registry.fieldHelp[fieldRef{lang: "fr", model: "User", field: "Active"}], "Lorsqu'il est inactif,\nun utilisateur ne sera pas autorisé à se connecter")
	assert.Contains(t, Registry.resource, resourceRef{lang: "fr", id: "user_view_id", source: "Profile Data"})
	assert.EqualValues(t, Registry.resource[resourceRef{lang: "fr", id: "user_view_id", source: "Profile Data"}], "Données du profil")
	assert.Contains(t, Registry.code, codeRef{lang: "fr", context: "base", source: "You are not allowed to perform this operation"})
	assert.EqualValues(t, Registry.code[codeRef{lang: "fr", context: "base", source: "You are not allowed to perform this operation"}], "Vous n'êtes pas autorisé à faire cette opération")
	assert.EqualValues(t, Registry.custom[customRef{lang: "fr", id: "Create", module: "testModule"}], "Créer")
	assert.EqualValues(t, Registry.custom[customRef{lang: "fr", id: "Warning", module: "testModule"}], "Attention")
}

func TestI18N(t *testing.T) {
	t.Run("Testing translation framework", func(t *testing.T) {
		t.Run("Loading translation from file", func(t *testing.T) {
			assert.NotPanics(t, func() { LoadPOFile("testdata/fr.po") })
			checkTranslation(t)
		})
		t.Run("Loading a second time the same file should not change anything", func(t *testing.T) {
			LoadPOFile("testdata/fr.po")
			checkTranslation(t)
		})
		t.Run("Translating field description should work", func(t *testing.T) {
			trans := TranslateFieldDescription("fr", "User", "Active", "")
			assert.EqualValues(t, trans, "Actif")
			trans = TranslateFieldDescription("de", "User", "Active", "Active")
			assert.EqualValues(t, trans, "Active")
			trans = TranslateFieldDescription("fr", "User", "Login", "Login")
			assert.EqualValues(t, trans, "Login")
		})
		t.Run("Translating field help should work", func(t *testing.T) {
			trans := TranslateFieldHelp("fr", "User", "Active", "")
			assert.EqualValues(t, trans, "Lorsqu'il est inactif,\nun utilisateur ne sera pas autorisé à se connecter")
			trans = TranslateFieldHelp("de", "User", "Active", "defaultValue")
			assert.EqualValues(t, trans, "defaultValue")
			trans = TranslateFieldHelp("fr", "User", "Login", "defaultHelp")
			assert.EqualValues(t, trans, "defaultHelp")
		})
		t.Run("Translating field selection should work", func(t *testing.T) {
			trans := TranslateFieldSelection("fr", "Profile", "State", types.Selection{"active": "Active", "inactive": "Inactive"})
			assert.Len(t, trans, 2)
			assert.EqualValues(t, trans["active"], "Actif")
			assert.EqualValues(t, trans["inactive"], "Inactif")
			trans = TranslateFieldSelection("de", "Profile", "State", types.Selection{"active": "Active", "inactive": "Inactive"})
			assert.Len(t, trans, 2)
			assert.EqualValues(t, trans["active"], "Active")
			assert.EqualValues(t, trans["inactive"], "Inactive")
			trans = TranslateFieldSelection("fr", "Profile", "State", types.Selection{"active": "Active", "inactive": "Unknown"})
			assert.Len(t, trans, 2)
			assert.EqualValues(t, trans["active"], "Actif")
			assert.EqualValues(t, trans["inactive"], "Unknown")
		})
		t.Run("Translating views should work", func(t *testing.T) {
			trans := TranslateResourceItem("fr", "user_view_id", "Profile Data")
			assert.EqualValues(t, trans, "Données du profil")
			trans = TranslateResourceItem("de", "user_view_id", "Profile Data")
			assert.EqualValues(t, trans, "Profile Data")
			trans = TranslateResourceItem("fr", "user_view2_id", "Profile Data")
			assert.EqualValues(t, trans, "Profile Data")
		})
		t.Run("Translating code should work", func(t *testing.T) {
			trans := TranslateCode("fr", "base", "You are not allowed to perform this operation")
			assert.EqualValues(t, trans, "Vous n'êtes pas autorisé à faire cette opération")
			trans = TranslateCode("de", "base", "You are not allowed to perform this operation")
			assert.EqualValues(t, trans, "You are not allowed to perform this operation")
			trans = TranslateCode("fr", "stock", "You are not allowed to perform this operation")
			assert.EqualValues(t, trans, "You are not allowed to perform this operation")
		})
		t.Run("Translating custom should work", func(t *testing.T) {
			trans := TranslateCustom("fr", "Create", "testModule")
			assert.EqualValues(t, trans, "Créer")
			trans = TranslateCustom("de", "Create", "testModule")
			assert.EqualValues(t, trans, "Create")
			trans = TranslateCustom("fr", "Stock", "testModule")
			assert.EqualValues(t, trans, "Stock")
		})
		t.Run("Testing translation overrides", func(t *testing.T) {
			LoadPOFile("testdata/fr-override.po")
			trans := TranslateFieldSelection("fr", "Profile", "State", types.Selection{"active": "Active", "inactive": "Inactive"})
			assert.Len(t, trans, 2)
			assert.EqualValues(t, trans["active"], "Activé")
			assert.EqualValues(t, trans["inactive"], "Inactif")
			transField := TranslateFieldDescription("fr", "User", "Active", "")
			assert.EqualValues(t, transField, "Actif")
			transView := TranslateResourceItem("fr", "user_view_id", "Profile Data")
			assert.EqualValues(t, transView, "Données du profil")
		})
		t.Run("Testing invalid PO files", func(t *testing.T) {
			assert.Panics(t, func() { LoadPOFile("testdata/invalid-po.txt") })
			assert.Panics(t, func() { LoadPOFile("testdata/no-lang.po") })
			assert.Panics(t, func() { LoadPOFile("testdata/invalid-field.po") })
			assert.Panics(t, func() { LoadPOFile("testdata/invalid-help.po") })
			assert.Panics(t, func() { LoadPOFile("testdata/invalid-selection.po") })
			assert.NotPanics(t, func() { LoadPOFile("testdata/invalid-comment.po") })
		})
	})
}

func TestLanguagesData(t *testing.T) {
	t.Run("Testing languages data", func(t *testing.T) {
		t.Run("Registering and overriding locales", func(t *testing.T) {
			assert.Len(t, locales, 78)
			assert.Contains(t, locales, "nl")
			assert.NotContains(t, locales, "wz")
			all := GetAllLanguageList()
			assert.Len(t, all, 78)
			assert.NotContains(t, all, "wz")
			assert.Nil(t, RegisterLocale(&Locale{
				Name:      "New Locale",
				ISOCode:   "wz",
				Direction: LangDirectionLTR,
			}))
			assert.Nil(t, OverrideLocale(&Locale{
				Name:      "New Dutch",
				ISOCode:   "nl",
				Direction: LangDirectionLTR,
			}))
			assert.Len(t, locales, 79)
			assert.Contains(t, locales, "nl")
			assert.Contains(t, locales, "wz")
			all = GetAllLanguageList()
			assert.Len(t, all, 79)
			assert.Contains(t, all, "wz")
			os.Remove("testdata/server/i18n/testModule")
		})
		t.Run("Registering/overriding invalid locale should fail", func(t *testing.T) {
			assert.NotNil(t, RegisterLocale(&Locale{}))
			assert.NotNil(t, OverrideLocale(&Locale{}))
			assert.NotNil(t, RegisterLocale(&Locale{
				ISOCode: "wz",
			}))
			assert.NotNil(t, RegisterLocale(&Locale{
				Name:    "New Locale",
				ISOCode: "wz",
			}))
		})
		t.Run("Registering existing locale should fail", func(t *testing.T) {
			assert.NotNil(t, RegisterLocale(&Locale{
				Name:      "New Locale",
				ISOCode:   "wz",
				Direction: LangDirectionLTR,
			}))
		})
		t.Run("Overriding non existing locale should fail", func(t *testing.T) {
			assert.NotNil(t, OverrideLocale(&Locale{
				Name:      "New Locale",
				ISOCode:   "zz",
				Direction: LangDirectionLTR,
			}))
		})
		t.Run("Checking existing language data", func(t *testing.T) {
			frParams := GetLocale("fr")
			assert.NotNil(t, frParams)
			assert.EqualValues(t, frParams.Name, "French / Français")
		})
		t.Run("Checking non-existing language data", func(t *testing.T) {
			frParams := GetLocale("noexists")
			assert.NotNil(t, frParams)
			assert.EqualValues(t, frParams.Name, "UNKNOWN_LOCALE (noexists)")
		})
	})
}

func TestCustomTranslations(t *testing.T) {
	t.Run("Testing retrieving custom translations", func(t *testing.T) {
		t.Run("Listing all custom translations", func(t *testing.T) {
			tr := GetAllCustomTranslations()
			assert.Len(t, tr, 1)
			assert.Contains(t, tr, "fr")
			assert.Len(t, tr["fr"], 1)
			assert.Contains(t, tr["fr"], "testModule")
			assert.Len(t, tr["fr"]["testModule"], 2)
		})
	})
}
