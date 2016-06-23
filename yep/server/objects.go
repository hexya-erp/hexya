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
func Execute(uid int64, params CallParams) (res interface{}, rError error) {
	var rs models.RecordSet
	defer func() {
		if r := recover(); r != nil {
			rs.Env().Cr().Rollback()
			res = nil
			rError = fmt.Errorf("%s", r)
			return
		}
		rError = rs.Env().Cr().Commit()
	}()
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

	// Create new Environment with new transaction
	env := models.NewEnvironment(uid, ctx)

	// Create RecordSet from Environment
	model := tools.ConvertModelName(params.Model)
	rs = *env.Pool(model)

	//// Try to parse the first argument of Args as id or ids.
	//var single bool
	//var idsParsed bool
	//var ids []float64
	//if len(params.Args) > 0 {
	//	if err := json.Unmarshal(params.Args[0], &ids); err != nil {
	//		// Unable to unmarshal in a list of IDs, trying with a single id
	//		var id float64
	//		if err := json.Unmarshal(params.Args[0], &id); err == nil {
	//			rs = rs.Filter("ID", id)
	//			single = true
	//			idsParsed = true
	//		}
	//	} else {
	//		rs = rs.Filter("ID__in", ids)
	//		idsParsed = true
	//	}
	//}
	//
	//parms := params.Args
	//if idsParsed {
	//	// We remove ids already parsed from args
	//	parms = parms[1:]
	//}
	//
	//methodName := tools.ConvertMethodName(params.Method)
	//
	//// Parse Args and KWArgs using the following logic:
	//// - If 2nd argument of the function is a struct, then:
	////     * Parse remaining Args in the struct fields
	////     * Parse KWArgs in the struct fields, possibly overwriting Args
	//// - Else:
	////     * Parse Args as the function args
	////     * Ignore KWArgs
	//fnArgs := make([]interface{}, rs.MethodType(methodName).NumIn()-1)
	//if rs.MethodType(methodName).NumIn() > 1 {
	//	fnSecondArgType := rs.MethodType(methodName).In(1)
	//	if fnSecondArgType.Kind() == reflect.Struct {
	//		// 2nd argument is a struct,
	//		argStructValue := reflect.New(fnSecondArgType).Elem()
	//		for i := 0; i < fnSecondArgType.NumField(); i++ {
	//			// We parse parms into the struct
	//			if len(parms) <= i {
	//				// We have less arguments than the size of the struct
	//				break
	//			}
	//			argsValue := reflect.ValueOf(parms[i])
	//			fieldPtrValue := reflect.New(argStructValue.Type().Field(i).Type)
	//			if err := tools.UnmarshalJSONValue(argsValue, fieldPtrValue); err != nil {
	//				// We deliberately continue here to have default value if there is an error
	//				// This is to manage cases where the given data type is inconsistent (such
	//				// false instead of [] or object{}).
	//				continue
	//			}
	//			argStructValue.Field(i).Set(fieldPtrValue.Elem())
	//		}
	//		for k, v := range params.KWArgs {
	//			// We parse params.KWArgs into the struct
	//			field := tools.GetStructFieldByJSONTag(argStructValue, k)
	//			if field.IsValid() {
	//				if err := tools.UnmarshalJSONValue(reflect.ValueOf(v), field); err != nil {
	//					// Same remark as above
	//					continue
	//				}
	//			}
	//		}
	//		fnArgs[0] = argStructValue.Interface()
	//	} else {
	//		// Second argument is not a struct, so we parse directly in the function args
	//		for i := 1; i < rs.MethodType(methodName).NumIn(); i++ {
	//			if len(parms) <= i-1 {
	//				// We have less arguments than the arguments of the method
	//				panic(fmt.Errorf("<Execute>Wrong number of args in non-struct function args"))
	//			}
	//			argsValue := reflect.ValueOf(parms[i-1])
	//			resValue := reflect.New(rs.MethodType(methodName).In(i))
	//			if err := tools.UnmarshalJSONValue(argsValue, resValue); err != nil {
	//				// Same remark as above
	//				continue
	//			}
	//			fnArgs[i-1] = resValue.Elem().Interface()
	//		}
	//	}
	//}
	//
	//res = rs.Call(methodName, fnArgs...)
	//
	//resVal := reflect.ValueOf(res)
	//if single && idsParsed && resVal.Kind() == reflect.Slice {
	//	// Return only the first element of the slice if called with only one id.
	//	newRes := reflect.New(resVal.Type().Elem()).Elem()
	//	if resVal.Len() > 0 {
	//		newRes.Set(resVal.Index(0))
	//	}
	//	res = newRes.Interface()
	//}

	// Return ID(s) if res is a RecordSet
	if rec, ok := res.(models.RecordSet); ok {
		if len(rec.Ids()) == 1 {
			res = rec.Ids()[0]
		} else {
			res = rec.Ids()
		}
	}

	return
}

/*
GetFieldValue retrieves the given field of the given model and id.
*/
func GetFieldValue(uid, id int64, model, field string) (res interface{}, rError error) {
	//var rs models.RecordSet
	//defer func() {
	//	if r := recover(); r != nil {
	//		if rs != nil {
	//			rs.Env().Cr().Rollback()
	//		}
	//		res = nil
	//		rError = fmt.Errorf("%s", r)
	//	}
	//}()
	//if uid == 0 {
	//	panic(fmt.Errorf("User must be logged in to retrieve field."))
	//}
	//model = tools.ConvertModelName(model)
	//field = models.GetFieldName(model, field)
	//env := models.NewCursorEnvironment(uid)
	//rs = models.NewRecordSet(env, model).Filter("ID", id).Search()
	//var parms orm.ParamsList
	//rs.ValuesFlat(&parms, field)
	//if len(parms) == 0 {
	//	res = nil
	//} else {
	//	res = parms[0]
	//}
	return
}

type SearchReadParams struct {
	Context tools.Context `json:"context"`
	Domain  models.Domain `json:"domain"`
	Fields  []string      `json:"fields"`
	Limit   interface{}   `json:"limit"`
	Model   string        `json:"model"`
	Offset  int           `json:"offset"`
	Sort    string        `json:"sort"`
}

type SearchReadResult struct {
	Records []models.FieldMap `json:"records"`
	Length  int64             `json:"length"`
}

/*
SearchRead retrieves database records according to the filters defined in params.
*/
func SearchRead(uid int64, params SearchReadParams) (res *SearchReadResult, rError error) {
	var rs models.RecordSet
	defer func() {
		if r := recover(); r != nil {
			rs.Env().Cr().Rollback()
			res = nil
			rError = fmt.Errorf("%s", r)
		}
	}()
	if uid == 0 {
		panic(fmt.Errorf("User must be logged in to search database."))
	}
	//model := tools.ConvertModelName(params.Model)
	//env := models.NewCursorEnvironment(uid)
	//rs = models.NewRecordSet(env, model)
	//srp := models.SearchParams{
	//	Domain: params.Domain,
	//	Fields: params.Fields,
	//	Offset: params.Offset,
	//	Limit:  params.Limit,
	//	Order:  params.Sort,
	//}
	//records := rs.Call("SearchRead", srp).([]orm.Params)
	//length := rs.SearchCount()
	//res = &SearchReadResult{
	//	Records: records,
	//	Length:  length,
	//}
	return
}
