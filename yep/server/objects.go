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

package server

import (
	"encoding/json"
	"fmt"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/tools"
	"reflect"
)

type CallParams struct {
	Model  string                     `json:"model"`
	Method string                     `json:"method"`
	Args   []json.RawMessage          `json:"args"`
	KWArgs map[string]json.RawMessage `json:"kwargs"`
}

/*
Execute executes a method on an object
*/
func Execute(uid int64, params CallParams) interface{} {
	ctxStr, ok := params.KWArgs["context"]
	var ctx models.Context
	if ok {
		if err := json.Unmarshal(ctxStr, &ctx); err != nil {
			ok = false
		}
	}
	env := models.NewCursorEnvironment(uid, ctx)

	model := tools.ConvertModelName(params.Model)
	rs := models.NewRecordSet(env, model)
	var single bool
	var ids []float64
	if err := json.Unmarshal(params.Args[0], &ids); err != nil {
		var id float64
		if err := json.Unmarshal(params.Args[0], &id); err != nil {
			panic(fmt.Errorf("Error while unmarshaling ids: %s", err))
		}
		rs = rs.Filter("ID", id)
		single = true
	} else {
		rs = rs.Filter("ID__in", ids)
	}

	methodName := tools.ConvertMethodName(params.Method)

	paramType := rs.MethodType(methodName).In(1).Elem()
	argsParamPtr := reflect.New(paramType).Interface()
	if err := json.Unmarshal(params.Args[1], &argsParamPtr); err != nil {
		panic(fmt.Errorf("Error while unmarshaling args: %s", err))
	}

	res := rs.Call(methodName, argsParamPtr)

	resVal := reflect.ValueOf(res)
	if single && resVal.Kind() == reflect.Slice {
		newRes := reflect.New(resVal.Type().Elem()).Elem()
		if resVal.Len() > 0 {
			newRes.Set(resVal.Index(0))
		}
		res = newRes.Interface()
	}
	return res
}
