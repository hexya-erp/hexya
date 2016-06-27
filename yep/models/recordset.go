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
)

/*
recordStruct implements RecordSet
*/
type RecordSet struct {
	query     Query
	mi        *modelInfo
	env       *Environment
	ids       []int64
	callStack []*methodLayer
}

func (rs RecordSet) String() string {
	idsStr := make([]string, len(rs.ids))
	for i, id := range rs.ids {
		idsStr[i] = strconv.Itoa(int(id))
		i++
	}
	rsIds := strings.Join(idsStr, ",")
	return fmt.Sprintf("%s(%s)", rs.mi.name, rsIds)
}

/*
Env returns the RecordSet's Environment
*/
func (rs RecordSet) Env() *Environment {
	return rs.env
}

/*
ModelName returns the model name of the RecordSet
*/
func (rs RecordSet) ModelName() string {
	return rs.mi.name
}

/*
Ids return the ids of the RecordSet
*/
func (rs RecordSet) Ids() []int64 {
	return rs.ids
}

func (rs RecordSet) RelatedDepth(depth int) *RecordSet {
	rs.query.relDepth = depth
	return &rs
}

/*
Search query the database with the current filter and fills the RecordSet with the queries ids.
Does nothing in case RecordSet already has Ids. It panics in case of error.
It returns a pointer to the same RecordSet.
*/
func (rs *RecordSet) Search() *RecordSet {
	if len(rs.Ids()) == 0 {
		return rs.ForceSearch()
	}
	return rs
}

/*
Search query the database with the current filter and fills the RecordSet with the queries ids.
Overwrite RecordSet Ids if any. It panics in case of error.
It returns a pointer to the same RecordSet.
*/
func (rs *RecordSet) ForceSearch() *RecordSet {
	var idsMap []FieldMap
	num := rs.ReadValues(&idsMap, "id")
	ids := make([]int64, num)
	for i := 0; i < int(num); i++ {
		ids[i] = idsMap[i]["id"].(int64)
	}
	return rs.withIds(ids)
}

// update updates the database with the given data and returns the number of updated rows.
// It panics in case of error.
// This function is private and low level. It should not be called directly.
// Instead use rs.Write() or rs.Call("Write")
func (rs RecordSet) update(data interface{}) bool {
	//_, err := rs.qs.Update(data)
	//if err != nil {
	//	panic(fmt.Errorf("recordSet `%s` Write error: %s", rs, err))
	//}
	//rs.updateStoredFields(data)
	return true
}

