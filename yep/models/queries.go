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
	"strings"
)

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
	sql, args := q.conditionSQLClause(q.cond)
	sql = "WHERE " + sql
	return sql, args
}

// sqlClauses returns the sql string and parameters corresponding to the
// WHERE clause of this Condition.
func (q *Query) conditionSQLClause(c *Condition) (string, SQLParams) {
	if c.IsEmpty() {
		return "", SQLParams{}
	}
	var (
		sql  string
		args SQLParams
	)

	first := true
	for _, val := range c.params {
		vSQL, vArgs := q.condValueSQLClause(val, first)
		first = false
		sql += vSQL
		args = args.Extend(vArgs)
	}
	return sql, args
}

// sqlClause returns the sql WHERE clause for this condValue.
// If 'first' is given and true, then the sql clause is not prefixed with
// 'AND' and panics if isOr is true.
func (q *Query) condValueSQLClause(cv condValue, first ...bool) (string, SQLParams) {
	var (
		sql     string
		args    SQLParams
		isFirst bool
		adapter dbAdapter = adapters[db.DriverName()]
	)
	if len(first) > 0 {
		isFirst = first[0]
	}
	if cv.isOr {
		if isFirst {
			panic(fmt.Errorf("First WHERE clause cannot be OR"))
		}
		sql += "OR "
	} else if !isFirst {
		sql += "AND "
	}
	if cv.isNot {
		sql += "NOT "
	}

	if cv.isCond {
		subSQL, subArgs := q.conditionSQLClause(cv.cond)
		sql += fmt.Sprintf(`(%s) `, subSQL)
		args = args.Extend(subArgs)
	} else {
		exprs := columnizeExpr(q.recordSet.mi, cv.exprs)
		field := q.joinedFieldExpression(exprs)
		sql += fmt.Sprintf(`%s %s `, field, adapter.operatorSQL(cv.operator))
		args = cv.args
	}
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
	adapter := adapters[db.DriverName()]
	sql, args := q.sqlWhereClause()
	delQuery := fmt.Sprintf(`DELETE FROM %s %s`, adapter.quoteTableName(q.recordSet.mi.tableName), sql)
	return delQuery, args
}

// countQuery returns the SQL query string and parameters to count
// the rows pointed at by this Query object.
func (q *Query) countQuery() (string, SQLParams) {
	sql, args := q.selectQuery([]string{"id"})
	delQuery := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) foo`, sql)
	return delQuery, args
}

// selectQuery returns the SQL query string and parameters to retrieve
// the rows pointed at by this Query object.
// fields is the list of fields to retrieve. Each field is a dot-separated
// expression pointing at the field, either as names or columns
// (e.g. 'User.Name' or 'user_id.name')
func (q *Query) selectQuery(fields []string) (string, SQLParams) {
	// Get all expressions, first given by fields
	fieldExprs := make([][]string, len(fields))
	for i, f := range fields {
		fieldExprs[i] = columnizeExpr(q.recordSet.mi, strings.Split(f, ExprSep))
	}
	// Then given by condition
	fExprs := append(fieldExprs, q.cond.getAllExpressions(q.recordSet.mi)...)
	// Build up the query
	// Fields
	fieldsSQL := q.fieldsSQL(fieldExprs)
	// Tables
	tablesSQL := q.tablesSQL(fExprs)
	// Where clause and args
	whereSQL, args := q.sqlWhereClause()
	whereSQL += q.sqlLimitOffsetClause()
	selQuery := fmt.Sprintf(`SELECT %s FROM %s %s`, fieldsSQL, tablesSQL, whereSQL)
	return selQuery, args
}

// fieldsSQL returns the SQL string for the given field expressions
// parameter must be with the following format (column names):
// [['user_id', 'name'] ['id'] ['profile_id', 'age']]
func (q *Query) fieldsSQL(fieldExprs [][]string) string {
	fStr := make([]string, len(fieldExprs))
	for i, field := range fieldExprs {
		fStr[i] = q.joinedFieldExpression(field)
	}
	return strings.Join(fStr, ", ")
}

// joinedFieldExpression joins the given expressions into a fields sql string
// ['profile_id' 'user_id' 'name'] => "profiles__users".name
// ['age'] => "mytable".age
func (q *Query) joinedFieldExpression(exprs []string) string {
	joins := q.generateTableJoins(exprs)
	num := len(joins)
	return fmt.Sprintf("%s.%s", joins[num-1].alias, exprs[num-1])
}

// generateTableJoins transforms a list of fields expression into a list of tableJoins
// ['user_id' 'profile_id' 'age'] => []tableJoins{CurrentTable User Profile}
func (q *Query) generateTableJoins(fieldExprs []string) []tableJoin {
	adapter := adapters[db.DriverName()]
	var joins []tableJoin

	// Create the tableJoin for the current table
	currentTableName := adapter.quoteTableName(q.recordSet.mi.tableName)
	currentTJ := tableJoin{
		tableName: currentTableName,
		joined:    false,
		alias:     currentTableName,
	}
	joins = append(joins, currentTJ)

	curMI := q.recordSet.mi
	curTJ := &currentTJ
	alias := curMI.tableName
	for _, expr := range fieldExprs {
		fi, ok := curMI.fields.get(expr)
		if !ok {
			panic(fmt.Errorf("Unparsable Expression: `%s`", strings.Join(fieldExprs, ExprSep)))
		}
		if fi.relatedModel == nil {
			break
		}
		var innerJoin bool
		if fi.required {
			innerJoin = true
		}
		linkedTableName := adapter.quoteTableName(fi.relatedModel.tableName)
		alias = fmt.Sprintf("%s%s%s", alias, sqlSep, fi.relatedModel.tableName)
		nextTJ := tableJoin{
			tableName:  linkedTableName,
			joined:     true,
			innerJoin:  innerJoin,
			field:      "id",
			otherTable: curTJ,
			otherField: expr,
			alias:      adapter.quoteTableName(alias),
		}
		joins = append(joins, nextTJ)
		curMI = fi.relatedModel
		curTJ = &nextTJ
	}
	return joins
}

// tablesSQL returns the SQL string for the FROM clause of our SQL query
// including all joins if any for the given expressions.
func (q *Query) tablesSQL(fExprs [][]string) string {
	var res string
	joinsMap := make(map[string]bool)
	// Get a list of unique table joins (by alias)
	for _, f := range fExprs {
		tJoins := q.generateTableJoins(f)
		for _, j := range tJoins {
			if _, exists := joinsMap[j.alias]; !exists {
				joinsMap[j.alias] = true
				res += j.sqlString()
			}
		}
	}
	return res
}
