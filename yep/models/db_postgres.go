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

	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/tools/logging"
)

type postgresAdapter struct{}

var pgOperators = map[DomainOperator]string{
	OPERATOR_EQUALS:        "= ?",
	OPERATOR_NOT_EQUALS:    "!= ?",
	OPERATOR_LIKE:          "LIKE ?",
	OPERATOR_NOT_LIKE:      "NOT LIKE ?",
	OPERATOR_LIKE_PATTERN:  "LIKE ?",
	OPERATOR_ILIKE:         "ILIKE ?",
	OPERATOR_NOT_ILIKE:     "NOT ILIKE ?",
	OPERATOR_ILIKE_PATTERN: "ILIKE ?",
	OPERATOR_IN:            "IN (?)",
	OPERATOR_NOT_IN:        "NOT IN (?)",
	OPERATOR_LOWER:         "< ?",
	OPERATOR_LOWER_EQUAL:   "< ?",
	OPERATOR_GREATER:       "> ?",
	OPERATOR_GREATER_EQUAL: ">= ?",
	//OPERATOR_CHILD_OF: "",
}

var pgTypes = map[types.FieldType]string{
	types.Boolean:   "bool",
	types.Char:      "varchar",
	types.Text:      "text",
	types.Date:      "date",
	types.DateTime:  "timestamp without time zone",
	types.Integer:   "integer",
	types.Float:     "double precision",
	types.HTML:      "text",
	types.Binary:    "bytea",
	types.Selection: "varchar",
	//types.REFERENCE: "varchar",
	types.Many2One: "integer",
	types.One2One:  "integer",
}

var pgDefaultValues = map[types.FieldType]string{
	types.Boolean:   "FALSE",
	types.Char:      "''",
	types.Text:      "''",
	types.Date:      "'0001-01-01'",
	types.DateTime:  "'0001-01-01 00:00:00'",
	types.Integer:   "0",
	types.Float:     "0.0",
	types.HTML:      "''",
	types.Binary:    "''",
	types.Selection: "''",
	//types.REFERENCE: "''",
}

// operatorSQL returns the sql string and placeholders for the given DomainOperator
// Also modifies the given args to match the syntax of the operator.
func (d *postgresAdapter) operatorSQL(do DomainOperator, arg interface{}) (string, interface{}) {
	op := pgOperators[do]
	switch do {
	case OPERATOR_LIKE, OPERATOR_ILIKE, OPERATOR_NOT_LIKE, OPERATOR_NOT_ILIKE:
		arg = fmt.Sprintf("%%%s%%", arg)
	}
	return op, arg
}

// typeSQL returns the sql type string for the given fieldInfo
func (d *postgresAdapter) typeSQL(fi *fieldInfo) string {
	typ, _ := pgTypes[fi.fieldType]
	return typ
}

// columnSQLDefinition returns the SQL type string, including columns constraints if any
func (d *postgresAdapter) columnSQLDefinition(fi *fieldInfo) string {
	var res string
	typ, ok := pgTypes[fi.fieldType]
	res = typ
	if !ok {
		logging.LogAndPanic(log, "Unknown column type", "type", fi.fieldType, "model", fi.model.name, "field", fi.name)
	}
	switch fi.fieldType {
	case types.Char:
		if fi.size > 0 {
			res = fmt.Sprintf("%s(%d)", res, fi.size)
		}
	case types.Float:
		emptyD := types.Digits{}
		if fi.digits != emptyD {
			res = fmt.Sprintf("numeric(%d, %d)", (fi.digits)[0], (fi.digits)[1])
		}
	}
	if d.fieldIsNotNull(fi) {
		res += " NOT NULL"
	}

	defValue := d.fieldSQLDefault(fi)
	if defValue != "" && !fi.required {
		res += fmt.Sprintf(" DEFAULT %v", defValue)
	}

	if fi.unique || fi.fieldType == types.One2One {
		res += " UNIQUE"
	}
	return res
}

// fieldIsNull returns true if the given fieldInfo results in a
// NOT NULL column in database.
func (d *postgresAdapter) fieldIsNotNull(fi *fieldInfo) bool {
	if fi.fieldType.IsStoredRelationType() {
		if fi.required {
			return true
		}
	} else {
		return true
	}
	return false
}

// fieldSQLDefault returns the SQL default value of the fieldInfo
func (d *postgresAdapter) fieldSQLDefault(fi *fieldInfo) string {
	return pgDefaultValues[fi.fieldType]
}

// tables returns a map of table names of the database
func (d *postgresAdapter) tables() map[string]bool {
	var resList []string
	query := "SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')"
	if err := db.Select(&resList, query); err != nil {
		logging.LogAndPanic(log, "Unable to get list of tables from database", "error", err)
	}
	res := make(map[string]bool, len(resList))
	for _, tableName := range resList {
		res[tableName] = true
	}
	return res
}

// quoteTableName returns the given table name with sql quotes
func (d *postgresAdapter) quoteTableName(tableName string) string {
	return fmt.Sprintf(`"%s"`, tableName)
}

// columns returns a list of ColumnData for the given tableName
func (d *postgresAdapter) columns(tableName string) map[string]ColumnData {
	query := fmt.Sprintf(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema') AND table_name = '%s'
	`, tableName)
	var colData []ColumnData
	if err := db.Select(&colData, query); err != nil {
		logging.LogAndPanic(log, "Unable to get list of columns for table", "table", tableName, "error", err)
	}
	res := make(map[string]ColumnData, len(colData))
	for _, col := range colData {
		res[col.ColumnName] = col
	}
	return res
}

// indexExists returns true if an index with the given name exists in the given table
func (d *postgresAdapter) indexExists(table string, name string) bool {
	query := fmt.Sprintf("SELECT COUNT(*) FROM pg_indexes WHERE tablename = '%s' AND indexname = '%s'", table, name)
	var cnt int
	dbGetNoTx(&cnt, query)
	return cnt > 0
}

// setTransactionIsolation returns the SQL string to set the
// transaction isolation level to serializable
func (d *postgresAdapter) setTransactionIsolation() string {
	return "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE"
}

var _ dbAdapter = new(postgresAdapter)
