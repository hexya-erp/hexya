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

package tools

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

/*
ConvertDotName converts an Odoo dotted style model name (e.g. res.partner) into
a YEP Pascal cased style (e.g. ResPartner).
*/
func ConvertDotName(val string) string {
	var res string
	tokens := strings.Split(val, ".")
	for _, token := range tokens {
		res += strings.Title(token)
	}
	return res
}

/*
ConvertUnderscoreName converts an Odoo snake style method name (e.g. search_read) into
a YEP Pascal cased style (e.g. SearchRead).
*/
func ConvertUnderscoreName(val string) string {
	var res string
	tokens := strings.Split(val, "_")
	for _, token := range tokens {
		res += strings.Title(token)
	}
	return res
}

/*
GetStructFieldByJSONTag returns a pointer value of the struct field of the given structValue
with the given JSON tag. If several are found, it returns the first one. If none are
found it returns the zero value. If structType is not a Type of Kind struct, then it panics.
*/
func GetStructFieldByJSONTag(structValue reflect.Value, tag string) (sf reflect.Value) {
	for i := 0; i < structValue.NumField(); i++ {
		sField := structValue.Field(i)
		sfTag := structValue.Type().Field(i).Tag.Get("json")
		if sfTag == tag {
			sf = sField.Addr()
			return
		}
	}
	return
}

/*
UnmarshalJSONValue unmarshals the given data as a Value of type []byte into
the dst Value. dst must be a pointer Value
*/
func UnmarshalJSONValue(data, dst reflect.Value) error {
	if dst.Type().Kind() != reflect.Ptr {
		panic(fmt.Errorf("<UnmarshalJSONValue>dst must be a pointer value"))
	}
	umArgs := []reflect.Value{data, reflect.New(dst.Type().Elem())}
	res := reflect.ValueOf(json.Unmarshal).Call(umArgs)[0]
	if res.Interface() != nil {
		return res.Interface().(error)
	}
	dst.Elem().Set(umArgs[1].Elem())
	return nil
}
