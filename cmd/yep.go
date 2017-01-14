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

	initVersion()
	initGenerate()
	initServer()
}
