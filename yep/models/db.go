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
	DB      *sqlx.DB
	drivers map[string]dbDriver
)

type dbDriver interface {
	// operatorSQL returns the sql string and placeholders for the given DomainOperator
	operatorSQL(DomainOperator) string
}

func registerDBDriver(name string, driver dbDriver) {
	drivers[name] = driver
}

// DBExecute is a wrapper around sqlx.MustExec
// It executes a query that returns no row
func DBExecute(cr *sqlx.Tx, query string, args ...interface{}) sql.Result {
	// TODO Add SQL debug logging here
	return cr.MustExec(query, args...)
}

// DBGet is a wrapper around sqlx.Get
// It gets the value of a single row found by the
// given query and arguments
func DBGet(cr *sqlx.Tx, dest interface{}, query string, args ...interface{}) error {
	// TODO Add SQL debug logging here
	return cr.Get(dest, query, args)
}
