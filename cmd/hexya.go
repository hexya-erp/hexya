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
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var log *logging.Logger

// HexyaCmd is the base 'hexya' command of the commander
var HexyaCmd = &cobra.Command{
	Use:   "hexya",
	Short: "Hexya is an open source modular ERP",
	Long: `Hexya is an open source modular ERP written in Go.
It is designed for high demand business data processing while being easily customizable`,
}

func init() {
	log = logging.GetLogger("init")
	cobra.OnInitialize(initConfig)

	HexyaCmd.PersistentFlags().StringP("config", "c", "", "Alternate configuration file to read. Defaults to $HOME/.hexya/")
	viper.BindPFlag("ConfigFileName", HexyaCmd.PersistentFlags().Lookup("config"))

	HexyaCmd.PersistentFlags().StringP("log-level", "L", "info", "Log level. Should be one of 'debug', 'info', 'warn', 'error' or 'crit'")
	viper.BindPFlag("LogLevel", HexyaCmd.PersistentFlags().Lookup("log-level"))
	HexyaCmd.PersistentFlags().String("log-file", "", "File to which the log will be written")
	viper.BindPFlag("LogFile", HexyaCmd.PersistentFlags().Lookup("log-file"))
	HexyaCmd.PersistentFlags().BoolP("log-stdout", "o", false, "Enable stdout logging. Use for development or debugging.")
	viper.BindPFlag("LogStdout", HexyaCmd.PersistentFlags().Lookup("log-stdout"))
	HexyaCmd.PersistentFlags().Bool("debug", false, "Enable server debug mode for development")
	viper.BindPFlag("Debug", HexyaCmd.PersistentFlags().Lookup("debug"))

	HexyaCmd.PersistentFlags().String("db-driver", "postgres", "Database driver to use")
	viper.BindPFlag("DB.Driver", HexyaCmd.PersistentFlags().Lookup("db-driver"))
	HexyaCmd.PersistentFlags().String("db-host", "/var/run/postgresql",
		"The database host to connect to. Values that start with / are for unix domain sockets directory")
	viper.BindPFlag("DB.Host", HexyaCmd.PersistentFlags().Lookup("db-host"))
	HexyaCmd.PersistentFlags().String("db-port", "5432", "Database port. Value is ignored if db-host is not set")
	viper.BindPFlag("DB.Port", HexyaCmd.PersistentFlags().Lookup("db-port"))
	HexyaCmd.PersistentFlags().String("db-user", "", "Database user. Defaults to current user")
	viper.BindPFlag("DB.User", HexyaCmd.PersistentFlags().Lookup("db-user"))
	HexyaCmd.PersistentFlags().String("db-password", "", "Database password. Leave empty when connecting through socket")
	viper.BindPFlag("DB.Password", HexyaCmd.PersistentFlags().Lookup("db-password"))
	HexyaCmd.PersistentFlags().String("db-name", "hexya", "Database name")
	viper.BindPFlag("DB.Name", HexyaCmd.PersistentFlags().Lookup("db-name"))

	initVersion()
	initGenerate()
	initServer()
	initUpdateDB()
	initI18n()
}

func initConfig() {
	cfgFile := viper.GetString("ConfigFileName")

	// Read config either from cfgFile or from home directory!
	viper.AddConfigPath("/etc/hexya/")
	viper.AddConfigPath("$HOME/.hexya")
	viper.AddConfigPath(".")
	if cfgFile == "" {
		cfgFile = "hexya"
	}
	viper.SetConfigFile(cfgFile)
	viper.ReadInConfig()
}
