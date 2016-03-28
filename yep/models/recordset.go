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
	"github.com/npiganeau/yep/yep/orm"
	"reflect"
	"strconv"
	"strings"
)

/*
recordStruct implements RecordSet
*/
type recordStruct struct {
	qs        orm.QuerySeter
	env       Environment
	ids       []int64
	callStack []*methodLayer
}

func (rs recordStruct) String() string {
	idsStr := make([]string, len(rs.ids))
	for i, id := range rs.ids {
		idsStr[i] = strconv.Itoa(int(id))
		i++
	}
	rsIds := strings.Join(idsStr, ",")
	return fmt.Sprintf("%s(%s)", rs.ModelName(), rsIds)
}

/*
Env returns the RecordSet's Environment
*/
func (rs recordStruct) Env() Environment {
	return rs.env
}

/*
ModelName returns the model name of the RecordSet
*/
func (rs recordStruct) ModelName() string {
	return rs.qs.ModelName()
}

/*
Ids return the ids of the RecordSet
*/
func (rs recordStruct) Ids() []int64 {
	return rs.ids
}

/*
Search query the database with the current filter and returns a new recordset with the queries ids.
Does nothing in case RecordSet already has Ids. It panics in case of error
*/
func (rs recordStruct) Search() RecordSet {
	if len(rs.Ids()) == 0 {
		return rs.ForceSearch()
	}
	return copyRecordStruct(rs)
}

/*
Search query the database with the current filter and returns a new recordset with the queries ids.
Overwrite RecordSet Ids if any. It panics in case of error
*/
func (rs recordStruct) ForceSearch() RecordSet {
	var idParams orm.ParamsList
	num := rs.ValuesFlat(&idParams, "ID")
	ids := make([]int64, num)
	for i := 0; i < int(num); i++ {
		ids[i] = idParams[i].(int64)
	}
	return copyRecordStruct(rs).withIds(ids)
}

/*
Write updates the database with the given data and returns the number of updated rows.
It panics in case of error.
*/
func (rs recordStruct) Write(data orm.Params) int64 {
	num, err := rs.qs.Update(data)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` Write error: %s", rs, err))
	}
	rs.updateStoredFields(data)
	return num
}

/*
Unlink deletes the database record of this RecordSet and returns the number of deleted rows.
*/
func (rs recordStruct) Unlink() int64 {
	num, err := rs.qs.Delete()
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` Unlink error: %s", rs, err))
	}
	return num
}

/*
Filter returns a new RecordSet with the given additional filter condition.
*/
func (rs recordStruct) Filter(cond string, data ...interface{}) RecordSet {
	newRs := copyRecordStruct(rs)
	newRs.qs = newRs.qs.Filter(cond, data...)
	return newRs
}

/*
Exclude returns a new RecordSet with the given additional NOT filter condition.
*/
func (rs recordStruct) Exclude(cond string, data ...interface{}) RecordSet {
	newRs := copyRecordStruct(rs)
	newRs.qs = newRs.qs.Exclude(cond, data...)
	return newRs
}

/*
SetCond returns a new RecordSet with the given additional condition
*/
func (rs recordStruct) SetCond(cond *orm.Condition) RecordSet {
	newRs := copyRecordStruct(rs)
	newRs.qs = newRs.qs.SetCond(cond)
	return newRs
}

/*
Limit returns a new RecordSet with the given limit as additional condition
*/
func (rs recordStruct) Limit(limit interface{}, args ...interface{}) RecordSet {
	newRs := copyRecordStruct(rs)
	newRs.qs = newRs.qs.Limit(limit, args...)
	return newRs
}

/*
Offset returns a new RecordSet with the given offset as additional condition
*/
func (rs recordStruct) Offset(offset interface{}) RecordSet {
	newRs := copyRecordStruct(rs)
	newRs.qs = newRs.qs.Offset(offset)
	return newRs
}

/*
OrderBy returns a new RecordSet with the given order as additional condition
*/
func (rs recordStruct) OrderBy(exprs ...string) RecordSet {
	newRs := copyRecordStruct(rs)
	newRs.qs = newRs.qs.OrderBy(exprs...)
	return newRs
}

/*
RelatedSel returns a new RecordSet that includes related models (table join) in its search
*/
func (rs recordStruct) RelatedSel(params ...interface{}) RecordSet {
	newRs := copyRecordStruct(rs)
	newRs.qs = newRs.qs.RelatedSel(params...)
	return newRs
}

