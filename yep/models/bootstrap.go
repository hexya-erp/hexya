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
)

type Option int

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
	orm.RegisterModelWithName(name, model)
	registerModelFields(name, model)
	declareBaseMethods(name)
}

func ExtendModel(name string, models ...interface{}) {
	orm.RegisterModelExtension(name, models...)
	for _, model := range models {
		registerModelFields(name, model)
	}
}

/*
BootStrap freezes model, fields and method caches and syncs the database structure
with the declared data.
*/
func BootStrap(force bool) {
	err := orm.RunSyncdb("default", force, orm.Debug)
	if err != nil {
		panic(fmt.Errorf("Unable to sync database: %s", err))
	}
	methodsCache.Lock()
	defer methodsCache.Unlock()
	methodsCache.done = true

	fieldsCache.Lock()
	defer fieldsCache.Unlock()
	processDepends()

	fieldsCache.done = true
}
