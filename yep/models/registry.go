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
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/tools"
	"github.com/npiganeau/yep/yep/tools/logging"
)

var modelRegistry *modelCollection

// Option describes a optional feature of a model
type Option int

type modelCollection struct {
	sync.RWMutex
	bootstrapped        bool
	commonMixins        []*modelInfo
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

// mustGet the given modelInfo by name or by table name.
// It panics if the modelInfo does not exist
func (mc *modelCollection) mustGet(nameOrJSON string) *modelInfo {
	mi, ok := mc.get(nameOrJSON)
	if !ok {
		logging.LogAndPanic(log, "Unknown model", "model", nameOrJSON)
	}
	return mi
}

// mustGetMixInModel returns the modelInfo of the given mixin name.
// It panics if the given name is not the name of a registered mixin
func (mc *modelCollection) mustGetMixInModel(name string) *modelInfo {
	mixInMI := mc.mustGet(name)
	if !mixInMI.isMixin() {
		logging.LogAndPanic(log, "Model is not a mixin model", "model", name)
	}
	return mixInMI
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
	name          string
	options       Option
	acl           *security.AccessControlList
	rulesRegistry *recordRuleRegistry
	tableName     string
	fields        *fieldsCollection
	methods       *methodsCollection
	mixins        []*modelInfo
}

// addFieldsFromStruct adds the fields of the given struct to our
// modelInfo
func (mi *modelInfo) addFieldsFromStruct(structPtr interface{}) {
	typ := reflect.TypeOf(structPtr)
	if typ.Kind() != reflect.Ptr {
		logging.LogAndPanic(log, "StructPtr must be a pointer to a struct", "model", mi.name, "received", structPtr)
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
// following path.
// - If skipLast is true, getRelatedModelInfo does not follow the last part of the path
// - If the last part of path is a non relational field, it is simply ignored, whatever
// the value of skipLast.
//
// Paths can be formed from field names or JSON names.
func (mi *modelInfo) getRelatedModelInfo(path string, skipLast ...bool) *modelInfo {
	if path == "" {
		return mi
	}
	var skip bool
	if len(skipLast) > 0 {
		skip = skipLast[0]
	}

	exprs := strings.Split(path, ExprSep)
	jsonizeExpr(mi, exprs)
	fi := mi.fields.mustGet(exprs[0])
	if fi.relatedModel == nil || (len(exprs) == 1 && skip) {
		// The field is a non relational field, so we are already
		// on the related modelInfo. Or we have only 1 exprs and we skip the last one.
		return mi
	}
	if len(exprs) > 1 {
		return fi.relatedModel.getRelatedModelInfo(strings.Join(exprs[1:], ExprSep), skipLast...)
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
		rmi = mi.getRelatedModelInfo(path, true)
	} else {
		rmi = mi
	}
	fi := rmi.fields.mustGet(colExprs[num-1])
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

	// Step 2: We populate our FieldMap with these values
	for i, dbValue := range dbValues {
		colName := strings.Replace(columns[i], sqlSep, ExprSep, -1)
		dbVal := reflect.ValueOf(dbValue).Elem().Interface()
		(*dest)[colName] = dbVal
	}

	// Step 3: We convert values with the type of the corresponding fieldInfo
	// if the value is not nil.
	mi.convertValuesToFieldType(dest)
	return r.Err()
}

// convertValuesToFieldType converts all values of the given FieldMap to
// their type in the modelInfo.
func (mi *modelInfo) convertValuesToFieldType(fMap *FieldMap) {
	destVals := reflect.ValueOf(fMap).Elem()
	for colName, fMapValue := range *fMap {
		if val, ok := fMapValue.(bool); ok && !val {
			// Hack to manage client returning false instead of nil
			fMapValue = nil
		}
		fi := mi.getRelatedFieldInfo(colName)
		fType := fi.structField.Type
		if fType == reflect.TypeOf(fMapValue) {
			// If we already have the good type, don't do anything
			continue
		}
		var val reflect.Value
		switch {
		case fMapValue == nil:
			// dbValue is null, we put the type zero value instead
			val = reflect.Zero(fType)
		case reflect.PtrTo(fType).Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()):
			// the type implements sql.Scanner, so we call Scan
			val = reflect.New(fType)
			scanFunc := val.MethodByName("Scan")
			inArgs := []reflect.Value{reflect.ValueOf(fMapValue)}
			scanFunc.Call(inArgs)
		default:
			rVal := reflect.ValueOf(fMapValue)
			if rVal.Type().Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
				// Our field is a related field
				ids := fMapValue.(RecordSet).Ids()
				if fType == reflect.TypeOf(int64(0)) {
					if len(ids) > 0 {
						val = reflect.ValueOf(ids[0])
					} else {
						val = reflect.ValueOf(nil)
					}
				} else if fType == reflect.TypeOf([]int64{}) {
					val = reflect.ValueOf(ids)
				} else {
					logging.LogAndPanic(log, "Non consistent type", "model", mi.name, "field", colName, "type", fType, "value", fMapValue)
				}
			} else {
				val = reflect.ValueOf(fMapValue)
				if val.IsValid() {
					if typ := val.Type(); typ.ConvertibleTo(fType) {
						val = val.Convert(fType)
					}
				}
			}
		}
		destVals.SetMapIndex(reflect.ValueOf(colName), val)
	}
}

// isMixin returns true if this is a mixin model.
func (mi *modelInfo) isMixin() bool {
	if mi.options&MixinModel > 0 {
		return true
	}
	return false
}

// CreateModel creates a new model with the given name and
// extends it with the given struct pointers.
func CreateModel(name string, structPtrs ...interface{}) {
	model := new(BaseModel)
	createModelInfo(name, model, Option(0))
	ExtendModel(name, structPtrs...)
}

// CreateMixinModel creates a new mixin model with the given name and
// extends it with the given struct pointers.
func CreateMixinModel(name string, structPtrs ...interface{}) {
	model := new(BaseModel)
	createModelInfo(name, model, MixinModel)
	ExtendModel(name, structPtrs...)
}

// CreateTransientModel creates a new mixin model with the given name and
// extends it with the given struct pointers.
func CreateTransientModel(name string, structPtrs ...interface{}) {
	model := new(BaseModel)
	createModelInfo(name, model, MixinModel)
	ExtendModel(name, structPtrs...)
}

// ExtendModel extends the model given by its name with the given struct pointers
func ExtendModel(name string, structPtrs ...interface{}) {
	mi := modelRegistry.mustGet(name)
	for _, structPtr := range structPtrs {
		mi.addFieldsFromStruct(structPtr)
	}
}

// MixInModel extends targetModel by importing all fields and methods of mixInModel.
// MixIn methods and fields have a lower priority than those of the model and are
// overridden by the them when applicable.
func MixInModel(targetModel, mixInModel string) {
	mi := modelRegistry.mustGet(targetModel)
	if mi.isMixin() {
		logging.LogAndPanic(log, "Trying to mixin a mixin model", "model", targetModel, "mixin", mixInModel)
	}
	mixInMI := modelRegistry.mustGetMixInModel(mixInModel)
	mi.mixins = append(mi.mixins, mixInMI)
}

// MixInAllModel extends all models with the given mixInModel.
// Mixins added with this method have lower priority than MixIns
// that are directly applied to a model, which have themselves a
// lower priority than the fields and methods of the model.
//
// Note that models extension will be deferred at bootstrap time,
// which means that all models, including those that are not yet
// defined at the time this function is called, will be extended.
func MixInAllModels(mixInModel string) {
	mixInMI := modelRegistry.mustGetMixInModel(mixInModel)
	modelRegistry.commonMixins = append(modelRegistry.commonMixins, mixInMI)
}

// createModelInfo creates and populates a new modelInfo with the given name
// by parsing the given struct pointer.
func createModelInfo(name string, model interface{}, options Option) {
	mi := &modelInfo{
		name:          name,
		options:       options,
		acl:           security.NewAccessControlList(),
		rulesRegistry: newRecordRuleRegistry(),
		tableName:     tools.SnakeCaseString(name),
		fields:        newFieldsCollection(),
		methods:       newMethodsCollection(),
	}
	pk := &fieldInfo{
		name:      "ID",
		json:      "id",
		acl:       security.NewAccessControlList(),
		mi:        mi,
		required:  true,
		noCopy:    true,
		fieldType: types.Integer,
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
