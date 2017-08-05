// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/spf13/viper"
)

var (
	log *logging.Logger
	// Langs is the list of all loaded languages in the application
	Langs []string
)

// BootStrap initializes available languages
func BootStrap() {
	Langs = viper.GetStringSlice("Server.Languages")
}

func init() {
	log = logging.GetLogger("i18n")
	Registry = NewTranslationsCollection()
}
