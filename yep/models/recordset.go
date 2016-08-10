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

	"github.com/npiganeau/yep/yep/tools"
)

// RecordSet is a generic struct representing several
// records of a model.
type RecordSet struct {
	query     Query
	relDepth  int
	mi        *modelInfo
	env       *Environment
	ids       []int64
	callStack []*methodLayer
}

// String returns the string representation of a RecordSet
func (rs RecordSet) String() string {
	idsStr := make([]string, len(rs.ids))
	for i, id := range rs.ids {
		idsStr[i] = strconv.Itoa(int(id))
		i++
	}
	rsIds := strings.Join(idsStr, ",")
	return fmt.Sprintf("%s(%s)", rs.mi.name, rsIds)
}

// Env returns the RecordSet's Environment
func (rs RecordSet) Env() *Environment {
	return rs.env
}

// ModelName returns the model name of the RecordSet
func (rs RecordSet) ModelName() string {
	return rs.mi.name
}

// Ids returns the ids of the RecordSet
func (rs RecordSet) Ids() []int64 {
	return rs.ids
}

// ID returns the ID of the unique record of this RecordSet
// It panics if rs is not a singleton.
func (rs RecordSet) ID() int64 {
	rs.EnsureOne()
	return rs.ids[0]
}

// RelatedDepth sets the depth at which to populate the structs with
// ReadOne and ReadAll.
func (rs RecordSet) RelatedDepth(depth int) *RecordSet {
	rs.relDepth = depth
	return &rs
}

// create inserts a new record in the database with the given data.
// data can be either a FieldMap or a struct pointer of the same model as rs.
// This function is private and low level. It should not be called directly.
// Instead use rs.Create(), rs.Call("Create") or env.Create()
func (rs RecordSet) create(data interface{}) *RecordSet {
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
	rs.updateStoredFields(fMap)
	if reflect.TypeOf(data).Kind() == reflect.Ptr {
		// set ID to the given struct
		idVal := reflect.ValueOf(data).Elem().FieldByName("ID")
		idVal.Set(reflect.ValueOf(createdId))
		// Update the given struct with its computed fields
		// FIXME: Add computed non stored field calculation here
		//rs.computeFields(data)
	}
	return rs.withIds([]int64{createdId})
}

// update updates the database with the given data and returns the number of updated rows.
// It panics in case of error.
// This function is private and low level. It should not be called directly.
// Instead use rs.Write() or rs.Call("Write")
func (rs RecordSet) update(data interface{}) bool {
	fMap := convertInterfaceToFieldMap(data)
	rs.mi.convertValuesToFieldType(&fMap)
	// clean our fMap from ID and non stored fields
	delete(fMap, "id")
	delete(fMap, "ID")
	for fName := range fMap {
		if fi := rs.mi.getRelatedFieldInfo(fName); !fi.isStored() {
			delete(fMap, fi.name)
			delete(fMap, fi.json)
		}
	}
	// update DB
	sql, args := rs.query.updateQuery(fMap)
	DBExecute(rs.env.cr, sql, args...)
	// compute stored fields
	rs.updateStoredFields(fMap)
	return true
}

// delete deletes the database record of this RecordSet and returns the number of deleted rows.
// This function is private and low level. It should not be called directly.
// Instead use rs.Unlink() or rs.Call("Unlink")
func (rs RecordSet) delete() int64 {
	sql, args := rs.query.deleteQuery()
	res := DBExecute(rs.env.cr, sql, args...)
	num, _ := res.RowsAffected()
	return num
}

/*
Filter returns a new RecordSet with the given additional filter condition.
*/
func (rs RecordSet) Filter(cond, op string, data interface{}) *RecordSet {
	rs.query.cond = rs.query.cond.And(cond, op, data)
	return &rs
}

