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

	"github.com/hexya-erp/hexya/hexya/models/security"
)

// A modelCouple holds a model and one of its mixin
type modelCouple struct {
	model *Model
	mixIn *Model
}

var mixed = map[modelCouple]bool{}

// BootStrap freezes model, fields and method caches and syncs the database structure
// with the declared data.
func BootStrap() {
	log.Info("Bootstrapping models")
	if Registry.bootstrapped == true {
		log.Panic("Trying to bootstrap models twice !")
	}
	Registry.Lock()
	defer Registry.Unlock()

	Registry.bootstrapped = true

	createModelLinks()
	inflateMixIns()
	inflateEmbeddings()
	syncRelatedFieldInfo()
	bootStrapMethods()
	processDepends()
	checkComputeMethodsSignature()
	setupSecurity()
}

// createModelLinks create links with related Model
// where applicable.
func createModelLinks() {
	for _, mi := range Registry.registryByName {
		for _, fi := range mi.fields.registryByName {
			var (
				relatedMI *Model
				ok        bool
			)
			if fi.fieldType.IsRelationType() {
				relatedMI, ok = Registry.Get(fi.relatedModelName)
				if !ok {
					log.Panic("Unknown related model in field declaration", "model", mi.name, "field", fi.name, "relatedName", fi.relatedModelName)
				}
			}
			fi.relatedModel = relatedMI
		}
		mi.fields.bootstrapped = true
	}
}

// inflateMixIns inserts fields and methods of mixed in models.
func inflateMixIns() {
	for _, mi := range Registry.registryByName {
		if mi.isM2MLink() {
			// We don"t mix in M2M link
			continue
		}
		for _, mixInMI := range mi.mixins {
			injectMixInModel(mixInMI, mi)
		}
	}
}

// injectMixInModel injects fields and methods of mixInMI in model
func injectMixInModel(mixInMI, mi *Model) {
	for _, mmm := range mixInMI.mixins {
		injectMixInModel(mmm, mixInMI)
	}
	if mixed[modelCouple{model: mi, mixIn: mixInMI}] {
		return
	}
	// Add mixIn fields
	for fName, fi := range mixInMI.fields.registryByName {
		if _, exists := mi.fields.registryByName[fName]; exists {
			// We do not add fields that already exist in the targetModel
			// since the target model should always override mixins.
			continue
		}
		newFI := *fi
		newFI.model = mi
		newFI.acl = security.NewAccessControlList()
		// TODO handle M2M fields
		mi.fields.add(&newFI)
		// We add the permissions of the mixin to the target model
		for group, perm := range fi.acl.Permissions() {
			newFI.acl.AddPermission(group, perm)
		}
	}
	// Add mixIn methods
	for methName, methInfo := range mixInMI.methods.registry {
		// Extract all method layers functions by inverse order
		layersInv := methInfo.invertedLayers()
		if emi, exists := mi.methods.registry[methName]; exists {
			// The method already exists in our target model
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
			for _, lf := range layersInv {
				ml := methodLayer{
					funcValue: wrapFunctionForMethodLayer(lf.funcValue),
					mixedIn:   true,
					method:    emi,
				}
				emi.nextLayer[&ml] = firstMixedLayer
				firstMixedLayer = &ml
			}
			if emi.topLayer == nil {
				// The existing method was empty
				emi.topLayer = firstMixedLayer
				emi.methodType = methInfo.methodType
			} else {
				emi.nextLayer[lastImplLayer] = firstMixedLayer
			}
		} else {
			// The method does not exist
			newMethInfo := copyMethod(mi, methInfo)
			for i := 0; i < len(layersInv); i++ {
				newMethInfo.addMethodLayer(layersInv[i].funcValue, layersInv[i].doc)
			}
			mi.methods.set(methName, newMethInfo)
		}
		// Copy groups to our methods in the target model
		for group := range methInfo.groups {
			mi.methods.MustGet(methName).groups[group] = true
		}
	}
	mixed[modelCouple{model: mi, mixIn: mixInMI}] = true
}

