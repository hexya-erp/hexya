/*   Copyright (C) 2016 by Nicolas Piganeau
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

import (
	"fmt"
	"strings"
)

/*
Query represents a backend independant query to fetch records of a model.
*/
type Query struct {
	Domain
	Limit  string
	Offset string
	Order  string
}

/*
Domain represents a set of search criteria, typically a WHERE clause in SQL.

Domain is recursive and is composed of a DomainOperator acting on two DomainOperand.
*/
type Domain struct {
	Operator     DomainOperator
	LeftOperand  *DomainOperand
	RightOperand *DomainOperand
}

func (d *Domain) String() string {
	var fmtString string
	if d.RightOperand != nil {
		if d.Operator == OR {
			fmtString = "\"|\", %s, %s"
		} else {
			fmtString = "\"&\", %s, %s"
		}
	} else{
		fmtString = "%s"
	}
	return fmt.Sprintf(fmtString, d.LeftOperand, d.RightOperand)
}

/*
ParseDomain parses the date byte slice into the given domain.
 */
func ParseDomain(data []byte, dom *Domain) error {
	trimmedData := strings.Trim(string(data), "[]")
	terms := strings.Split(trimmedData, ",")

	return nil
}

/*
DomainOperand can be either a Leaf or a Domain.
*/
type DomainOperand struct {
	Type        OperandType
	LeafValue   *Leaf
	DomainValue *Domain
}

func (do *DomainOperand) String() string {
	if do.Type == LEAF {
		return do.LeafValue.String()
	} else {
		return do.DomainValue.String()
	}
}

/*
OperandType is the type of DomainOperand which can be either Domain or Leaf.
*/
type OperandType uint8

const (
	LEAF   OperandType = 0
	DOMAIN OperandType = 1
)

/*
Leaf is a single condition of a Domain stating that FieldName compares to Value through Operator.
*/
type Leaf struct {
	FieldName string
	Operator  LeafOperator
	Value     interface{}
}

func (l *Leaf) String() string {
	return fmt.Sprintf("(\"%s\", \"%s\", %s)", l.FieldName, l.Operator, l.Value)
}

/*
DomainOperator are used to link Leaves conditions together into a Domain.
*/
type DomainOperator string

const (
	AND DomainOperator = "&"
	OR  DomainOperator = "|"
)

/*
LeafOperator is an operator that can be found inside a Leaf
*/
type LeafOperator string

const (
	EQ        LeafOperator = "="
	NE        LeafOperator = "!="
	LE        LeafOperator = "<="
	LT        LeafOperator = "<"
	GT        LeafOperator = ">"
	GE        LeafOperator = ">="
	EQ_LIKE   LeafOperator = "=like"
	EQ_ILIKE  LeafOperator = "=ilike"
	LIKE      LeafOperator = "like"
	NOT_LIKE  LeafOperator = "not like"
	ILIKE     LeafOperator = "ilike"
	NOT_ILIKE LeafOperator = "not ilike"
	IN        LeafOperator = "in"
	NOT_IN    LeafOperator = "not in"
	CHILD_OF  LeafOperator = "child_of"
)
