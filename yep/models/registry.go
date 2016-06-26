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
	"sync"

	"fmt"
	"github.com/npiganeau/yep/yep/tools"
	"reflect"
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
		panic(fmt.Errorf("StructPtr must be a pointer to a struct"))
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
	//orm.RegisterModelWithName(name, model)
	//registerModelFields(name, model)
	declareBaseMethods(name)
}

// ExtendModel extends the model given by its name with the given struct pointers
func ExtendModel(name string, structPtrs ...interface{}) {
	mi, ok := modelRegistry.get(name)
	if !ok {
		panic(fmt.Errorf("Unknown model `%s`", name))
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
	}
	mi.fields.add(pk)
	mi.addFieldsFromStruct(model)
	modelRegistry.add(mi)
}
