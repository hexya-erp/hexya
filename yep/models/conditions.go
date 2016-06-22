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

// ExprSep define the expression separation
const (
	ExprSep = "."
	sqlSep  = "__"
)

type condValue struct {
	exprs    []string
	operator DomainOperator
	args     SQLParams
	cond     *Condition
	isOr     bool
	isNot    bool
	isCond   bool
}

// Condition struct.
// work for WHERE conditions.
type Condition struct {
	params []condValue
}

// NewCondition return new condition struct
func NewCondition() *Condition {
	c := &Condition{}
	return c
}

// checkArgs check expressions, operator and args and panics if
// they are not valid
func checkArgs(expr, op string, args ...interface{}) {
	if expr == "" || len(args) == 0 {
		panic(fmt.Errorf("<Condition> args cannot empty"))
	}
	dop := DomainOperator(op)
	if !allowedOperators[dop] {
		panic(fmt.Errorf("<Condition> unknown operator `%s`", op))
	}
}

// And add expression to condition
func (c Condition) And(expr string, op string, args ...interface{}) *Condition {
	checkArgs(expr, op, args...)
	c.params = append(c.params, condValue{
		exprs:    strings.Split(expr, ExprSep),
		operator: DomainOperator(op),
		args:     SQLParams(args),
	})
	return &c
}

// AndNot add NOT expression to condition
func (c Condition) AndNot(expr string, op string, args ...interface{}) *Condition {
	checkArgs(expr, op, args...)
	c.params = append(c.params, condValue{
		exprs:    strings.Split(expr, ExprSep),
		operator: DomainOperator(op),
		args:     SQLParams(args),
		isNot:    true,
	})
	return &c
}

// AndCond combine a condition to current condition
func (c *Condition) AndCond(cond *Condition) *Condition {
	c = c.clone()
	if c == cond {
		panic(fmt.Errorf("<Condition.AndCond> cannot use self as sub cond"))
	}
	if cond != nil {
		c.params = append(c.params, condValue{cond: cond, isCond: true})
	}
	return c
}

// Or add OR expression to condition
func (c Condition) Or(expr string, op string, args ...interface{}) *Condition {
	checkArgs(expr, op, args...)
	c.params = append(c.params, condValue{
		exprs:    strings.Split(expr, ExprSep),
		operator: DomainOperator(op),
		args:     SQLParams(args),
		isOr:     true,
	})
	return &c
}

// OrNot add OR NOT expression to condition
func (c Condition) OrNot(expr string, op string, args ...interface{}) *Condition {
	checkArgs(expr, op, args...)
	c.params = append(c.params, condValue{
		exprs:    strings.Split(expr, ExprSep),
		operator: DomainOperator(op),
		args:     SQLParams(args),
		isNot:    true,
		isOr:     true,
	})
	return &c
}

// OrCond combine a OR condition to current condition
func (c *Condition) OrCond(cond *Condition) *Condition {
	c = c.clone()
	if c == cond {
		panic(fmt.Errorf("<Condition.OrCond> cannot use self as sub cond"))
	}
	if cond != nil {
		c.params = append(c.params, condValue{cond: cond, isCond: true, isOr: true})
	}
	return c
}

// IsEmpty check the condition arguments are empty or not.
func (c *Condition) IsEmpty() bool {
	return len(c.params) == 0
}

// clone clone a condition
func (c Condition) clone() *Condition {
	return &c
}

// getAllExpressions returns a list of all exprs used in this condition,
// and recursively in all subconditions.
// Expressions are given in column name format
func (c Condition) getAllExpressions(mi *modelInfo) [][]string {
	var res [][]string
	for _, cv := range c.params {
		res = append(res, columnizeExpr(mi, cv.exprs))
		if cv.cond != nil {
			res = append(res, cv.cond.getAllExpressions(mi)...)
		}
	}
	return res
}
