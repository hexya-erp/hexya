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

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/operator"
	"github.com/hexya-erp/hexya/hexya/tools/nbutils"
	"github.com/lib/pq"
)

type postgresAdapter struct{}

var pgOperators = map[operator.Operator]string{
	operator.Equals:         "= ?",
	operator.NotEquals:      "!= ?",
	operator.Contains:       "LIKE ?",
	operator.NotContains:    "NOT LIKE ?",
	operator.Like:           "LIKE ?",
	operator.IContains:      "ILIKE ?",
	operator.NotIContains:   "NOT ILIKE ?",
	operator.ILike:          "ILIKE ?",
	operator.In:             "IN (?)",
	operator.NotIn:          "NOT IN (?)",
	operator.Lower:          "< ?",
	operator.LowerOrEqual:   "<= ?",
	operator.Greater:        "> ?",
	operator.GreaterOrEqual: ">= ?",
}

var pgTypes = map[fieldtype.Type]string{
	fieldtype.Boolean:   "boolean",
	fieldtype.Char:      "character varying",
	fieldtype.Text:      "text",
	fieldtype.Date:      "date",
	fieldtype.DateTime:  "timestamp without time zone",
	fieldtype.Integer:   "integer",
	fieldtype.Float:     "numeric",
	fieldtype.HTML:      "text",
	fieldtype.Binary:    "bytea",
	fieldtype.Selection: "varchar",
	fieldtype.Many2One:  "integer",
	fieldtype.One2One:   "integer",
}

var pgDefaultValues = map[fieldtype.Type]string{
	fieldtype.Boolean:   "FALSE",
	fieldtype.Char:      "''",
	fieldtype.Text:      "''",
	fieldtype.Date:      "'0001-01-01'",
	fieldtype.DateTime:  "'0001-01-01 00:00:00'",
	fieldtype.Integer:   "0",
	fieldtype.Float:     "0.0",
	fieldtype.HTML:      "''",
	fieldtype.Binary:    "''",
	fieldtype.Selection: "''",
}

// operatorSQL returns the sql string and placeholders for the given DomainOperator
// Also modifies the given args to match the syntax of the operator.
func (d *postgresAdapter) operatorSQL(do operator.Operator, arg interface{}) (string, interface{}) {
	op := pgOperators[do]
	switch do {
	case operator.Contains, operator.IContains, operator.NotContains, operator.NotIContains:
		arg = fmt.Sprintf("%%%s%%", arg)
	}
	return op, arg
}

// typeSQL returns the sql type string for the given Field
func (d *postgresAdapter) typeSQL(fi *Field) string {
	typ, _ := pgTypes[fi.fieldType]
	return typ
}

// columnSQLDefinition returns the SQL type string, including columns constraints if any
func (d *postgresAdapter) columnSQLDefinition(fi *Field) string {
	var res string
	typ, ok := pgTypes[fi.fieldType]
	res = typ
	if !ok {
		log.Panic("Unknown column type", "type", fi.fieldType, "model", fi.model.name, "field", fi.name)
	}
	switch fi.fieldType {
	case fieldtype.Char:
		if fi.size > 0 {
			res = fmt.Sprintf("%s(%d)", res, fi.size)
		}
	case fieldtype.Float:
		emptyD := nbutils.Digits{}
		if fi.digits != emptyD {
			res = fmt.Sprintf("numeric(%d, %d)", fi.digits.Precision, fi.digits.Scale)
		}
	}
	if d.fieldIsNotNull(fi) {
		res += " NOT NULL"
	}

	defValue := d.fieldSQLDefault(fi)
	if defValue != "" && !fi.required {
		res += fmt.Sprintf(" DEFAULT %v", defValue)
	}

	if fi.unique || fi.fieldType == fieldtype.One2One {
		res += " UNIQUE"
	}
	return res
}

// fieldIsNull returns true if the given Field results in a
// NOT NULL column in database.
func (d *postgresAdapter) fieldIsNotNull(fi *Field) bool {
	if fi.fieldType.IsFKRelationType() {
		if fi.required {
			return true
		}
		return false
	}
	return true
}

// fieldSQLDefault returns the SQL default value of the Field
func (d *postgresAdapter) fieldSQLDefault(fi *Field) string {
	return pgDefaultValues[fi.fieldType]
}

// tables returns a map of table names of the database
func (d *postgresAdapter) tables() map[string]bool {
	var resList []string
	query := "SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')"
	if err := db.Select(&resList, query); err != nil {
		log.Panic("Unable to get list of tables from database", "error", err)
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
		log.Panic("Unable to get list of columns for table", "table", tableName, "error", err)
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

// constraintExists returns true if a constraint with the given name exists in the given table
func (d *postgresAdapter) constraintExists(name string) bool {
	query := fmt.Sprintf("SELECT COUNT(*) FROM pg_constraint WHERE conname = '%s'", name)
	var cnt int
	dbGetNoTx(&cnt, query)
	return cnt > 0
}

// constraints returns a list of all constraints matching the given SQL pattern
func (d *postgresAdapter) constraints(pattern string) []string {
	query := "SELECT conname FROM pg_constraint WHERE conname ILIKE ?"
	var res []string
	dbSelectNoTx(&res, query, pattern)
	return res
}

// createSequence creates a DB sequence with the given name
func (d *postgresAdapter) createSequence(name string) {
	query := fmt.Sprintf("CREATE SEQUENCE %s", name)
	dbExecuteNoTx(query)
}

// dropSequence drops the DB sequence with the given name
func (d *postgresAdapter) dropSequence(name string) {
	query := fmt.Sprintf("DROP SEQUENCE IF EXISTS %s", name)
	dbExecuteNoTx(query)
}

// nextSequenceValue returns the next value of the given given sequence
func (d *postgresAdapter) nextSequenceValue(name string) int64 {
	query := fmt.Sprintf("SELECT nextval('%s')", name)
	var val int64
	dbGetNoTx(&val, query)
	return val
}

// sequences returns a list of all sequences matching the given SQL pattern
func (d *postgresAdapter) sequences(pattern string) []string {
	query := "SELECT sequence_name FROM information_schema.sequences WHERE sequence_name ILIKE ?"
	var res []string
	dbSelectNoTx(&res, query, pattern)
	return res
}

// setTransactionIsolation returns the SQL string to set the
// transaction isolation level to serializable
func (d *postgresAdapter) setTransactionIsolation() string {
	return "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE"
}

// childrenIdsQuery returns a query that finds all descendant of the given
// a record from table including itself. The query has a placeholder for the
// record's ID
func (d *postgresAdapter) childrenIdsQuery(table string) string {
	res := fmt.Sprintf(`
WITH RECURSIVE "recursive_query_children_ids" AS
(
	SELECT  id
	FROM    %s "m1"
	WHERE   id = ?
UNION ALL
	SELECT  "m2".id
	FROM    %s "m2"
	JOIN    "recursive_query_children_ids"
	ON      "m2".parent_id = "recursive_query_children_ids".id
)
SELECT  id
FROM    recursive_query_children_ids`, d.quoteTableName(table), d.quoteTableName(table))
	return res
}

// substituteErrorMessage substitutes the given error's message by newMsg
func (d *postgresAdapter) substituteErrorMessage(err error, newMsg string) error {
	pgError, ok := err.(*pq.Error)
	if !ok {
		return err
	}
	pgError.Message = newMsg
	return pgError
}

// isSerializationError returns true if the given error is a serialization error
// and that the failed transaction should be retried.
func (d *postgresAdapter) isSerializationError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Class() == "40" {
		return true
	}
	return false
}

var _ dbAdapter = new(postgresAdapter)
