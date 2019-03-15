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
	"reflect"
	"sort"
	"strings"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	"github.com/hexya-erp/hexya/src/tools/strutils"
)

const maxSQLidentifierLength = 63

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
	recordSet  *RecordCollection
	cond       *Condition
	ctxCond    *Condition
	fetchAll   bool
	limit      int
	offset     int
	noDistinct bool
	groups     []string
	ctxGroups  []string
	orders     []string
	ctxOrders  []string
}

// clone returns a pointer to a deep copy of this Query
//
// rc is the RecordCollection the new query will be bound to.
func (q Query) clone(rc *RecordCollection) *Query {
	newCond := *q.cond
	q.cond = &newCond
	newCtxCond := *q.ctxCond
	q.ctxCond = &newCtxCond
	q.noDistinct = false
	q.recordSet = rc
	return &q
}

// sqlWhereClause returns the sql string and parameters corresponding to the
// WHERE clause of this Query
//
// If withCtx is set, the extra conditions are included
func (q *Query) sqlWhereClause(withCtx bool) (string, SQLParams) {
	sql, args := q.conditionSQLClause(q.cond)
	extraSQL, extraArgs := q.conditionSQLClause(q.ctxCond)
	if sql == "" && extraSQL == "" {
		return "", SQLParams{}
	}
	resSQL := "WHERE "
	var resArgs SQLParams
	switch {
	case extraSQL == "" || !withCtx:
		resSQL += sql
		resArgs = args
	case sql == "":
		resSQL += extraSQL
		resArgs = extraArgs
	default:
		resSQL += fmt.Sprintf("(%s) AND (%s)", sql, extraSQL)
		resArgs = args.Extend(extraArgs)
	}
	return resSQL, resArgs
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
	for _, p := range c.predicates {
		op := "AND"
		if p.isOr {
			op = "OR"
		}
		if p.isNot {
			op += " NOT"
		}

		vSQL, vArgs := q.predicateSQLClause(p)
		switch {
		case first:
			sql = vSQL
			if p.isNot {
				sql = "NOT " + sql
			}
		case p.isCond:
			sql = fmt.Sprintf("(%s) %s (%s)", sql, op, vSQL)
		default:
			sql = fmt.Sprintf("%s %s %s", sql, op, vSQL)
		}
		args = args.Extend(vArgs)
		first = false
	}
	return sql, args
}

// sqlClause returns the sql WHERE clause and arguments for this predicate.
func (q *Query) predicateSQLClause(p predicate) (string, SQLParams) {
	if p.isCond {
		return q.conditionSQLClause(p.cond)
	}

	exprs := jsonizeExpr(q.recordSet.model, p.exprs)
	fi := q.recordSet.model.getRelatedFieldInfo(strings.Join(exprs, ExprSep))
	if fi.fieldType.IsFKRelationType() {
		// If we have a relation type with a 0 as foreign key, we substitute for nil
		if valInt, err := nbutils.CastToInteger(p.arg); err == nil && valInt == 0 {
			p.arg = nil
		}
	}

	var (
		sql  string
		args SQLParams
	)
	field, _, _ := q.joinedFieldExpression(exprs, false, 0)

	var isNull bool
	switch v := p.arg.(type) {
	case nil:
		isNull = true
	case string:
		if v == "" {
			isNull = true
		}
	case bool:
		if !v {
			isNull = true
		}
	}
	if s, ok := p.arg.(string); ok && s == "" {
		isNull = true
	}
	if isNull {
		switch p.operator {
		case operator.Equals:
			sql = fmt.Sprintf(`%s IS NULL`, field)
			if !fi.isRelationField() {
				sql = fmt.Sprintf(`(%s OR %s = ?)`, sql, field)
				args = SQLParams{reflect.Zero(fi.fieldType.DefaultGoType()).Interface()}
			}
		case operator.NotEquals:
			sql = fmt.Sprintf(`%s IS NOT NULL`, field)
			if !fi.isRelationField() {
				sql = fmt.Sprintf(`(%s AND %s != ?)`, sql, field)
				args = SQLParams{reflect.Zero(fi.fieldType.DefaultGoType()).Interface()}
			}
		default:
			log.Panic("Null argument can only be used with = and != operators", "operator", p.operator)
		}
		return sql, args
	}
	adapter := adapters[db.DriverName()]
	arg := q.evaluateConditionArgFunctions(p)
	opSql, arg := adapter.operatorSQL(p.operator, arg)
	sql = fmt.Sprintf(`%s %s`, field, opSql)
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

// prepareOrderByExprs returns the 'orders' of this query as an expressions slice on
// the one side and a slice of directions (ASC or DESC) on the other side.
func (q *Query) prepareOrderByExprs() ([][]string, []string) {
	fExprs := make([][]string, len(q.orders))
	directions := make([]string, len(q.orders))
	for i, order := range q.orders {
		fieldOrder := strings.Split(strings.TrimSpace(order), " ")
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(fieldOrder[0], ExprSep))
		fExprs[i] = oExprs
		if len(fieldOrder) > 1 {
			directions[i] = fieldOrder[1]
		}
	}
	ctxExprs, ctxDirs := q.prepareCtxOrderByExprs(false)
	fExprs = append(fExprs, ctxExprs...)
	directions = append(directions, ctxDirs...)
	return fExprs, directions
}

