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
	"reflect"
	"strconv"
	"strings"

	"github.com/npiganeau/yep/yep/tools/logging"
)

// RecordCollection is a generic struct representing several
// records of a model.
type RecordCollection struct {
	mi        *modelInfo
	callStack []*methodLayer
	query     *Query
	env       *Environment
	ids       []int64
}

// String returns the string representation of a RecordSet
func (rs RecordCollection) String() string {
	idsStr := make([]string, len(rs.ids))
	for i, id := range rs.ids {
		idsStr[i] = strconv.Itoa(int(id))
		i++
	}
	rsIds := strings.Join(idsStr, ",")
	return fmt.Sprintf("%s(%s)", rs.mi.name, rsIds)
}

// Env returns the RecordSet's Environment
func (rs RecordCollection) Env() Environment {
	res := *rs.env
	return res
}

// ModelName returns the model name of the RecordSet
func (rs RecordCollection) ModelName() string {
	return rs.mi.name
}

// Ids returns the ids of the RecordSet
func (rs RecordCollection) Ids() []int64 {
	return rs.ids
}

// create inserts a new record in the database with the given data.
// data can be either a FieldMap or a struct pointer of the same model as rs.
// This function is private and low level. It should not be called directly.
// Instead use rs.Create(), rs.Call("Create") or env.Create()
func (rs RecordCollection) create(data interface{}) RecordCollection {
	fMap := convertInterfaceToFieldMap(data)
	rs.mi.convertValuesToFieldType(&fMap)
	// clean our fMap from ID and non stored fields
	if idl, ok := fMap["id"]; ok && idl.(int64) == 0 {
		delete(fMap, "id")
	}
	if idu, ok := fMap["ID"]; ok && idu.(int64) == 0 {
		delete(fMap, "ID")
	}

	for _, cf := range rs.mi.fields.registryByJSON {
		if !cf.isStored() {
			delete(fMap, cf.name)
			delete(fMap, cf.json)
		}
	}
	// insert in DB
	sql, args := rs.query.insertQuery(fMap)
	var createdId int64
	DBGet(rs.env.cr, &createdId, sql, args...)
	// compute stored fields
	rSet := rs.withIds([]int64{createdId})
	rSet.updateStoredFields(fMap)
	return rSet
}

// update updates the database with the given data and returns the number of updated rows.
// It panics in case of error.
// This function is private and low level. It should not be called directly.
// Instead use rs.Write() or rs.Call("Write")
func (rc RecordCollection) update(data interface{}, fieldsToUpdate ...string) bool {
	fMap := convertInterfaceToFieldMap(data)
	if _, ok := data.(FieldMap); !ok {
		for _, f := range fieldsToUpdate {
			if _, exists := fMap[f]; !exists {
				fMap[f] = nil
			}
		}
	}
	rc.mi.convertValuesToFieldType(&fMap)
	// clean our fMap from ID and non stored fields
	delete(fMap, "id")
	delete(fMap, "ID")
	for fName := range fMap {
		if fi := rc.mi.getRelatedFieldInfo(fName); !fi.isStored() {
			delete(fMap, fi.name)
			delete(fMap, fi.json)
		}
	}
	// invalidate cache
	// We do it before the actual write on purpose so that we are sure it
	// is invalidated, even in case of error.
	for _, id := range rc.Ids() {
		rc.env.cache.invalidateRecord(rc.mi, id)
	}
	// update DB
	sql, args := rc.query.updateQuery(fMap)
	DBExecute(rc.env.cr, sql, args...)
	// compute stored fields
	rc.updateStoredFields(fMap)
	return true
}

// delete deletes the database record of this RecordSet and returns the number of deleted rows.
// This function is private and low level. It should not be called directly.
// Instead use rs.Unlink() or rs.Call("Unlink")
func (rs RecordCollection) delete() int64 {
	sql, args := rs.query.deleteQuery()
	res := DBExecute(rs.env.cr, sql, args...)
	num, _ := res.RowsAffected()
	return num
}

// Filter returns a new RecordSet filtered on records matching the given additional condition.
func (rs RecordCollection) Filter(fieldName, op string, data interface{}) RecordCollection {
	rs.query.cond = rs.query.cond.And(fieldName, op, data)
	return rs
}

// Exclude returns a new RecordSet filtered on records NOT matching the given additional condition.
func (rs RecordCollection) Exclude(fieldName, op string, data interface{}) RecordCollection {
	rs.query.cond = rs.query.cond.AndNot(fieldName, op, data)
	return rs
}

