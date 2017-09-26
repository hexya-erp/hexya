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

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/operator"
	"github.com/hexya-erp/hexya/hexya/tools/nbutils"
)

// An SQLParams is a list of parameters that are passed to the
// DB server with the query string and that will be used in the
// placeholders.
type SQLParams []interface{}

// Extend returns a new SQLParams with both params of this SQLParams and
// of p2 SQLParams.
func (p SQLParams) Extend(p2 SQLParams) SQLParams {
	pi := []interface{}(p)
	pi2 := []interface{}(p2)
	res := append(pi, pi2...)
	return SQLParams(res)
}

// A Query defines the common part an SQL Query, i.e. all that come
// after the FROM keyword.
type Query struct {
	recordSet RecordCollection
	cond      *Condition
	fetchAll  bool
	limit     int
	offset    int
	groups    []string
	orders    []string
}

// clone returns a pointer to a deep copy of this Query
func (q Query) clone() *Query {
	newCond := *q.cond
	q.cond = &newCond
	return &q
}

// sqlWhereClause returns the sql string and parameters corresponding to the
// WHERE clause of this Query
func (q *Query) sqlWhereClause() (string, SQLParams) {
	q.evaluateConditionArgFunctions()
	sql, args := q.conditionSQLClause(q.cond)
	if sql != "" {
		sql = "WHERE " + sql
	}
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
	for _, val := range c.predicates {
		vSQL, vArgs := q.predicateSQLClause(val, first)
		first = false
		sql += vSQL
		args = args.Extend(vArgs)
	}
	return sql, args
}

// sqlClause returns the sql WHERE clause for this predicate.
// If 'first' is given and true, then the sql clause is not prefixed with
// 'AND' and panics if isOr is true.
func (q *Query) predicateSQLClause(p predicate, first ...bool) (string, SQLParams) {
	var (
		sql     string
		args    SQLParams
		isFirst bool
		adapter dbAdapter = adapters[db.DriverName()]
	)
	if len(first) > 0 {
		isFirst = first[0]
	}
	if p.isOr && !isFirst {
		sql += "OR "
	} else if !isFirst {
		sql += "AND "
	}
	if p.isNot {
		sql += "NOT "
	}

	if p.isCond {
		subSQL, subArgs := q.conditionSQLClause(p.cond)
		sql += fmt.Sprintf(`(%s) `, subSQL)
		args = args.Extend(subArgs)
		return sql, args
	}

	exprs := jsonizeExpr(q.recordSet.model, p.exprs)
	fi := q.recordSet.model.getRelatedFieldInfo(strings.Join(exprs, ExprSep))
	if fi.fieldType.IsFKRelationType() {
		// If we have a relation type with a 0 as foreign key, we substitute for nil
		if valInt, err := nbutils.CastToInteger(p.arg); err == nil && valInt == 0 {
			p.arg = nil
		}
	}
	field := q.joinedFieldExpression(exprs)
	if p.arg == nil {
		switch p.operator {
		case operator.Equals:
			sql += fmt.Sprintf(`%s IS NULL `, field)
		case operator.NotEquals:
			sql += fmt.Sprintf(`%s IS NOT NULL `, field)
		default:
			log.Panic("Null argument can only be used with = and != operators", "operator", p.operator)
		}
		return sql, args
	}

	opSql, arg := adapter.operatorSQL(p.operator, p.arg)
	sql += fmt.Sprintf(`%s %s `, field, opSql)
	args = append(args, arg)
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

// sqlOrderByClause returns the sql string for the ORDER BY clause
// of this Query
func (q *Query) sqlOrderByClause() string {
	if len(q.orders) == 0 {
		return "ORDER BY id"
	}

	var fExprs [][]string
	directions := make([]string, len(q.orders))
	for i, order := range q.orders {
		fieldOrder := strings.Split(strings.TrimSpace(order), " ")
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(fieldOrder[0], ExprSep))
		fExprs = append(fExprs, oExprs)
		if len(fieldOrder) > 1 {
			directions[i] = fieldOrder[1]
		}
	}
	resSlice := make([]string, len(q.orders))
	for i, field := range fExprs {
		resSlice[i] = q.joinedFieldExpression(field)
		resSlice[i] += fmt.Sprintf(" %s", directions[i])
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(resSlice, ", "))
}

