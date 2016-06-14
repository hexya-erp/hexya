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

	"github.com/npiganeau/yep/yep/tools"
)

type Option int

type modelCollection struct {
	sync.RWMutex
	registryByName      map[string]*modelInfo
	registryByTableName map[string]*modelInfo
}

type modelInfo struct {
	name      string
	tableName string
	fields    *fieldsCollection
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

}

// createModelInfo creates and populates a new modelInfo with the given name
// by parsing the given struct pointer.
func createModelInfo(name string, model interface{}) {
	typ := getStructType(model)
	mi := &modelInfo{
		name:      name,
		tableName: tools.SnakeCaseString(name),
		fields:    newFieldsCollection(),
	}
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		fi := createFieldInfo(sf, mi)
		if fi == nil {
			// unexported field
			continue
		}
		mi.fields.add(fi)
	}
}