// prepareCtxOrderByExprs returns the 'ctxOrders' of this query as an expressions slice on
// the one side and a slice of directions (ASC or DESC) on the other side.
//
// if inv is true, the orders are inverted
func (q *Query) prepareCtxOrderByExprs(inv bool) ([][]string, []string) {
	fExprs := make([][]string, len(q.ctxOrders))
	directions := make([]string, len(q.ctxOrders))
	for i, order := range q.ctxOrders {
		fieldOrder := strings.Split(strings.TrimSpace(order), " ")
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(fieldOrder[0], ExprSep))
		fExprs[i] = oExprs
		if len(fieldOrder) > 1 {
			directions[i] = fieldOrder[1]
		}
		if inv {
			if strings.TrimSpace(strings.ToUpper(directions[i])) == "DESC" {
				directions[i] = "ASC"
			} else {
				directions[i] = "DESC"
			}
		}
	}
	return fExprs, directions
}

// sqlOrderByClause returns the sql string for the ORDER BY clause
// of this Query
func (q *Query) sqlOrderByClause() string {
	fExprs, directions := q.prepareOrderByExprs()
	resSlice := make([]string, len(q.orders)+len(q.ctxOrders))
	for i, field := range fExprs {
		resSlice[i], _, _ = q.joinedFieldExpression(field, false, 0)
		resSlice[i] += fmt.Sprintf(" %s", directions[i])
	}
	if len(resSlice) == 0 {
		return ""
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(resSlice, ", "))
}

// sqlCtxOrderByClause returns the sql string for the ORDER BY clause of the ctx fields
// of this Query with inversed order.
func (q *Query) sqlCtxOrderBy() string {
	fExprs, directions := q.prepareCtxOrderByExprs(true)
	resSlice := make([]string, len(q.ctxOrders))
	for i, field := range fExprs {
		resSlice[i], _, _ = q.joinedFieldExpression(field, false, 0)
		resSlice[i] += fmt.Sprintf(" %s", directions[i])
	}
	if len(resSlice) == 0 {
		return ""
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(resSlice, ", "))
}

// sqlOrderByClauseForGroupBy returns the sql string for the ORDER BY clause
// of this Query, which should be a group by clause.
func (q *Query) sqlOrderByClauseForGroupBy(aggFncts map[string]string) string {
	fExprs, directions := q.prepareOrderByExprs()
	resSlice := make([]string, len(q.orders)+len(q.ctxOrders))
	for i, field := range fExprs {
		aggFnct := aggFncts[strings.Join(field, ExprSep)]
		if aggFnct == "" {
			jfe, _, _ := q.joinedFieldExpression(field, false, 0)
			resSlice[i] = fmt.Sprintf("%s %s", jfe, directions[i])
			continue
		}
		jfe, _, _ := q.joinedFieldExpression(field, false, 0)
		resSlice[i] = fmt.Sprintf("%s(%s) %s", aggFnct, jfe, directions[i])
	}
	if len(resSlice) == 0 {
		return ""
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(resSlice, ", "))
}

