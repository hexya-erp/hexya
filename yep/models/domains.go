/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package models

import "fmt"

/*
Domain is a list of search criteria (DomainTerm) in the form of a tuplet (field_name, operator, value).
Domain criteria (DomainTerm) can be combined using logical operators in prefix form (DomainPrefixOperator)
*/
type Domain []interface{}

type DomainTerm []interface{}

type DomainPrefixOperator string

const (
	PREFIX_AND DomainPrefixOperator = "&"
	PREFIX_OR  DomainPrefixOperator = "|"
	PREFIX_NOT DomainPrefixOperator = "!"
)

type DomainOperator string

const (
	OPERATOR_EQUALS        DomainOperator = "="
	OPERATOR_NOT_EQUALS    DomainOperator = "!="
	OPERATOR_GREATER       DomainOperator = ">"
	OPERATOR_GREATER_EQUAL DomainOperator = ">="
	OPERATOR_LOWER         DomainOperator = "<"
	OPERATOR_LOWER_EQUAL   DomainOperator = "<="
	OPERATOR_UNSET_EQUALS  DomainOperator = "=?"
	OPERATOR_LIKE_PATTERN  DomainOperator = "=like"
	OPERATOR_LIKE          DomainOperator = "like"
	OPERATOR_NOT_LIKE      DomainOperator = "not like"
	OPERATOR_ILIKE         DomainOperator = "ilike"
	OPERATOR_NOT_ILIKE     DomainOperator = "not ilike"
	OPERATOR_ILIKE_PATTERN DomainOperator = "=ilike"
	OPERATOR_IN            DomainOperator = "in"
	OPERATOR_NOT_IN        DomainOperator = "not in"
	OPERATOR_CHILD_OF      DomainOperator = "child_of"
)

var allowedOperators = map[DomainOperator]bool{
	OPERATOR_EQUALS:        true,
	OPERATOR_NOT_EQUALS:    true,
	OPERATOR_GREATER:       true,
	OPERATOR_GREATER_EQUAL: true,
	OPERATOR_LOWER:         true,
	OPERATOR_LOWER_EQUAL:   true,
	OPERATOR_UNSET_EQUALS:  true,
	OPERATOR_LIKE_PATTERN:  true,
	OPERATOR_LIKE:          true,
	OPERATOR_NOT_LIKE:      true,
	OPERATOR_ILIKE:         true,
	OPERATOR_NOT_ILIKE:     true,
	OPERATOR_ILIKE_PATTERN: true,
	OPERATOR_IN:            true,
	OPERATOR_NOT_IN:        true,
	OPERATOR_CHILD_OF:      true,
}

/*
ParseDomain gets an Odoo domain and parses it into an orm.Condition.
*/
func ParseDomain(dom Domain) *Condition {
	res := parseDomain(&dom)
	for len(dom) > 0 {
		res = NewCondition().AndCond(res).AndCond(parseDomain(&dom))
	}
	return res
}

// parseDomain is the internal recursive function making all the job of
// ParseDomain. The given domain through pointer is deleted during operation.
func parseDomain(dom *Domain) *Condition {
	res := NewCondition()
	if len(*dom) == 0 {
		return res
	}

	currentOp := PREFIX_AND

	operatorTerm := (*dom)[0]
	firstTerm := (*dom)[0]
	if ftStr, ok := operatorTerm.(string); ok {
		currentOp = DomainPrefixOperator(ftStr)
		*dom = (*dom)[1:]
		firstTerm = (*dom)[0]
	}

	switch ft := firstTerm.(type) {
	case string:
		// We have a unary operator '|' or '&', so this is an included condition
		// We have AndCond because this is the first term.
		res = res.AndCond(parseDomain(dom))
	case []interface{}:
		// We have a domain leaf ['field', 'op', value]
		term := DomainTerm(ft)
		res = addTerm(res, term, currentOp)
		*dom = (*dom)[1:]
	}

	// dom has been reduced in previous step
	// check if we still have terms to add
	if len(*dom) > 0 {
		secondTerm := (*dom)[0]
		switch secondTerm.(type) {
		case string:
			// We have a unary operator '|' or '&', so this is an included condition
			switch currentOp {
			case PREFIX_OR:
				res = res.OrCond(parseDomain(dom))
			default:
				res = res.AndCond(parseDomain(dom))
			}
		case []interface{}:
			term := DomainTerm(secondTerm.([]interface{}))
			res = addTerm(res, term, currentOp)
			*dom = (*dom)[1:]
		}
	}
	return res
}

/*
addTerm parses the given DomainTerm and adds it to the given condition with the given
prefix operator. Returns the new condition.
*/
func addTerm(cond *Condition, term DomainTerm, op DomainPrefixOperator) *Condition {
	if len(term) != 3 {
		panic(fmt.Errorf("Malformed domain term: %v", term))
	}
	fieldName := term[0].(string)
	operator := DomainOperator(term[1].(string))
	value := term[2]
	meth := getConditionMethod(cond, op)
	cond = meth(fieldName, string(operator), value)
	return cond
}

/*
getConditionMethod returns the condition method to use on the given condition
for the given prefix operator and negation condition.
*/
func getConditionMethod(cond *Condition, op DomainPrefixOperator) func(string, string, ...interface{}) *Condition {
	switch op {
	case PREFIX_AND:
		return cond.And
	case PREFIX_OR:
		return cond.Or
	default:
		panic(fmt.Errorf("Unknown prefix operator: %s", op))
	}
}
