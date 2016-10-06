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

package tests

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/npiganeau/yep/yep/models"
	"testing"
)

var DBARGS = struct {
	Driver   string
	User     string
	Password string
	DB       string
	Debug    string
}{
	os.Getenv("YEP_DB_DRIVER"),
	os.Getenv("YEP_DB_USER"),
	os.Getenv("YEP_DB_PASSWORD"),
	os.Getenv("YEP_DB_PREFIX") + "_tests",
	os.Getenv("YEP_DEBUG"),
}

func TestMain(m *testing.M) {
	initializeTests()
	res := m.Run()
	os.Exit(res)
}

func initializeTests() {
	if DBARGS.Driver == "" || DBARGS.DB == "" || DBARGS.User == "" {
		fmt.Println(`need driver and credentials!

Default DB Drivers.

postgres: https://github.com/lib/pq


usage:

go get -u github.com/lib/pq

#### PostgreSQL
psql -c 'create database yep_test_tests;' -U postgres
export YEP_DB_DRIVER=postgres
export YEP_DB_USER=postgres
export YEP_DB_PREFIX=yep_test
export YEP_DB_PASSWORD=secret
go test -v github.com/npiganeau/yep/yep/tests

`)
		os.Exit(2)
	}

	models.DBConnect(DBARGS.Driver, fmt.Sprintf("dbname=%s sslmode=disable user=%s password=%s", DBARGS.DB, DBARGS.User, DBARGS.Password))
	models.Testing = true
}