// sqlGroupByClause returns the sql string for the GROUP BY clause
// of this Query (without the GROUP BY keywords)
func (q *Query) sqlGroupByClause() string {
	var fExprs [][]string
	for _, group := range q.groups {
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(group, ExprSep))
		fExprs = append(fExprs, oExprs)
	}
	resSlice := make([]string, len(q.groups))
	for i, field := range fExprs {
		resSlice[i], _, _ = q.joinedFieldExpression(field, false, 0)
	}
	res := strings.Join(resSlice, ", ")
	ctxStr := strings.TrimSpace(q.sqlCtxGroupByClause())
	if ctxStr != "" {
		res = fmt.Sprintf("%s, %s", res, ctxStr)
	}
	return res
}

// sqlCtxGroupByClause returns the sql string for the GROUP BY clause
// of contexted fields for this Query (without the GROUP BY keywords)
func (q *Query) sqlCtxGroupByClause() string {
	var fExprs [][]string
	for _, group := range q.ctxGroups {
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(group, ExprSep))
		fExprs = append(fExprs, oExprs)
	}
	resSlice := make([]string, len(q.ctxGroups))
	for i, field := range fExprs {
		resSlice[i], _, _ = q.joinedFieldExpression(field, false, 0)
	}
	return strings.Join(resSlice, ", ")
}

// deleteQuery returns the SQL query string and parameters to unlink
// the rows pointed at by this Query object.
func (q *Query) deleteQuery() (string, SQLParams) {
	adapter := adapters[db.DriverName()]
	sql, args := q.sqlWhereClause(false)
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
	sql, args, _ := q.selectQuery([]string{"id"})
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
func (q *Query) selectQuery(fields []string) (string, SQLParams, map[string]string) {
	if len(q.groups) > 0 {
		log.Panic("Calling selectQuery on a Group By query")
	}
	fieldExprs, allExprs := q.selectData(fields, true)
	// Build up the query
	// Fields
	fieldsSQL, fieldSubsts := q.fieldsSQL(fieldExprs)
	// Tables
	tablesSQL, joinsMap := q.tablesSQL(allExprs)
	// Where clause and args
	whereSQL, args := q.sqlWhereClause(true)
	orderSQL := q.sqlOrderByClause()
	limitSQL := q.sqlLimitOffsetClause()
	var distinct string
	if !q.noDistinct {
		distinct = "DISTINCT"
	}
	selQuery := fmt.Sprintf(`SELECT %s %s FROM %s %s %s %s`, distinct, fieldsSQL, tablesSQL, whereSQL, orderSQL, limitSQL)
	selQuery = strutils.Substitute(selQuery, joinsMap)
	return selQuery, args, fieldSubsts
}

// selectGroupQuery returns the SQL query string and parameters to retrieve
// the result of this Query object, which must include a Group By.
// fields is the list of fields to retrieve.
//
// This query must have a Group By clause.
//
// fields keys are a dot-separated expression pointing at the field, either
// as names or columns (e.g. 'User.Name' or 'user_id.name').
// fields values are the SQL aggregate function to use for the field or ""
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
	fieldExprs, allExprs := q.selectData(fieldsList, false)
	// Build up the query
	// Fields
	fieldsSQL := q.fieldsGroupSQL(fieldExprs, fields)
	// Tables
	tablesSQL, joinsMap := q.tablesSQL(allExprs)
	// Where clause and args
	whereSQL, args := q.sqlWhereClause(true)
	// Group by clause
	groupSQL := q.sqlGroupByClause()
	orderSQL := q.sqlOrderByClauseForGroupBy(fields)
	limitSQL := q.sqlLimitOffsetClause()
	selQuery := fmt.Sprintf(`WITH group_query as (SELECT %s FROM %s %s GROUP BY %s %s %s) SELECT * FROM group_query WHERE __rank = 1`, fieldsSQL, tablesSQL, whereSQL, groupSQL, orderSQL, limitSQL)
	selQuery = strutils.Substitute(selQuery, joinsMap)
	return selQuery, args
}

// selectData returns for this query:
// - Expressions defined by the given fields and that must appear in the field list of the select clause.
// - All expressions that also include expressions used in the where clause.
func (q *Query) selectData(fields []string, withCtx bool) ([][]string, [][]string) {
	q.substituteChildOfPredicates()
	// Get all expressions, first given by fields
	fieldExprs := make([][]string, len(fields))
	for i, f := range fields {
		fieldExprs[i] = jsonizeExpr(q.recordSet.model, strings.Split(f, ExprSep))
	}
	// Add 'order by' exprs
	fieldExprs = append(fieldExprs, q.getOrderByExpressions(withCtx)...)
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
	whereSQL, args := q.sqlWhereClause(false)
	sql = fmt.Sprintf("UPDATE %s SET %s %s", tableName, updates, whereSQL)
	vals = append(vals, args...)
	return sql, vals
}

