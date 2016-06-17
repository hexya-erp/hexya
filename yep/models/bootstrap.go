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
	"github.com/npiganeau/yep/yep/tools"
	"reflect"
)

/*
BootStrap freezes model, fields and method caches and syncs the database structure
with the declared data.
*/
func BootStrap() {
	modelRegistry.Lock()
	defer modelRegistry.Unlock()
	modelRegistry.bootstrapped = true

	createModelLinks()
	syncDatabase()

	//fieldsCache.Lock()
	//defer fieldsCache.Unlock()
	//processDepends()
	//
	//fieldsCache.done = true
}

// createModelLinks create links with related modelInfo
// where applicable.
func createModelLinks() {
	for _, mi := range modelRegistry.registryByName {
		for _, fi := range mi.fields.registryByName {
			sfType := fi.structField.Type
			var (
				relatedMI *modelInfo
				ok bool = true
			)
			switch fi.fieldType {
			case tools.MANY2ONE:
				if sfType.Kind() != reflect.Ptr {
					panic(fmt.Errorf("Many2one fields must be pointers"))
				}
				relatedMI, ok = modelRegistry.get(sfType.Elem().Name())
				if !ok {
					panic(fmt.Errorf("Unknown model `%s`", sfType.Elem().Name()))
				}
				fi.relatedModel = relatedMI
			case tools.ONE2MANY, tools.MANY2MANY:
				if sfType.Kind() != reflect.Slice {
					panic(fmt.Errorf("One2many or Many2many fields must be slice of pointers"))
				}
				if sfType.Elem().Kind() != reflect.Ptr {
					panic(fmt.Errorf("One2many or Many2many fields must be slice of pointers"))
				}
				relatedMI, ok = modelRegistry.get(sfType.Elem().Elem().Name())
			}
			if !ok {
				panic(fmt.Errorf("Unknown model `%s`", sfType.Elem().Name()))
			}
			fi.relatedModel = relatedMI
		}
	}
}

// syncDatabase creates or updates database tables with the data in the model registry
func syncDatabase() {
	adapter := adapters[db.DriverName()]
	dbTables := adapter.tables()
	// Create or update existing tables
	for tableName, mi := range modelRegistry.registryByTableName {
		var tableExists bool
		for _, dbTable := range dbTables {
			if tableName != dbTable {
				continue
			}
			tableExists = true
			break
		}
		if !tableExists {
			createDBTable(mi)
		}
		updateDBColumns(mi)
	}
	// Drop DB tables that are not in the models
	for _, dbTable := range adapter.tables() {
		var modelExists bool
		for tableName, _ := range modelRegistry.registryByTableName {
			if dbTable != tableName {
				continue
			}
			modelExists = true
			break
		}
		if !modelExists {
			dropDBTable(dbTable)
		}
	}
}

// createDBTable creates a table in the database from the given modelInfo
// It only creates the primary key. Call updateDBColumns to create columns.
func createDBTable(mi *modelInfo) {
	query := fmt.Sprintf(`
	CREATE TABLE %s (
		id serial NOT NULL PRIMARY KEY
	)
	`, mi.tableName)
	db.MustExec(query)
}

// updateDBColumns synchronizes the colums of the database with the
// given modelInfo.
func updateDBColumns(mi *modelInfo) {

}

// dropDBTable drops the given table in the database
func dropDBTable(tableName string) {
	query := fmt.Sprintf(`DROP TABLE %s`, tableName)
	db.MustExec(query)
}
