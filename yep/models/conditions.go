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
	"reflect"
	"strings"

	"github.com/npiganeau/yep/yep/models/operator"
	"github.com/npiganeau/yep/yep/tools/logging"
)

// ExprSep define the expression separation
const (
	ExprSep = "."
	sqlSep  = "__"
)

type condValue struct {
	exprs    []string
	operator operator.Operator
	arg      interface{}
	cond     *Condition
	isOr     bool
	isNot    bool
	isCond   bool
}

// A Condition represents a WHERE clause of an SQL query.
type Condition struct {
	params []condValue
}

// newCondition returns a new condition struct
func newCondition() *Condition {
	c := &Condition{}
	return c
}

// And completes the current condition with a simple AND clause : c.And().nextCond => c AND nextCond
func (c Condition) And() *ConditionStart {
	res := ConditionStart{cond: c}
	return &res
}

// AndCond completes the current condition with the given cond as an AND clause
// between brackets : c.And(cond) => c AND (cond)
func (c Condition) AndCond(cond *Condition) *Condition {
	c.params = append(c.params, condValue{cond: cond, isCond: true})
	return &c
}

// AndNot completes the current condition with a simple AND NOT clause :
// c.AndNot().nextCond => c AND NOT nextCond
func (c Condition) AndNot() *ConditionStart {
	res := ConditionStart{cond: c}
	res.nextIsNot = true
	return &res
}

// AndNotCond completes the current condition with an AND NOT clause between
// brackets : c.AndNot(cond) => c AND NOT (cond)
func (c Condition) AndNotCond(cond *Condition) *Condition {
	c.params = append(c.params, condValue{cond: cond, isCond: true, isNot: true})
	return &c
}

// Or completes the current condition both with a simple OR clause : c.Or().nextCond => c OR nextCond
func (c Condition) Or() *ConditionStart {
	res := ConditionStart{cond: c}
	res.nextIsOr = true
	return &res
}

// OrCond completes the current condition both with an OR clause between
// brackets : c.Or(cond) => c OR (cond)
func (c Condition) OrCond(cond *Condition) *Condition {
	c.params = append(c.params, condValue{cond: cond, isCond: true, isOr: true})
	return &c
}

// OrNot completes the current condition both with a simple OR NOT clause : c.OrNot().nextCond => c OR NOT nextCond
func (c Condition) OrNot() *ConditionStart {
	res := ConditionStart{cond: c}
	res.nextIsNot = true
	res.nextIsOr = true
	return &res
}

// OrNotCond completes the current condition both with an OR NOT clause between
// brackets : c.OrNot(cond) => c OR NOT (cond)
func (c Condition) OrNotCond(cond *Condition) *Condition {
	c.params = append(c.params, condValue{cond: cond, isCond: true, isOr: true, isNot: true})
	return &c
}

// A ConditionStart is an object representing a Condition when
// we just added a logical operator (AND, OR, ...) and we are
// about to add a predicate.
type ConditionStart struct {
	cond      Condition
	nextIsOr  bool
	nextIsNot bool
}

// Field adds a field path (dot separated) to this condition
func (cs ConditionStart) Field(name string) *ConditionField {
	newExprs := strings.Split(name, ExprSep)
	cp := ConditionField{cs: cs}
	cp.exprs = append(cp.exprs, newExprs...)
	return &cp
}

// FilteredOn adds a condition with a table join on the given field and
// filters the result with the given condition
func (cs ConditionStart) FilteredOn(field string, condition *Condition) *Condition {
	res := cs.cond
	for i, p := range condition.params {
		condition.params[i].exprs = append([]string{field}, p.exprs...)
	}
	res.params = append(res.params, condition.params...)
	return &res
}

// A ConditionField is a partial Condition when we have set
// a field name in a predicate and are about to add an operator.
type ConditionField struct {
	cs    ConditionStart
	exprs []string
}

// FieldName returns the field name of this ConditionField
func (c ConditionField) FieldName() FieldName {
	return FieldName(strings.Join(c.exprs, ExprSep))
}

var _ FieldNamer = ConditionField{}

// addOperator adds a condition value to the condition with the given operator and data
func (c ConditionField) addOperator(op operator.Operator, data interface{}) *Condition {
	cond := c.cs.cond
	cond.params = append(cond.params, condValue{
		exprs:    c.exprs,
		operator: op,
		arg:      data,
		isNot:    c.cs.nextIsNot,
		isOr:     c.cs.nextIsOr,
	})
	return &cond
}

// sanitizeArgs returns the given args suitable for SQL query
// In particular, retrieves the ids of a recordset if args is one.
// If multi is true, a recordset will be converted into a slice of int64
// otherwise, it will return an int64 and panic if the recordset is not
// a singleton
func sanitizeArgs(args interface{}, multi bool) interface{} {
	if rs, ok := args.(RecordSet); ok {
		if multi {
			return rs.Ids()
		}
		if len(rs.Ids()) > 1 {
			logging.LogAndPanic(log, "Trying to extract a single ID from a non singleton", "args", args)
		}
		return rs.Ids()[0]
	}
	return args
}