// fieldsSQL returns the SQL string for the given field expressions
// parameter must be with the following format (column names):
// [['user_id', 'name'] ['id'] ['profile_id', 'age']]
//
// Second returned field is a map with the aliases used if the nominal "user_id__name"
// alias type gives a string longer than 64 chars
func (q *Query) fieldsSQL(fieldExprs [][]string) (string, map[string]string) {
	fStr := make([]string, len(fieldExprs))
	substs := make(map[string]string)
	for i, field := range fieldExprs {
		res, natAlias, realAlias := q.joinedFieldExpression(field, true, i)
		fStr[i] = res
		substs[realAlias] = natAlias
	}
	return strings.Join(fStr, ", "), substs
}

// fieldsGroupSQL returns the SQL string for the given field expressions
// in a select query with a GROUP BY clause.
// Parameter must be with the following format (column names):
// [['user_id', 'name'] ['id'] ['profile_id', 'age']]
func (q *Query) fieldsGroupSQL(fieldExprs [][]string, aggFncts map[string]string) string {
	fStr := make([]string, len(fieldExprs)+2)
	for i, exprs := range fieldExprs {
		aggFnct := aggFncts[strings.Join(exprs, ExprSep)]
		joins := q.generateTableJoins(exprs)
		lastJoin := joins[len(joins)-1]
		if aggFnct == "" {
			fStr[i] = fmt.Sprintf("%s.%s AS %s", lastJoin.alias, lastJoin.expr, strings.Join(exprs, sqlSep))
			continue
		}
		fStr[i] = fmt.Sprintf("%s(%s.%s) AS %s", aggFnct, lastJoin.alias, lastJoin.expr, strings.Join(exprs, sqlSep))
	}
	fStr[len(fieldExprs)] = "count(1) AS __count"
	curTable := q.generateTableJoins([]string{"id"})[0]
	fStr[len(fieldExprs)+1] = fmt.Sprintf("rank() OVER (PARTITION BY min(%s.id) %s) AS __rank", curTable.alias, q.sqlCtxOrderBy())
	return strings.Join(fStr, ", ")
}

// joinedFieldExpression joins the given expressions into a fields sql string
//     ['profile_id' 'user_id' 'name'] => "profiles__users".name
//     ['age'] => "mytable".age
//
// If withAlias is true, then returns fields with its alias. In this case, aliasIndex is used
// to define aliases when the nominal "profile_id__user_id__name" is longer than 64 chars.
// Returned second argument is the nominal alias and third argument is the alias actually used.
func (q *Query) joinedFieldExpression(exprs []string, withAlias bool, aliasIndex int) (string, string, string) {
	joins := q.generateTableJoins(exprs)
	lastJoin := joins[len(joins)-1]
	if withAlias {
		fAlias := strings.Join(exprs, sqlSep)
		oldAlias := fAlias
		if len(fAlias) > maxSQLidentifierLength {
			fAlias = fmt.Sprintf("f%d", aliasIndex)
		}
		return fmt.Sprintf("%s.%s AS %s", lastJoin.alias, lastJoin.expr, fAlias), oldAlias, fAlias
	}
	return fmt.Sprintf("%s.%s", lastJoin.alias, lastJoin.expr), "", ""
}

