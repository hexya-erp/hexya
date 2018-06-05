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
	"sort"
)

// FieldMap is a map of interface{} specifically used for holding model
// fields values.
type FieldMap map[string]interface{}

// Keys returns the FieldMap keys as a slice of strings
func (fm FieldMap) Keys() (res []string) {
	for k := range fm {
		res = append(res, k)
	}
	return
}

// OrderedKeys returns the keys of this FieldMap ordered.
//
// This has the convenient side effect of having shorter paths come before longer paths,
// which is particularly useful when creating or updating related records.
func (fm FieldMap) OrderedKeys() []string {
	keys := fm.Keys()
	sort.Strings(keys)
	return keys
}

// FieldNames returns the FieldMap keys as a slice of FieldNamer.
// As within a FieldMap, the result can be field names or JSON names
// or a mix of both.
func (fm FieldMap) FieldNames() (res []FieldNamer) {
	for k := range fm {
		res = append(res, FieldName(k))
	}
	return
}

// Values returns the FieldMap values as a slice of interface{}
func (fm FieldMap) Values() (res []interface{}) {
	for _, v := range fm {
		res = append(res, v)
	}
	return
}

// RemovePK removes the entries of our FieldMap which
// references the ID field.
func (fm *FieldMap) RemovePK() {
	delete(*fm, "id")
	delete(*fm, "ID")
}

// RemovePKIfZero removes the entries of our FieldMap which
// references the ID field if the referenced id is 0.
func (fm *FieldMap) RemovePKIfZero() {
	if idl, ok := (*fm)["id"]; ok && idl.(int64) == 0 {
		delete(*fm, "id")
	}
	if idu, ok := (*fm)["ID"]; ok && idu.(int64) == 0 {
		delete(*fm, "ID")
	}
}

// JSONized returns a new field map identical to this one but
// with all its keys switched to the JSON name of the fields
func (fm *FieldMap) JSONized(model *Model) FieldMap {
	res := make(FieldMap)
	for f, v := range *fm {
		jsonFieldName := model.JSONizeFieldName(f)
		res[jsonFieldName] = v
	}
	return res
}

// Get returns the value of the given field referring to the given model.
// field can be either a field name (or path) or a field JSON name (or path).
// The second returned value is true if the field has been found in the FieldMap
func (fm FieldMap) Get(field string, model *Model) (interface{}, bool) {
	fi := model.getRelatedFieldInfo(field)
	val, ok := fm[fi.name]
	if !ok {
		val, ok = fm[fi.json]
		if !ok {
			return nil, false
		}
	}
	return val, true
}

// MustGet returns the value of the given field referring to the given model.
// field can be either a field name (or path) or a field JSON name (or path).
// It panics if the field is not found.
func (fm FieldMap) MustGet(field string, model *Model) interface{} {
	val, ok := fm.Get(field, model)
	if !ok {
		log.Panic("Field not found in FieldMap", "field", field, "fMap", fm, "model", model)
	}
	return val
}

// Set sets the given field with the given value.
// If the field already exists, then it is updated with value.
// Otherwise, a new entry is inserted in the FieldMap with the
// JSON name of the field.
func (fm *FieldMap) Set(field string, value interface{}, model *Model) {
	fi := model.getRelatedFieldInfo(field)
	key := fi.name
	_, ok := (*fm)[key]
	if !ok {
		key = fi.json
	}
	(*fm)[key] = value
}

// Delete removes the given field from this FieldMap.
// Calling Del on a non existent field is a no op.
func (fm *FieldMap) Delete(field string, model *Model) {
	fi := model.getRelatedFieldInfo(field)
	key := fi.name
	_, ok := (*fm)[key]
	if !ok {
		key = fi.json
	}
	delete(*fm, key)
}

// MergeWith updates this FieldMap with the given other FieldMap
// If a key of the other FieldMap already exists here, the value is overridden,
// otherwise, the key is inserted with its json name.
func (fm *FieldMap) MergeWith(other FieldMap, model *Model) {
	for field, value := range other {
		fm.Set(field, value, model)
	}
}

// FieldMap returns the object converted to a FieldMap
// i.e. itself
func (fm FieldMap) FieldMap(fields ...FieldNamer) FieldMap {
	return fm
}

var _ FieldMapper = FieldMap{}

// KeySubstitution defines a key substitution in a FieldMap
type KeySubstitution struct {
	Orig string
	New  string
	Keep bool
}

// SubstituteKeys changes the column names of the given field map with the
// given substitutions.
func (fm *FieldMap) SubstituteKeys(substs []KeySubstitution) {
	for _, subs := range substs {
		value, exists := (*fm)[subs.Orig]
		if exists {
			if !subs.Keep {
				delete(*fm, subs.Orig)
			}
			(*fm)[subs.New] = value
		}
	}
}

// Copy returns a shallow copy of this FieldMap
func (fm FieldMap) Copy() FieldMap {
	res := make(FieldMap, len(fm))
	for k, v := range fm {
		res[k] = v
	}
	return res
}

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
	// FieldMap returns the object converted to a FieldMap.
	//
	// If this is a struct, only non zero fields will be added to the result.
	// To put additional fields with zero value in the map, specify them in fields.
	//
	// If this is already a FieldMap, then fields are ignored and the map is
	// returned as is.
	FieldMap(fields ...FieldNamer) FieldMap
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
