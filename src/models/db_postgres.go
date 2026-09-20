// Copyright 2016 Nicolas Piganeau. All Rights Reserved.
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
	"strings"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
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
	fieldtype.Selection: "character varying",
	fieldtype.Many2One:  "integer",
	fieldtype.One2One:   "integer",
	fieldtype.JSON:      "jsonb",
}

// pgContextedType is the type of the columns holding contexted values
const pgContextedType = "jsonb"

// pgCasts is the cast to apply to the values of contexted fields when they are
// written to or read from their jsonb column.
var pgCasts = map[fieldtype.Type]string{
	fieldtype.Boolean:  "::boolean",
	fieldtype.Date:     "::date",
	fieldtype.DateTime: "::timestamp without time zone",
	fieldtype.Integer:  "::integer",
	fieldtype.Float:    "::numeric",
	fieldtype.Many2One: "::integer",
	fieldtype.One2One:  "::integer",
	fieldtype.Binary:   "::bytea",
}

// connectionString returns the connection string for the given parameters
func (d *postgresAdapter) connectionString(params ConnectionParams) string {
	connectString := fmt.Sprintf("dbname=%s", params.DBName)
	if params.SSLMode != "" {
		connectString += fmt.Sprintf(" sslmode=%s", params.SSLMode)
	}
	if params.SSLCert != "" {
		connectString += fmt.Sprintf(" sslcert=%s", params.SSLCert)
	}
	if params.SSLKey != "" {
		connectString += fmt.Sprintf(" sslkey=%s", params.SSLKey)
	}
	if params.SSLCA != "" {
		connectString += fmt.Sprintf(" sslrootcert=%s", params.SSLCA)
	}
	if params.User != "" {
		connectString += fmt.Sprintf(" user=%s", params.User)
	}
	if params.Password != "" {
		connectString += fmt.Sprintf(" password=%s", params.Password)
	}
	if params.Host != "" {
		connectString += fmt.Sprintf(" host=%s", params.Host)
	}
	if params.Port != "" && params.Port != "5432" {
		connectString += fmt.Sprintf(" port=%s", params.Port)
	}
	return connectString
}

// operatorSQL returns the sql string and placeholders for the given DomainOperator
// Also modifies the given args to match the syntax of the operator.
func (d *postgresAdapter) operatorSQL(do operator.Operator, arg any) (string, any) {
	op := pgOperators[do]
	switch do {
	case operator.Contains, operator.IContains, operator.NotContains, operator.NotIContains:
		arg = fmt.Sprintf("%%%s%%", arg)
	}
	return op, arg
}

// typeSQL returns the sql type string for the given Field
func (d *postgresAdapter) typeSQL(fi *Field) string {
	if fi.isContextedField() {
		return pgContextedType
	}
	typ, _ := pgTypes[fi.fieldType]
	return typ
}

// columnSQLDefinition returns the SQL type string, including columns constraints if any
//
// If null is true, then the column will be nullable, whatever the field defines
func (d *postgresAdapter) columnSQLDefinition(fi *Field, null bool) string {
	if fi.isContextedField() {
		// Contexted values are stored in a jsonb document which holds the value
		// of the field for each context. Such column cannot be constrained,
		// nor be given a SQL UNIQUE constraint: uniqueness of unique contexted
		// fields is checked per context by the ORM.
		return pgContextedType
	}
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
	if d.fieldIsNotNull(fi) && !null {
		res += " NOT NULL"
	}

	if fi.unique || fi.fieldType == fieldtype.One2One {
		res += " UNIQUE"
	}
	return res
}

// fieldIsNull returns true if the given Field results in a
// NOT NULL column in database.
func (d *postgresAdapter) fieldIsNotNull(fi *Field) bool {
	if fi.isContextedField() {
		return false
	}
	if fi.required {
		return true
	}
	return false
}

// contextedValueSQL returns the SQL expression that returns the value of the
// given contexted field for the given context values.
//
// The returned expression looks up each possible branch of the contexted value
// document, from the most specific to the most generic, and returns the first
// one that is set.
func (d *postgresAdapter) contextedValueSQL(fi *Field, colExpr string, ctxValues map[string]string) string {
	paths := contextedPaths(fi.contextNames(), ctxValues)
	exprs := make([]string, len(paths))
	for i, path := range paths {
		var expr strings.Builder
		expr.WriteString(colExpr)
		for _, key := range path[:len(path)-1] {
			expr.WriteString("->")
			expr.WriteString(pgQuoteString(key))
		}
		expr.WriteString("->>")
		expr.WriteString(pgQuoteString(ctxValueKey))
		exprs[i] = expr.String()
	}
	res := exprs[0]
	if len(exprs) > 1 {
		res = fmt.Sprintf("COALESCE(%s)", strings.Join(exprs, ", "))
	}
	if cast := pgCasts[fi.fieldType]; cast != "" {
		res = fmt.Sprintf("(%s)%s", res, cast)
	}
	return res
}

