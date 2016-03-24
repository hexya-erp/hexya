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

type IdStruct struct {
	ID int64
}

/*
RecordSet is both a set of database records and an entrypoint to the models API for CRUD operations.
RecordSet are immutable.
*/
type RecordSet interface {
	// returns the Environment of the RecordSet
	Env() Environment
	// returns the model name of the RecordSet
	ModelName() string
	// returns the ids of this RecordSet
	Ids() []int64
	// creates a record in database from the given data and returns the corresponding recordset.
	// data can be either a ptrStruct, a slice of ptrStruct or an orm.Params map.
	Create(interface{}) RecordSet
	// query the database with the current filter and returns a new recordset with the queries ids
	// Does nothing if RecordSet already has Ids in cache.
	Search() RecordSet
	// query the database with the current filter and returns a new recordset with the queries ids
	// Overwrite existing Ids if any
	ForceSearch() RecordSet
	// updates the database with the given data and returns the number of updated rows.
	// data can be either
	// - a ptrStruct for a single update. In this case, the RecordSet is discarded and the pk of
	// the ptrStruct is used to determine the record to update.
	// - an orm.Params map for multi update. In this case, the records of this RecordSet are updated.
	Write(interface{}) int64
	// deletes the database record of this RecordSet and returns the number of deleted rows.
	Unlink() int64
	// returns a new RecordSet with the given additional filter condition
	Filter(string, ...interface{}) RecordSet
	// returns a new RecordSet with the given additional NOT condition
	Exclude(string, ...interface{}) RecordSet
	// returns a new RecordSet with the given additional condition
	SetCond(*orm.Condition) RecordSet
	// returns a new RecordSet with the given limit as additional condition
	Limit(limit interface{}, args ...interface{}) RecordSet
	// returns a new RecordSet with the given offset as additional condition
	Offset(offset interface{}) RecordSet
	// returns a new RecordSet with the given order as additional condition
	OrderBy(exprs ...string) RecordSet
	// returns a new RecordSet that includes related models (table join) in its search
	RelatedSel(params ...interface{}) RecordSet
	// fetch from the database the number of records that match the RecordSet conditions
	SearchCount() int64
	// query all data pointed by the RecordSet and map to containers.
	ReadAll(container interface{}, cols ...string) int64
	// query the RecordSet row and map to containers.
	// returns error if the RecordSet does not contain exactly one row.
	ReadOne(container interface{}, cols ...string)
	// query all data of the RecordSet and map to []map[string]interface.
	// expres means condition expression.
	// it converts data to []map[column]value.
	Values(results *[]orm.Params, exprs ...string) int64
	// query all data of the RecordSet and map to [][]interface
	// it converts data to [][column_index]value
	ValuesList(results *[]orm.ParamsList, exprs ...string) int64
	// query all data and map to []interface.
	// it's designed for one column record set, auto change to []value, not [][column]value.
	ValuesFlat(result *orm.ParamsList, expr string) int64
	// Call the given method by name with the given arguments
	Call(methName string, args ...interface{}) interface{}
	// Super is called from inside a method to call its parent, passing itself as fnctPtr
	Super(args ...interface{}) interface{}
	// Returns a slice of RecordSet singleton that constitute this RecordSet
	Records() []RecordSet
}

/*
recordStruct implements RecordSet
*/
type recordStruct struct {
	qs        orm.QuerySeter
	env       Environment
	idMap     map[int64]bool
	callStack []*methodLayer
}

