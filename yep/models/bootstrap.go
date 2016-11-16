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
	"reflect"

	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/tools/logging"
)

// BootStrap freezes model, fields and method caches and syncs the database structure
// with the declared data.
func BootStrap() {
	log.Info("Bootstrapping models")
	if modelRegistry.bootstrapped == true {
		logging.LogAndPanic(log, "Trying to bootstrap models twice !")
	}
	modelRegistry.Lock()
	defer modelRegistry.Unlock()

	modelRegistry.bootstrapped = true

	createModelLinks()
	inflateMixIns()
	inflateEmbeddings()
	syncRelatedFieldInfo()
	syncDatabase()
	bootStrapMethods()
	processDepends()
	setupSecurity()
}

// createModelLinks create links with related modelInfo
// where applicable.
func createModelLinks() {
	for _, mi := range modelRegistry.registryByName {
		for _, fi := range mi.fields.registryByName {
			var (
				relatedMI *modelInfo
				ok        bool
			)
			if fi.fieldType.IsRelationType() {
				relatedMI, ok = modelRegistry.get(fi.relatedModelName)
				if !ok {
					logging.LogAndPanic(log, "Unknown related model in field declaration", "model", mi.name, "field", fi.name, "relatedName", fi.relatedModelName)
				}
			}
			fi.relatedModel = relatedMI
		}
		mi.fields.bootstrapped = true
	}
}

// inflateMixIns inserts fields and methods of mixed in models.
func inflateMixIns() {
	for _, mi := range modelRegistry.registryByName {
		for _, mixInMI := range mi.mixins {
			// Add mixIn fields
			for fName, fi := range mixInMI.fields.registryByName {
				if _, exists := mi.fields.registryByName[fName]; exists {
					// We do not add fields that already exist in the targetModel
					// since the target model should always override mixins.
					continue
				}
				newFI := *fi
				newFI.mi = mi
				mi.fields.add(&newFI)
			}
			// Add mixIn methods
			for methName, methInfo := range mixInMI.methods.registry {
				// Extract all method layers functions by inverse order
				var layerFuncsInv []reflect.Value
				for cl := methInfo.topLayer; cl != nil; cl = methInfo.getNextLayer(cl) {
					layerFuncsInv = append([]reflect.Value{cl.funcValue}, layerFuncsInv...)
				}
				if emi, exists := mi.methods.registry[methName]; exists {
					// The method already exists in our target model.
					// We insert our new method layers above previous mixins layers
					// but below the target model implementations.
					lastImplLayer := emi.topLayer
					firstMixedLayer := emi.getNextLayer(lastImplLayer)
					for firstMixedLayer != nil {
						if firstMixedLayer.mixedIn {
							break
						}
						lastImplLayer = firstMixedLayer
						firstMixedLayer = emi.getNextLayer(lastImplLayer)
					}
					for _, lf := range layerFuncsInv {
						ml := methodLayer{
							funcValue: wrapFunctionForMethodLayer(lf),
							mixedIn:   true,
							methInfo:  emi,
						}
						emi.nextLayer[&ml] = firstMixedLayer
						firstMixedLayer = &ml
					}
					emi.nextLayer[lastImplLayer] = firstMixedLayer
				} else {
					newMethInfo := &methodInfo{
						mi:         mi,
						name:       methName,
						methodType: methInfo.methodType,
						nextLayer:  make(map[*methodLayer]*methodLayer),
					}
					for i := 0; i < len(layerFuncsInv); i++ {
						newMethInfo.addMethodLayer(layerFuncsInv[i])
					}
					mi.methods.set(methName, newMethInfo)
				}
			}
		}
	}
}

// inflateEmbeddings creates related fields for all fields of embedded models.
func inflateEmbeddings() {
	for _, mi := range modelRegistry.registryByName {
		for _, fi := range mi.fields.registryByName {
			if !fi.embed {
				continue
			}
			for relName, relFI := range fi.relatedModel.fields.registryByName {
				if _, ok := mi.fields.get(relName); ok {
					// Don't add the field if we have a field with the same name
					// in our model (shadowing).
					continue
				}
				fInfo := fieldInfo{
					name:        relName,
					json:        relFI.json,
					acl:         security.NewAccessControlList(),
					mi:          mi,
					stored:      fi.stored,
					structField: relFI.structField,
					noCopy:      true,
					relatedPath: fmt.Sprintf("%s%s%s", fi.name, ExprSep, relName),
				}
				mi.fields.add(&fInfo)
			}
		}
	}
}

// syncRelatedFieldInfo overwrites the fieldInfo data of the related fields
// with the data of the fieldInfo of the target.
func syncRelatedFieldInfo() {
	for _, mi := range modelRegistry.registryByName {
		for _, fi := range mi.fields.registryByName {
			if !fi.isRelatedField() {
				continue
			}
			newFI := *mi.getRelatedFieldInfo(fi.relatedPath)
			newFI.name = fi.name
			newFI.json = fi.json
			newFI.relatedPath = fi.relatedPath
			newFI.stored = fi.stored
			newFI.mi = mi
			newFI.noCopy = true
			*fi = newFI
		}
	}
}