// Equals appends the '=' operator to the current Condition
func (c ConditionField) Equals(data interface{}) *Condition {
	return c.addOperator(operator.Equals, sanitizeArgs(data, false))
}

// NotEquals appends the '!=' operator to the current Condition
func (c ConditionField) NotEquals(data interface{}) *Condition {
	return c.addOperator(operator.NotEquals, sanitizeArgs(data, false))
}

// Greater appends the '>' operator to the current Condition
func (c ConditionField) Greater(data interface{}) *Condition {
	return c.addOperator(operator.Greater, sanitizeArgs(data, false))
}

// GreaterOrEqual appends the '>=' operator to the current Condition
func (c ConditionField) GreaterOrEqual(data interface{}) *Condition {
	return c.addOperator(operator.GreaterOrEqual, sanitizeArgs(data, false))
}

// Lower appends the '<' operator to the current Condition
func (c ConditionField) Lower(data interface{}) *Condition {
	return c.addOperator(operator.Lower, sanitizeArgs(data, false))
}

// LowerOrEqual appends the '<=' operator to the current Condition
func (c ConditionField) LowerOrEqual(data interface{}) *Condition {
	return c.addOperator(operator.LowerOrEqual, sanitizeArgs(data, false))
}

// LikePattern appends the 'LIKE' operator to the current Condition
func (c ConditionField) LikePattern(data interface{}) *Condition {
	return c.addOperator(operator.LikePattern, sanitizeArgs(data, false))
}

// ILikePattern appends the 'ILIKE' operator to the current Condition
func (c ConditionField) ILikePattern(data interface{}) *Condition {
	return c.addOperator(operator.ILikePattern, sanitizeArgs(data, false))
}

// Like appends the 'LIKE %%' operator to the current Condition
func (c ConditionField) Like(data interface{}) *Condition {
	return c.addOperator(operator.Like, sanitizeArgs(data, false))
}

// NotLike appends the 'NOT LIKE %%' operator to the current Condition
func (c ConditionField) NotLike(data interface{}) *Condition {
	return c.addOperator(operator.NotLike, sanitizeArgs(data, false))
}

// ILike appends the 'ILIKE %%' operator to the current Condition
func (c ConditionField) ILike(data interface{}) *Condition {
	return c.addOperator(operator.ILike, sanitizeArgs(data, false))
}

// NotILike appends the 'NOT ILIKE %%' operator to the current Condition
func (c ConditionField) NotILike(data interface{}) *Condition {
	return c.addOperator(operator.NotILike, sanitizeArgs(data, false))
}

// In appends the 'IN' operator to the current Condition
func (c ConditionField) In(data interface{}) *Condition {
	return c.addOperator(operator.In, sanitizeArgs(data, true))
}

// NotIn appends the 'NOT IN' operator to the current Condition
func (c ConditionField) NotIn(data interface{}) *Condition {
	return c.addOperator(operator.NotIn, sanitizeArgs(data, true))
}

// ChildOf appends the 'child of' operator to the current Condition
func (c ConditionField) ChildOf(data interface{}) *Condition {
	return c.addOperator(operator.ChildOf, sanitizeArgs(data, false))
}

// IsEmpty check the condition arguments are empty or not.
func (c *Condition) IsEmpty() bool {
	return len(c.params) == 0
}

// getAllExpressions returns a list of all exprs used in this condition,
// and recursively in all subconditions.
// Expressions are given in field json format
func (c Condition) getAllExpressions(mi *Model) [][]string {
	var res [][]string
	for _, cv := range c.params {
		res = append(res, jsonizeExpr(mi, cv.exprs))
		if cv.cond != nil {
			res = append(res, cv.cond.getAllExpressions(mi)...)
		}
	}
	return res
}

// substituteExprs recursively replaces condition exprs that match substs keys
// with the corresponding substs values.
func (c *Condition) substituteExprs(mi *Model, substs map[string][]string) {
	for i, cv := range c.params {
		for k, v := range substs {
			if len(cv.exprs) > 0 && jsonizeExpr(mi, cv.exprs)[0] == k {
				c.params[i].exprs = v
			}
		}
		if cv.cond != nil {
			cv.cond.substituteExprs(mi, substs)
		}
	}
}

// evaluateArgFunctions recursively evaluates all args in the queries that are
// functions and substitute it with the result.
func (c *Condition) evaluateArgFunctions(rc RecordCollection) {
	for i, cv := range c.params {
		if cv.cond != nil {
			cv.cond.evaluateArgFunctions(rc)
		}

		fnctVal := reflect.ValueOf(cv.arg)
		if fnctVal.Kind() != reflect.Func {
			continue
		}

		firstArgType := fnctVal.Type().In(0)
		if !firstArgType.Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
			continue
		}
		argValue := reflect.ValueOf(rc)
		if firstArgType != reflect.TypeOf(RecordCollection{}) {
			newArgValue := reflect.New(firstArgType).Elem()
			newArgValue.FieldByName("RecordCollection").Set(argValue)
			argValue = newArgValue
		}

		res := fnctVal.Call([]reflect.Value{argValue})
		c.params[i].arg = res[0].Interface()
	}
}