// Write is a shortcut for rs.Call("Write") on the current RecordSet.
// Data can be either a struct pointer or a FieldMap.
func (rs RecordSet) Write(data interface{}) bool {
	return rs.Call("Write", data).(bool)
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

// Unlink is a shortcut for rs.Call("Unlink") on the current RecordSet.
func (rs RecordSet) Unlink() int64 {
	return rs.Call("Unlink").(int64)
}

/*
Filter returns a new RecordSet with the given additional filter condition.
*/
func (rs RecordSet) Filter(cond, op string, data ...interface{}) *RecordSet {
	rs.query.cond = rs.query.cond.And(cond, op, data...)
	return &rs
}

/*
Exclude returns a new RecordSet with the given additional NOT filter condition.
*/
func (rs RecordSet) Exclude(cond, op string, data ...interface{}) *RecordSet {
	rs.query.cond = rs.query.cond.AndNot(cond, op, data...)
	return &rs
}

/*
SetCond returns a new RecordSet with the given additional condition
*/
func (rs RecordSet) SetCond(cond *Condition) *RecordSet {
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

/*
SearchCount fetch from the database the number of records that match the RecordSet conditions
It panics in case of error
*/
func (rs RecordSet) SearchCount() int {
	sql, args := rs.query.countQuery()
	var res int
	DBGet(rs.env.cr, res, sql, args...)
	return res
}

// ReadAll query all data pointed by the RecordSet and map to containers.
// If cols are given, retrieve only the given fields.
// Returns the number of rows fetched.
// It panics in case of error
func (rs RecordSet) ReadAll(container interface{}, cols ...string) int64 {
	if err := checkStructSlicePtr(container); err != nil {
		panic(fmt.Errorf("recordSet `%s` ReadAll() error: %s", rs, err))
	}
	typ := reflect.TypeOf(container).Elem().Elem().Elem()
	structCtn := reflect.New(typ).Interface()
	sfMap := structToMap(structCtn, rs.query.relDepth)
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
	rs.EnsureOne()
	if err := checkStructPtr(container); err != nil {
		panic(fmt.Errorf("recordSet `%s` ReadOne() error: %s", rs, err))
	}
	sfMap := structToMap(container, rs.query.relDepth)
	fields := filterFields(rs.mi, sfMap.Keys(), cols)
	dbFields := filterOnDBFields(rs.mi, fields)
	var fMap FieldMap
	rs.ReadValue(&fMap, dbFields...)
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
	dbFields := filterOnDBFields(rs.mi, fields)
	sql, args := rs.query.selectQuery(dbFields)
	rows := DBQuery(rs.env.cr, sql, args...)
	defer rows.Close()
	for rows.Next() {
		line := make(FieldMap)
		err := rows.MapScan(line)
		if err != nil {
			panic(err)
		}
		*results = append(*results, line)
	}

	for i, rec := range rs.Records() {
		rec.computeFieldValues(&(*results)[i], fields...)
	}
	return int64(len(*results))
}

/*
Call calls the given method name methName with the given arguments and return the
result as interface{}.
*/
func (rs RecordSet) Call(methName string, args ...interface{}) interface{} {
	methInfo, ok := rs.mi.methods.get(methName)
	if !ok {
		panic(fmt.Errorf("Unknown method `%s` in model `%s`", methName, rs.ModelName()))
	}
	methLayer := methInfo.topLayer

	rs.callStack = append([]*methodLayer{methLayer}, rs.callStack...)
	return rs.call(methLayer, args...)
}

/*
call is a wrapper around reflect.Value.Call() to use with interface{} type.
*/
func (rs RecordSet) call(methLayer *methodLayer, args ...interface{}) interface{} {
	fnVal := methLayer.funcValue
	fnTyp := fnVal.Type()

	rsVal := reflect.ValueOf(rs)
	inVals := []reflect.Value{rsVal}
	methName := fmt.Sprintf("%s.%s()", methLayer.methInfo.mi.name, methLayer.methInfo.name)
	for i := 1; i < fnTyp.NumIn(); i++ {
		if i > len(args) {
			panic(fmt.Errorf("Not enough argument when Calling `%s`", methName))
		}
		inVals = append(inVals, reflect.ValueOf(args[i-1]))
	}
	retVal := fnVal.Call(inVals)
	if len(retVal) == 0 {
		return nil
	}
	return retVal[0].Interface()
}

/*
Super calls the next method Layer after the given funcPtr.
This method is meant to be used inside a method layer function to call its parent,
passing itself as funcPtr.
*/
func (rs RecordSet) Super(args ...interface{}) interface{} {
	if len(rs.callStack) == 0 {
		panic(fmt.Errorf("Internal error: empty call stack !"))
	}
	methLayer := rs.callStack[0]
	methInfo := methLayer.methInfo
	methLayer = methInfo.getNextLayer(methLayer)
	if methLayer == nil {
		// No parent
		return nil
	}

	rs.callStack[0] = methLayer
	return rs.call(methLayer, args...)
}

/*
MethodType returns the type of the method given by methName
*/
func (rs RecordSet) MethodType(methName string) reflect.Type {
	methInfo, ok := rs.mi.methods.get(methName)
	if !ok {
		panic(fmt.Errorf("Unknown method `%s` in model `%s`", methName, rs.ModelName()))
	}
	return methInfo.methodType
}

/*
Records returns the slice of RecordSet singletons that constitute this RecordSet
*/
func (rs RecordSet) Records() []*RecordSet {
	res := make([]*RecordSet, len(rs.Ids()))
	for i, id := range rs.Ids() {
		res[i] = rs.withIds([]int64{id})
	}
	return res
}

/*
EnsureOne panics if rs is not a singleton
*/
func (rs RecordSet) EnsureOne() {
	rs.Search()
	if len(rs.Ids()) != 1 {
		panic(fmt.Errorf("Expected singleton, got : %s", rs))
	}
}

// create inserts a new record in the database with the given data.
// data can be either a FieldMap or a struct pointer of the same model as rs.
// This function is private and low level. It should not be called directly.
// Instead use rs.Create(), rs.Call("Create") or env.Create()
func (rs RecordSet) create(data interface{}) *RecordSet {
	// create our FieldMap
	var fMap FieldMap
	if _, ok := data.(FieldMap); ok {
		fMap = data.(FieldMap)
	} else {
		if err := checkStructPtr(data); err != nil {
			panic(err)
		}
		fMap = structToMap(data, 0)
	}
	// clean our fMap from ID and non stored fields
	delete(fMap, "id")
	delete(fMap, "ID")
	for _, cf := range rs.mi.fields.getComputedFields() {
		delete(fMap, cf.name)
		delete(fMap, cf.json)
	}
	// insert in DB
	sql, args := rs.query.insertQuery(fMap)
	var createdId int64
	DBGet(rs.env.cr, &createdId, sql, args...)
	// compute stored fields
	//rs.updateStoredFields()
	if _, ok := data.(FieldMap); !ok {
		// set ID to the given struct
		idVal := reflect.ValueOf(data).Elem().FieldByName("ID")
		idVal.Set(reflect.ValueOf(createdId))
		// Update the given struct with its computed fields
		//rs.computeFields(data)
	}
	return rs.withIds([]int64{createdId})
}

// Create is a shortcut function for rs.Call("Create") on the current RecordSet.
// Data can be either a struct pointer or a FieldMap.
func (rs RecordSet) Create(data interface{}) *RecordSet {
	return rs.Call("Create", data).(*RecordSet)
}

/*
withIdMap returns a copy of rs filtered on the given ids slice (overwriting current query).
*/
func (rs RecordSet) withIds(ids []int64) *RecordSet {
	rs.ids = ids
	if len(ids) > 0 {
		rs.query = newQuery(&rs)
		rs = *rs.Filter("ID", "in", ids)
	}
	return &rs
}

// computeFieldValues updates the given params with the given computed (non stored) fields
// or all the computed fields of the model if not given.
// Returned fieldMap keys are field's JSON name
func (rs RecordSet) computeFieldValues(params *FieldMap, fields ...string) {
	for _, fInfo := range rs.mi.fields.getComputedFields(fields...) {
		if _, exists := (*params)[fInfo.name]; exists {
			// We already have the value we need in params
			// probably because it was computed with another field
			continue
		}
		newParams := rs.Call(fInfo.compute).(FieldMap)
		for k, v := range newParams {
			key, _ := rs.mi.fields.get(k)
			(*params)[key.json] = v
		}
	}
}

///*
//updateStoredFields updates all dependent fields of rs that are included in structPtrOrParams.
//*/
//func (rs RecordSet) updateStoredFields(structPtrOrParams interface{}) {
//	// First get list of fields that have been passed through structPtrOrParams
//	var fieldNames []string
//	if params, ok := structPtrOrParams.(FieldMap); ok {
//		cpsf, _ := fieldsCache.getComputedStoredFields(rs.ModelName())
//		fieldNames = make([]string, len(params)+len(cpsf))
//		i := 0
//		for k, _ := range params {
//			fieldNames[i] = k
//			i++
//		}
//		for _, v := range cpsf {
//			fieldNames[i] = v.name
//			i++
//		}
//	} else {
//		val := reflect.ValueOf(structPtrOrParams)
//		typ := reflect.Indirect(val).Type()
//		fieldNames = make([]string, typ.NumField())
//		for i := 0; i < typ.NumField(); i++ {
//			fieldNames[i] = typ.Field(i).Name
//		}
//	}
//	// Then get all fields to update
//	var toUpdate []computeData
//	for _, fieldName := range fieldNames {
//		refField := fieldRef{modelName: rs.ModelName(), name: fieldName}
//		targetFields, ok := fieldsCache.getDependentFields(refField)
//		if !ok {
//			continue
//		}
//		toUpdate = append(toUpdate, targetFields...)
//	}
//	// Compute all that must be computed and store the values
//	computed := make(map[string]bool)
//	rs = rs.Search()
//	for _, cData := range toUpdate {
//		methUID := fmt.Sprintf("%s.%s", cData.modelName, cData.compute)
//		if _, ok := computed[methUID]; ok {
//			continue
//		}
//		recs := NewRecordSet(rs.env, cData.modelName)
//		if cData.path != "" {
//			domainString := fmt.Sprintf("%s%s%s", cData.path, orm.ExprSep, "in")
//			recs.Filter(domainString, rs.Ids())
//		} else {
//			recs = rs
//		}
//		for _, rec := range recs.Records() {
//			vals := rec.Call(cData.compute)
//			if len(vals.(FieldMap)) > 0 {
//				rec.Write(vals.(FieldMap))
//			}
//		}
//	}
//}

// newRecordSet returns a new empty RecordSet in the given environment for the given modelName
func newRecordSet(env *Environment, modelName string) *RecordSet {
	mi, ok := modelRegistry.get(modelName)
	if !ok {
		panic(fmt.Errorf("Unknown model name `%s`", modelName))
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

///*
//newRecordStructFromData returns a recordStruct pointing to data.
//*/
//func newRecordStructFromData(env Environment, data interface{}) *RecordSet {
//	rs := newRecordStruct(env, data)
//	if err := checkStructPtr(data); err != nil {
//		panic(fmt.Errorf("newRecordStructFromData: %s", err))
//	}
//	val := reflect.ValueOf(data)
//	ind := reflect.Indirect(val)
//	id := ind.FieldByName("ID").Int()
//	return rs.withIds([]int64{id})
//}
