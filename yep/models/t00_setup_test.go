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

package models

import (
	"fmt"
	"github.com/npiganeau/yep/yep/orm"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type argAny []interface{}

// get interface by index from interface slice
func (a argAny) Get(i int, args ...interface{}) (r interface{}) {
	if i >= 0 && i < len(a) {
		r = a[i]
	}
	if len(args) > 0 {
		r = args[0]
	}
	return
}

var DBARGS = struct {
	Driver string
	Source string
	Debug  string
}{
	os.Getenv("ORM_DRIVER"),
	os.Getenv("ORM_SOURCE"),
	os.Getenv("ORM_DEBUG"),
}

var (
	dORM orm.Ormer
)

func init() {
	orm.Debug, _ = orm.StrTo(DBARGS.Debug).Bool()

	if DBARGS.Driver == "" || DBARGS.Source == "" {
		fmt.Println(`need driver and source!

Default DB Drivers.

  driver: url
   mysql: https://github.com/go-sql-driver/mysql
 sqlite3: https://github.com/mattn/go-sqlite3
postgres: https://github.com/lib/pq
tidb: https://github.com/pingcap/tidb

usage:

go get -u github.com/npiganeau/yep/yep/orm
go get -u github.com/go-sql-driver/mysql
go get -u github.com/mattn/go-sqlite3
go get -u github.com/lib/pq
go get -u github.com/pingcap/tidb

#### MySQL
mysql -u root -e 'create database orm_test;'
export ORM_DRIVER=mysql
export ORM_SOURCE="root:@/orm_test?charset=utf8"
go test -v github.com/npiganeau/yep/yep/orm


#### Sqlite3
export ORM_DRIVER=sqlite3
export ORM_SOURCE='file:memory_test?mode=memory'
go test -v github.com/npiganeau/yep/yep/orm


#### PostgreSQL
psql -c 'create database orm_test;' -U postgres
export ORM_DRIVER=postgres
export ORM_SOURCE="user=postgres dbname=orm_test sslmode=disable"
go test -v github.com/npiganeau/yep/yep/orm

#### TiDB
export ORM_DRIVER=tidb
export ORM_SOURCE='memory://test/test'
go test -v github.com/npiganeau/yep/yep/orm

`)
		os.Exit(2)
	}

	orm.RegisterDataBase("default", DBARGS.Driver, DBARGS.Source, 20)
}
