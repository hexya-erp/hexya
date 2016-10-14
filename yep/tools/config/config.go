// Copyright 2016 NDP Systèmes. All Rights Reserved.
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

package config

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var Config *viper.Viper

func init() {
	Config = viper.New()
	setConfigDefaults()

	// Enable env vars configuration
	Config.SetEnvPrefix("YEP")
	Config.AutomaticEnv()

	// Read configuration file (if any)
	Config.SetConfigName("yep")
	Config.AddConfigPath("/etc/yep/")
	Config.AddConfigPath("$HOME/.yep")
	Config.AddConfigPath("./config/")
	Config.ReadInConfig()

	setConfigFlags()
	flag.Parse()
}

// setConfigDefaults sets the defaults of YEP configuration
func setConfigDefaults() {
	Config.SetDefault("DBDriver", "postgres")
	Config.SetDefault("DBSource", "dbname=yep sslmode=disable password=yep user=yep")
}

// setConfigFlags defines YEP command line flags and bind them with the Config
func setConfigFlags() {
	flag.StringP("log-level", "L", "info", "Log level. Should be one of 'debug', 'info', 'warn', 'error' or 'crit'")
	Config.BindPFlag("LogLevel", flag.Lookup("log-level"))
	flag.StringP("log-file", "l", "", "File to which the log will be written")
	Config.BindPFlag("LogFile", flag.Lookup("log-file"))
	flag.BoolP("log-stdout", "o", false, "Enable stdout logging. Use for development or debugging.")
	Config.BindPFlag("LogStdout", flag.Lookup("log-stdout"))

	flag.StringP("db-driver", "d", "postgres", "Database driver to use")
	Config.BindPFlag("DBDriver", flag.Lookup("db-driver"))
	flag.StringP("db-source", "s", "", "Database source string (e.g. 'dbname=yep sslmode=disable password=yep user=yep'")
	Config.BindPFlag("DBSource", flag.Lookup("db-source"))
}