// inflateEmbeddings creates related fields for all fields of embedded models.
func inflateEmbeddings() {
	for _, mi := range Registry.registryByName {
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
				fInfo := Field{
					name:        relName,
					json:        relFI.json,
					acl:         security.NewAccessControlList(),
					model:       mi,
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

// syncRelatedFieldInfo overwrites the Field data of the related fields
// with the data of the Field of the target.
func syncRelatedFieldInfo() {
	for _, mi := range Registry.registryByName {
		for _, fi := range mi.fields.registryByName {
			if !fi.isRelatedField() {
				continue
			}
			newFI := *mi.getRelatedFieldInfo(fi.relatedPath)
			newFI.name = fi.name
			newFI.json = fi.json
			newFI.relatedPath = fi.relatedPath
			newFI.stored = fi.stored
			newFI.model = mi
			newFI.noCopy = true
			newFI.onChange = ""
			newFI.index = false
			newFI.compute = ""
			newFI.depends = nil
			*fi = newFI
		}
	}
}

// SyncDatabase creates or updates database tables with the data in the model registry
func SyncDatabase() {
	adapter := adapters[db.DriverName()]
	dbTables := adapter.tables()
	// Create or update existing tables
	for tableName, model := range Registry.registryByTableName {
		if model.isMixin() {
			// Don't create table for mixin models
			continue
		}
		if model.isManual() {
			// Don't create table for manual models
			continue
		}
		if _, ok := dbTables[tableName]; !ok {
			createDBTable(model.tableName)
		}
		updateDBColumns(model)
		updateDBIndexes(model)
	}
	// Setup constraints
	for _, model := range Registry.registryByTableName {
		if model.isMixin() {
			continue
		}
		buildSQLErrorSubstitutionMap(model)
		updateDBForeignKeyConstraints(model)
		updateDBConstraints(model)
	}
	// Drop DB tables that are not in the models
	for dbTable := range adapter.tables() {
		var modelExists bool
		for tableName, model := range Registry.registryByTableName {
			if dbTable != tableName {
				continue
			}
			if model.isMixin() {
				// We don't want a table for mixin models
				continue
			}
			modelExists = true
			break
		}
		if !modelExists {
			dropDBTable(dbTable)
		}
	}
	updateDBSequences()
}

// buildSQLErrorSubstitutionMap populates the sqlErrors map of the
// model with the appropriate error message substitution
func buildSQLErrorSubstitutionMap(model *Model) {
	for sqlConstraintName, sqlConstraint := range model.sqlConstraints {
		model.sqlErrors[sqlConstraintName] = sqlConstraint.errorString
	}
	for _, field := range model.fields.registryByJSON {
		if field.unique {
			cName := fmt.Sprintf("%s_%s_key", model.tableName, field.json)
			model.sqlErrors[cName] = fmt.Sprintf("%s must be unique", field.name)
		}
		if field.fieldType.IsFKRelationType() {
			cName := fmt.Sprintf("%s_%s_fkey", model.tableName, field.json)
			model.sqlErrors[cName] = fmt.Sprintf("%s must reference an existing %s record", field.name, field.relatedModelName)
		}
	}
}

// updateDBSequences synchronizes sequences between the DB
// and the registry.
func updateDBSequences() {
	adapter := adapters[db.DriverName()]
	// Create sequences
	for _, sequence := range Registry.sequences {
		exists := false
		for _, dbSeq := range adapter.sequences("%_manseq") {
			if sequence.JSON == dbSeq {
				exists = true
			}
		}
		if !exists {
			adapter.createSequence(sequence.JSON)
		}
	}
	// Drop unused sequences
	for _, dbSeq := range adapter.sequences("%_manseq") {
		var sequenceExists bool
		for _, sequence := range Registry.sequences {
			if sequence.JSON != dbSeq {
				continue
			}
			sequenceExists = true
			break
		}
		if !sequenceExists {
			adapter.dropSequence(dbSeq)
		}
	}
}

// createDBTable creates a table in the database from the given Model
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
// given Model.
func updateDBColumns(mi *Model) {
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

// createDBColumn insert the column described by Field in the database
func createDBColumn(fi *Field) {
	if !fi.isStored() {
		log.Panic("createDBColumn should not be called on non stored fields", "model", fi.model.name, "field", fi.json)
	}
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ADD COLUMN %s %s
	`, adapter.quoteTableName(fi.model.tableName), fi.json, adapter.columnSQLDefinition(fi))
	dbExecuteNoTx(query)
}

// updateDBColumnDataType updates the data type in database for the given Field
func updateDBColumnDataType(fi *Field) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ALTER COLUMN %s SET DATA TYPE %s
	`, adapter.quoteTableName(fi.model.tableName), fi.json, adapter.typeSQL(fi))
	dbExecuteNoTx(query)
}

// updateDBColumnNullable updates the NULL/NOT NULL data in database for the given Field
func updateDBColumnNullable(fi *Field) {
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
	`, adapter.quoteTableName(fi.model.tableName), fi.json, verb)
	dbExecuteNoTx(query)
}

// updateDBColumnDefault updates the default value in database for the given Field
func updateDBColumnDefault(fi *Field) {
	adapter := adapters[db.DriverName()]
	defValue := adapter.fieldSQLDefault(fi)
	var query string
	if defValue == "" {
		query = fmt.Sprintf(`
			ALTER TABLE %s
			ALTER COLUMN %s DROP DEFAULT
		`, adapter.quoteTableName(fi.model.tableName), fi.json)
	} else {
		query = fmt.Sprintf(`
			ALTER TABLE %s
			ALTER COLUMN %s SET DEFAULT %s
		`, adapter.quoteTableName(fi.model.tableName), fi.json, adapter.fieldSQLDefault(fi))
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

// updateDBForeignKeyConstraints creates or updates fk constraints
// based on the data of the given Model
func updateDBForeignKeyConstraints(m *Model) {
	adapter := adapters[db.DriverName()]
	for colName, fi := range m.fields.registryByJSON {
		fkContraintInDB := adapter.constraintExists(fmt.Sprintf("%s_%s_fkey", m.tableName, colName))
		fieldIsFK := fi.fieldType.IsFKRelationType() && fi.isStored()
		switch {
		case fieldIsFK && !fkContraintInDB:
			createFKConstraint(m.tableName, colName, fi.relatedModel.tableName, string(fi.onDelete))
		case !fieldIsFK && fkContraintInDB:
			dropFKConstraint(m.tableName, colName)
		}
	}
}

// updateDBConstraints creates or updates sql constraints
// based on the data of the given Model
func updateDBConstraints(m *Model) {
	adapter := adapters[db.DriverName()]
	for constraintName, constraint := range m.sqlConstraints {
		if !adapter.constraintExists(constraintName) {
			createConstraint(m.tableName, constraintName, constraint.sql)
		}
	}
dbConLoop:
	for _, dbConstraintName := range adapter.constraints(fmt.Sprintf("%%_%s_mancon", m.tableName)) {
		for constraintName := range m.sqlConstraints {
			if constraintName == dbConstraintName {
				continue dbConLoop
			}
		}
		dropConstraint(m.tableName, dbConstraintName)
	}
}

// createFKConstraint creates an FK constraint for the given column that references the given targetTable
func createFKConstraint(tableName, colName, targetTable, ondelete string) {
	adapter := adapters[db.DriverName()]
	constraint := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s ON DELETE %s", colName, adapter.quoteTableName(targetTable), ondelete)
	createConstraint(tableName, fmt.Sprintf("%s_%s_fkey", tableName, colName), constraint)
}

// dropFKConstraint drops an FK constraint for colName in the given table
func dropFKConstraint(tableName, colName string) {
	dropConstraint(tableName, fmt.Sprintf("%s_%s_fkey", tableName, colName))
}

// createConstraint creates a constraint in the given table
func createConstraint(tableName, constraintName, sql string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s ADD CONSTRAINT %s %s
	`, adapter.quoteTableName(tableName), constraintName, sql)
	dbExecuteNoTx(query)
}

// dropConstraint drops a constraint with the given name
func dropConstraint(tableName, constraintName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s
	`, adapter.quoteTableName(tableName), constraintName)
	dbExecuteNoTx(query)
}

// updateDBIndexes creates or updates indexes based on the data of
// the given Model
func updateDBIndexes(m *Model) {
	adapter := adapters[db.DriverName()]
	for colName, fi := range m.fields.registryByJSON {
		indexInDB := adapter.indexExists(m.tableName, fmt.Sprintf("%s_%s_index", m.tableName, colName))
		switch {
		case fi.index && !indexInDB:
			createColumnIndex(m.tableName, colName)
		case !fi.index && indexInDB:
			dropColumnIndex(m.tableName, colName)
		}
	}
}

// createColumnIndex creates an column index for colName in the given table
func createColumnIndex(tableName, colName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		CREATE INDEX %s ON %s (%s)
	`, fmt.Sprintf("%s_%s_index", tableName, colName), adapter.quoteTableName(tableName), colName)
	dbExecuteNoTx(query)
}

// dropColumnIndex drops a column index for colName in the given table
func dropColumnIndex(tableName, colName string) {
	query := fmt.Sprintf(`
		DROP INDEX IF EXISTS %s
	`, fmt.Sprintf("%s_%s_index", tableName, colName))
	dbExecuteNoTx(query)
}

// bootStrapMethods freezes the methods of the models.
func bootStrapMethods() {
	for _, model := range Registry.registryByName {
		model.methods.bootstrapped = true
	}
}

// setupSecurity adds execution permission to:
// - the admin group for all methods
// - all methods of a model for groups that have been granted all rights
func setupSecurity() {
	for _, model := range Registry.registryByName {
		for _, meth := range model.methods.registry {
			meth.groups[security.GroupAdmin] = true
			for group := range model.methods.powerGroups {
				meth.groups[group] = true
			}
		}
	}
}
