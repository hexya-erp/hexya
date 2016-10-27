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

import "github.com/npiganeau/yep/yep/tools/logging"

// A Domain is a list of search criteria (DomainTerm) in the form of
// a tuplet (field_name, operator, value).
// Domain criteria (DomainTerm) can be combined using logical operators
// in prefix form (DomainPrefixOperator)
type Domain []interface{}

// A DomainTerm is a search criterion in the form of
// a tuplet (field_name, operator, value).
type DomainTerm []interface{}

// A DomainPrefixOperator is used to combine DomainTerms
type DomainPrefixOperator string

// Domain prefix operators
const (
	PREFIX_AND DomainPrefixOperator = "&"
	PREFIX_OR  DomainPrefixOperator = "|"
	PREFIX_NOT DomainPrefixOperator = "!"
)

// A DomainOperator is an operator inside a DomainTerm.
// Each operator can be matched with an equivalent SQL operator.
type DomainOperator string

// Domain operators
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

// ParseDomain gets Domain and parses it into a RecordSet query Condition.
// Returns nil if the domain is []
func ParseDomain(dom Domain) *Condition {
	res := parseDomain(&dom)
	if res == nil {
		return nil
	}
	for len(dom) > 0 {
		res = NewCondition().AndCond(res).AndCond(parseDomain(&dom))
	}
	return res
}

// parseDomain is the internal recursive function making all the job of
// ParseDomain. The given domain through pointer is deleted during operation.
func parseDomain(dom *Domain) *Condition {
	if len(*dom) == 0 {
		return nil
	}

	res := NewCondition()
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

// addTerm parses the given DomainTerm and adds it to the given condition with the given
// prefix operator. Returns the new condition.
func addTerm(cond *Condition, term DomainTerm, op DomainPrefixOperator) *Condition {
	if len(term) != 3 {
		logging.LogAndPanic(log, "Malformed domain term", "term", term)
	}
	fieldName := term[0].(string)
	operator := DomainOperator(term[1].(string))
	value := term[2]
	meth := getConditionMethod(cond, op)
	cond = meth(fieldName, string(operator), value)
	return cond
}

// getConditionMethod returns the condition method to use on the given condition
// for the given prefix operator and negation condition.
func getConditionMethod(cond *Condition, op DomainPrefixOperator) func(string, string, interface{}) *Condition {
	switch op {
	case PREFIX_AND:
		return cond.And
	case PREFIX_OR:
		return cond.Or
	default:
		logging.LogAndPanic(log, "Unknown prefix operator", "operator", op)
	}
	return nil
}
