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
	"github.com/npiganeau/yep/yep/tools"
	"reflect"
)

/*
Environment holds the context data for a transaction.
*/
type Environment interface {
	Cr() orm.Ormer
	Uid() int64
	Context() tools.Context
	WithContext(ctx tools.Context, replace ...bool) Environment
	Sudo(...int64) Environment
	Pool(interface{}) RecordSet
	Create(interface{}) RecordSet
	Sync(interface{}, ...string) int64
}

/*
RecordSet is both a set of database records and an entrypoint to the models API for CRUD operations.
RecordSet are immutable.
*/
type RecordSet interface {
	fmt.Stringer
	// returns the Environment of the RecordSet
	Env() Environment
	// returns the model name of the RecordSet
	ModelName() string
	// returns the ids of this RecordSet
	Ids() []int64
	// query the database with the current filter and returns a new recordset with the queries ids
	// Does nothing if RecordSet already has Ids in cache.
	Search() RecordSet
	// query the database with the current filter and returns a new recordset with the queries ids
	// Overwrite existing Ids if any
	ForceSearch() RecordSet
	// updates the database with the given data and returns the number of updated rows.
	Write(orm.Params) int64
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
	// MethodType returns the type of the method given by methName
	MethodType(methName string) reflect.Type
	// Returns a slice of RecordSet singleton that constitute this RecordSet
	Records() []RecordSet
	// Panics if RecordSet is not a singleton.
	EnsureOne()
}