func (rs recordStruct) String() string {
	idsStr := make([]string, len(rs.idMap))
	i := 0
	for id, _ := range rs.idMap {
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
	return ids(rs.idMap)
}

/*
Create creates a new record in database from the given data and returns the corresponding RecordSet
Data can be either a struct pointer or an orm.Params map.
*/
func (rs recordStruct) Create(data interface{}) RecordSet {
	val := reflect.ValueOf(data)
	ind := reflect.Indirect(val)
	if ind.Kind() != reflect.Struct {
		panic(orm.ErrNotImplement)
	}
	if getModelName(ind.Type()) != rs.ModelName() {
		panic(fmt.Errorf("Data type mismatch: received `%s` object to create `%s` record set",
			getModelName(ind.Type()), rs))
	}
	id, err := rs.env.Cr().Insert(data)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` Create error: %s", rs, err))
	}
	return copyRecordStruct(rs).withIdMap(map[int64]bool{id: true})
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
	var Ids orm.ParamsList
	num := rs.ValuesFlat(&Ids, "ID")
	idMap := make(map[int64]bool, num)
	for _, id := range Ids {
		idMap[id.(int64)] = true
	}
	return copyRecordStruct(rs).withIdMap(idMap)
}

/*
Write updates the database with the given data and returns the number of updated rows.
data can be either a ptrStruct (single update) or an orm.Params map (multi-update).
It panics in case of error.
*/
func (rs recordStruct) Write(data interface{}) int64 {
	val := reflect.ValueOf(data)
	ind := reflect.Indirect(val)
	indType := ind.Type()
	var num int64
	var err error
	if ind.Kind() == reflect.Struct {
		if getModelName(indType) != rs.ModelName() {
			panic(fmt.Errorf("Data type mismatch: received `%s` object(s) to write `%s` record set",
				getModelName(indType), rs))
		}
		params := structToMap(data)
		delete(params, "ID")
		num, err = rs.qs.Update(orm.Params(params))
	} else if indType == reflect.TypeOf(orm.Params{}) {
		num, err = rs.qs.Update(data.(orm.Params))
	} else {
		panic(fmt.Errorf("Wrong data type for writing `%s`", data))
	}
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` Write error: %s", rs, err))
	}
	rs.updateRelatedFields(data)
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
	err := rs.qs.One(container, cols...)
	if err != nil {
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
		res[i] = rs.withIdMap(map[int64]bool{id: true})
	}
	return res
}

var _ RecordSet = recordStruct{}

/*
withIdMap returns a copy of rs filtered on the given idMap (overwriting current queryset).
*/
func (rs recordStruct) withIdMap(idMap map[int64]bool) recordStruct {
	newRs := copyRecordStruct(rs)
	newRs.idMap = idMap
	newRs.qs = rs.env.Cr().QueryTable(rs.ModelName())
	if len(idMap) > 0 {
		domStr := fmt.Sprintf("id%sin", orm.ExprSep)
		newRs.qs = newRs.qs.Filter(domStr, ids(idMap))
	}
	return newRs
}

/*
computeFields sets the value of the computed fields of structPtr.
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
updateRelatedFields updates all dependent fields of rs that are included in structPtrOrParams.
*/
func (rs recordStruct) updateRelatedFields(structPtrOrParams interface{}) {
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
				rec.Write(vals)
			}
		}
	}
}

/*
newRecordStruct returns a new empty recordStruct.
*/
func newRecordStruct(env Environment, ptrStructOrTableName interface{}) recordStruct {
	qs := env.Cr().QueryTable(ptrStructOrTableName)
	rs := recordStruct{
		qs:    qs,
		env:   NewEnvironment(env.Cr(), env.Uid(), env.Context()),
		idMap: make(map[int64]bool),
	}
	return rs
}

func copyRecordStruct(rs recordStruct) recordStruct {
	newRs := newRecordStruct(rs.env, rs.ModelName())
	newRs.qs = rs.qs
	for k, v := range rs.idMap {
		newRs.idMap[k] = v
	}
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

/*
getName returns Model name from reflectType (splitting on _)
*/
func getModelName(typ reflect.Type) string {
	name := strings.SplitN(typ.Name(), "_", 2)[0]
	return name
}

/*
ids returns the ids of the given idMap
*/
func ids(idMap map[int64]bool) []int64 {
	keys := make([]int64, len(idMap))
	i := 0
	for k := range idMap {
		keys[i] = k
		i++
	}
	return keys
}
