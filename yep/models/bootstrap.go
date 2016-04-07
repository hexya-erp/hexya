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
	"time"
)

type Option int

func ComputeWriteDate(rs RecordSet) orm.Params {
	return orm.Params{"WriteDate": time.Now()}
}

type BaseModel struct {
	ID         int64     `orm:"column(id)"`
	CreateDate time.Time `orm:"auto_now_add"`
	CreateUid  int64
	WriteDate  time.Time `yep:"compute(ComputeWriteDate),store,depends(ID)" orm:"null"`
	WriteUid   int64
}

type BaseTransientModel struct {
	ID int64 `orm:"column(id)"`
}

const (
	TRANSIENT_MODEL Option = 1 << iota
)

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
	DeclareMethod(name, "ComputeWriteDate", ComputeWriteDate)
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
func BootStrap() {
	err := orm.RunSyncdb("default", true, orm.Debug)
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
