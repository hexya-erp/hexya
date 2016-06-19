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

	"github.com/jmoiron/sqlx"
)

var (
	db       *sqlx.DB
	adapters map[string]dbAdapter
)

type dbAdapter interface {
	// operatorSQL returns the sql string and placeholders for the given DomainOperator
	operatorSQL(DomainOperator) string
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
}

// registerDBAdapter adds a adapter to the adapters registry
// name of the adapter should match the database/sql driver name
func registerDBAdapter(name string, adapter dbAdapter) {
	adapters[name] = adapter
}

// DBConnect is a wrapper around sqlx.MustConnect
// It connects to a database using the given driver and
// connection data.
func DBConnect(driver, connData string) {
	// TODO Add log
	db = sqlx.MustConnect(driver, connData)
}

// DBExecute is a wrapper around sqlx.MustExec
// It executes a query that returns no row
func DBExecute(cr *sqlx.Tx, query string, args ...interface{}) sql.Result {
	// TODO Add SQL debug logging here
	return cr.MustExec(query, args...)
}

// dbExecuteNoTx simply executes the given query in the database without any transaction
func dbExecuteNoTx(query string, args ...interface{}) sql.Result {
	// TODO Add SQL debug logging
	return db.MustExec(query, args...)
}

// DBGet is a wrapper around sqlx.Get
// It gets the value of a single row found by the
// given query and arguments
func DBGet(cr *sqlx.Tx, dest interface{}, query string, args ...interface{}) error {
	// TODO Add SQL debug logging here
	return cr.Get(dest, query, args...)
}

// dbGetNoTx is a wrapper around sqlx.Get outside a transaction
// It gets the value of a single row found by the
// given query and arguments
func dbGetNoTx(dest interface{}, query string, args ...interface{}) error {
	// TODO Add SQL debug logging
	return db.Get(dest, query, args...)
}
