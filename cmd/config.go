// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration file utilities",
	Long:  `Hexya configuration file (hexya.toml) utilities`,
}

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Scaffold a Hexya configuration file",
	Long: `Create a Hexya configuration file hexya.toml in the current directory. Use the -c flag to specify another destination file.
All configuration parameters passed as environment variables or as flags will be set in the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		cfgFile := viper.GetString("ConfigFileName")
		if cfgFile == "" {
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			cfgFile = filepath.Join(cwd, "hexya.toml")
		}
		if err := viper.WriteConfigAs(cfgFile); err != nil {
			panic(err)
		}
	},
}

func init() {
	HexyaCmd.AddCommand(configCmd)
	configCmd.AddCommand(scaffoldCmd)
}
