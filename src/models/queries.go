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

	"github.com/hexya-erp/hexya/src/models/field"
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
	return res
}

// An orderPredicate in a query. e.g. "name ASC".
type orderPredicate struct {
	field FieldName
	desc  bool
}

// A Query defines the common part an SQL Query, i.e. all that come
// after the FROM keyword.
type Query struct {
	recordSet *RecordCollection
	cond      *Condition
	ctxCond   *Condition
	fetchAll  bool
	limit     int
	offset    int
	groups    []FieldName
	ctxGroups []FieldName
	orders    []orderPredicate
	ctxOrders []orderPredicate
}

// clone returns a pointer to a deep copy of this Query
//
// rc is the RecordCollection the new query will be bound to.
func (q Query) clone(rc *RecordCollection) *Query {
	newCond := *q.cond
	q.cond = &newCond
	newCtxCond := *q.ctxCond
	q.ctxCond = &newCtxCond
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

	fi := q.recordSet.model.getRelatedFieldInfo(joinFieldNames(p.exprs, ExprSep))
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
	field, _, _ := q.joinedFieldExpression(p.exprs, false, 0)

	adapter := adapters[db.DriverName()]
	arg := q.evaluateConditionArgFunctions(p)
	opSql, arg := adapter.operatorSQL(p.operator, arg)

	var isNull bool
	switch v := arg.(type) {
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
	if isNull {
		return nullSQLClause(field, p.operator, fi)
	}

	sql = fmt.Sprintf(`%s %s`, field, opSql)
	if p.operator.IsNegative() {
		sql = fmt.Sprintf(`(%s IS NULL OR %s)`, field, sql)
	}

	args = append(args, arg)
	return sql, args
}

//nullSQLClause returns the sql string and arguments for searching the given field with an empty argument
func nullSQLClause(field string, op operator.Operator, fi *Field) (string, SQLParams) {
	var (
		sql  string
		args SQLParams
	)
	switch op {
	case operator.Equals, operator.Like, operator.ILike, operator.Contains, operator.IContains:
		sql = fmt.Sprintf(`%s IS NULL`, field)
		if !fi.isRelationField() {
			sql = fmt.Sprintf(`(%s OR %s = ?)`, sql, field)
			args = SQLParams{reflect.Zero(fi.fieldType.DefaultGoType()).Interface()}
		}
	case operator.NotEquals, operator.NotContains, operator.NotIContains:
		sql = fmt.Sprintf(`%s IS NOT NULL`, field)
		if !fi.isRelationField() {
			sql = fmt.Sprintf(`(%s AND %s != ?)`, sql, field)
			args = SQLParams{reflect.Zero(fi.fieldType.DefaultGoType()).Interface()}
		}
	default:
		log.Panic("Null argument can only be used with = and != operators", "operator", op)
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

// sqlOrderByClause returns the sql string for the ORDER BY clause
// of this Query
func (q *Query) sqlOrderByClause() string {
	resSlice := make([]string, len(q.orders))
	for i, order := range q.orders {
		_, _, resSlice[i] = q.joinedFieldExpression(splitFieldNames(order.field, ExprSep), true, i)
		if order.desc {
			resSlice[i] += " DESC"
		}
	}
	if len(resSlice) == 0 {
		return ""
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(resSlice, ", "))
}

// sqlCtxOrderByClause returns the sql string for the ORDER BY clause of the ctx fields
// of this Query.
func (q *Query) sqlCtxOrderBy() string {
	resSlice := make([]string, len(q.ctxOrders))
	for i, order := range q.ctxOrders {
		resSlice[i], _, _ = q.joinedFieldExpression(splitFieldNames(order.field, ExprSep), false, 0)
		if order.desc {
			resSlice[i] += " DESC"
		}
	}
	if len(resSlice) == 0 {
		return ""
	}
	return fmt.Sprintf("%s", strings.Join(resSlice, ", "))
}

// sqlOrderByClauseForGroupBy returns the sql string for the ORDER BY clause
// of this Query, which should be a group by clause.
func (q *Query) sqlOrderByClauseForGroupBy(aggFncts map[string]string) string {
	resSlice := make([]string, len(q.orders))
	for i, order := range q.orders {
		aggFnct := aggFncts[order.field.JSON()]
		if aggFnct == "" {
			_, _, jfe := q.joinedFieldExpression(splitFieldNames(order.field, ExprSep), true, i)
			if order.desc {
				jfe += " DESC"
			}
			resSlice[i] = jfe
			continue
		}
		_, _, jfe := q.joinedFieldExpression(splitFieldNames(order.field, ExprSep), true, i)
		resSlice[i] = fmt.Sprintf("%s(%s)", aggFnct, jfe)
		if order.desc {
			resSlice[i] += " DESC"
		}
	}
	if len(resSlice) == 0 {
		return ""
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(resSlice, ", "))
}

// sqlGroupByClause returns the sql string for the GROUP BY clause
// of this Query (without the GROUP BY keywords)
func (q *Query) sqlGroupByClause() string {
	var fExprs [][]FieldName
	for _, group := range q.groups {
		oExprs := splitFieldNames(group, ExprSep)
		fExprs = append(fExprs, oExprs)
	}
	resSlice := make([]string, len(q.groups))
	for i, field := range fExprs {
		_, _, resSlice[i] = q.joinedFieldExpression(field, true, i)
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
	var fExprs [][]FieldName
	for _, group := range q.ctxGroups {
		oExprs := splitFieldNames(group, ExprSep)
		fExprs = append(fExprs, oExprs)
	}
	resSlice := make([]string, len(q.ctxGroups))
	for i, field := range fExprs {
		_, _, resSlice[i] = q.joinedFieldExpression(field, true, i)
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
	sql, args, _ := q.selectQuery([]FieldName{ID})
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) foo`, sql)
	return countQuery, args
}

// selectCommonQuery returns the SQL query string and parameters to retrieve
// the rows pointed at by this Query object.
// This subquery will be used in selectQuery and selectGroupQuery
// fields is the list of fields to retrieve.
//
// Each field is a dot-separated
// expression pointing at the field, either as names or columns
// (e.g. 'User.Name' or 'user_id.name')
func (q *Query) selectCommonQuery(fields []FieldName) (string, SQLParams, map[string]string) {
	fieldExprs, allExprs := q.selectData(fields, true)
	// Build up the query
	// Fields
	fieldsSQL, fieldSubsts := q.fieldsSQL(fieldExprs)
	// Tables
	tablesSQL, joinsMap := q.tablesSQL(allExprs)
	// Where clause and args
	whereSQL, args := q.sqlWhereClause(true)
	ctxOrderSQL := q.sqlCtxOrderBy()
	if ctxOrderSQL != "" {
		ctxOrderSQL = fmt.Sprintf(", %s", ctxOrderSQL)
	}
	selQuery := fmt.Sprintf(`SELECT DISTINCT ON (%s.id) %s FROM %s %s ORDER BY %s.id %s`,
		q.thisTable(), fieldsSQL, tablesSQL, whereSQL, q.thisTable(), ctxOrderSQL)
	selQuery = strutils.Substitute(selQuery, joinsMap)
	return selQuery, args, fieldSubsts
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
func (q *Query) selectQuery(fields []FieldName) (string, SQLParams, map[string]string) {
	if len(q.groups) > 0 {
		log.Panic("Calling selectQuery on a Group By query")
	}
	subQuery, args, substs := q.selectCommonQuery(fields)
	orderSQL := q.sqlOrderByClause()
	limitSQL := q.sqlLimitOffsetClause()
	selQuery := fmt.Sprintf(`SELECT * FROM (%s) foo %s %s`,
		subQuery, orderSQL, limitSQL)
	return selQuery, args, substs
}

// selectGroupQuery returns the SQL query string and parameters to retrieve
// the result of this Query object, which must include a Group By.
// fields is the list of fields to retrieve.
//
// This query must have a Group By clause.
func (q *Query) selectGroupQuery(fieldsList []FieldName, aggFncts map[string]string) (string, SQLParams) {
	if len(q.groups) == 0 {
		log.Panic("Calling selectGroupQuery on a query without Group By clause")
	}
	// Recompute fieldsList, addy group bys
	fieldExprs, _ := q.selectData(fieldsList, true)
	fieldsList = []FieldName{}
	for _, fe := range fieldExprs {
		fieldsList = append(fieldsList, joinFieldNames(fe, ExprSep))
	}
	// Get base query
	baseQuery, baseArgs, _ := q.selectCommonQuery(fieldsList)
	// Build up the query
	// Fields
	fieldsSQL := q.fieldsGroupSQL(fieldExprs, aggFncts)
	// Group by clause
	groupSQL := q.sqlGroupByClause()
	orderSQL := q.sqlOrderByClauseForGroupBy(aggFncts)
	limitSQL := q.sqlLimitOffsetClause()
	selQuery := fmt.Sprintf(`SELECT %s, count(1) AS __count FROM (%s) base GROUP BY %s %s %s`,
		fieldsSQL, baseQuery, groupSQL, orderSQL, limitSQL)
	return selQuery, baseArgs
}

// selectData returns for this query:
// - Expressions defined by the given fields and that must appear in the field list of the select clause.
// - All expressions that also include expressions used in the where clause.
func (q *Query) selectData(fields []FieldName, withCtx bool) ([][]FieldName, [][]FieldName) {
	q.substituteChildOfPredicates()
	// Get all expressions, first given by fields
	fieldExprs := make([][]FieldName, len(fields))
	fieldsExprsMap := make(map[string][]FieldName)
	for i, f := range fields {
		fieldExprs[i] = splitFieldNames(f, ExprSep)
		fieldsExprsMap[joinFieldNames(fieldExprs[i], ExprSep).JSON()] = fieldExprs[i]
	}
	// Add 'order by' exprs removing duplicates
	oExprs := q.getOrderByExpressions(withCtx)
	for _, oExpr := range oExprs {
		if _, ok := fieldsExprsMap[joinFieldNames(oExpr, ExprSep).JSON()]; !ok {
			fieldExprs = append(fieldExprs, oExpr)
			fieldsExprsMap[joinFieldNames(oExpr, ExprSep).JSON()] = oExpr
		}
	}
	// Add 'group by' exprs removing duplicates
	gExprs := q.getGroupByExpressions()
	for _, gExpr := range gExprs {
		if _, ok := fieldsExprsMap[joinFieldNames(gExpr, ExprSep).JSON()]; !ok {
			fieldExprs = append(fieldExprs, gExpr)
		}
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
func (q *Query) fieldsSQL(fieldExprs [][]FieldName) (string, map[string]string) {
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
func (q *Query) fieldsGroupSQL(fieldExprs [][]FieldName, aggFncts map[string]string) string {
	fStr := make([]string, len(fieldExprs))
	for i, exprs := range fieldExprs {
		aggFnct := aggFncts[joinFieldNames(exprs, ExprSep).JSON()]
		if aggFnct == "" {
			fStr[i] = joinFieldNames(exprs, sqlSep).JSON()
			continue
		}
		fStr[i] = fmt.Sprintf("%s(%s) AS %s", aggFnct, joinFieldNames(exprs, sqlSep).JSON(), joinFieldNames(exprs, sqlSep).JSON())
	}
	return strings.Join(fStr, ", ")
}

// joinedFieldExpression joins the given expressions into a fields sql string
//     ['profile_id' 'user_id' 'name'] => "profiles__users".name
//     ['age'] => "mytable".age
//
// If withAlias is true, then returns fields with its alias. In this case, aliasIndex is used
// to define aliases when the nominal "profile_id__user_id__name" is longer than 64 chars.
// Returned second argument is the nominal alias and third argument is the alias actually used.
func (q *Query) joinedFieldExpression(exprs []FieldName, withAlias bool, aliasIndex int) (string, string, string) {
	joins := q.generateTableJoins(exprs)
	lastJoin := joins[len(joins)-1]
	if withAlias {
		fAlias := joinFieldNames(exprs, sqlSep).JSON()
		oldAlias := fAlias
		if len(fAlias) > maxSQLidentifierLength {
			fAlias = fmt.Sprintf("f%d", aliasIndex)
		}
		return fmt.Sprintf("%s.%s AS %s", lastJoin.alias, lastJoin.expr.JSON(), fAlias), oldAlias, fAlias
	}
	return fmt.Sprintf("%s.%s", lastJoin.alias, lastJoin.expr.JSON()), "", ""
}

// generateTableJoins transforms a list of fields expression into a list of tableJoins
// ['user_id' 'profile_id' 'age'] => []tableJoins{CurrentTable User Profile}
func (q *Query) generateTableJoins(fieldExprs []FieldName) []tableJoin {
	adapter := adapters[db.DriverName()]
	var joins []tableJoin
	curMI := q.recordSet.model
	// Create the tableJoin for the current table
	currentTableName := adapter.quoteTableName(curMI.tableName)
	var curExpr FieldName
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
		fi, ok := curMI.fields.Get(expr.JSON())
		if !ok {
			log.Panic("Unparsable Expression", "expr", joinFieldNames(fieldExprs, ExprSep))
		}
		if fi.relatedModel == nil || (i == exprsLen-1 && fi.fieldType.IsFKRelationType()) {
			// Don't create an extra join if our field is not a relation field
			// or if it is the last field of our expressions
			break
		}

		var field, otherField FieldName
		var tjExpr FieldName
		if i < exprsLen-1 {
			tjExpr = fieldExprs[i+1]
		}
		switch fi.fieldType {
		case field.Many2One, field.One2One:
			field, otherField = ID, expr
		case field.One2Many, field.Rev2One:
			field, otherField = fi.relatedModel.FieldName(fi.reverseFK), ID
			if tjExpr == nil {
				tjExpr = ID
			}
		case field.Many2Many:
			// Add relation table join
			relationTableName := adapter.quoteTableName(fi.m2mRelModel.tableName)
			alias = fmt.Sprintf("%s%s%s", alias, sqlSep, fi.m2mRelModel.tableName)
			tj := tableJoin{
				tableName:  relationTableName,
				joined:     true,
				field:      fi.m2mRelModel.FieldName(fi.m2mOurField.name),
				otherTable: curTJ,
				otherField: ID,
				alias:      adapter.quoteTableName(alias),
				expr:       fi.m2mRelModel.FieldName(fi.m2mTheirField.name),
			}
			joins = append(joins, tj)
			curTJ = &tj
			// Add relation to other table
			field, otherField = ID, fi.m2mRelModel.FieldName(fi.m2mTheirField.name)
			if tjExpr == nil {
				tjExpr = ID
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
func (q *Query) tablesSQL(fExprs [][]FieldName) (string, map[string]string) {
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

// thisTable returns the quoted table name of this query's recordset table
func (q *Query) thisTable() string {
	adapter := adapters[db.DriverName()]
	return adapter.quoteTableName(q.recordSet.model.tableName)
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
func (q *Query) substituteConditionExprs(substMap map[FieldName][]FieldName) {
	q.cond.substituteExprs(q.recordSet.model, substMap)
	for i, order := range q.orders {
		for k, v := range substMap {
			if order.field.JSON() == k.JSON() {
				q.orders[i].field = joinFieldNames(v, ExprSep)
				break
			}
		}
	}
	for i, group := range q.groups {
		for k, v := range substMap {
			if group.JSON() == k.JSON() {
				q.groups[i] = joinFieldNames(v, ExprSep)
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
func (q *Query) getAllExpressions() [][]FieldName {
	return append(q.getOrderByExpressions(true),
		append(q.getGroupByExpressions(), q.cond.getAllExpressions(q.recordSet.model)...)...)
}

// getOrderByExpressions returns all expressions used in order by clause of this query.
//
// If withCtx is true, ctxOrder expressions are also returned
func (q *Query) getOrderByExpressions(withCtx bool) [][]FieldName {
	var exprs [][]FieldName
	for _, order := range q.orders {
		oExprs := splitFieldNames(order.field, ExprSep)
		exprs = append(exprs, oExprs)
	}
	if withCtx {
		exprs = append(exprs, q.getCtxOrderByExpressions()...)
	}
	return exprs
}

// getOrderByExpressions returns expressions used in context order by clause of this query.
func (q *Query) getCtxOrderByExpressions() [][]FieldName {
	var exprs [][]FieldName
	for _, order := range q.ctxOrders {
		oExprs := splitFieldNames(order.field, ExprSep)
		exprs = append(exprs, oExprs)
	}
	return exprs
}

// getGroupByExpressions returns all expressions used in group by clause of this query.
func (q *Query) getGroupByExpressions() [][]FieldName {
	var exprs [][]FieldName
	for _, group := range q.groups {
		exprs = append(exprs, splitFieldNames(group, ExprSep))
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
