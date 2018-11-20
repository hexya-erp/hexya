// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"strings"

	"github.com/hexya-erp/hexya/src/tools/logging"
	"github.com/spf13/viper"
)

var (
	log logging.Logger
	// Langs is the list of all loaded languages in the application
	Langs []string
)

// BootStrap initializes available languages
func BootStrap() {
	Langs = viper.GetStringSlice("Server.Languages")
	for i, lang := range Langs {
		if strings.ToUpper(lang) == "ALL" {
			Langs = append(Langs[:i], append(GetAllLanguageList(), Langs[i+1:]...)...)
		}
	}
}

func init() {
	log = logging.GetLogger("i18n")
	Registry = NewTranslationsCollection()
}
