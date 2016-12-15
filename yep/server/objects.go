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
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/tools"
	"github.com/npiganeau/yep/yep/tools/logging"
)

// CallParams is the arguments' struct for the Execute function.
// It defines a method to call on a model with the given args and keyword args.
type CallParams struct {
	Model  string                     `json:"model"`
	Method string                     `json:"method"`
	Args   []json.RawMessage          `json:"args"`
	KWArgs map[string]json.RawMessage `json:"kwargs"`
}

// Execute executes a method on an object
func Execute(uid int64, params CallParams) (res interface{}, rError error) {
	checkUser(uid)

	// Create new Environment with new transaction
	rError = models.ExecuteInNewEnvironment(uid, func(env models.Environment) {

		// Create RecordSet from Environment
		rs, parms, single := createRecordCollection(env, params)
		ctx := extractContext(params)
		rs = rs.WithNewContext(&ctx)

		methodName := tools.ConvertMethodName(params.Method)

		// Parse Args and KWArgs using the following logic:
		// - If 2nd argument of the function is a struct, then:
		//     * Parse remaining Args in the struct fields
		//     * Parse KWArgs in the struct fields, possibly overwriting Args
		// - Else:
		//     * Parse Args as the function args
		//     * Ignore KWArgs
		var fnArgs []interface{}
		if rs.MethodType(methodName).NumIn() > 1 {
			fnSecondArgType := rs.MethodType(methodName).In(1)
			if fnSecondArgType.Kind() == reflect.Struct {
				// 2nd argument is a struct,
				fnArgs = make([]interface{}, 1)
				argStructValue := reflect.New(fnSecondArgType).Elem()
				putParamsValuesInStruct(&argStructValue, parms)
				putKWValuesInStruct(&argStructValue, params.KWArgs)
				fnArgs[0] = argStructValue.Interface()
			} else {
				// Second argument is not a struct, so we parse directly in the function args
				fnArgs = make([]interface{}, len(parms))
				err := putParamsValuesInArgs(&fnArgs, rs.MethodType(methodName), parms)
				if err != nil {
					logging.LogAndPanic(log, err.Error(), "method", methodName, "args", parms)
				}
			}
		}

		res = rs.Call(methodName, fnArgs...)

		resVal := reflect.ValueOf(res)
		if single && resVal.Kind() == reflect.Slice {
			// Return only the first element of the slice if called with only one id.
			newRes := reflect.New(resVal.Type().Elem()).Elem()
			if resVal.Len() > 0 {
				newRes.Set(resVal.Index(0))
			}
			res = newRes.Interface()
		}
		// Return ID(s) if res is a *RecordSet
		if rec, ok := res.(models.RecordCollection); ok {
			if len(rec.Ids()) == 1 {
				res = rec.Ids()[0]
			} else {
				res = rec.Ids()
			}
		}
	})

	return
}

// putParamsValuesInStruct decodes parms and sets the fields of the structValue
// with the values of parms, in order.
func putParamsValuesInStruct(structValue *reflect.Value, parms []json.RawMessage) {
	argStructValue := *structValue
	for i := 0; i < argStructValue.NumField(); i++ {
		if len(parms) <= i {
			// We have less arguments than the size of the struct
			break
		}
		argsValue := reflect.ValueOf(parms[i])
		fieldPtrValue := reflect.New(argStructValue.Type().Field(i).Type)
		if err := tools.UnmarshalJSONValue(argsValue, fieldPtrValue); err != nil {
			// We deliberately continue here to have default value if there is an error
			// This is to manage cases where the given data type is inconsistent (such
			// false instead of [] or object{}).
			continue
		}
		argStructValue.Field(i).Set(fieldPtrValue.Elem())
	}
}

// putKWValuesInStruct decodes kwArgs and sets the fields of the structValue
// with the values of kwArgs, mapping each field with its entry in kwArgs.
func putKWValuesInStruct(structValue *reflect.Value, kwArgs map[string]json.RawMessage) {
	argStructValue := *structValue
	for k, v := range kwArgs {
		field := tools.GetStructFieldByJSONTag(argStructValue, k)
		if field.IsValid() {
			if err := tools.UnmarshalJSONValue(reflect.ValueOf(v), field); err != nil {
				// We deliberately continue here to have default value if there is an error
				// This is to manage cases where the given data type is inconsistent (such
				// false instead of [] or object{}).
				continue
			}
		}
	}
}