/*
SearchCount fetch from the database the number of records that match the RecordSet conditions
It panics in case of error
*/
func (rs recordStruct) SearchCount() int64 {
	num, err := rs.qs.Count()
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` SearchCount() error: %s", rs, err))
	}
	return num
}

/*
All query all data pointed by the RecordSet and map to containers.
It panics in case of error
*/
func (rs recordStruct) ReadAll(container interface{}, cols ...string) int64 {
	if err := checkStructPtr(container, true); err != nil {
		panic(fmt.Errorf("recordSet `%s` ReadAll() error: %s", rs, err))
	}
	num, err := rs.qs.OrderBy("ID").All(container, cols...)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` ReadAll() error: %s", rs, err))
	}
	val := reflect.ValueOf(container)
	ind := reflect.Indirect(val)
	if ind.Kind() == reflect.Slice {
		contSlice := make([]interface{}, ind.Len())
		for i := 0; i < ind.Len(); i++ {
			csIndex := reflect.ValueOf(contSlice).Index(i)
			csIndex.Set(ind.Index(i))
		}
		rs = rs.Search().(recordStruct)
		for i, item := range rs.Records() {
			item.(recordStruct).computeFields(contSlice[i])
		}
		return num
	}
	rs.computeFields(container)
	return 1
}

/*
One query the RecordSet row and map to containers.
it panics if the RecordSet does not contain exactly one row.
*/
func (rs recordStruct) ReadOne(container interface{}, cols ...string) {
	if err := checkStructPtr(container); err != nil {
		panic(fmt.Errorf("recordSet `%s` ReadOne() error: %s", rs, err))
	}
	if err := rs.qs.One(container, cols...); err != nil {
		panic(fmt.Errorf("recordSet `%s` ReadOne() error: %s", rs, err))
	}
	rs.computeFields(container)
}

/*
Values query all data of the RecordSet and map to []map[string]interface.
exprs means condition expression.
it converts data to []map[column]value.
*/
func (rs recordStruct) Values(results *[]orm.Params, exprs ...string) int64 {
	num, err := rs.qs.Values(results, exprs...)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` Values() error: %s", rs, err))
	}
	return num

}

/*
ValuesList query all data of the RecordSet and map to [][]interface
it converts data to [][column_index]value
*/
func (rs recordStruct) ValuesList(results *[]orm.ParamsList, exprs ...string) int64 {
	num, err := rs.qs.ValuesList(results, exprs...)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` ValuesList() error: %s", rs, err))
	}
	return num
}

