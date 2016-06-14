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
	sqlSep = "__"
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

// sqlClause returns the sql WHERE clause for this condValue.
// First argument is the database/sql driver name
// If 'first' is given and true, then the sql clause is not prefixed with
// 'AND' and panics if isOr is true.
func (cv *condValue) sqlClause(driverName string, first ...bool) (string, SQLParams) {
	var (
		sql string
		args SQLParams
		isFirst bool
	)
	if len(first) > 0 {
		isFirst = first[0]
		if cv.isOr {
			panic(fmt.Errorf("First WHERE clause cannot be OR"))
		}
	}

	if cv.isOr {
		sql += "OR "
	} else if !isFirst {
		sql += "AND "
	}
	if cv.isNot {
		sql += "NOT "
	}

	if cv.isCond {
		sql, args = cv.cond.sqlClause(driverName)
	} else {
		var field string
		num := len(cv.exprs)
		if len(cv.exprs) > 1 {
			field = fmt.Sprintf("%s.%s", strings.Join(cv.exprs[:num - 1], sqlSep), cv.exprs[num - 1])
		} else {
			field = cv.exprs[0]
		}
		sql = fmt.Sprintf(`%s %s`, field, drivers[driverName].operatorSQL(cv.operator))
		args = cv.args
	}
	return sql, args
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

// sqlClauses returns the sql string and parameters corresponding to the
// WHERE clause of this Condition.
// Argument is the database/sql driver name
func (c Condition) sqlClause(driver string) (string, SQLParams) {
	if c.IsEmpty() {
		return "", SQLParams{}
	}

	return "", SQLParams{}
}