// contextedSetSQL returns the SQL expression that sets the value of the given
// path in the contexted value document given by docExpr.
func (d *postgresAdapter) contextedSetSQL(fi *Field, docExpr, lookupExpr string, path []string) string {
	res := fmt.Sprintf("COALESCE(%s, %s)", docExpr, d.emptyContextedValueSQL())
	// We must create the intermediate nodes of the path ourselves since
	// jsonb_set only creates the last one.
	for i := 1; i < len(path); i++ {
		node := pgArrayLiteral(path[:i])
		res = fmt.Sprintf("jsonb_set(%s, %s, COALESCE(%s#>%s, %s), true)",
			res, node, lookupExpr, node, d.emptyContextedValueSQL())
	}
	cast := pgCasts[fi.fieldType]
	if cast == "" {
		// to_jsonb needs to know the type of its argument
		cast = "::text"
	}
	return fmt.Sprintf("jsonb_set(%s, %s, COALESCE(to_jsonb(?%s), 'null'::jsonb), true)",
		res, pgArrayLiteral(path), cast)
}

// emptyContextedValueSQL returns the SQL expression of an empty contexted
// value document.
func (d *postgresAdapter) emptyContextedValueSQL() string {
	return "'{}'::jsonb"
}

// contextedIndexSQL returns the statement to create an index with the given
// name on all the context values of the given contexted field.
//
// Text-ish fields are indexed with a trigram index on the array of all the
// values of the document, so that a single index serves every context. Such
// an index only speeds up 'like' and 'ilike' conditions.
//
// This method creates the pg_trgm extension if needed and returns an empty
// string if it is not available.
func (d *postgresAdapter) contextedIndexSQL(fi *Field, table, indexName string) string {
	if !pgTrigramTypes[fi.fieldType] {
		return fmt.Sprintf(`CREATE INDEX %s ON %s USING gin (%s)`,
			indexName, d.quoteTableName(table), fi.json)
	}
	if err := dbTryExecuteNoTx("CREATE EXTENSION IF NOT EXISTS pg_trgm"); err != nil {
		log.Warn("Unable to create the pg_trgm extension. Contexted field will not be indexed",
			"model", fi.model.name, "field", fi.name, "error", err)
		return ""
	}
	return fmt.Sprintf(`CREATE INDEX %s ON %s USING gin ((jsonb_path_query_array(%s, '$.**._'::jsonpath)::text) gin_trgm_ops)`,
		indexName, d.quoteTableName(table), fi.json)
}

// pgTrigramTypes are the field types whose contexted values are indexed with
// a trigram index.
var pgTrigramTypes = map[fieldtype.Type]bool{
	fieldtype.Char:      true,
	fieldtype.Text:      true,
	fieldtype.HTML:      true,
	fieldtype.Selection: true,
}

// pgQuoteString returns the given string as a single quoted SQL literal
func pgQuoteString(str string) string {
	return fmt.Sprintf("'%s'", strings.Replace(str, "'", "''", -1))
}

// pgArrayLiteral returns the given strings as a postgres text array literal
func pgArrayLiteral(values []string) string {
	res := make([]string, len(values))
	for i, val := range values {
		val = strings.Replace(val, `\`, `\\`, -1)
		val = strings.Replace(val, `"`, `\"`, -1)
		res[i] = fmt.Sprintf(`"%s"`, val)
	}
	return pgQuoteString(fmt.Sprintf("{%s}", strings.Join(res, ",")))
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
func (d *postgresAdapter) createSequence(name string, increment, start int64) {
	query := fmt.Sprintf("CREATE SEQUENCE %s INCREMENT BY %d START WITH %d", name, increment, start)
	dbExecuteNoTx(query)
}

// dropSequence drops the DB sequence with the given name
func (d *postgresAdapter) dropSequence(name string) {
	query := fmt.Sprintf("DROP SEQUENCE IF EXISTS %s", name)
	dbExecuteNoTx(query)
}

// alterSequence modifies the DB sequence given by name
func (d *postgresAdapter) alterSequence(name string, increment, restart int64) {
	query := fmt.Sprintf(`ALTER SEQUENCE %s`, name)
	if increment != 0 {
		query += fmt.Sprintf(` INCREMENT BY %d`, increment)
	}
	if restart != 0 {
		query += fmt.Sprintf(` RESTART WITH %d`, restart)
	}
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
func (d *postgresAdapter) sequences(pattern string) []seqData {
	query := "SELECT sequence_name, start_value, increment FROM information_schema.sequences WHERE sequence_name ILIKE ?"
	var res []seqData
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