// sqlGroupByClause returns the sql string for the GROUP BY clause
// of this Query
func (q *Query) sqlGroupByClause() string {
	var fExprs [][]string
	for _, group := range q.groups {
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(group, ExprSep))
		fExprs = append(fExprs, oExprs)
	}
	resSlice := make([]string, len(q.groups))
	for i, field := range fExprs {
		resSlice[i] = q.joinedFieldExpression(field)
	}
	return fmt.Sprintf("GROUP BY %s", strings.Join(resSlice, ", "))
}

// deleteQuery returns the SQL query string and parameters to unlink
// the rows pointed at by this Query object.
func (q *Query) deleteQuery() (string, SQLParams) {
	adapter := adapters[db.DriverName()]
	sql, args := q.sqlWhereClause()
	delQuery := fmt.Sprintf(`DELETE FROM %s %s`, adapter.quoteTableName(q.recordSet.model.tableName), sql)
	return delQuery, args
}

// insertQuery returns the SQL query string and parameters to insert
// a row with the given data.
func (q *Query) insertQuery(data FieldMap) (string, SQLParams) {
	adapter := adapters[db.DriverName()]
	if len(data) == 0 {
		log.Panic("No data given for insert")
	}
	var (
		cols []string
		vals SQLParams
		i    int
		sql  string
	)
	for k, v := range data {
		fi := q.recordSet.model.fields.MustGet(k)
		if fi.fieldType.IsFKRelationType() && !fi.required {
			if _, ok := v.(*interface{}); ok {
				// We have a null fk field
				continue
			}
		}
		cols = append(cols, fi.json)
		vals = append(vals, v)
		i++
	}
	tableName := adapter.quoteTableName(q.recordSet.model.tableName)
	fields := strings.Join(cols, ", ")
	values := "?" + strings.Repeat(", ?", i-1)
	sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", tableName, fields, values)
	return sql, vals
}

