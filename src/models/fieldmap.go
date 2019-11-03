// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
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

// FieldNames returns the keys of this FieldMap as FieldNames of the given model
func (fm FieldMap) FieldNames(model *Model) FieldNames {
	res := make(FieldNames, len(fm))
	var i int
	for k := range fm {
		res[i] = model.FieldName(k)
		i++
	}
	return res
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

// Get returns the value of the given field referring to the given model.
// field can be either a field name (or path) or a field JSON name (or path).
// The second returned value is true if the field has been found in the FieldMap
func (fm FieldMap) Get(field FieldName) (interface{}, bool) {
	val, ok := fm[field.Name()]
	if !ok {
		val, ok = fm[field.JSON()]
		if !ok {
			return nil, false
		}
	}
	return val, true
}

// MustGet returns the value of the given field referring to the given model.
// field can be either a field name (or path) or a field JSON name (or path).
// It panics if the field is not found.
func (fm FieldMap) MustGet(field FieldName) interface{} {
	val, ok := fm.Get(field)
	if !ok {
		log.Panic("Field not found in FieldMap", "field", field.Name(), "fMap", fm)
	}
	return val
}

// Set sets the given field with the given value.
// If the field already exists, then it is updated with value.
// Otherwise, a new entry is inserted in the FieldMap with the
// JSON name of the field.
func (fm *FieldMap) Set(field FieldName, value interface{}) {
	key := field.Name()
	_, ok := (*fm)[key]
	if !ok {
		key = field.JSON()
	}
	(*fm)[key] = value
}

// Delete removes the given field from this FieldMap.
// Calling Del on a non existent field is a no op.
func (fm *FieldMap) Delete(field FieldName) {
	key := field.Name()
	_, ok := (*fm)[key]
	if !ok {
		key = field.JSON()
	}
	delete(*fm, key)
}

// MergeWith updates this FieldMap with the given other FieldMap
// If a key of the other FieldMap already exists here, the value is overridden,
// otherwise, the key is inserted with its json name.
func (fm *FieldMap) MergeWith(other FieldMap, model *Model) {
	for field, value := range other {
		fm.Set(model.FieldName(field), value)
	}
}

// Underlying returns the object converted to a FieldMap
// i.e. itself
func (fm FieldMap) Underlying() FieldMap {
	return fm
}

var _ FieldMapper = FieldMap{}

// Copy returns a shallow copy of this FieldMap
func (fm FieldMap) Copy() FieldMap {
	res := make(FieldMap, len(fm))
	for k, v := range fm {
		res[k] = v
	}
	return res
}
