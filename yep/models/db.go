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
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/npiganeau/yep/yep/models/operator"
	"github.com/npiganeau/yep/yep/tools/logging"
)

var (
	db       *sqlx.DB
	adapters map[string]dbAdapter
)

// A ColumnData holds information from the db schema about one column
type ColumnData struct {
	ColumnName    string
	DataType      string
	IsNullable    string
	ColumnDefault sql.NullString
}

type dbAdapter interface {
	// operatorSQL returns the sql string and placeholders for the given DomainOperator
	operatorSQL(operator.Operator, interface{}) (string, interface{})
	// typeSQL returns the SQL type string, including columns constraints if any
	typeSQL(fi *fieldInfo) string
	// columnSQLDefinition returns the SQL type string, including columns constraints if any
	columnSQLDefinition(fi *fieldInfo) string
	// fieldSQLDefault returns the SQL default value of the fieldInfo
	fieldSQLDefault(fi *fieldInfo) string
	// tables returns a map of table names of the database
	tables() map[string]bool
	// columns returns a list of ColumnData for the given tableName
	columns(tableName string) map[string]ColumnData
	// fieldIsNull returns true if the given fieldInfo results in a
	// NOT NULL column in database.
	fieldIsNotNull(fi *fieldInfo) bool
	// quoteTableName returns the given table name with sql quotes
	quoteTableName(string) string
	// indexExists returns true if an index with the given name exists in the given table
	indexExists(table string, name string) bool
	// setTransactionIsolation returns the SQL string to set the transaction isolation
	// level to serializable
	setTransactionIsolation() string
}

// registerDBAdapter adds a adapter to the adapters registry
// name of the adapter should match the database/sql driver name
func registerDBAdapter(name string, adapter dbAdapter) {
	adapters[name] = adapter
}

// Cursor is a wrapper around a database transaction
type Cursor struct {
	tx *sqlx.Tx
}

// Execute a query without returning any rows. It panics in case of error.
// The args are for any placeholder parameters in the query.
func (c *Cursor) Execute(query string, args ...interface{}) sql.Result {
	return dbExecute(c.tx, query, args...)
}

// Get queries a row into the database and maps the result into dest.
// The query must return only one row. Get panics on errors
func (c *Cursor) Get(dest interface{}, query string, args ...interface{}) {
	dbGet(c.tx, dest, query, args...)
}

// Select queries multiple rows and map the result into dest which must be a slice.
// Select panics on errors.
func (c *Cursor) Select(dest interface{}, query string, args ...interface{}) {
	dbSelect(c.tx, dest, query, args...)
}

// newCursor returns a new db cursor on the given database
func newCursor(db *sqlx.DB) *Cursor {
	adapter := adapters[db.DriverName()]
	tx := db.MustBegin()
	dbExecute(tx, adapter.setTransactionIsolation())
	return &Cursor{
		tx: tx,
	}
}

// DBConnect is a wrapper around sqlx.MustConnect
// It connects to a database using the given driver and
// connection data.
func DBConnect(driver, connData string) {
	db = sqlx.MustConnect(driver, connData)
	log.Info("Connected to database", "driver", driver, "connData", connData)
}

// dbExecute is a wrapper around sqlx.MustExec
// It executes a query that returns no row
func dbExecute(cr *sqlx.Tx, query string, args ...interface{}) sql.Result {
	query, args = sanitizeQuery(query, args...)
	t := time.Now()
	res := cr.MustExec(query, args...)
	logSQLResult(nil, t, query, args...)
	return res
}

// dbExecuteNoTx simply executes the given query in the database without any transaction
func dbExecuteNoTx(query string, args ...interface{}) sql.Result {
	query, args = sanitizeQuery(query, args...)
	t := time.Now()
	res := db.MustExec(query, args...)
	logSQLResult(nil, t, query, args...)
	return res
}

// dbGet is a wrapper around sqlx.Get
// It gets the value of a single row found by the given query and arguments
// It panics in case of error
func dbGet(cr *sqlx.Tx, dest interface{}, query string, args ...interface{}) {
	query, args = sanitizeQuery(query, args...)
	t := time.Now()
	err := cr.Get(dest, query, args...)
	logSQLResult(err, t, query, args)
}

// dbGetNoTx is a wrapper around sqlx.Get outside a transaction
// It gets the value of a single row found by the
// given query and arguments
func dbGetNoTx(dest interface{}, query string, args ...interface{}) {
	query, args = sanitizeQuery(query, args...)
	t := time.Now()
	err := db.Get(dest, query, args...)
	logSQLResult(err, t, query, args)
}

// dbSelect is a wrapper around sqlx.Select
// It gets the value of a multiple rows found by the given query and arguments
// dest must be a slice. It panics in case of error
func dbSelect(cr *sqlx.Tx, dest interface{}, query string, args ...interface{}) {
	query, args = sanitizeQuery(query, args...)
	t := time.Now()
	err := cr.Select(dest, query, args...)
	logSQLResult(err, t, query, args)
}

// dbQuery is a wrapper around sqlx.Queryx
// It returns a sqlx.Rowsx found by the given query and arguments
// It panics in case of error
func dbQuery(cr *sqlx.Tx, query string, args ...interface{}) *sqlx.Rows {
	query, args = sanitizeQuery(query, args...)
	t := time.Now()
	rows, err := cr.Queryx(query, args...)
	logSQLResult(err, t, query, args)
	return rows
}

// sanitizeQuery calls 'In' expansion and 'Rebind' on the given query and
// returns the new values to use. It panics in case of error
func sanitizeQuery(query string, args ...interface{}) (string, []interface{}) {
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		logging.LogAndPanic(log, "Unable to expand 'IN' statement", "error", err, "query", query, "args", args)
	}
	query = sqlx.Rebind(sqlx.BindType(db.DriverName()), query)
	return query, args
}

// Log the result of the given sql query started at start time with the
// given args, and error. This function panics after logging if error is not nil.
func logSQLResult(err error, start time.Time, query string, args ...interface{}) {
	logCtx := log.New("query", query, "args", args, "duration", time.Now().Sub(start))
	if err != nil {
		logging.LogAndPanic(logCtx, "Error while executing query", "error", err, "query", query, "args", args)
	}
	logCtx.Debug("Query executed")
}
