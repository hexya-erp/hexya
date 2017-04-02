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
	LikePattern    Operator = "=like"
	Like           Operator = "like"
	NotLike        Operator = "not like"
	ILike          Operator = "ilike"
	NotILike       Operator = "not ilike"
	ILikePattern   Operator = "=ilike"
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
	LikePattern:    true,
	Like:           true,
	NotLike:        true,
	ILike:          true,
	NotILike:       true,
	ILikePattern:   true,
	In:             true,
	NotIn:          true,
	ChildOf:        true,
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