// syncDatabase creates or updates database tables with the data in the model registry
func syncDatabase() {
	adapter := adapters[db.DriverName()]
	dbTables := adapter.tables()
	// Create or update existing tables
	for tableName, mi := range modelRegistry.registryByTableName {
		if _, ok := dbTables[tableName]; !ok {
			createDBTable(mi.tableName)
		}
		updateDBColumns(mi)
		updateDBIndexes(mi)
	}
	// Drop DB tables that are not in the models
	for dbTable := range adapter.tables() {
		var modelExists bool
		for tableName := range modelRegistry.registryByTableName {
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
func createDBTable(tableName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
	CREATE TABLE %s (
		id serial NOT NULL PRIMARY KEY
	)
	`, adapter.quoteTableName(tableName))
	dbExecuteNoTx(query)
}

// dropDBTable drops the given table in the database
func dropDBTable(tableName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`DROP TABLE %s`, adapter.quoteTableName(tableName))
	dbExecuteNoTx(query)
}

// updateDBColumns synchronizes the colums of the database with the
// given modelInfo.
func updateDBColumns(mi *modelInfo) {
	adapter := adapters[db.DriverName()]
	dbColumns := adapter.columns(mi.tableName)
	// create or update columns from registry data
	for colName, fi := range mi.fields.registryByJSON {
		if colName == "id" || !fi.isStored() {
			continue
		}
		dbColData, ok := dbColumns[colName]
		if !ok {
			createDBColumn(fi)
		}
		if dbColData.DataType != adapter.typeSQL(fi) {
			updateDBColumnDataType(fi)
		}
		if (dbColData.IsNullable == "NO" && !adapter.fieldIsNotNull(fi)) ||
			(dbColData.IsNullable == "YES" && adapter.fieldIsNotNull(fi)) {
			updateDBColumnNullable(fi)
		}
		if dbColData.ColumnDefault.Valid &&
			dbColData.ColumnDefault.String != adapter.fieldSQLDefault(fi) {
			updateDBColumnDefault(fi)
		}
	}
	// drop columns that no longer exist
	for colName := range dbColumns {
		if _, ok := mi.fields.registryByJSON[colName]; !ok {
			dropDBColumn(mi.tableName, colName)
		}
	}
}

// createDBColumn insert the column described by fieldInfo in the database
func createDBColumn(fi *fieldInfo) {
	if !fi.isStored() {
		logging.LogAndPanic(log, "createDBColumn should not be called on non stored fields", "model", fi.mi.name, "field", fi.json)
	}
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ADD COLUMN %s %s
	`, adapter.quoteTableName(fi.mi.tableName), fi.json, adapter.columnSQLDefinition(fi))
	dbExecuteNoTx(query)
}

// updateDBColumnDataType updates the data type in database for the given fieldInfo
func updateDBColumnDataType(fi *fieldInfo) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ALTER COLUMN %s SET DATA TYPE %s
	`, adapter.quoteTableName(fi.mi.tableName), fi.json, adapter.typeSQL(fi))
	dbExecuteNoTx(query)
}

// updateDBColumnNullable updates the NULL/NOT NULL data in database for the given fieldInfo
func updateDBColumnNullable(fi *fieldInfo) {
	adapter := adapters[db.DriverName()]
	var verb string
	if adapter.fieldIsNotNull(fi) {
		verb = "SET"
	} else {
		verb = "DROP"
	}
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ALTER COLUMN %s %s NOT NULL
	`, adapter.quoteTableName(fi.mi.tableName), fi.json, verb)
	dbExecuteNoTx(query)
}

// updateDBColumnDefault updates the default value in database for the given fieldInfo
func updateDBColumnDefault(fi *fieldInfo) {
	adapter := adapters[db.DriverName()]
	defValue := adapter.fieldSQLDefault(fi)
	var query string
	if defValue == "" {
		query = fmt.Sprintf(`
			ALTER TABLE %s
			ALTER COLUMN %s DROP DEFAULT
		`, adapter.quoteTableName(fi.mi.tableName), fi.json)
	} else {
		query = fmt.Sprintf(`
			ALTER TABLE %s
			ALTER COLUMN %s SET DEFAULT %s
		`, adapter.quoteTableName(fi.mi.tableName), fi.json, adapter.fieldSQLDefault(fi))
	}
	dbExecuteNoTx(query)
}

// dropDBColumn drops the column colName from table tableName in database
func dropDBColumn(tableName, colName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s
		DROP COLUMN %s
	`, adapter.quoteTableName(tableName), colName)
	dbExecuteNoTx(query)
}

// updateDBIndexes creates or updates indexes based on the data of
// the given modelInfo
func updateDBIndexes(mi *modelInfo) {
	adapter := adapters[db.DriverName()]
	// update column indexes
	for colName, fi := range mi.fields.registryByJSON {
		if !fi.index {
			continue
		}
		if !adapter.indexExists(mi.tableName, fmt.Sprintf("%s_%s_index", mi.tableName, colName)) {
			createColumnIndex(mi.tableName, colName)
		}
	}
}

// createIndex creates an column index for colName in the given table
func createColumnIndex(tableName, colName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		CREATE INDEX %s ON %s (%s)
	`, fmt.Sprintf("%s_%s_index", tableName, colName), adapter.quoteTableName(tableName), colName)
	dbExecuteNoTx(query)
}

// bootStrapMethods freezes the methods of the models.
func bootStrapMethods() {
	for _, mi := range modelRegistry.registryByName {
		mi.methods.bootstrapped = true
	}
}

// setupSecurity adds all permissions to the admin group for all models
func setupSecurity() {
	for modelName := range modelRegistry.registryByName {
		AllowModelAccess(ModelName(modelName), security.AdminGroup, security.All)
	}
}
