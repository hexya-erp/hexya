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
	"reflect"
	"strings"
	"unicode"
)

/*
ConvertModelName converts an Odoo dotted style model name (e.g. res.partner) into
a YEP Pascal cased style (e.g. ResPartner).
*/
func ConvertModelName(val string) string {
	var res string
	tokens := strings.Split(val, ".")
	for _, token := range tokens {
		res += strings.Title(token)
	}
	return res
}

/*
ConvertMethodName converts an Odoo snake style method name (e.g. search_read) into
a YEP Pascal cased style (e.g. SearchRead).
*/
func ConvertMethodName(val string) string {
	var res string
	tokens := strings.Split(val, "_")
	for _, token := range tokens {
		res += strings.Title(token)
	}
	return res
}

// SnakeCaseString convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func SnakeCaseString(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// TitleString convert the given camelCase string to a title string.
// eg. MyHTMLData => My HTML Data
func TitleString(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, ' ')
		}
		out = append(out, runes[i])
	}

	return string(out)
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
		log.Panic("dst must be a pointer value", "data", data, "dst", dst)
	}
	umArgs := []reflect.Value{data, reflect.New(dst.Type().Elem())}
	res := reflect.ValueOf(json.Unmarshal).Call(umArgs)[0]
	if res.Interface() != nil {
		return res.Interface().(error)
	}
	dst.Elem().Set(umArgs[1].Elem())
	return nil
}

// GetDefaultString returns str if it is not an empty string or def otherwise
func GetDefaultString(str, def string) string {
	if str == "" {
		return def
	}
	return str
}
