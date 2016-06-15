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

import "fmt"

type SQLParams []interface{}

// Extend returns a new SQLParams with both params of this SQLParams and
// of p2 SQLParams.
func (p SQLParams) Extend(p2 SQLParams) SQLParams {
	pi := []interface{}(p)
	pi2 := []interface{}(p2)
	res := append(pi, pi2...)
	return SQLParams(res)
}

type Query struct {
	recordSet *RecordSet
	cond      *Condition
	related   []string
	relDepth  int
	limit     int
	offset    int
	groups    []string
	orders    []string
	distinct  bool
}

// sqlWhereClause returns the sql string and parameters corresponding to the
// WHERE clause of this Query
func (q *Query) sqlWhereClause() (string, SQLParams) {
	sql, args := q.cond.sqlClause()
	sql = "WHERE " + sql
	return sql, args
}

// sqlLimitClause returns the sql string for the LIMIT and OFFSET clauses
// of this Query
func (q *Query) sqlLimitOffsetClause() string {
	var res string
	if q.limit > 0 {
		res = fmt.Sprintf(`LIMIT %d `, q.limit)
	}
	if q.offset > 0 {
		res += fmt.Sprintf(`OFFSET %d`, q.offset)
	}
	return res
}

// deleteQuery returns the SQL query string and parameters to delete
// the rows pointed at by this Query object.
func (q *Query) deleteQuery() (string, SQLParams) {
	sql, args := q.sqlWhereClause()
	delQuery := fmt.Sprintf(`DELETE FROM %s %s`, q.recordSet.mi.tableName, sql)
	return delQuery, args
}

// countQuery returns the SQL query string and parameters to count
// the rows pointed at by this Query object.
func (q *Query) countQuery() (string, SQLParams) {
	sql, args := q.selectQuery()
	delQuery := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) foo`, sql)
	return delQuery, args
}

// selectQuery returns the SQL query string and parameters to retrieve
// the rows pointed at by this Query object.
func (q *Query) selectQuery() (string, SQLParams) {
	sql, args := q.sqlWhereClause()
	sql += q.sqlLimitOffsetClause()
	delQuery := fmt.Sprintf(`SELECT * FROM %s`, sql)
	return delQuery, args
}
