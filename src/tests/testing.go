// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/controllers"
	"github.com/hexya-erp/hexya/src/menus"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/hexya/src/templates"
	"github.com/hexya-erp/hexya/src/tools/logging"
	"github.com/hexya-erp/hexya/src/views"
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
//         "github.com/hexya-erp/hexya/src/tests"
//     )
//
//     func TestMain(m *testing.M) {
//	       tests.RunTests(m, "my_module")
//     }
func RunTests(m *testing.M, moduleName string, preHookFnct func()) {
	var res int
	defer func() {
		TearDownTests(moduleName)
		if r := recover(); r != nil {
			panic(r)
		}
		os.Exit(res)
	}()
	InitializeTests(moduleName)
	if preHookFnct != nil {
		preHookFnct()
	}
	res = m.Run()

}

// InitializeTests initializes a database for the tests of the given module.
// You probably want to use RunTests instead.
func InitializeTests(moduleName string) {
	fmt.Printf("Initializing tests for module %s\n", moduleName)
	driver = os.Getenv("HEXYA_DB_DRIVER")
	if driver == "" {
		driver = "postgres"
	}
	user = os.Getenv("HEXYA_DB_USER")
	if user == "" {
		user = "hexya"
	}
	password = os.Getenv("HEXYA_DB_PASSWORD")
	if password == "" {
		password = "hexya"
	}
	prefix = os.Getenv("HEXYA_DB_PREFIX")
	if prefix == "" {
		prefix = "hexya"
	}
	dbName := fmt.Sprintf("%s_%s_tests", prefix, moduleName)
	debug = os.Getenv("HEXYA_DEBUG")
	logTests := os.Getenv("HEXYA_LOG")

	viper.Set("LogLevel", "panic")
	if logTests != "" {
		viper.Set("LogLevel", "info")
		viper.Set("LogStdout", true)
	}
	if debug != "" {
		viper.Set("Debug", true)
		viper.Set("LogLevel", "debug")
		viper.Set("LogStdout", true)
	}
	logging.Initialize()

	db := sqlx.MustConnect(driver, fmt.Sprintf("dbname=postgres sslmode=disable user=%s password=%s", user, password))
	keepDB := os.Getenv("HEXYA_KEEP_TEST_DB") != ""
	var dbExists bool
	err := db.Get(&dbExists, fmt.Sprintf("SELECT TRUE FROM pg_database WHERE datname = '%s'", dbName))
	if err != nil {
		fmt.Println(err)
	}
	if !dbExists || !keepDB {
		fmt.Println("Creating database", dbName)
		db.MustExec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
		db.MustExec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	}
	db.Close()

	server.PreInit()
	models.DBConnect(driver, models.ConnectionParams{
		DBName:   dbName,
		User:     user,
		Password: password,
		SSLMode:  "disable",
	})
	models.BootStrap()
	resourceDir, _ := filepath.Abs(filepath.Join(".", "res"))
	server.ResourceDir = resourceDir
	if !dbExists || !keepDB {
		fmt.Println("Upgrading schemas in database", dbName)
		models.SyncDatabase()
		fmt.Println("Loading resources into database", dbName)
		server.LoadDataRecords(resourceDir)
		server.LoadDemoRecords(resourceDir)
	}
	server.LoadInternalResources(resourceDir)
	views.BootStrap()
	templates.BootStrap()
	actions.BootStrap()
	controllers.BootStrap()
	menus.BootStrap()
	server.PostInit()
}

// TearDownTests tears down the tests for the given module
func TearDownTests(moduleName string) {
	models.DBClose()
	keepDB := os.Getenv("HEXYA_KEEP_TEST_DB")
	if keepDB != "" {
		return
	}
	fmt.Printf("Tearing down database for module %s...", moduleName)
	dbName := fmt.Sprintf("%s_%s_tests", prefix, moduleName)
	db := sqlx.MustConnect(driver, fmt.Sprintf("dbname=postgres sslmode=disable user=%s password=%s", user, password))
	db.MustExec(fmt.Sprintf("DROP DATABASE %s", dbName))
	db.Close()
	fmt.Println("Ok")
}