/*
ValuesFlat query all data and map to []interface.
it's designed for one column record set, auto change to []value, not [][column]value.
*/
func (rs recordStruct) ValuesFlat(result *orm.ParamsList, expr string) int64 {
	num, err := rs.qs.ValuesFlat(result, expr)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` ValuesFlat() error: %s", rs, err))
	}
	return num
}

/*
Call calls the given method name methName with the given arguments and return the
result as interface{}.
*/
func (rs recordStruct) Call(methName string, args ...interface{}) interface{} {
	methInfo, ok := methodsCache.get(method{modelName: rs.ModelName(), name: methName})
	if !ok {
		panic(fmt.Errorf("Unknown method `%s` in model `%s`", methName, rs.ModelName()))
	}
	methLayer := methInfo.topLayer

	rsCopy := copyRecordStruct(rs)
	rsCopy.callStack = append([]*methodLayer{methLayer}, rsCopy.callStack...)
	return rsCopy.call(methLayer, args...)
}

/*
call is a wrapper around reflect.Value.Call() to use with interface{} type.
*/
func (rs recordStruct) call(methLayer *methodLayer, args ...interface{}) interface{} {
	fnVal := methLayer.funcValue
	fnTyp := fnVal.Type()

	rsVal := reflect.ValueOf(rs)
	inVals := []reflect.Value{rsVal}
	for i := 1; i < fnTyp.NumIn(); i++ {
		if i > len(args) {
			panic(fmt.Errorf("Not enough argument when Calling `%s`", fnVal))
		}
		argTyp := reflect.TypeOf(args[i-1])
		if argTyp != fnTyp.In(i) {
			panic(fmt.Errorf("Wrong argument type for argument %d: %s instead of %s", i, argTyp.Name(), fnTyp.Name()))
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
func (rs recordStruct) Super(args ...interface{}) interface{} {
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

	rsCopy := copyRecordStruct(rs)
	rsCopy.callStack[0] = methLayer
	return rsCopy.call(methLayer, args...)
}

/*
Records returns the slice of RecordSet singletons that constitute this RecordSet
*/
func (rs recordStruct) Records() []RecordSet {
	rs = rs.Search().(recordStruct)
	res := make([]RecordSet, len(rs.Ids()))
	for i, id := range rs.Ids() {
		res[i] = rs.withIds([]int64{id})
	}
	return res
}

var _ RecordSet = recordStruct{}

/*
withIdMap returns a copy of rs filtered on the given ids slice (overwriting current queryset).
*/
func (rs recordStruct) withIds(ids []int64) recordStruct {
	newRs := copyRecordStruct(rs)
	newRs.ids = ids
	newRs.qs = rs.env.Cr().QueryTable(rs.ModelName())
	if len(ids) > 0 {
		domStr := fmt.Sprintf("id%sin", orm.ExprSep)
		newRs.qs = newRs.qs.Filter(domStr, ids)
	}
	return newRs
}

/*
computeFields sets the value of the computed (non stored) fields of structPtr.
*/
func (rs recordStruct) computeFields(structPtr interface{}) {
	val := reflect.ValueOf(structPtr)
	ind := reflect.Indirect(val)

	fInfos, _ := fieldsCache.getComputedFields(rs.ModelName())
	params := make(orm.Params)
	for _, fInfo := range fInfos {
		sf := ind.FieldByName(fInfo.name)
		if !sf.IsValid() {
			// Computed field is not present in structPtr
			continue
		}
		if _, exists := params[fInfo.name]; exists {
			// We already have the value we need in params
			continue
		}
		newParams := rs.Call(fInfo.compute).(orm.Params)
		for k, v := range newParams {
			params[k] = v
		}
		structField := ind.FieldByName(fInfo.name)
		structField.Set(reflect.ValueOf(params[fInfo.name]))
	}
}

/*
updateStoredFields updates all dependent fields of rs that are included in structPtrOrParams.
*/
func (rs recordStruct) updateStoredFields(structPtrOrParams interface{}) {
	// First get list of fields that have been passed through structPtrOrParams
	var fieldNames []string
	if params, ok := structPtrOrParams.(orm.Params); ok {
		fieldNames = make([]string, len(params))
		i := 0
		for k, _ := range params {
			fieldNames[i] = k
			i++
		}
	} else {
		val := reflect.ValueOf(structPtrOrParams)
		typ := reflect.Indirect(val).Type()
		fieldNames = make([]string, typ.NumField())
		for i := 0; i < typ.NumField(); i++ {
			fieldNames[i] = typ.Field(i).Name
		}
	}
	// Then get all fields to update
	var toUpdate []computeData
	for _, fieldName := range fieldNames {
		refField := field{modelName: rs.ModelName(), name: fieldName}
		targetFields, ok := fieldsCache.getDependentFields(refField)
		if !ok {
			continue
		}
		toUpdate = append(toUpdate, targetFields...)
	}
	// Compute all that must be computed and store the values
	computed := make(map[string]bool)
	rs = rs.Search().(recordStruct)
	for _, cData := range toUpdate {
		methUID := fmt.Sprintf("%s.%s", cData.modelName, cData.compute)
		if _, ok := computed[methUID]; ok {
			continue
		}
		recs := NewRecordSet(rs.env, cData.modelName)
		if cData.path != "" {
			domainString := fmt.Sprintf("%s%s%s", cData.path, orm.ExprSep, "in")
			recs.Filter(domainString, rs.Ids())
		} else {
			recs = rs
		}
		for _, rec := range recs.Records() {
			vals := rec.Call(cData.compute)
			if len(vals.(orm.Params)) > 0 {
				rec.Write(vals.(orm.Params))
			}
		}
	}
}

/*
newRecordStruct returns a new empty recordStruct.
*/
func newRecordStruct(env Environment, ptrStructOrTableName interface{}) recordStruct {
	modelName := getModelName(ptrStructOrTableName)
	qs := env.Cr().QueryTable(modelName)
	rs := recordStruct{
		qs:  qs,
		env: NewEnvironment(env.Cr(), env.Uid(), env.Context()),
		ids: make([]int64, 0),
	}
	return rs
}

/*
newRecordStructFromData returns a recordStruct pointing to data.
*/
func newRecordStructFromData(env Environment, data interface{}) recordStruct {
	rs := newRecordStruct(env, data)
	if err := checkStructPtr(data); err != nil {
		panic(fmt.Errorf("newRecordStructFromData: %s", err))
	}
	val := reflect.ValueOf(data)
	ind := reflect.Indirect(val)
	id := ind.FieldByName("ID").Int()
	return rs.withIds([]int64{id})
}

func copyRecordStruct(rs recordStruct) recordStruct {
	newRs := newRecordStruct(rs.env, rs.ModelName())
	newRs.qs = rs.qs
	newRs.ids = make([]int64, len(rs.ids))
	copy(newRs.ids, rs.ids)
	newRs.callStack = make([]*methodLayer, len(rs.callStack))
	copy(newRs.callStack, rs.callStack)
	return newRs
}

/*
NewRecordSet returns a new empty Recordset on the model given by ptrStructOrTableName and the
given Environment.
*/
func NewRecordSet(env Environment, ptrStructOrTableName interface{}) RecordSet {
	return newRecordStruct(env, ptrStructOrTableName)
}
