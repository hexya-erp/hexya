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
)

// A RecordRef uniquely identifies a Record by giving its model and ID.
type RecordRef struct {
	ModelName string
	ID        int64
}

// RecordSet identifies a type that holds a set of records of
// a given model.
type RecordSet interface {
	// ModelName returns the name of the model of this RecordSet
	ModelName() string
	// Ids returns the ids in this set of Records
	Ids() []int64
	// Env returns the current Environment of this RecordSet
	Env() Environment
	// Len returns the number of records in this RecordSet
	Len() int
	// IsEmpty returns true if this RecordSet has no records
	IsEmpty() bool
	// Call executes the given method (as string) with the given arguments
	Call(string, ...interface{}) interface{}
	// Collection returns the underlying RecordCollection instance
	Collection() *RecordCollection
}

// A FieldName is a type representing field names in models.
type FieldName string

// FieldName makes a FieldName instance a FieldNamer
func (fn FieldName) FieldName() FieldName {
	return fn
}

// String function for FieldName
func (fn FieldName) String() string {
	return string(fn)
}

var _ FieldNamer = FieldName("")

// A FieldNamer is a type that can yield a FieldName through
// its FieldName() method
type FieldNamer interface {
	fmt.Stringer
	FieldName() FieldName
}

// A GroupAggregateRow holds a row of results of a query with a group by clause
// - Values holds the values of the actual query
// - Count is the number of lines aggregated into this one
// - Condition can be used to query the aggregated rows separately if needed
type GroupAggregateRow struct {
	Values    FieldMap
	Count     int
	Condition *Condition
}

// FieldContexts define the different contexts for a field, that will define different
// values for this field.
//
// The key is a context name and the value is a function that returns the context
// value for the given recordset.
type FieldContexts map[string]func(RecordSet) string

// A FieldMapper is an object that can convert itself into a FieldMap
type FieldMapper interface {
	// Underlying returns the object converted to a FieldMap.
	Underlying() FieldMap
}

// A Methoder can return a Method data object through its Underlying() method
type Methoder interface {
	Underlying() *Method
}

// A Modeler can return a Model data object through its Underlying() method
type Modeler interface {
	Underlying() *Model
}

// A Conditioner can return a Condition object through its Underlying() method
type Conditioner interface {
	Underlying() *Condition
}

// A ModelData is used to hold values of an object instance for creating or
// updating a RecordSet. It is mainly designed to be embedded in a type-safe
// struct.
type ModelData struct {
	FieldMap
	model *Model
}

var _ FieldMapper = ModelData{}
var _ FieldMapper = new(ModelData)

// Get returns the value of the given field.
// The second returned value is true if the value exists.
//
// The field can be either its name or is JSON name.
func (md *ModelData) Get(field string) (interface{}, bool) {
	return md.FieldMap.Get(field, md.model)
}

// Set sets the given field with the given value.
// If the field already exists, then it is updated with value.
// Otherwise, a new entry is inserted.
func (md *ModelData) Set(field string, value interface{}) {
	md.FieldMap.Set(field, value, md.model)
}

// Unset removes the value of the given field if it exists.
func (md *ModelData) Unset(field string) {
	md.FieldMap.Delete(field, md.model)
}

// NewModelData returns a pointer to a new instance of ModelData
// for the given model.
func NewModelData(model Modeler) *ModelData {
	return &ModelData{
		FieldMap: make(FieldMap),
		model:    model.Underlying(),
	}
}
