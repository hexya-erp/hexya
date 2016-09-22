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
	"testing"

	_ "github.com/lib/pq"
	"github.com/npiganeau/yep/yep/models"
)

var DBARGS = struct {
	Driver string
	Source string
	Debug  string
}{
	os.Getenv("YEP_DB_DRIVER"),
	os.Getenv("YEP_DB_SOURCE"),
	os.Getenv("YEP_DEBUG"),
}

func init() {
	if DBARGS.Driver == "" || DBARGS.Source == "" {
		fmt.Println(`need driver and source!

Default DB Drivers.

postgres: https://github.com/lib/pq


usage:

go get -u github.com/lib/pq

#### PostgreSQL
psql -c 'create database orm_test;' -U postgres
export YEP_DB_DRIVER=postgres
export YEP_DB_SOURCE="user=postgres dbname=orm_test sslmode=disable"
go test -v github.com/npiganeau/yep/yep/models

`)
		os.Exit(2)
	}

	models.DBConnect(DBARGS.Driver, DBARGS.Source)
	models.Testing = true
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
