// Copyright 2017 NDP Systèmes. All Rights Reserved.
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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// YEPCmd is the base 'yep' command of the commander
var YEPCmd = &cobra.Command{
	Use:   "yep",
	Short: "YEP is an open source modular ERP",
	Long: `YEP is an open source modular ERP written in Go.
It is designed for high demand business data processing while being easily customizable`,
}

func init() {
	YEPCmd.PersistentFlags().StringP("config", "c", "", "Alternate configuration file to read. Defaults to $HOME/.yep/")
	viper.BindPFlag("ConfigFileName", YEPCmd.PersistentFlags().Lookup("config"))

	YEPCmd.PersistentFlags().StringP("log-level", "L", "info", "Log level. Should be one of 'debug', 'info', 'warn', 'error' or 'crit'")
	viper.BindPFlag("LogLevel", YEPCmd.PersistentFlags().Lookup("log-level"))
	YEPCmd.PersistentFlags().StringP("log-file", "l", "", "File to which the log will be written")
	viper.BindPFlag("LogFile", YEPCmd.PersistentFlags().Lookup("log-file"))
	YEPCmd.PersistentFlags().BoolP("log-stdout", "o", false, "Enable stdout logging. Use for development or debugging.")
	viper.BindPFlag("LogStdout", YEPCmd.PersistentFlags().Lookup("log-stdout"))
	YEPCmd.PersistentFlags().Bool("debug", false, "Enable server debug mode for development")
	viper.BindPFlag("Debug", YEPCmd.PersistentFlags().Lookup("debug"))

	YEPCmd.PersistentFlags().String("db-driver", "postgres", "Database driver to use")
	viper.BindPFlag("DB.Driver", YEPCmd.PersistentFlags().Lookup("db-driver"))
	YEPCmd.PersistentFlags().String("db-host", "", "Database hostname or IP. Leave empty to connect through socket.")
	viper.BindPFlag("DB.Host", YEPCmd.PersistentFlags().Lookup("db-host"))
	YEPCmd.PersistentFlags().String("db-port", "5432", "Database port. Value is ignored if db-host is not set.")
	viper.BindPFlag("DB.Port", YEPCmd.PersistentFlags().Lookup("db-port"))
	YEPCmd.PersistentFlags().String("db-user", "", "Database user. Defaults to current user")
	viper.BindPFlag("DB.User", YEPCmd.PersistentFlags().Lookup("db-user"))
	YEPCmd.PersistentFlags().String("db-password", "", "Database password. Leave empty when connecting through socket.")
	viper.BindPFlag("DB.Password", YEPCmd.PersistentFlags().Lookup("db-password"))
	YEPCmd.PersistentFlags().String("db-name", "yep", "Database name. Defaults to 'yep'")
	viper.BindPFlag("DB.Name", YEPCmd.PersistentFlags().Lookup("db-name"))

	initVersion()
	initGenerate()
	initServer()
	initUpdateDB()
}
