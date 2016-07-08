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
	"database/sql"
	"reflect"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/npiganeau/yep/yep/tools"
)

var modelRegistry *modelCollection

type Option int

type modelCollection struct {
	sync.RWMutex
	bootstrapped        bool
	registryByName      map[string]*modelInfo
	registryByTableName map[string]*modelInfo
}

// get the given modelInfo by name or by table name
func (mc *modelCollection) get(nameOrJSON string) (mi *modelInfo, ok bool) {
	mi, ok = mc.registryByName[nameOrJSON]
	if !ok {
		mi, ok = mc.registryByTableName[nameOrJSON]
	}
	return
}

// add the given modelInfo to the modelCollection
func (mc *modelCollection) add(mi *modelInfo) {
	mc.registryByName[mi.name] = mi
	mc.registryByTableName[mi.tableName] = mi
}

// newModelCollection returns a pointer to a new modelCollection
func newModelCollection() *modelCollection {
	return &modelCollection{
		registryByName:      make(map[string]*modelInfo),
		registryByTableName: make(map[string]*modelInfo),
	}
}

type modelInfo struct {
	name      string
	tableName string
	fields    *fieldsCollection
	methods   *methodsCollection
}

// addFieldsFromStruct adds the fields of the given struct to our
// modelInfo
func (mi *modelInfo) addFieldsFromStruct(structPtr interface{}) {
	typ := reflect.TypeOf(structPtr)
	if typ.Kind() != reflect.Ptr {
		tools.LogAndPanic(log, "StructPtr must be a pointer to a struct", "model", mi.name, "received", structPtr)
	}
	typ = typ.Elem()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		fi := createFieldInfo(sf, mi)
		if fi == nil {
			// unexported field
			continue
		}
		if fi.json == "id" {
			// do not change primary key
			continue
		}
		mi.fields.add(fi)
	}
}

// getRelatedModelInfo returns the modelInfo of the related model when
// following path. If the last part of path is a non relational field,
// it is simply ignored. Path can be formed from field names or JSON names.
func (mi *modelInfo) getRelatedModelInfo(path string) *modelInfo {
	if path == "" {
		return mi
	}
	exprs := strings.Split(path, ExprSep)
	jsonizeExpr(mi, exprs)
	fi, ok := mi.fields.get(exprs[0])
	if !ok {
		tools.LogAndPanic(log, "Unknown field in model", "field", exprs[0], "model", mi.name)
	}
	if fi.relatedModel == nil {
		// The field is a non relational field, so we are already
		// on the related modelInfo.
		return mi
	}
	if len(exprs) > 1 {
		return fi.relatedModel.getRelatedModelInfo(strings.Join(exprs[1:], ExprSep))
	}
	return fi.relatedModel
}

// getRelatedFieldIfo returns the fieldInfo of the related field when
// following path. Path can be formed from field names or JSON names.
func (mi *modelInfo) getRelatedFieldInfo(path string) *fieldInfo {
	colExprs := strings.Split(path, ExprSep)
	var rmi *modelInfo
	num := len(colExprs)
	if len(colExprs) > 1 {
		rmi = mi.getRelatedModelInfo(path)
	} else {
		rmi = mi
	}
	fi, ok := rmi.fields.get(colExprs[num-1])
	if !ok {
		tools.LogAndPanic(log, "Unknown field in model", "field", colExprs[num-1], "model", rmi.name)
	}
	return fi
}

// scanToFieldMap scans the db query result r into the given FieldMap.
// Unlike slqx.MapScan, the returned interface{} values are of the type
// of the modelInfo fields instead of the database types.
func (mi *modelInfo) scanToFieldMap(r sqlx.ColScanner, dest *FieldMap) error {
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	// Step 1: We create a []interface{} which is in fact a []*interface{}
	// and we scan our DB row into it. This enables us to get null values
	// without panic, since null values will map to nil.
	dbValues := make([]interface{}, len(columns))
	for i := range dbValues {
		dbValues[i] = new(interface{})
	}

	err = r.Scan(dbValues...)
	if err != nil {
		return err
	}

	// Step 2: We scan values with the type of the corresponding fieldInfo
	// if the value is not nil.
	destVals := reflect.ValueOf(dest).Elem()
	for i, dbValue := range dbValues {
		fi := mi.getRelatedFieldInfo(strings.Replace(columns[i], sqlSep, ExprSep, -1))
		fType := fi.structField.Type
		var val reflect.Value
		switch {
		case dbValue == nil:
			val = reflect.Zero(fType)
		case reflect.PtrTo(fType).Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()):
			val = reflect.New(fType)
			scanFunc := val.MethodByName("Scan")
			inArgs := []reflect.Value{reflect.ValueOf(dbValue).Elem()}
			scanFunc.Call(inArgs)
		default:
			if fType.Kind() == reflect.Ptr {
				// Scan foreign keys into int64
				fType = reflect.TypeOf(int64(0))
			}
			val = reflect.ValueOf(dbValue).Elem().Elem()
			if val.IsValid() {
				if typ := val.Type(); typ.ConvertibleTo(fType) {
					val = val.Convert(fType)
				}
			}
		}
		destVals.SetMapIndex(reflect.ValueOf(columns[i]), val)
	}

	return r.Err()
}

// CreateModel creates a new model with the given name
// Available options are
// - TRANSIENT_MODEL: each instance of the model will have a limited lifetime in database (used for wizards)
func CreateModel(name string, options ...Option) {
	var opts Option
	for _, o := range options {
		opts |= o
	}
	var model interface{}
	if opts&TRANSIENT_MODEL > 0 {
		model = new(BaseTransientModel)
	} else {
		model = new(BaseModel)
	}
	createModelInfo(name, model)
	declareBaseMethods(name)
}

// ExtendModel extends the model given by its name with the given struct pointers
func ExtendModel(name string, structPtrs ...interface{}) {
	mi, ok := modelRegistry.get(name)
	if !ok {
		tools.LogAndPanic(log, "Unknown model", "model", name)
	}
	for _, structPtr := range structPtrs {
		mi.addFieldsFromStruct(structPtr)
	}
}

// createModelInfo creates and populates a new modelInfo with the given name
// by parsing the given struct pointer.
func createModelInfo(name string, model interface{}) {
	mi := &modelInfo{
		name:      name,
		tableName: tools.SnakeCaseString(name),
		fields:    newFieldsCollection(),
		methods:   newMethodsCollection(),
	}
	pk := &fieldInfo{
		name:      "ID",
		json:      "id",
		mi:        mi,
		required:  true,
		fieldType: tools.INTEGER,
		structField: reflect.TypeOf(
			struct {
				ID int64
			}{},
		).Field(0),
	}
	mi.fields.add(pk)
	mi.addFieldsFromStruct(model)
	modelRegistry.add(mi)
}