/*
Exclude returns a new RecordSet with the given additional NOT filter condition.
*/
func (rs RecordSet) Exclude(cond, op string, data interface{}) *RecordSet {
	rs.query.cond = rs.query.cond.AndNot(cond, op, data)
	return &rs
}

/*
SetCond returns a new RecordSet with the given additional condition
*/
func (rs RecordSet) Condition(cond *Condition) *RecordSet {
	rs.query.cond = rs.query.cond.AndCond(cond)
	return &rs
}

/*
Limit returns a new RecordSet with the given limit as additional condition
*/
func (rs RecordSet) Limit(limit int, args ...int) *RecordSet {
	rs.query.limit = limit
	if len(args) > 0 {
		rs.query.offset = args[0]
	}
	return &rs
}

/*
Offset returns a new RecordSet with the given offset as additional condition
*/
func (rs RecordSet) Offset(offset int) *RecordSet {
	rs.query.offset = offset
	return &rs
}

/*
OrderBy returns a new RecordSet with the given ORDER BY clause in its Query
*/
func (rs RecordSet) OrderBy(exprs ...string) *RecordSet {
	rs.query.orders = append(rs.query.orders, exprs...)
	return &rs
}

/*
GroupBy returns a new RecordSet with the given GROUP BY clause in its Query
*/
func (rs RecordSet) GroupBy(exprs ...string) *RecordSet {
	rs.query.groups = append(rs.query.groups, exprs...)
	return &rs
}

// Distinct returns a new RecordSet with its Query filtering duplicates
func (rs RecordSet) Distinct() *RecordSet {
	rs.query.distinct = true
	return &rs
}

// Search query the database with the current filter and fills the RecordSet with the queries ids.
// Does nothing in case RecordSet already has Ids. It panics in case of error.
// It returns a pointer to the same RecordSet.
func (rs *RecordSet) Search() *RecordSet {
	if len(rs.Ids()) == 0 {
		return rs.ForceSearch()
	}
	return rs
}

// ForceSearch query the database with the current filter and fills the RecordSet with the queries ids.
// Overwrite RecordSet Ids if any. It panics in case of error.
// It returns a pointer to the same RecordSet.
func (rs *RecordSet) ForceSearch() *RecordSet {
	var idsMap []FieldMap
	num := rs.ReadValues(&idsMap, "id")
	ids := make([]int64, num)
	for i := 0; i < int(num); i++ {
		ids[i] = idsMap[i]["id"].(int64)
	}
	return rs.withIds(ids)
}

/*
SearchCount fetch from the database the number of records that match the RecordSet conditions
It panics in case of error
*/
func (rs RecordSet) SearchCount() int {
	sql, args := rs.query.countQuery()
	var res int
	DBGet(rs.env.cr, &res, sql, args...)
	return res
}

// ReadAll query all data pointed by the RecordSet and map to containers.
// If cols are given, retrieve only the given fields.
// Returns the number of rows fetched.
// It panics in case of error
func (rs RecordSet) ReadAll(container interface{}, cols ...string) int64 {
	rs = *rs.Search()
	if err := checkStructSlicePtr(container); err != nil {
		tools.LogAndPanic(log, err.Error(), "container", container)
	}
	typ := reflect.TypeOf(container).Elem().Elem().Elem()
	structCtn := reflect.New(typ).Interface()
	sfMap := structToMap(structCtn, rs.relDepth)
	fields := filterFields(rs.mi, sfMap.Keys(), cols)

	var fMaps []FieldMap
	rs.ReadValues(&fMaps, fields...)

	ptrVal := reflect.ValueOf(container)
	sliceVal := ptrVal.Elem()
	sliceVal.Set(reflect.MakeSlice(ptrVal.Type().Elem(), len(fMaps), len(fMaps)))

	for i, fMap := range fMaps {
		structPtr := reflect.New(typ).Interface()
		mapToStruct(rs.mi, structPtr, fMap)
		sliceVal.Index(i).Set(reflect.ValueOf(structPtr))
	}
	return int64(len(fMaps))
}