// Search returns a new RecordSet filtering on the current one with the
// additional given Condition
func (rs RecordCollection) Search(cond *Condition) RecordCollection {
	rs.query.cond = rs.query.cond.AndCond(cond)
	return rs
}

// Limit returns a new RecordSet with only the first 'limit' records.
func (rs RecordCollection) Limit(limit int) RecordCollection {
	rs.query.limit = limit
	return rs
}

// Offset returns a new RecordSet with only the records starting at offset
func (rs RecordCollection) Offset(offset int) RecordCollection {
	rs.query.offset = offset
	return rs
}

// OrderBy returns a new RecordSet ordered by the given ORDER BY expressions
func (rs RecordCollection) OrderBy(exprs ...string) RecordCollection {
	rs.query.orders = append(rs.query.orders, exprs...)
	return rs
}

// GroupBy returns a new RecordSet grouped with the given GROUP BY expressions
func (rs RecordCollection) GroupBy(exprs ...string) RecordCollection {
	rs.query.groups = append(rs.query.groups, exprs...)
	return rs
}

// Distinct returns a new RecordSet without duplicates
func (rs RecordCollection) Distinct() RecordCollection {
	rs.query.distinct = true
	return rs
}

// Fetch query the database with the current filter and returns a RecordSet
// with the queries ids. Fetch is lazy and only return ids. Use Load() instead
// if you want to fetch all fields.
func (rs RecordCollection) Fetch() RecordCollection {
	if len(rs.Ids()) == 0 && !rs.query.isEmpty() {
		// We do not load empty queries to keep empty record sets empty
		// Call Load instead to load all the records of the table
		return rs.Load("id")
	}
	return rs
}

/*
SearchCount fetch from the database the number of records that match the RecordSet conditions
It panics in case of error
*/
func (rs RecordCollection) SearchCount() int {
	sql, args := rs.query.countQuery()
	var res int
	DBGet(rs.env.cr, &res, sql, args...)
	return res
}

// Load query all data of the RecordCollection and store in cache.
// fields are the fields to retrieve in the expression format,
// i.e. "User.Profile.Age" or "user_id.profile_id.age".
// If no fields are given, all DB columns of the RecordCollection's
// model are retrieved.
func (rc RecordCollection) Load(fields ...string) RecordCollection {
	var results []FieldMap
	if len(fields) == 0 {
		fields = rc.mi.fields.storedFieldNames()
	}
	subFields, substs := rc.substituteRelatedFields(fields)
	dbFields := filterOnDBFields(rc.mi, subFields)
	sql, args := rc.query.selectQuery(dbFields)
	rows := DBQuery(rc.env.cr, sql, args...)
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		line := make(FieldMap)
		err := rc.mi.scanToFieldMap(rows, &line)
		line.SubstituteKeys(substs)
		if err != nil {
			logging.LogAndPanic(log, err.Error(), "model", rc.ModelName(), "fields", fields)
		}
		results = append(results, line)
		rc.env.cache.addRecord(rc.mi, line["id"].(int64), line)
		ids = append(ids, line["id"].(int64))
	}

	rSet := rc.withIds(ids)
	return rSet
}

// Get returns the value of the given fieldName for the first record of this RecordCollection.
// It returns the type's zero value if the RecordCollection is empty.
func (rc RecordCollection) Get(fieldName string) interface{} {
	rSet := rc.Fetch()
	fi, ok := rSet.mi.fields.get(fieldName)
	if !ok {
		logging.LogAndPanic(log, "Unknown field in model", "model", rSet.ModelName(), "field", fieldName)
	}
	var res interface{}
	if rSet.IsEmpty() {
		res = reflect.Zero(fi.structField.Type).Interface()
	} else if fi.isComputedField() && !fi.isStored() {
		fMap := make(FieldMap)
		rSet.computeFieldValues(&fMap, fi.json)
		res = fMap[fi.json]
	} else if fi.isRelatedField() && !fi.isStored() {
		rSet.Load(fi.relatedPath)
		res = rSet.env.cache.get(rSet.mi, rSet.ids[0], fi.relatedPath)
	} else {
		if !rSet.env.cache.checkIfInCache(rSet.mi, []int64{rSet.ids[0]}, []string{fi.json}) {
			// If value is not in cache we fetch the whole model to speed up
			// later calls to Get. The user can call Read with fields beforehand
			// in order not to have this behaviour.
			rSet.Load()
		}
		res = rSet.env.cache.get(rSet.mi, rSet.ids[0], fi.json)
	}
	if fi.isRelationField() {
		switch r := res.(type) {
		case int64:
			res = newRecordCollection(rSet.Env(), fi.relatedModel.name)
			if r != 0 {
				res = res.(RecordCollection).withIds([]int64{r})
			}
		case []int64:
			res = newRecordCollection(rSet.Env(), fi.relatedModel.name).withIds(r)
		}
	}
	return res
}

