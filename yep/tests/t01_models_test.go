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
	"testing"

	"github.com/npiganeau/yep/yep/models"
	_ "github.com/npiganeau/yep/yep/tests/test_module"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateDB(t *testing.T) {
	Convey("Creating DataBase...", t, func() {
		// Creating a dummy table to check that it is correctly removed by Bootstrap
		//db.MustExec("CREATE TABLE IF NOT EXISTS shouldbedeleted (id serial NOT NULL PRIMARY KEY)")
	})

	Convey("Database creation should run fine", t, func() {
		//Convey("Dummy table should exist", func() {
		//	So(testAdapter.tables(), ShouldContainKey, "shouldbedeleted")
		//})
		Convey("Bootstrap should not panic", func() {
			So(models.BootStrap, ShouldNotPanic)
		})
		//Convey("All models should have a DB table", func() {
		//	dbTables := testAdapter.tables()
		//	for tableName := range modelRegistry.registryByTableName {
		//		So(dbTables[tableName], ShouldBeTrue)
		//	}
		//})
		//Convey("All DB tables should have a model", func() {
		//	for dbTable := range testAdapter.tables() {
		//		So(modelRegistry.registryByTableName, ShouldContainKey, dbTable)
		//	}
		//})
	})
	//Convey("Truncating all tables...", t, func() {
	//	for tn := range modelRegistry.registryByTableName {
	//		dbExecuteNoTx(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, tn))
	//	}
	//})
}
