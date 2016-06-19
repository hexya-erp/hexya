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

/*
Context is a map of objects that is passed along from function to function
during a transaction.
*/
type Context map[string]interface{}

// Digits is a tuple of 2 ints specifying respectively:
// - The precision: the total number of digits
// - The scale: the number of digits to the right of the decimal point
// (PostgresSQL definitions)
type Digits [2]int

type FieldType string

const (
	NO_TYPE   FieldType = ""
	BINARY    FieldType = "binary"
	BOOLEAN   FieldType = "boolean"
	CHAR      FieldType = "char"
	DATE      FieldType = "date"
	DATETIME  FieldType = "datetime"
	FLOAT     FieldType = "float"
	HTML      FieldType = "html"
	INTEGER   FieldType = "integer"
	MANY2MANY FieldType = "many2many"
	MANY2ONE  FieldType = "many2one"
	ONE2MANY  FieldType = "one2many"
	ONE2ONE   FieldType = "one2one"
	REV2ONE FieldType = "rev2one"
	REFERENCE FieldType = "reference"
	SELECTION FieldType = "selection"
	TEXT      FieldType = "text"
)