// generateTableJoins transforms a list of fields expression into a list of tableJoins
// ['user_id' 'profile_id' 'age'] => []tableJoins{CurrentTable User Profile}
func (q *Query) generateTableJoins(fieldExprs []string) []tableJoin {
	adapter := adapters[db.DriverName()]
	var joins []tableJoin
	curMI := q.recordSet.model
	// Create the tableJoin for the current table
	currentTableName := adapter.quoteTableName(curMI.tableName)
	var curExpr string
	if len(fieldExprs) > 0 {
		curExpr = fieldExprs[0]
	}
	curTJ := &tableJoin{
		tableName: currentTableName,
		joined:    false,
		alias:     currentTableName,
		expr:      curExpr,
	}
	joins = append(joins, *curTJ)
	alias := curMI.tableName
	exprsLen := len(fieldExprs)
	for i, expr := range fieldExprs {
		fi, ok := curMI.fields.Get(expr)
		if !ok {
			log.Panic("Unparsable Expression", "expr", strings.Join(fieldExprs, ExprSep))
		}
		if fi.relatedModel == nil || (i == exprsLen-1 && fi.fieldType.IsFKRelationType()) {
			// Don't create an extra join if our field is not a relation field
			// or if it is the last field of our expressions
			break
		}

		var field, otherField string
		var tjExpr string
		if i < exprsLen-1 {
			tjExpr = fieldExprs[i+1]
		}
		switch fi.fieldType {
		case fieldtype.Many2One, fieldtype.One2One:
			field, otherField = "id", expr
		case fieldtype.One2Many, fieldtype.Rev2One:
			field, otherField = jsonizePath(fi.relatedModel, fi.reverseFK), "id"
			if tjExpr == "" {
				tjExpr = "id"
			}
		case fieldtype.Many2Many:
			// Add relation table join
			relationTableName := adapter.quoteTableName(fi.m2mRelModel.tableName)
			alias = fmt.Sprintf("%s%s%s", alias, sqlSep, fi.m2mRelModel.tableName)
			tj := tableJoin{
				tableName:  relationTableName,
				joined:     true,
				field:      jsonizePath(fi.m2mRelModel, fi.m2mOurField.name),
				otherTable: curTJ,
				otherField: "id",
				alias:      adapter.quoteTableName(alias),
				expr:       jsonizePath(fi.m2mRelModel, fi.m2mTheirField.name),
			}
			joins = append(joins, tj)
			curTJ = &tj
			// Add relation to other table
			field, otherField = "id", jsonizePath(fi.m2mRelModel, fi.m2mTheirField.name)
			if tjExpr == "" {
				tjExpr = "id"
			}
		}

		linkedTableName := adapter.quoteTableName(fi.relatedModel.tableName)
		alias = fmt.Sprintf("%s%s%s", alias, sqlSep, fi.relatedModel.tableName)
		nextTJ := tableJoin{
			tableName:  linkedTableName,
			joined:     true,
			field:      field,
			otherTable: curTJ,
			otherField: otherField,
			alias:      adapter.quoteTableName(alias),
			expr:       tjExpr,
		}
		joins = append(joins, nextTJ)
		curMI = fi.relatedModel
		curTJ = &nextTJ
	}
	return joins
}

// tablesSQL returns the SQL string for the FROM clause of our SQL query
// including all joins if any for the given expressions.
//
// Returned FROM clause uses table alias such as "Tn" and second argument is the
// mapping between aliases in tableJoin objects and the new "Tn" aliases. This
// mapping is necessary to keep table alias < 63 chars which is postgres limit.
func (q *Query) tablesSQL(fExprs [][]string) (string, map[string]string) {
	adapter := adapters[db.DriverName()]
	var (
		res        string
		aliasIndex int
	)
	joinsMap := make(map[string]string)
	// Get a list of unique table joins (by alias)
	for _, f := range fExprs {
		tJoins := q.generateTableJoins(f)
		for _, j := range tJoins {
			if _, exists := joinsMap[j.alias]; !exists {
				joinsMap[j.alias] = adapter.quoteTableName(fmt.Sprintf("T%d", aliasIndex))
				if aliasIndex == 0 {
					joinsMap[j.alias] = j.alias
				}
				aliasIndex++
				res += j.sqlString()
			}
		}
	}
	return res, joinsMap
}

// isEmpty returns true if this query is empty
// i.e. this query will search all the database.
func (q *Query) isEmpty() bool {
	if !q.cond.IsEmpty() {
		return false
	}
	return q.sideDataIsEmpty()
}

