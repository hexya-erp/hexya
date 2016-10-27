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

package tools

// A Context is a map of objects that is passed along from function to function
// during a transaction.
type Context map[string]interface{}

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
