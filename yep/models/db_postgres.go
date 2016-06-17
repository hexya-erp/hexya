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
)

type postgresAdapter struct{}

var pgOperators = map[DomainOperator]string{
	OPERATOR_EQUALS:        "= ?",
	OPERATOR_NOT_EQUALS:    "!= ?",
	OPERATOR_LIKE:          "LIKE %?%",
	OPERATOR_NOT_LIKE:      "NOT LIKE %?%",
	OPERATOR_LIKE_PATTERN:  "LIKE ?",
	OPERATOR_ILIKE:         "ILIKE %?%",
	OPERATOR_NOT_ILIKE:     "NOT ILIKE %?%",
	OPERATOR_ILIKE_PATTERN: "ILIKE ?",
	OPERATOR_IN:            "IN (?)",
	OPERATOR_NOT_IN:        "NOT IN (?)",
	OPERATOR_LOWER:         "< ?",
	OPERATOR_LOWER_EQUAL:   "< ?",
	OPERATOR_GREATER:       "> ?",
	OPERATOR_GREATER_EQUAL: ">= ?",
	//OPERATOR_CHILD_OF: "",
}

var pgTypes = map[tools.FieldType]string{
	tools.BOOLEAN:   "bool",
	tools.CHAR:      "varchar",
	tools.TEXT:      "text",
	tools.DATE:      "date",
	tools.DATETIME:  "timestamp without time zone",
	tools.INTEGER:   "integer",
	tools.FLOAT:     "double precision",
	tools.HTML:      "text",
	tools.BINARY:    "bytea",
	tools.SELECTION: "varchar",
	//tools.REFERENCE: "varchar",
	tools.MANY2ONE: "integer",
	tools.ONE2ONE:  "integer",
}

// operatorSQL returns the sql string and placeholders for the given DomainOperator
func (d *postgresAdapter) operatorSQL(do DomainOperator) string {
	return pgOperators[do]
}

// typeSQL returns the SQL type string, including columns constraints if any
func (d *postgresAdapter) typeSQL(fi *fieldInfo) string {
	res, ok := pgTypes[fi.fieldType]
	if !ok {
		panic(fmt.Errorf("Unknown column type `%s`", fi.fieldType))
	}
	switch fi.fieldType {
	case tools.CHAR:
		if fi.size > 0 {
			res = fmt.Sprintf("%s(%d)", res, fi.size)
		}
	case tools.FLOAT:
		emptyD := tools.Digits{}
		if fi.digits != emptyD {
			res = fmt.Sprintf("numeric(%d, %d)", (fi.digits)[0], (fi.digits)[1])
		}
	}
	if fi.required {
		res += " NOT NULL"
	}
	if fi.unique || fi.fieldType == tools.ONE2ONE {
		res += " UNIQUE"
	}
	return res
}

// tables returns a slice of table names of the database
func (d *postgresAdapter) tables() []string {
	var res []string
	query := "SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')"
	if err := db.Select(&res, query); err != nil {
		panic(fmt.Errorf("Unable to get list of tables from database"))
	}
	return res
}

var _ dbAdapter = new(postgresAdapter)