// ReadOne query the RecordSet row and map to container.
// If cols are given, retrieve only the given fields.
// It panics if the RecordSet does not contain exactly one row.
func (rs RecordSet) ReadOne(container interface{}, cols ...string) {
	rs = *rs.Search()
	rs.EnsureOne()
	if err := checkStructPtr(container); err != nil {
		tools.LogAndPanic(log, err.Error(), "container", container)
	}

	sfMap := structToMap(container, rs.relDepth)
	fields := filterFields(rs.mi, sfMap.Keys(), cols)
	var fMap FieldMap
	rs.ReadValue(&fMap, fields...)
	mapToStruct(rs.mi, container, fMap)
}

// Value query a single line of data in the database and maps the
// result to the result FieldMap
func (rs RecordSet) ReadValue(result *FieldMap, fields ...string) {
	rs.EnsureOne()
	var fieldsMap []FieldMap
	rs.ReadValues(&fieldsMap, fields...)
	*result = make(FieldMap)
	*result = fieldsMap[0]
}

// Values query all data of the RecordSet and map to []FieldMap.
// fields are the fields to retrieve in the expression format,
// i.e. "User.Profile.Age" or "user_id.profile_id.age".
// If no fields are given, all columns of the RecordSet's model are retrieved.
func (rs RecordSet) ReadValues(results *[]FieldMap, fields ...string) int64 {
	if len(fields) == 0 {
		fields = rs.mi.fields.nonRelatedFieldJSONNames()
	}
	subFields, substs := rs.substituteRelatedFields(fields)
	dbFields := filterOnDBFields(rs.mi, subFields)
	sql, args := rs.query.selectQuery(dbFields)
	rows := DBQuery(rs.env.cr, sql, args...)
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		line := make(FieldMap)
		err := rs.mi.scanToFieldMap(rows, &line)
		line.SubstituteKeys(substs)
		if err != nil {
			tools.LogAndPanic(log, err.Error(), "model", rs.ModelName(), "fields", fields)
		}
		*results = append(*results, line)
		ids = append(ids, line["id"].(int64))
	}

	// Call withIds directly and not ForceSearch to avoid infinite recursion
	rs = *rs.withIds(ids)
	for i, rec := range rs.Records() {
		rec.computeFieldValues(&(*results)[i], fields...)
	}
	return int64(len(*results))
}

// Records returns the slice of RecordSet singletons that constitute this RecordSet
func (rs RecordSet) Records() []*RecordSet {
	res := make([]*RecordSet, len(rs.Ids()))
	for i, id := range rs.Ids() {
		res[i] = rs.withIds([]int64{id})
	}
	return res
}

// EnsureOne panics if rs is not a singleton
func (rs *RecordSet) EnsureOne() {
	rs.Search()
	if len(rs.Ids()) != 1 {
		tools.LogAndPanic(log, "Expected singleton", "model", rs.ModelName(), "received", rs)
	}
}

// withIdMap sets the given RecordSet ids to the given ids slice (overwriting current query).
// This method both replaces in place and returns a pointer to the same RecordSet.
func (rs *RecordSet) withIds(ids []int64) *RecordSet {
	rs.ids = ids
	if len(ids) > 0 {
		rs.query.cond = NewCondition()
		rs = rs.Filter("ID", "in", ids)
	}
	return rs
}

// newRecordSet returns a new empty RecordSet in the given environment for the given modelName
func newRecordSet(env *Environment, modelName string) *RecordSet {
	mi, ok := modelRegistry.get(modelName)
	if !ok {
		tools.LogAndPanic(log, "Unknown model", "model", modelName)
	}
	rs := RecordSet{
		mi:    mi,
		query: newQuery(),
		env:   env,
		ids:   make([]int64, 0),
	}
	rs.query.recordSet = &rs
	return &rs
}
