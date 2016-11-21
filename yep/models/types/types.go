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

package types

// A Context is a map of objects that is passed along from function to function
// during a transaction. A Context is read only.
type Context struct {
	values map[string]interface{}
}

// Copy returns a shallow copy of the Context
func (c Context) Copy() *Context {
	newCtx := NewContext()
	for k, v := range c.values {
		newCtx.values[k] = v
	}
	return newCtx
}

// Get returns the value of this Context for the given key
func (c *Context) Get(key string) interface{} {
	value := c.values[key]
	return value
}

// HasKey returns true if this Context has the given key
func (c *Context) HasKey(key string) bool {
	_, exists := c.values[key]
	return exists
}

// SetEntry returns a copy of this context with the given key/value.
// If key already exists, it is overwritten.
func (c Context) SetEntry(key string, value interface{}) *Context {
	c.values[key] = value
	return &c
}

// IsEmpty returns true if this Context has no entries.
func (c Context) IsEmpty() bool {
	if len(c.values) == 0 {
		return true
	}
	return false
}

// NewContext returns a new Context instance
func NewContext() *Context {
	return &Context{
		values: make(map[string]interface{}),
	}
}

// Digits is a tuple of 2 ints specifying respectively:
// - The precision: the total number of digits
// - The scale: the number of digits to the right of the decimal point
// (PostgresSQL definitions)
type Digits [2]int

// A FieldType defines a type of a model's field
type FieldType string

// Types for model fields
const (
	NoType    FieldType = ""
	Binary    FieldType = "binary"
	Boolean   FieldType = "boolean"
	Char      FieldType = "char"
	Date      FieldType = "date"
	DateTime  FieldType = "datetime"
	Float     FieldType = "float"
	HTML      FieldType = "html"
	Integer   FieldType = "integer"
	Many2Many FieldType = "many2many"
	Many2One  FieldType = "many2one"
	One2Many  FieldType = "one2many"
	One2One   FieldType = "one2one"
	Rev2One   FieldType = "rev2one"
	Reference FieldType = "reference"
	Selection FieldType = "selection"
	Text      FieldType = "text"
)

// IsRelationType returns true if this type is a relation.
func (t FieldType) IsRelationType() bool {
	return t == Many2Many || t == Many2One || t == One2Many || t == One2One || t == Rev2One
}

// IsStoredRelationType returns true for relation types
// that are stored in the model's table (i.e. M2O and O2O)
func (t FieldType) IsStoredRelationType() bool {
	return t == Many2One || t == One2One
}

// IsNonStoredRelationType returns true for relation types
// that are not stored in the model's table (i.e. M2M, O2M and R2O)
func (t FieldType) IsNonStoredRelationType() bool {
	return t == Many2Many || t == One2Many || t == Rev2One
}

// IsReverseRelationType returns true for relation types
// that are stored in the comodel's table (i.e. O2M and R2O)
func (t FieldType) IsReverseRelationType() bool {
	return t == One2Many || t == Rev2One
}

// Is2OneRelationType returns true for relation types
// that point to a single comodel record (i.e. M2O, O2O and R2O)
func (t FieldType) Is2OneRelationType() bool {
	return t == Many2One || t == One2One || t == Rev2One
}

// Is2ManyRelationType returns true for relation types
// that point to multiple comodel records (i.e. M2M and O2M)
func (t FieldType) Is2ManyRelationType() bool {
	return t == Many2Many || t == One2Many
}
