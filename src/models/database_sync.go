// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"strings"

	"github.com/hexya-erp/hexya/src/models/security"
)

// SyncDatabase creates or updates database tables with the data in the model registry
func SyncDatabase() {
	log.Info("Updating database schema")
	adapter := adapters[db.DriverName()]
	dbTables := adapter.tables()
	// Create or update sequences
	updateDBSequences()
	// Create or update existing tables
	for tableName, model := range Registry.registryByTableName {
		if model.IsMixin() || model.IsManual() {
			continue
		}
		if _, ok := dbTables[tableName]; !ok {
			createDBTable(model)
		}
		updateDBColumns(model)
		updateDBIndexes(model)
	}
	// Setup constraints
	for _, model := range Registry.registryByTableName {
		if model.IsMixin() || model.IsManual() {
			continue
		}
		buildSQLErrorSubstitutionMap(model)
		updateDBForeignKeyConstraints(model)
		updateDBConstraints(model)
	}
	// Run init method on each model
	for _, model := range Registry.registryByTableName {
		if model.IsMixin() {
			continue
		}
		runInit(model)
	}

	// Drop DB tables that are not in the models
	for dbTable := range adapter.tables() {
		var modelExists bool
		for tableName, model := range Registry.registryByTableName {
			if dbTable != tableName || model.IsMixin() {
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

// buildSQLErrorSubstitutionMap populates the sqlErrors map of the
// model with the appropriate error message substitution
func buildSQLErrorSubstitutionMap(model *Model) {
	for sqlConstrName, sqlConstr := range model.sqlConstraints {
		model.sqlErrors[sqlConstrName] = sqlConstr.errorString
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

// updateDBSequences creates sequences in the DB from data in the registry.
func updateDBSequences() {
	adapter := adapters[db.DriverName()]
	// Create or alter boot sequences
	for _, sequence := range Registry.sequences {
		if !sequence.boot {
			continue
		}
		exists := false
		for _, dbSeq := range adapter.sequences("%_bootseq") {
			if sequence.JSON == dbSeq.Name {
				exists = true
			}
		}
		if !exists {
			adapter.createSequence(sequence.JSON, sequence.Increment, sequence.Start)
			continue
		}
		adapter.alterSequence(sequence.JSON, sequence.Increment, sequence.Start)
	}
	// Drop unused boot sequences
	for _, dbSeq := range adapter.sequences("%_bootseq") {
		var sequenceExists bool
		for _, sequence := range Registry.sequences {
			if sequence.JSON == dbSeq.Name {
				sequenceExists = true
				break
			}
		}
		if !sequenceExists {
			adapter.dropSequence(dbSeq.Name)
		}
	}
}

// createDBTable creates a table in the database from the given Model
// It only creates the primary key. Call updateDBColumns to create columns.
func createDBTable(m *Model) {
	adapter := adapters[db.DriverName()]
	var columns []string
	for colName, fi := range m.fields.registryByJSON {
		if colName == "id" || !fi.isStored() {
			continue
		}
		col := fmt.Sprintf("%s %s", colName, adapter.columnSQLDefinition(fi, false))
		columns = append(columns, col)
	}
	query := fmt.Sprintf(`
CREATE TABLE %s (
	id serial NOT NULL PRIMARY KEY`,
		adapter.quoteTableName(m.tableName))
	if len(columns) > 0 {
		query += ",\n\t" + strings.Join(columns, ",\n\t")
	}
	query += "\n)"
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
			continue
		}
		if dbColData.DataType != adapter.typeSQL(fi) {
			updateDBColumnDataType(fi)
		}
		if (dbColData.IsNullable == "NO" && !adapter.fieldIsNotNull(fi)) ||
			(dbColData.IsNullable == "YES" && adapter.fieldIsNotNull(fi)) {
			updateDBColumnNullable(fi)
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
	// Add column without not null
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ADD COLUMN %s %s
	`, adapter.quoteTableName(fi.model.tableName), fi.json, adapter.columnSQLDefinition(fi, true))
	dbExecuteNoTx(query)
	// Set default value if defined
	if fi.defaultFunc != nil {
		updateQuery := fmt.Sprintf(`
			UPDATE %s SET %s = ? WHERE %s IS NULL
		`, adapter.quoteTableName(fi.model.tableName), fi.json, fi.json)
		var defaultValue interface{}
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			defaultValue = fi.defaultFunc(env)
		})
		dbExecuteNoTx(updateQuery, defaultValue)
	}
	// Add not null if required
	updateDBColumnNullable(fi)
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
	query, _ = sanitizeQuery(query)
	_, err := db.Exec(query)
	if err != nil {
		log.Warn("unable to change NOT NULL constraint", "model", fi.model.name, "field", fi.name, "verb", verb)
	}
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
		case indexInDB && !fi.index:
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
