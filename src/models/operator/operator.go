// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package operator

// An Operator inside an SQL WHERE clause
type Operator string

// Operators
const (
	Equals         Operator = "="
	NotEquals      Operator = "!="
	Greater        Operator = ">"
	GreaterOrEqual Operator = ">="
	Lower          Operator = "<"
	LowerOrEqual   Operator = "<="
	Like           Operator = "=like"
	Contains       Operator = "like"
	NotContains    Operator = "not like"
	IContains      Operator = "ilike"
	NotIContains   Operator = "not ilike"
	ILike          Operator = "=ilike"
	In             Operator = "in"
	NotIn          Operator = "not in"
	ChildOf        Operator = "child_of"
)

var allowedOperators = map[Operator]bool{
	Equals:         true,
	NotEquals:      true,
	Greater:        true,
	GreaterOrEqual: true,
	Lower:          true,
	LowerOrEqual:   true,
	Like:           true,
	Contains:       true,
	NotContains:    true,
	IContains:      true,
	NotIContains:   true,
	ILike:          true,
	In:             true,
	NotIn:          true,
	ChildOf:        true,
}

var negativeOperators = map[Operator]bool{
	NotEquals:    true,
	NotContains:  true,
	NotIContains: true,
	NotIn:        true,
}

var positiveOperators = map[Operator]bool{
	Equals:    true,
	IContains: true,
	ILike:     true,
	Contains:  true,
	Like:      true,
	In:        true,
}

var multiOperator = map[Operator]bool{
	In:    true,
	NotIn: true,
}

// IsMulti returns true if the operator expects a array as arguments
func (o Operator) IsMulti() bool {
	return multiOperator[o]
}

// IsValid returns true if o is a known operator.
func (o Operator) IsValid() bool {
	_, res := allowedOperators[o]
	return res
}

// IsNegative returns true if this is a negative operator
func (o Operator) IsNegative() bool {
	_, res := negativeOperators[o]
	return res
}

// IsPositive returns true if this is a positive operator
func (o Operator) IsPositive() bool {
	_, res := positiveOperators[o]
	return res
}
