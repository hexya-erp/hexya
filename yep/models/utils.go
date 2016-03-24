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
	"reflect"
	"strings"
)

const (
	defaultStructTagName   = "yep"
	defaultStructTagDelim  = ";"
	defaultDependsTagDelim = ","
)

var (
	supportedTag = map[string]int{
		"store":   1,
		"string":  2,
		"help":    2,
		"compute": 2,
		"depends": 2,
	}
)

/*
parseStructTag parses the given structField tag string and fills:
- attrs if the individual tag is boolean
- tags if the individual tag has a string value
*/
func parseStructTag(data string, attrs *map[string]bool, tags *map[string]string) {
	attr := make(map[string]bool)
	tag := make(map[string]string)
	for _, v := range strings.Split(data, defaultStructTagDelim) {
		v = strings.TrimSpace(v)
		if supportedTag[v] == 1 {
			attr[v] = true
		} else if i := strings.Index(v, "("); i > 0 && strings.Index(v, ")") == len(v)-1 {
			name := v[:i]
			if supportedTag[name] == 2 {
				v = v[i+1 : len(v)-1]
				tag[name] = v
			}
		}
	}
	*attrs = attr
	*tags = tag
}

/*
structToMap returns the fields and values of the given struct pointer in a map.
*/
func structToMap(structPtr interface{}) map[string]interface{} {
	val := reflect.ValueOf(structPtr)
	ind := reflect.Indirect(val)
	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		panic(fmt.Errorf("structPtr must be a pointer to a struct"))
	}
	res := make(map[string]interface{})
	for i := 0; i < ind.NumField(); i++ {
		fieldName := ind.Type().Field(i).Name
		fieldValue := ind.Field(i)
		var resVal interface{}
		if fieldValue.Kind() == reflect.Ptr {
			relInd := reflect.Indirect(fieldValue)
			if relInd.Kind() != reflect.Struct {
				continue
			}
			if !relInd.FieldByName("ID").IsValid() {
				continue
			}
			resVal = relInd.FieldByName("ID").Interface()
		} else {
			resVal = ind.Field(i).Interface()
		}
		res[fieldName] = resVal
	}
	return res
}
