// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package fieldtype

import (
	"reflect"

	"github.com/hexya-erp/hexya/src/models/types/dates"
)

// A Type defines a type of a model's field
type Type string

// Types for model fields
const (
	NoType    Type = ""
	Binary    Type = "binary"
	Boolean   Type = "boolean"
	Char      Type = "char"
	Date      Type = "date"
	DateTime  Type = "datetime"
	Float     Type = "float"
	HTML      Type = "html"
	Integer   Type = "integer"
	Many2Many Type = "many2many"
	Many2One  Type = "many2one"
	One2Many  Type = "one2many"
	One2One   Type = "one2one"
	Rev2One   Type = "rev2one"
	Reference Type = "reference"
	Selection Type = "selection"
	Text      Type = "text"
)

// IsRelationType returns true if this type is a relation.
func (t Type) IsRelationType() bool {
	return t == Many2Many || t == Many2One || t == One2Many || t == One2One || t == Rev2One
}

// IsFKRelationType returns true for relation types
// that are stored in the model's table (i.e. M2O and O2O)
func (t Type) IsFKRelationType() bool {
	return t == Many2One || t == One2One
}

// IsNonStoredRelationType returns true for relation types
// that are not stored in the model's table (i.e. M2M, O2M and R2O)
func (t Type) IsNonStoredRelationType() bool {
	return t == Many2Many || t == One2Many || t == Rev2One
}

// IsReverseRelationType returns true for relation types
// that are stored in the comodel's table (i.e. O2M and R2O)
func (t Type) IsReverseRelationType() bool {
	return t == One2Many || t == Rev2One
}

// Is2OneRelationType returns true for relation types
// that point to a single comodel record (i.e. M2O, O2O and R2O)
func (t Type) Is2OneRelationType() bool {
	return t == Many2One || t == One2One || t == Rev2One
}

// Is2ManyRelationType returns true for relation types
// that point to multiple comodel records (i.e. M2M and O2M)
func (t Type) Is2ManyRelationType() bool {
	return t == Many2Many || t == One2Many
}

// IsNullInDB returns true if this type's zero value is
// saved as null in database.
func (t Type) IsNullInDB() bool {
	return t.IsFKRelationType() || t == Binary || t == Char || t == Text || t == HTML || t == Selection || t == Date || t == DateTime
}

// DefaultGoType returns this Type's default Go type
func (t Type) DefaultGoType() reflect.Type {
	switch t {
	case NoType:
		return reflect.TypeOf(nil)
	case Binary, Char, Text, HTML, Selection:
		return reflect.TypeOf(*new(string))
	case Boolean:
		return reflect.TypeOf(true)
	case Date:
		return reflect.TypeOf(*new(dates.Date))
	case DateTime:
		return reflect.TypeOf(*new(dates.DateTime))
	case Float:
		return reflect.TypeOf(*new(float64))
	case Integer, Many2One, One2One, Rev2One:
		return reflect.TypeOf(*new(int64))
	case One2Many, Many2Many:
		return reflect.TypeOf(*new([]int64))
	}
	return reflect.TypeOf(nil)
}
