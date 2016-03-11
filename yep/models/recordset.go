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
	Search() RecordSet
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
}

/*
recordStruct implements RecordSet
*/
type recordStruct struct {
	qs    orm.QuerySeter
	env   Environment
	idMap map[int64]bool
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
	newRs := newRecordStruct(rs.env, rs.ModelName(), map[int64]bool{id: true})
	return newRs
}

/*
Search query the database with the current filter and returns a new recordset with the queries ids.
It panics in case of error
*/
func (rs recordStruct) Search() RecordSet {
	var recIds []*IdStruct
	num, err := rs.qs.All(&recIds)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` Search error: %s", rs, err))
	}
	idMap := make(map[int64]bool, num)
	for _, idStruct := range recIds {
		idMap[idStruct.ID] = true
	}
	return newRecordStruct(rs.env, rs.ModelName(), idMap)
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
		num, err = rs.env.Cr().Update(data)
	} else if indType == reflect.TypeOf(orm.Params{}) {
		num, err = rs.qs.Update(data.(orm.Params))
	} else {
		panic(fmt.Errorf("Wrong data type for writing `%s`", data))
	}
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` Write error: %s", rs, err))
	}
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
	newRs := newRecordStruct(rs.env, rs.ModelName(), rs.idMap)
	newRs.qs = newRs.qs.Filter(cond, data...)
	return newRs
}

/*
Exclude returns a new RecordSet with the given additional NOT filter condition.
*/
func (rs recordStruct) Exclude(cond string, data ...interface{}) RecordSet {
	newRs := newRecordStruct(rs.env, rs.ModelName(), rs.idMap)
	newRs.qs = newRs.qs.Exclude(cond, data...)
	return newRs
}

/*
SetCond returns a new RecordSet with the given additional condition
*/
func (rs recordStruct) SetCond(cond *orm.Condition) RecordSet {
	newRs := newRecordStruct(rs.env, rs.ModelName(), rs.idMap)
	newRs.qs = newRs.qs.SetCond(cond)
	return newRs
}

/*
Limit returns a new RecordSet with the given limit as additional condition
*/
func (rs recordStruct) Limit(limit interface{}, args ...interface{}) RecordSet {
	newRs := newRecordStruct(rs.env, rs.ModelName(), rs.idMap)
	newRs.qs = newRs.qs.Limit(limit, args...)
	return newRs
}

/*
Offset returns a new RecordSet with the given offset as additional condition
*/
func (rs recordStruct) Offset(offset interface{}) RecordSet {
	newRs := newRecordStruct(rs.env, rs.ModelName(), rs.idMap)
	newRs.qs = newRs.qs.Offset(offset)
	return newRs
}

/*
OrderBy returns a new RecordSet with the given order as additional condition
*/
func (rs recordStruct) OrderBy(exprs ...string) RecordSet {
	newRs := newRecordStruct(rs.env, rs.ModelName(), rs.idMap)
	newRs.qs = newRs.qs.OrderBy(exprs...)
	return newRs
}

/*
RelatedSel returns a new RecordSet that includes related models (table join) in its search
*/
func (rs recordStruct) RelatedSel(params ...interface{}) RecordSet {
	newRs := newRecordStruct(rs.env, rs.ModelName(), rs.idMap)
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
	num, err := rs.qs.All(container, cols...)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` All() error: %s", rs, err))
	}
	return num
}

/*
One query the RecordSet row and map to containers.
it panics if the RecordSet does not contain exactly one row.
*/
func (rs recordStruct) ReadOne(container interface{}, cols ...string) {
	err := rs.qs.One(container, cols...)
	if err != nil {
		panic(fmt.Errorf("recordSet `%s` One() error: %s", rs, err))
	}
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

var _ RecordSet = recordStruct{}

/*
newRecordStruct returns a new recordStruct with the given parameters
*/
func newRecordStruct(env Environment, ptrStructOrTableName interface{}, idMap map[int64]bool) recordStruct {
	qs := env.Cr().QueryTable(ptrStructOrTableName)
	if len(idMap) > 0 {
		qs = qs.Filter("id__in", ids(idMap))
	}
	rs := recordStruct{
		qs:    qs,
		env:   NewEnvironment(env.Cr(), env.Uid(), env.Context()),
		idMap: idMap,
	}
	return rs
}

/*
NewRecordSet returns a new empty Recordset on the model given by ptrStructOrTableName and the
given Environment.
*/
func NewRecordSet(env Environment, ptrStructOrTableName interface{}) RecordSet {
	return newRecordStruct(env, ptrStructOrTableName, make(map[int64]bool))
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
