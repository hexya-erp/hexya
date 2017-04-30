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

type predicate struct {
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
	predicates []predicate
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
	c.predicates = append(c.predicates, predicate{cond: cond, isCond: true})
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
	c.predicates = append(c.predicates, predicate{cond: cond, isCond: true, isNot: true})
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
	c.predicates = append(c.predicates, predicate{cond: cond, isCond: true, isOr: true})
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
	c.predicates = append(c.predicates, predicate{cond: cond, isCond: true, isOr: true, isNot: true})
	return &c
}

// Serialize returns the condition as a list which mimics Odoo domains.
func (c Condition) Serialize() []interface{} {
	return serializePredicates(c.predicates)
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
	for i, p := range condition.predicates {
		condition.predicates[i].exprs = append([]string{field}, p.exprs...)
	}
	res.predicates = append(res.predicates, condition.predicates...)
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

// AddOperator adds a condition value to the condition with the given operator and data
// If multi is true, a recordset will be converted into a slice of int64
// otherwise, it will return an int64 and panic if the recordset is not
// a singleton.
//
// This method is low level and should be avoided. Use operator methods such as Equals()
// instead.
func (c ConditionField) AddOperator(op operator.Operator, data interface{}) *Condition {
	cond := c.cs.cond
	data = sanitizeArgs(data, op.IsMulti())
	if data != nil && op.IsMulti() && reflect.ValueOf(data).Kind() == reflect.Slice && reflect.ValueOf(data).Len() == 0 {
		return &cond
	}
	cond.predicates = append(cond.predicates, predicate{
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
		if len(rs.Ids()) == 0 {
			return nil
		}
		return rs.Ids()[0]
	}
	return args
}

// Equals appends the '=' operator to the current Condition
func (c ConditionField) Equals(data interface{}) *Condition {
	return c.AddOperator(operator.Equals, data)
}

// NotEquals appends the '!=' operator to the current Condition
func (c ConditionField) NotEquals(data interface{}) *Condition {
	return c.AddOperator(operator.NotEquals, data)
}

// Greater appends the '>' operator to the current Condition
func (c ConditionField) Greater(data interface{}) *Condition {
	return c.AddOperator(operator.Greater, data)
}

// GreaterOrEqual appends the '>=' operator to the current Condition
func (c ConditionField) GreaterOrEqual(data interface{}) *Condition {
	return c.AddOperator(operator.GreaterOrEqual, data)
}

// Lower appends the '<' operator to the current Condition
func (c ConditionField) Lower(data interface{}) *Condition {
	return c.AddOperator(operator.Lower, data)
}

// LowerOrEqual appends the '<=' operator to the current Condition
func (c ConditionField) LowerOrEqual(data interface{}) *Condition {
	return c.AddOperator(operator.LowerOrEqual, data)
}

// LikePattern appends the 'LIKE' operator to the current Condition
func (c ConditionField) LikePattern(data interface{}) *Condition {
	return c.AddOperator(operator.LikePattern, data)
}

// ILikePattern appends the 'ILIKE' operator to the current Condition
func (c ConditionField) ILikePattern(data interface{}) *Condition {
	return c.AddOperator(operator.ILikePattern, data)
}

// Like appends the 'LIKE %%' operator to the current Condition
func (c ConditionField) Like(data interface{}) *Condition {
	return c.AddOperator(operator.Like, data)
}

// NotLike appends the 'NOT LIKE %%' operator to the current Condition
func (c ConditionField) NotLike(data interface{}) *Condition {
	return c.AddOperator(operator.NotLike, data)
}

// ILike appends the 'ILIKE %%' operator to the current Condition
func (c ConditionField) ILike(data interface{}) *Condition {
	return c.AddOperator(operator.ILike, data)
}

// NotILike appends the 'NOT ILIKE %%' operator to the current Condition
func (c ConditionField) NotILike(data interface{}) *Condition {
	return c.AddOperator(operator.NotILike, data)
}

// In appends the 'IN' operator to the current Condition
func (c ConditionField) In(data interface{}) *Condition {
	return c.AddOperator(operator.In, data)
}

// NotIn appends the 'NOT IN' operator to the current Condition
func (c ConditionField) NotIn(data interface{}) *Condition {
	return c.AddOperator(operator.NotIn, data)
}

// ChildOf appends the 'child of' operator to the current Condition
func (c ConditionField) ChildOf(data interface{}) *Condition {
	return c.AddOperator(operator.ChildOf, data)
}

// IsEmpty check the condition arguments are empty or not.
func (c *Condition) IsEmpty() bool {
	switch {
	case c == nil:
		return false
	case len(c.predicates) == 0:
		return true
	case len(c.predicates) == 1 && c.predicates[0].cond.IsEmpty():
		return true
	}
	return false
}

// getAllExpressions returns a list of all exprs used in this condition,
// and recursively in all subconditions.
// Expressions are given in field json format
func (c Condition) getAllExpressions(mi *Model) [][]string {
	var res [][]string
	for _, p := range c.predicates {
		res = append(res, jsonizeExpr(mi, p.exprs))
		if p.cond != nil {
			res = append(res, p.cond.getAllExpressions(mi)...)
		}
	}
	return res
}

// substituteExprs recursively replaces condition exprs that match substs keys
// with the corresponding substs values.
func (c *Condition) substituteExprs(mi *Model, substs map[string][]string) {
	for i, p := range c.predicates {
		for k, v := range substs {
			if len(p.exprs) > 0 && jsonizeExpr(mi, p.exprs)[0] == k {
				c.predicates[i].exprs = v
			}
		}
		if p.cond != nil {
			p.cond.substituteExprs(mi, substs)
		}
	}
}

// evaluateArgFunctions recursively evaluates all args in the queries that are
// functions and substitute it with the result.
func (c *Condition) evaluateArgFunctions(rc RecordCollection) {
	for i, p := range c.predicates {
		if p.cond != nil {
			p.cond.evaluateArgFunctions(rc)
		}

		fnctVal := reflect.ValueOf(p.arg)
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
		c.predicates[i].arg = res[0].Interface()
	}
}
