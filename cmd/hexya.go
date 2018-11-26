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
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hexya-erp/hexya/src/tools/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var log logging.Logger

// HexyaCmd is the base 'hexya' command of the commander
var HexyaCmd = &cobra.Command{
	Use:   "hexya",
	Short: "Hexya is an open source modular ERP",
	Long: `Hexya is an open source modular ERP written in Go.
It is designed for high demand business data processing while being easily customizable`,
}

// SetHexyaFlags adds the Hexya flags to the given cobra command
func SetHexyaFlags(c *cobra.Command) {
	c.PersistentFlags().StringP("config", "c", "", "Alternate configuration file to read. Defaults to $HOME/.hexya/")
	viper.BindPFlag("ConfigFileName", c.PersistentFlags().Lookup("config"))
	c.PersistentFlags().StringSliceP("modules", "m", []string{"github.com/hexya-addons/web"}, "List of module paths to load. Defaults to ['github.com/hexya-addons/web']")
	viper.BindPFlag("Modules", c.PersistentFlags().Lookup("modules"))
	c.PersistentFlags().StringP("log-level", "L", "info", "Log level. Should be one of 'debug', 'info', 'warn', 'error' or 'panic'")
	viper.BindPFlag("LogLevel", c.PersistentFlags().Lookup("log-level"))
	c.PersistentFlags().String("log-file", "", "File to which the log will be written")
	viper.BindPFlag("LogFile", c.PersistentFlags().Lookup("log-file"))
	c.PersistentFlags().BoolP("log-stdout", "o", false, "Enable stdout logging. Use for development or debugging.")
	viper.BindPFlag("LogStdout", c.PersistentFlags().Lookup("log-stdout"))
	c.PersistentFlags().Bool("debug", false, "Enable server debug mode for development")
	viper.BindPFlag("Debug", c.PersistentFlags().Lookup("debug"))
	c.PersistentFlags().Bool("demo", false, "Load demo data for evaluating or tests")
	viper.BindPFlag("Demo", c.PersistentFlags().Lookup("demo"))
	c.PersistentFlags().String("data-dir", "", "Path to the directory where Hexya should store its data")
	viper.BindPFlag("DataDir", c.PersistentFlags().Lookup("data-dir"))
	c.PersistentFlags().String("resource-dir", "./res", "Path to the directory where Hexya should read its resources. Defaults to 'res' subdirectory of current directory")
	viper.BindPFlag("ResourceDir", c.PersistentFlags().Lookup("resource-dir"))
	c.PersistentFlags().String("db-driver", "postgres", "Database driver to use")
	viper.BindPFlag("DB.Driver", c.PersistentFlags().Lookup("db-driver"))
	c.PersistentFlags().String("db-host", "/var/run/postgresql",
		"The database host to connect to. Values that start with / are for unix domain sockets directory")
	viper.BindPFlag("DB.Host", c.PersistentFlags().Lookup("db-host"))
	c.PersistentFlags().String("db-port", "5432", "Database port. Value is ignored if db-host is not set")
	viper.BindPFlag("DB.Port", c.PersistentFlags().Lookup("db-port"))
	c.PersistentFlags().String("db-user", "", "Database user. Defaults to current user")
	viper.BindPFlag("DB.User", c.PersistentFlags().Lookup("db-user"))
	c.PersistentFlags().String("db-password", "", "Database password. Leave empty when connecting through socket")
	viper.BindPFlag("DB.Password", c.PersistentFlags().Lookup("db-password"))
	c.PersistentFlags().String("db-name", "hexya", "Database name")
	viper.BindPFlag("DB.Name", c.PersistentFlags().Lookup("db-name"))
	c.PersistentFlags().String("db-ssl-mode", "disable", "SSL mode to connect to the database. Must be one of 'disable' (default), 'require', 'verify-ca' or 'verify-full'")
	viper.BindPFlag("DB.SSLMode", c.PersistentFlags().Lookup("db-ssl-mode"))
	c.PersistentFlags().String("db-ssl-cert", "", "Path to client certificate file")
	viper.BindPFlag("DB.SSLCert", c.PersistentFlags().Lookup("db-ssl-cert"))
	c.PersistentFlags().String("db-ssl-key", "", "Path to client private key file")
	viper.BindPFlag("DB.SSLKey", c.PersistentFlags().Lookup("db-ssl-key"))
	c.PersistentFlags().String("db-ssl-ca", "", "Path to certificate authority certificate(s) file")
	viper.BindPFlag("DB.SSLCA", c.PersistentFlags().Lookup("db-ssl-ca"))
}

// InitConfig initializes Hexya configuration system (viper).
func InitConfig() {
	viper.SetEnvPrefix("hexya")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	cfgFile := viper.GetString("ConfigFileName")
	if runtime.GOOS != "windows" {
		viper.AddConfigPath("/etc/hexya")
	}

	if osUser, err := user.Current(); err == nil {
		defaultHexyaDir := filepath.Join(osUser.HomeDir, ".hexya")
		viper.SetDefault("DataDir", defaultHexyaDir)
		viper.AddConfigPath(defaultHexyaDir)
	} else {
		fmt.Println(fmt.Errorf("unable to retrieve current user. Error: %s", err))
	}
	viper.AddConfigPath(".")

	viper.SetConfigName("hexya")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	log = logging.GetLogger("init")
	cobra.OnInitialize(InitConfig)
	SetHexyaFlags(HexyaCmd)
}
