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