// putParamsValuesInArgs decodes parms and sets fnArgs with the types of methodType arguments
// and the values of parms, in order.
func putParamsValuesInArgs(fnArgs *[]interface{}, methodType reflect.Type, parms []json.RawMessage) error {
	numArgs := methodType.NumIn() - 1
	for i := 0; i < len(parms); i++ {
		if (methodType.IsVariadic() && len(parms) < numArgs-1) ||
			(!methodType.IsVariadic() && len(parms) < numArgs) {
			// We have less arguments than the arguments of the method
			return fmt.Errorf("Wrong number of args in non-struct function args (%d instead of %d)", len(parms), numArgs)
		}
		argsValue := reflect.ValueOf(parms[i])
		resValue := reflect.New(methodType.In(i + 1))
		if err := tools.UnmarshalJSONValue(argsValue, resValue); err != nil {
			// Same remark as above
			log.Debug("Unable to unmarshal argument", "error", err)
			continue
		}
		(*fnArgs)[i] = resValue.Elem().Interface()
	}
	return nil
}

// createRecordCollection creates a RecordCollection instance from the given environment, based
// on the given params. If the first argument given in params can be parsed as an id or a slice
// of ids, then it is used to populate the RecordCollection. Otherwise, it returns an empty
// RecordCollection. This function also returns the remaining arguments after id(s) have been
// parsed, and a boolean value set to true if the RecordSet has only one ID.
func createRecordCollection(env models.Environment, params CallParams) (rc models.RecordCollection, remainingParams []json.RawMessage, single bool) {
	modelName := tools.ConvertModelName(params.Model)
	rc = env.Pool(modelName)

	// Try to parse the first argument of Args as id or ids.
	var idsParsed bool
	var ids []float64
	if len(params.Args) > 0 {
		if err := json.Unmarshal(params.Args[0], &ids); err != nil {
			// Unable to unmarshal in a list of IDs, trying with a single id
			var id float64
			if err := json.Unmarshal(params.Args[0], &id); err == nil {
				rc = rc.Filter("ID", "=", id)
				single = true
				idsParsed = true
			}
		} else {
			rc = rc.Filter("ID", "in", ids)
			idsParsed = true
		}
	}

	remainingParams = params.Args
	if idsParsed {
		// We remove ids already parsed from args
		remainingParams = remainingParams[1:]
	}
	return
}

// extractContext extracts the context from the given params and returns it.
func extractContext(params CallParams) types.Context {
	ctxStr, ok := params.KWArgs["context"]
	var ctx types.Context
	if ok {
		if err := json.Unmarshal(ctxStr, &ctx); err != nil {
			logging.LogAndPanic(log, "Unable to JSON unmarshal context", "context_string", ctxStr)
		}
	}
	return ctx
}

// checkUser panics if the given uid is 0 (i.e. no user is logged in).
func checkUser(uid int64) {
	if uid == 0 {
		logging.LogAndPanic(log, "User must be logged in to call model method")
	}
}

// GetFieldValue retrieves the given field of the given model and id.
func GetFieldValue(uid, id int64, model, field string) (res interface{}, rError error) {
	checkUser(uid)
	rError = models.ExecuteInNewEnvironment(uid, func(env models.Environment) {
		model = tools.ConvertModelName(model)
		res = env.Pool(model).Filter("ID", "=", id).Get(field)
	})

	return
}

// SearchReadParams is the args struct for the SearchRead function.
type SearchReadParams struct {
	Context types.Context `json:"context"`
	Domain  models.Domain `json:"domain"`
	Fields  []string      `json:"fields"`
	Limit   interface{}   `json:"limit"`
	Model   string        `json:"model"`
	Offset  int           `json:"offset"`
	Sort    string        `json:"sort"`
}

// SearchReadResult is the result struct for the SearchRead function.
type SearchReadResult struct {
	Records []models.FieldMap `json:"records"`
	Length  int               `json:"length"`
}

// SearchRead retrieves database records according to the filters defined in params.
func SearchRead(uid int64, params SearchReadParams) (res *SearchReadResult, rError error) {
	checkUser(uid)
	rError = models.ExecuteInNewEnvironment(uid, func(env models.Environment) {
		model := tools.ConvertModelName(params.Model)
		rs := env.Pool(model)
		srp := models.SearchParams{
			Domain: params.Domain,
			Fields: params.Fields,
			Offset: params.Offset,
			Limit:  params.Limit,
			Order:  params.Sort,
		}
		records := rs.Call("SearchRead", srp).([]models.FieldMap)
		length := rs.SearchCount()
		res = &SearchReadResult{
			Records: records,
			Length:  length,
		}
	})
	return
}