// sideDataIsEmpty returns true if all side data of the query is empty.
// By side data, we mean everything but the condition itself.
func (q *Query) sideDataIsEmpty() bool {
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

// substituteConditionExprs substitutes all occurrences of each substMap keys in
// its conditions 1st exprs with the corresponding substMap value.
func (q *Query) substituteConditionExprs(substMap map[string][]string) {
	q.cond.substituteExprs(q.recordSet.model, substMap)
	for i, order := range q.orders {
		orderPath := strings.Split(strings.TrimSpace(order), " ")[0]
		jsonPath := jsonizePath(q.recordSet.model, orderPath)
		for k, v := range substMap {
			if jsonPath == k {
				q.orders[i] = strings.Replace(q.orders[i], orderPath, strings.Join(v, ExprSep), -1)
				break
			}
		}
	}
	for i, group := range q.groups {
		jsonPath := jsonizePath(q.recordSet.model, group)
		for k, v := range substMap {
			if jsonPath == k {
				q.groups[i] = strings.Replace(q.groups[i], group, strings.Join(v, ExprSep), -1)
				break
			}
		}
	}
}

// evaluateConditionArgFunctions evaluates all args in the queries that are functions and
// substitute it with the result.
//
// multi should be true if the operator of the predicate is IN
func (q *Query) evaluateConditionArgFunctions(p predicate) interface{} {
	fnctVal := reflect.ValueOf(p.arg)
	if fnctVal.Kind() != reflect.Func {
		return p.arg
	}
	firstArgType := fnctVal.Type().In(0)
	if !firstArgType.Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
		return p.arg
	}
	argValue := reflect.ValueOf(q.recordSet)
	res := fnctVal.Call([]reflect.Value{argValue})
	return sanitizeArgs(res[0].Interface(), p.operator.IsMulti())
}

// getAllExpressions returns all expressions used in this query,
// both in the condition and the order by clause.
func (q *Query) getAllExpressions() [][]string {
	return append(q.getOrderByExpressions(true),
		append(q.getGroupByExpressions(), q.cond.getAllExpressions(q.recordSet.model)...)...)
}

// getOrderByExpressions returns all expressions used in order by clause of this query.
//
// If withCtx is true, ctxOrder expressions are also returned
func (q *Query) getOrderByExpressions(withCtx bool) [][]string {
	var exprs [][]string
	for _, order := range q.orders {
		orderField := strings.Split(strings.TrimSpace(order), " ")[0]
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(orderField, ExprSep))
		exprs = append(exprs, oExprs)
	}
	if withCtx {
		exprs = append(exprs, q.getCtxOrderByExpressions()...)
	}
	return exprs
}

// getOrderByExpressions returns expressions used in context order by clause of this query.
func (q *Query) getCtxOrderByExpressions() [][]string {
	var exprs [][]string
	for _, order := range q.ctxOrders {
		orderField := strings.Split(strings.TrimSpace(order), " ")[0]
		oExprs := jsonizeExpr(q.recordSet.model, strings.Split(orderField, ExprSep))
		exprs = append(exprs, oExprs)
	}
	return exprs
}

// getGroupByExpressions returns all expressions used in group by clause of this query.
func (q *Query) getGroupByExpressions() [][]string {
	var exprs [][]string
	for _, group := range q.groups {
		gExprs := jsonizeExpr(q.recordSet.model, strings.Split(group, ExprSep))
		exprs = append(exprs, gExprs)
	}
	return exprs
}

// ctxArgsSlug returns a slug of the arguments of the context condition of this query
func (q *Query) ctxArgsSlug() string {
	return q.argsSlug(q.ctxCond)
}

// argsSlug returns a slug of the given condition arguments
func (q *Query) argsSlug(c *Condition) string {
	var (
		res  string
		args []string
	)
	for _, p := range c.predicates {
		if p.isCond {
			res += q.argsSlug(p.cond)
			continue
		}
		arg := fmt.Sprintf("%v", q.evaluateConditionArgFunctions(p))
		arg = strings.Replace(arg, ExprSep, "-", -1)
		arg = strings.Replace(arg, ContextSep, "-", -1)
		arg = strings.Replace(arg, "<nil>", "", -1)
		args = append(args, arg)
	}
	sort.Strings(args)
	res += strings.Join(args, "")
	return res
}

// newQuery returns a new empty query
// If rs is given, bind this query to the given RecordSet.
func newQuery(rs ...*RecordCollection) *Query {
	var rset *RecordCollection
	if len(rs) > 0 {
		rset = rs[0]
	}
	return &Query{
		cond:      newCondition(),
		ctxCond:   newCondition(),
		recordSet: rset,
	}
}