// Set sets field given by fieldName to the given value. If the RecordSet has several
// Records, all of them will be updated. Each call to Set makes an update query in the
// database. It panics if it is called on an empty RecordSet.
func (rc RecordCollection) Set(fieldName string, value interface{}) {
	rSet := rc.Fetch()
	if rSet.IsEmpty() {
		logging.LogAndPanic(log, "Call to Set on empty RecordSet", "model", rSet.ModelName(), "field", fieldName, "value", value)
	}
	fMap := make(FieldMap)
	fMap[fieldName] = value
	rSet.Call("Write", fMap)
}

// First populates structPtr with a copy of the first Record of the RecordCollection.
// structPtr must a pointer to a struct.
func (rc RecordCollection) First(structPtr interface{}) {
	rSet := rc.Fetch()
	if err := checkStructPtr(structPtr); err != nil {
		logging.LogAndPanic(log, "Invalid structPtr given", "error", err, "model", rSet.ModelName(), "received", structPtr)
	}
	if rSet.IsEmpty() {
		return
	}
	typ := reflect.TypeOf(structPtr).Elem()
	fields := make([]string, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		fields[i] = typ.Field(i).Name
	}
	rSet.Load(fields...)
	fMap := rSet.env.cache.getRecord(rSet.ModelName(), rSet.ids[0])
	mapToStruct(rSet, structPtr, fMap)
}

// All Returns a copy of all records of the RecordCollection.
// It returns an empty slice if the RecordSet is empty.
func (rc RecordCollection) All(structSlicePtr interface{}) {
	rSet := rc.Fetch()
	if err := checkStructSlicePtr(structSlicePtr); err != nil {
		logging.LogAndPanic(log, "Invalid structPtr given", "error", err, "model", rSet.ModelName(), "received", structSlicePtr)
	}
	val := reflect.ValueOf(structSlicePtr)
	// sspType is []*struct
	sspType := val.Type().Elem()
	// structType is struct
	structType := sspType.Elem().Elem()
	val.Elem().Set(reflect.MakeSlice(sspType, rSet.Len(), rSet.Len()))
	recs := rSet.Records()
	for i := 0; i < rSet.Len(); i++ {
		fMap := rSet.env.cache.getRecord(rSet.ModelName(), recs[i].ids[0])
		newStructPtr := reflect.New(structType).Interface()
		mapToStruct(rSet, newStructPtr, fMap)
		val.Elem().Index(i).Set(reflect.ValueOf(newStructPtr))
	}
}

// Records returns the slice of RecordCollection singletons that constitute this
// RecordCollection.
func (rc RecordCollection) Records() []RecordCollection {
	rSet := rc.Fetch()
	res := make([]RecordCollection, len(rSet.Ids()))
	if rSet.IsEmpty() {
		return res
	}
	rSet.Load()
	for i, id := range rSet.Ids() {
		newRC := newRecordCollection(rSet.Env(), rSet.ModelName())
		res[i] = newRC.withIds([]int64{id})
	}
	return res
}

// EnsureOne panics if rc is not a singleton
func (rc RecordCollection) EnsureOne() {
	if rc.Len() != 1 {
		logging.LogAndPanic(log, "Expected singleton", "model", rc.ModelName(), "received", rc)
	}
}

// IsEmpty returns true if rc is an empty RecordCollection
func (rc RecordCollection) IsEmpty() bool {
	return rc.Len() == 0
}

// Len returns the number of records in this RecordCollection
func (rc RecordCollection) Len() int {
	rSet := rc.Fetch()
	return len(rSet.ids)
}

// withIdMap returns a new RecordCollection pointing to the given ids.
// It overrides the current query with ("ID", "in", ids).
func (rc RecordCollection) withIds(ids []int64) RecordCollection {
	rSet := rc
	rSet.ids = ids
	if len(ids) > 0 {
		rSet.query.cond = NewCondition().And("ID", "in", ids)
	}
	return rSet
}

var _ RecordSet = RecordCollection{}

// newRecordCollection returns a new empty RecordCollection in the
// given environment for the given modelName
func newRecordCollection(env Environment, modelName string) RecordCollection {
	mi, ok := modelRegistry.get(modelName)
	if !ok {
		logging.LogAndPanic(log, "Unknown model", "model", modelName)
	}
	rc := RecordCollection{
		mi:    mi,
		query: newQuery(),
		env:   &env,
		ids:   make([]int64, 0),
	}
	rc.query.recordSet = &rc
	return rc
}
