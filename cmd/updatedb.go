// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package cmd

import (
	"path/filepath"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var updateDBCmd = &cobra.Command{
	Use:   "updatedb",
	Short: "Update the database schema",
	Long:  `Synchronize the database schema with the models definitions.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectDir := "."
		if len(args) > 0 {
			projectDir = args[0]
		}
		runProject(projectDir, "updatedb", args)
	},
}

// UpdateDB updates the database schema. It is meant to be called from
// a project start file which imports all the project's module.
func UpdateDB() {
	setupLogger()
	setupDebug()
	server.PreInit()
	connectToDB()
	models.BootStrap()
	models.SyncDatabase()
	resourceDir, err := filepath.Abs(viper.GetString("ResourceDir"))
	if err != nil {
		log.Panic("Unable to find Resource directory", "error", err)
	}
	server.ResourceDir = resourceDir
	server.LoadDataRecords(resourceDir)
	if viper.GetBool("Demo") {
		log.Info("Demo mode detected: loading demo data")
		server.LoadDemoRecords(resourceDir)
	}
	log.Info("Database updated successfully")
}

func init() {
	HexyaCmd.AddCommand(updateDBCmd)
}
