// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var driver, user, password, prefix, debug string

// RunTests initializes the database, run the tests given by m and
// tears the database down.
//
// It is meant to be used for modules testing. Initialize your module's
// tests with:
//
//     import (
//         "testing"
//         "github.com/hexya-erp/hexya/hexya/tests"
//     )
//
//     func TestMain(m *testing.M) {
//	       tests.RunTests(m, "my_module")
//     }
func RunTests(m *testing.M, moduleName string) {
	InitializeTests(moduleName)
	res := m.Run()
	TearDownTests(moduleName)
	os.Exit(res)
}

// InitializeTests initializes a database for the tests of the given module.
// You probably want to use RunTests instead.
func InitializeTests(moduleName string) {
	fmt.Printf("Initializing database for module %s\n", moduleName)
	driver = os.Getenv("Hexya_DB_DRIVER")
	if driver == "" {
		driver = "postgres"
	}
	user = os.Getenv("Hexya_DB_USER")
	if user == "" {
		user = "hexya"
	}
	password = os.Getenv("Hexya_DB_PASSWORD")
	if password == "" {
		password = "hexya"
	}
	prefix = os.Getenv("Hexya_DB_PREFIX")
	if prefix == "" {
		prefix = "hexya"
	}
	dbName := fmt.Sprintf("%s_%s_tests", prefix, moduleName)
	debug = os.Getenv("Hexya_DEBUG")

	viper.Set("LogLevel", "crit")
	if debug != "" {
		viper.Set("LogLevel", "debug")
		viper.Set("LogStdout", true)
	}
	logging.Initialize()

	db := sqlx.MustConnect(driver, fmt.Sprintf("dbname=postgres sslmode=disable user=%s password=%s", user, password))
	db.MustExec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	db.Close()

	models.DBConnect(driver, fmt.Sprintf("dbname=%s sslmode=disable user=%s password=%s", dbName, user, password))
	models.BootStrap()
	models.SyncDatabase()

	server.PostInitModules()
}

// TearDownTests tears down the tests for the given module
func TearDownTests(moduleName string) {
	models.DBClose()
	fmt.Printf("Tearing down database for module %s\n", moduleName)
	dbName := fmt.Sprintf("%s_%s_tests", prefix, moduleName)
	db := sqlx.MustConnect(driver, fmt.Sprintf("dbname=postgres sslmode=disable user=%s password=%s", user, password))
	db.MustExec(fmt.Sprintf("DROP DATABASE %s", dbName))
	db.Close()
}
