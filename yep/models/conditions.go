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
	"strings"

	"github.com/npiganeau/yep/yep/tools/logging"
)

// ExprSep define the expression separation
const (
	ExprSep = "."
	sqlSep  = "__"
)

type condValue struct {
	exprs    []string
	operator DomainOperator
	arg      interface{}
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
func checkArgs(expr, op string, arg interface{}) {
	if expr == "" || op == "" || arg == nil {
		logging.LogAndPanic(log, "Condition arguments cannot empty", "expr", expr, "operator", op, "arg", arg)
	}
	dop := DomainOperator(op)
	if !allowedOperators[dop] {
		logging.LogAndPanic(log, "Unknown operator", "operator", op)
	}
}

// And add expression to condition
func (c Condition) And(expr string, op string, arg interface{}) *Condition {
	checkArgs(expr, op, arg)
	c.params = append(c.params, condValue{
		exprs:    strings.Split(expr, ExprSep),
		operator: DomainOperator(op),
		arg:      arg,
	})
	return &c
}

// AndNot add NOT expression to condition
func (c Condition) AndNot(expr string, op string, arg interface{}) *Condition {
	checkArgs(expr, op, arg)
	c.params = append(c.params, condValue{
		exprs:    strings.Split(expr, ExprSep),
		operator: DomainOperator(op),
		arg:      arg,
		isNot:    true,
	})
	return &c
}

// AndCond combine a condition to current condition
func (c *Condition) AndCond(cond *Condition) *Condition {
	c = c.clone()
	if c == cond {
		logging.LogAndPanic(log, "Cannot use self as sub condition", "condition", c)
	}
	if cond != nil {
		c.params = append(c.params, condValue{cond: cond, isCond: true})
	}
	return c
}

// Or add OR expression to condition
func (c Condition) Or(expr string, op string, arg interface{}) *Condition {
	checkArgs(expr, op, arg)
	c.params = append(c.params, condValue{
		exprs:    strings.Split(expr, ExprSep),
		operator: DomainOperator(op),
		arg:      arg,
		isOr:     true,
	})
	return &c
}

// OrNot add OR NOT expression to condition
func (c Condition) OrNot(expr string, op string, arg interface{}) *Condition {
	checkArgs(expr, op, arg)
	c.params = append(c.params, condValue{
		exprs:    strings.Split(expr, ExprSep),
		operator: DomainOperator(op),
		arg:      arg,
		isNot:    true,
		isOr:     true,
	})
	return &c
}

// OrCond combine a OR condition to current condition
func (c *Condition) OrCond(cond *Condition) *Condition {
	c = c.clone()
	if c == cond {
		logging.LogAndPanic(log, "Cannot use self as sub condition", "condition", c)
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
// Expressions are given in field json format
func (c Condition) getAllExpressions(mi *modelInfo) [][]string {
	var res [][]string
	for _, cv := range c.params {
		res = append(res, jsonizeExpr(mi, cv.exprs))
		if cv.cond != nil {
			res = append(res, cv.cond.getAllExpressions(mi)...)
		}
	}
	return res
}