// countQuery returns the SQL query string and parameters to count
// the rows pointed at by this Query object.
func (q *Query) countQuery() (string, SQLParams) {
	sql, args := q.selectQuery([]string{"id"})
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) foo`, sql)
	return countQuery, args
}

// selectQuery returns the SQL query string and parameters to retrieve
// the rows pointed at by this Query object.
// fields is the list of fields to retrieve.
//
// This query must not have a Group By clause.
//
// Each field is a dot-separated
// expression pointing at the field, either as names or columns
// (e.g. 'User.Name' or 'user_id.name')
func (q *Query) selectQuery(fields []string) (string, SQLParams) {
	if len(q.groups) > 0 {
		log.Panic("Calling selectQuery on a Group By query")
	}
	fieldExprs, allExprs := q.selectData(fields)
	// Build up the query
	// Fields
	fieldsSQL := q.fieldsSQL(fieldExprs)
	// Tables
	tablesSQL := q.tablesSQL(allExprs)
	// Where clause and args
	whereSQL, args := q.sqlWhereClause()
	orderSQL := q.sqlOrderByClause()
	limitSQL := q.sqlLimitOffsetClause()
	selQuery := fmt.Sprintf(`SELECT DISTINCT %s FROM %s %s %s %s`, fieldsSQL, tablesSQL, whereSQL, orderSQL, limitSQL)
	return selQuery, args
}

// selectGroupQuery returns the SQL query string and parameters to retrieve
// the result of this Query object, which must include a Group By.
// fields is the list of fields to retrieve.
//
// This query must have a Group By clause.
//
// fields keys are a dot-separated expression pointing at the field, either
// as names or columns (e.g. 'User.Name' or 'user_id.name').
// fields values are
func (q *Query) selectGroupQuery(fields map[string]string) (string, SQLParams) {
	if len(q.groups) == 0 {
		log.Panic("Calling selectGroupQuery on a query without Group By clause")
	}
	fieldsList := make([]string, len(fields))
	i := 0
	for f := range fields {
		fieldsList[i] = f
		i++
	}
	fieldExprs, allExprs := q.selectData(fieldsList)
	// Build up the query
	// Fields
	fieldsSQL := q.fieldsGroupSQL(fieldExprs, fields)
	// Tables
	tablesSQL := q.tablesSQL(allExprs)
	// Where clause and args
	whereSQL, args := q.sqlWhereClause()
	// Group by clause
	groupSQL := q.sqlGroupByClause()
	orderSQL := q.sqlOrderByClause()
	limitSQL := q.sqlLimitOffsetClause()
	selQuery := fmt.Sprintf(`SELECT DISTINCT %s FROM %s %s %s %s %s`, fieldsSQL, tablesSQL, whereSQL, groupSQL, orderSQL, limitSQL)
	return selQuery, args
}

// selectData returns for this query:
// - Expressions defined by the given fields and that must appear in the field list of the select clause.
// - All expressions that also include expressions used in the where clause.
func (q *Query) selectData(fields []string) ([][]string, [][]string) {
	q.substituteChildOfPredicates()
	// Get all expressions, first given by fields
	fieldExprs := make([][]string, len(fields))
	for i, f := range fields {
		fieldExprs[i] = jsonizeExpr(q.recordSet.model, strings.Split(f, ExprSep))
	}
	// Add 'order by' exprs
	for _, order := range q.orders {
		orderField := strings.Split(strings.TrimSpace(order), " ")[0]
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(orderField, ExprSep))
		fieldExprs = append(fieldExprs, oExprs)
	}
	// Then given by condition
	allExprs := append(fieldExprs, q.cond.getAllExpressions(q.recordSet.model)...)
	return fieldExprs, allExprs
}

// substituteChildOfPredicates replaces in the query the predicates with ChildOf
// operator by the predicates to actually execute.
func (q *Query) substituteChildOfPredicates() {
	q.cond.substituteChildOfOperator(q.recordSet)
}

// updateQuery returns the SQL update string and parameters to update
// the rows pointed at by this Query object with the given FieldMap.
func (q *Query) updateQuery(data FieldMap) (string, SQLParams) {
	adapter := adapters[db.DriverName()]
	if len(data) == 0 {
		log.Panic("No data given for update")
	}
	cols := make([]string, len(data))
	vals := make(SQLParams, len(data))
	var (
		i   int
		sql string
	)
	for k, v := range data {
		fi := q.recordSet.model.fields.MustGet(k)
		cols[i] = fmt.Sprintf("%s = ?", fi.json)
		vals[i] = v
		i++
	}
	tableName := adapter.quoteTableName(q.recordSet.model.tableName)
	updates := strings.Join(cols, ", ")
	whereSQL, args := q.sqlWhereClause()
	sql = fmt.Sprintf("UPDATE %s SET %s %s", tableName, updates, whereSQL)
	vals = append(vals, args...)
	return sql, vals
}

// fieldsSQL returns the SQL string for the given field expressions
// parameter must be with the following format (column names):
// [['user_id', 'name'] ['id'] ['profile_id', 'age']]
func (q *Query) fieldsSQL(fieldExprs [][]string) string {
	fStr := make([]string, len(fieldExprs))
	for i, field := range fieldExprs {
		fStr[i] = q.joinedFieldExpression(field, true)
	}
	return strings.Join(fStr, ", ")
}

// fieldsGroupSQL returns the SQL string for the given field expressions
// in a select query with a GROUP BY clause.
// Parameter must be with the following format (column names):
// [['user_id', 'name'] ['id'] ['profile_id', 'age']]
func (q *Query) fieldsGroupSQL(fieldExprs [][]string, fields map[string]string) string {
	fStr := make([]string, len(fieldExprs)+1)
	for i, exprs := range fieldExprs {
		aggFnct := fields[strings.Join(exprs, ExprSep)]
		joins := q.generateTableJoins(exprs)
		num := len(joins)
		fStr[i] = fmt.Sprintf("%s(%s.%s) AS %s", aggFnct, joins[num-1].alias, exprs[num-1], strings.Join(exprs, sqlSep))
	}
	fStr[len(fieldExprs)] = "count(1) AS __count"
	return strings.Join(fStr, ", ")
}

// joinedFieldExpression joins the given expressions into a fields sql string
// ['profile_id' 'user_id' 'name'] => "profiles__users".name
// ['age'] => "mytable".age
// If withAlias is true, then returns fields with its alias
func (q *Query) joinedFieldExpression(exprs []string, withAlias ...bool) string {
	joins := q.generateTableJoins(exprs)
	num := len(joins)
	if len(withAlias) > 0 && withAlias[0] {
		return fmt.Sprintf("%s.%s AS %s", joins[num-1].alias, exprs[num-1], strings.Join(exprs, sqlSep))
	}
	return fmt.Sprintf("%s.%s", joins[num-1].alias, exprs[num-1])
}

// generateTableJoins transforms a list of fields expression into a list of tableJoins
// ['user_id' 'profile_id' 'age'] => []tableJoins{CurrentTable User Profile}
func (q *Query) generateTableJoins(fieldExprs []string) []tableJoin {
	adapter := adapters[db.DriverName()]
	var joins []tableJoin
	curMI := q.recordSet.model
	// Create the tableJoin for the current table
	currentTableName := adapter.quoteTableName(curMI.tableName)
	curTJ := &tableJoin{
		tableName: currentTableName,
		joined:    false,
		alias:     currentTableName,
	}
	joins = append(joins, *curTJ)
	alias := curMI.tableName
	exprsLen := len(fieldExprs)
	for i, expr := range fieldExprs {
		fi, ok := curMI.fields.get(expr)
		if !ok {
			log.Panic("Unparsable Expression", "expr", strings.Join(fieldExprs, ExprSep))
		}
		if fi.relatedModel == nil || i == exprsLen-1 {
			// Don't create an extra join if our field is not a relation field
			// or if it is the last field of our expressions
			break
		}
		var innerJoin bool
		if fi.required {
			innerJoin = true
		}
		linkedTableName := adapter.quoteTableName(fi.relatedModel.tableName)
		alias = fmt.Sprintf("%s%s%s", alias, sqlSep, fi.relatedModel.tableName)

		var field, otherField string
		switch fi.fieldType {
		case fieldtype.Many2One, fieldtype.One2One:
			field, otherField = "id", expr
		case fieldtype.One2Many, fieldtype.Rev2One:
			field, otherField = jsonizePath(fi.relatedModel, fi.reverseFK), "id"
		}

		nextTJ := tableJoin{
			tableName:  linkedTableName,
			joined:     true,
			innerJoin:  innerJoin,
			field:      field,
			otherTable: curTJ,
			otherField: otherField,
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

// isEmpty returns true if this query is empty
// i.e. this query will search all the database.
func (q *Query) isEmpty() bool {
	if !q.cond.IsEmpty() {
		return false
	}
	if q.fetchAll {
		return false
	}
	if q.limit != 0 {
		return false
	}
	if q.offset != 0 {
		return false
	}
	if len(q.groups) > 0 {
		return false
	}
	if len(q.orders) > 0 {
		return false
	}
	return true
}

// inferIds tries to return the list of ids this query points to without calling the database.
// If it can't, the second argument will be false.
func (q *Query) inferIds() ([]int64, bool) {
	if q.fetchAll {
		return []int64{}, false
	}
	if q.limit != 0 {
		return []int64{}, false
	}
	if q.offset != 0 {
		return []int64{}, false
	}
	if len(q.groups) > 0 {
		return []int64{}, false
	}
	if q.cond.IsEmpty() {
		return []int64{}, false
	}
	if len(q.cond.predicates) != 1 {
		return []int64{}, false
	}
	predicate := q.cond.predicates[0]
	if len(predicate.exprs) == 0 && !predicate.cond.IsEmpty() {
		predicate = predicate.cond.predicates[0]
	}
	if !predicate.cond.IsEmpty() {
		return []int64{}, false
	}
	if len(predicate.exprs) != 1 {
		return []int64{}, false
	}
	if predicate.exprs[0] != "id" && predicate.exprs[0] != "ID" {
		return []int64{}, false
	}

	switch predicate.operator {
	case operator.Equals:
		res, ok := predicate.arg.(int64)
		return []int64{res}, ok
	case operator.In:
		res, ok := predicate.arg.([]int64)
		return res, ok
	default:
		return []int64{}, false
	}
}

// substituteConditionExprs substitutes all occurrences of each substMap keys in
// its conditions 1st exprs with the corresponding substMap value.
func (q *Query) substituteConditionExprs(substMap map[string][]string) {
	q.cond.substituteExprs(q.recordSet.model, substMap)
}

// evaluateConditionArgFunctions evaluates all args in the queries that are functions and
// substitute it with the result.
func (q *Query) evaluateConditionArgFunctions() {
	q.cond.evaluateArgFunctions(q.recordSet)
}

// newQuery returns a new empty query
// If rs is given, bind this query to the given RecordSet.
func newQuery(rs ...RecordCollection) *Query {
	var rset RecordCollection
	if len(rs) > 0 {
		rset = rs[0]
	}
	return &Query{
		cond:      newCondition(),
		recordSet: rset,
	}
}
