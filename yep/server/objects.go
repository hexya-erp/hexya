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
	"reflect"

	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/orm"
	"github.com/npiganeau/yep/yep/tools"
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
	if uid == 0 {
		panic(fmt.Errorf("User must be logged in to call model method."))
	}
	ctxStr, ok := params.KWArgs["context"]
	var ctx tools.Context
	if ok {
		if err := json.Unmarshal(ctxStr, &ctx); err != nil {
			ok = false
		}
	}
	env := models.NewCursorEnvironment(uid, ctx)

	model := tools.ConvertDotName(params.Model)
	rs := models.NewRecordSet(env, model)

	// Try to parse the first argument of Args as id or ids.
	var single bool
	var idsParsed bool
	var ids []float64
	if len(params.Args) > 0 {
		if err := json.Unmarshal(params.Args[0], &ids); err != nil {
			var id float64
			if err := json.Unmarshal(params.Args[0], &id); err == nil {
				rs = rs.Filter("ID", id)
				single = true
				idsParsed = true
			}
		} else {
			rs = rs.Filter("ID__in", ids)
			idsParsed = true
		}
	}

	parms := params.Args
	if idsParsed {
		// We remove ids already parsed from args
		parms = parms[1:]
	}

	methodName := tools.ConvertUnderscoreName(params.Method)

	// Parse Args and KWArgs using the following logic:
	// - If 2nd argument of the function is a struct, then:
	//     * Parse remaining Args in the struct fields
	//     * Parse KWArgs in the struct fields, possibly overwriting Args
	// - Else:
	//     * Parse Args as the function args
	//     * Ignore KWArgs
	fnArgs := make([]interface{}, rs.MethodType(methodName).NumIn()-1)
	if rs.MethodType(methodName).NumIn() > 1 {
		fnSecondArgType := rs.MethodType(methodName).In(1)
		if fnSecondArgType.Kind() == reflect.Struct {
			// 2nd argument is a struct,
			argStructValue := reflect.New(fnSecondArgType).Elem()
			for i := 0; i < fnSecondArgType.NumField(); i++ {
				// We parse parmss into the struct
				if len(parms) <= i {
					// We have less arguments than the size of the struct
					break
				}
				argsValue := reflect.ValueOf(parms[i])
				if err := tools.UnmarshalJSONValue(argsValue, argStructValue.Field(i)); err != nil {
					panic(fmt.Errorf("<Execute>Unable to unmarshal arg %s: %s", parms[i], err))
				}
			}
			for k, v := range params.KWArgs {
				// We parse params.KWArgs into the struct
				field := tools.GetStructFieldByJSONTag(argStructValue, k)
				if field.IsValid() {
					if err := tools.UnmarshalJSONValue(reflect.ValueOf(v), field); err != nil {
						panic(fmt.Errorf("<Execute>Unable to unmarshal kwarg %s: %s", v, err))
					}
				}
			}
			fnArgs[0] = argStructValue.Interface()
		} else {
			// Second argument is not a struct, so we parse directly in the function args
			for i := 1; i < rs.MethodType(methodName).NumIn(); i++ {
				if len(parms) <= i-1 {
					// We have less arguments than the arguments of the method
					panic(fmt.Errorf("<Execute>Wrong number of args in non-struct function args"))
				}
				argsValue := reflect.ValueOf(parms[i-1])
				resValue := reflect.New(rs.MethodType(methodName).In(i))
				if err := tools.UnmarshalJSONValue(argsValue, resValue); err != nil {
					panic(fmt.Errorf("<Execute>Unable to unmarshal arg %s: %s", parms[i-1], err))
				}
				fnArgs[i-1] = resValue.Elem().Interface()
			}
		}
	}

	res := rs.Call(methodName, fnArgs...)

	resVal := reflect.ValueOf(res)
	if single && resVal.Kind() == reflect.Slice {
		// Return only the first element of the slice if called with only one id.
		newRes := reflect.New(resVal.Type().Elem()).Elem()
		if resVal.Len() > 0 {
			newRes.Set(resVal.Index(0))
		}
		res = newRes.Interface()
	}
	return res
}

/*
GetFieldValue retrieves the given field of the given model and id.
*/
func GetFieldValue(uid, id int64, model, field string) interface{} {
	if uid == 0 {
		panic(fmt.Errorf("User must be logged in to retrieve field."))
	}
	model = tools.ConvertDotName(model)
	field = tools.ConvertUnderscoreName(field)
	env := models.NewCursorEnvironment(uid)
	rs := models.NewRecordSet(env, model).Filter("ID", id).Search()
	var res orm.ParamsList
	rs.ValuesFlat(&res, field)
	if len(res) == 0 {
		return nil
	}
	return res[0]
}

type SearchParams struct {
	Context tools.Context `json:"context"`
	Domain  string        `json:"domain"`
	Fields  []string      `json:"fields"`
	Limit   int           `json:"limit"`
	Model   string        `json:"model"`
	Offset  int           `json:"offset"`
	Sort    string        `json:"sort"`
}

type SearchReadResult struct {
	Records []orm.Params `json:"records"`
	Length  int64        `json:"length"`
}

/*
SearchRead retrieves database records according to the filters defined in params.
*/
func SearchRead(uid int64, params SearchParams) *SearchReadResult {
	if uid == 0 {
		panic(fmt.Errorf("User must be logged in to search database."))
	}
	model := tools.ConvertDotName(params.Model)
	var fields []string
	for _, f := range params.Fields {
		fields = append(fields, tools.ConvertUnderscoreName(f))
	}
	env := models.NewCursorEnvironment(uid)
	// TODO Add support for domain & filtering
	limit := 80
	if params.Limit != 0 {
		limit = params.Limit
	}
	rs := models.NewRecordSet(env, model).Limit(limit).Search()
	records := rs.Call("Read", fields).([]orm.Params)
	length := rs.SearchCount()
	return &SearchReadResult{
		Records: records,
		Length:  length,
	}
}
