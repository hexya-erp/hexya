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
	"errors"
	"reflect"
	"strings"

	"github.com/npiganeau/yep/yep/tools"
)

const (
	defaultStructTagName  = "yep"
	defaultStructTagDelim = ";"
	defaultTagDataDelim   = ","
)

var (
	supportedTag = map[string]int{
		"type":           2,
		"fk":             2,
		"comodel":        2,
		"selection":      2,
		"size":           2,
		"digits":         2,
		"json":           2,
		"string":         2,
		"help":           2,
		"required":       1,
		"optional":       1,
		"unique":         1,
		"not-unique":     1,
		"index":          1,
		"nocopy":         1,
		"copy":           1,
		"group_operator": 2,
		"compute":        2,
		"related":        2,
		"store":          1,
		"unstore":        1,
		"depends":        2,
		"inherits":       1,
	}
	Testing = false
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
checkStructPtr checks that the given data is a struct ptr valid for receiving data from
the database through the RecordSet API. That is:
- data must be a struct pointer
- The struct must contain an ID field of type int64
It returns an error if it not the case.
*/
func checkStructPtr(data interface{}) error {
	val := reflect.ValueOf(data)
	ind := reflect.Indirect(val)
	indType := ind.Type()
	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		return errors.New("Argument must be a struct pointer")
	}

	if _, ok := indType.FieldByName("ID"); !ok {
		return errors.New("Struct must have an ID field")
	}
	if f, _ := indType.FieldByName("ID"); f.Type != reflect.TypeOf(int64(1.0)) {
		return errors.New("Struct ID field must be of type `int64`")
	}
	return nil
}

// checkStructSlicePtr checks that the given data is a pointer to a slice of
// struct pointers valid for receiving data from the database through the RecordSet API.
// That is:
// - data must be a pointer to a slice of struct pointers
// - The struct must contain an ID field of type int64
// It returns an error if it not the case.
func checkStructSlicePtr(data interface{}) error {
	val := reflect.ValueOf(data)
	ind := reflect.Indirect(val)
	indType := ind.Type()
	if val.Kind() != reflect.Ptr ||
		ind.Kind() != reflect.Slice ||
		indType.Elem().Kind() != reflect.Ptr ||
		indType.Elem().Elem().Kind() != reflect.Struct {

		return errors.New("Argument must be a pointer to a slice of struct pointers")
	}
	structType := indType.Elem().Elem()

	if _, ok := structType.FieldByName("ID"); !ok {
		return errors.New("Struct must have an ID field")
	}
	if f, _ := structType.FieldByName("ID"); f.Type != reflect.TypeOf(int64(1.0)) {
		return errors.New("Struct ID field must be of type `int64`")
	}
	return nil
}

// jsonizeExpr returns an expression slice with field names changed to the fields json names
// Computation is made relatively to the given modelInfo
// e.g. [User Profile Name] -> [user_id profile_id name]
func jsonizeExpr(mi *modelInfo, exprs []string) []string {
	if len(exprs) == 0 {
		return []string{}
	}
	var res []string
	fi, ok := mi.fields.get(exprs[0])
	if !ok {
		tools.LogAndPanic(log, "Unknown expression for model", "expression", exprs, "model", mi.name)
	}
	res = append(res, fi.json)
	if len(exprs) > 1 {
		if fi.relatedModel != nil {
			res = append(res, jsonizeExpr(fi.relatedModel, exprs[1:])...)
		} else {
			tools.LogAndPanic(log, "Field is not a relation in model", "field", exprs[0], "model", mi.name)
		}
	}
	return res
}

// jsonizePath returns a path with field names changed to the field json names
// Computation is made relatively to the given modelInfo
// e.g. User.Profile.Name -> user_id.profile_id.name
func jsonizePath(mi *modelInfo, path string) string {
	exprs := strings.Split(path, ExprSep)
	exprs = jsonizeExpr(mi, exprs)
	return strings.Join(exprs, ExprSep)
}

// structToMap returns the fields and values of the given struct pointer in a map.
// struct fields that equals their type's Go zero value will not be set in the map.
func structToMap(structPtr interface{}) FieldMap {
	val := reflect.ValueOf(structPtr)
	ind := reflect.Indirect(val)
	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		tools.LogAndPanic(log, "structPtr must be a pointer to a struct", "structPtr", structPtr)
	}
	res := make(FieldMap)
	for i := 0; i < ind.NumField(); i++ {
		fieldName := ind.Type().Field(i).Name
		fieldValue := ind.Field(i)

		if reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(fieldValue.Type()).Interface()) {
			// Omit field if its value equals its type's zero value
			continue
		}

		res[fieldName] = ind.Field(i).Interface()
	}
	return res
}

// mapToStruct populates the given structPtr with the values in fMap.
func mapToStruct(rc RecordCollection, structPtr interface{}, fMap FieldMap) {
	fMap = nestMap(fMap)
	val := reflect.ValueOf(structPtr)
	ind := reflect.Indirect(val)
	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		tools.LogAndPanic(log, "structPtr must be a pointer to a struct", "structPtr", structPtr)
	}
	for i := 0; i < ind.NumField(); i++ {
		fVal := ind.Field(i)
		sf := ind.Type().Field(i)
		fi, ok := rc.mi.fields.get(sf.Name)
		if !ok {
			tools.LogAndPanic(log, "Unregistered field in model", "field", sf.Name, "model", rc.ModelName())
		}

		mValue, mValExists := fMap[fi.json]
		if mValExists && mValue != nil {
			var convertedValue reflect.Value
			if sf.Type.Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
				var relRC RecordCollection
				switch r := mValue.(type) {
				case int64:
					relRC = newRecordCollection(rc.Env(), fi.relatedModel.name).withIds([]int64{r})
				case []int64:
					relRC = newRecordCollection(rc.Env(), fi.relatedModel.name).withIds(r)
				}
				if sf.Type == reflect.TypeOf(RecordCollection{}) {
					convertedValue = reflect.ValueOf(relRC)
				} else {
					// We have a generated RecordSet Type
					convertedValue = reflect.New(sf.Type).Elem()
					convertedValue.FieldByName("RecordCollection").Set(reflect.ValueOf(relRC))
				}
			} else {
				convertedValue = reflect.ValueOf(mValue).Convert(fVal.Type())
			}
			fVal.Set(convertedValue)
		}
	}
}

// nestMap returns a nested FieldMap from a flat FieldMap with dotted
// field names. nestMap is lazy and only nests the first level.
func nestMap(fMap FieldMap) FieldMap {
	res := make(FieldMap)
	nested := make(map[string]FieldMap)
	for k, v := range fMap {
		exprs := strings.Split(k, ExprSep)
		if len(exprs) == 1 {
			// We are in the top map here
			res[k] = v
			continue
		}
		if _, exists := nested[exprs[0]]; !exists {
			nested[exprs[0]] = make(FieldMap)
		}
		nested[exprs[0]][strings.Join(exprs[1:], ExprSep)] = v
	}
	// Get nested FieldMap and assign to top key
	for k, fm := range nested {
		res[k] = fm
	}
	return res
}

/*
getFieldType returns the FieldType corresponding to the given reflect.Type.
*/
func getFieldType(typ reflect.Type) tools.FieldType {
	k := typ.Kind()
	switch {
	case k == reflect.Bool:
		return tools.BOOLEAN
	case k >= reflect.Int && k <= reflect.Uint64:
		return tools.INTEGER
	case k == reflect.Float32 || k == reflect.Float64:
		return tools.FLOAT
	case k == reflect.String:
		return tools.CHAR
	case k == reflect.Ptr:
		indTyp := typ.Elem()
		switch indTyp.Kind() {
		case reflect.Struct:
			return tools.MANY2ONE
		case reflect.Slice:
			return tools.ONE2MANY
		}
	}
	switch typ {
	case reflect.TypeOf(DateTime{}):
		return tools.DATETIME
	case reflect.TypeOf(Date{}):
		return tools.DATE
	}
	tools.LogAndPanic(log, "Unable to match field type with go Type. Please specify 'type()' in struct tag", "type", typ)
	return tools.NO_TYPE
}

// filterFields returns the fields slice with only the fields that appear in
// the filters slice or all fields if filters is an empty slice.
// Fields or Filters can contain field names or JSON names.
// The result has only JSON names
func filterFields(mi *modelInfo, fields, filters []string) []string {
	var res []string
	for _, field := range fields {
		fieldExprs := jsonizeExpr(mi, strings.Split(field, ExprSep))
		field = strings.Join(fieldExprs, ExprSep)
		if len(filters) == 0 {
			res = append(res, field)
			continue
		}
		for _, filter := range filters {
			filterExprs := jsonizeExpr(mi, strings.Split(filter, ExprSep))
			filter = strings.Join(filterExprs, ExprSep)
			if field == filter {
				res = append(res, field)
			}
		}
	}
	return res
}

// filterOnDBFields returns the given fields slice with only stored fields
// This function also adds the "id" field to the list if not present
func filterOnDBFields(mi *modelInfo, fields []string) []string {
	var res []string
	// Check if fields are stored
	for _, field := range fields {
		fieldExprs := jsonizeExpr(mi, strings.Split(field, ExprSep))
		fi, ok := mi.fields.get(fieldExprs[0])
		if !ok {
			tools.LogAndPanic(log, "Unknown Field in model", "field", fieldExprs[0], "model", mi.name)
		}
		var resExprs []string
		if fi.isStored() {
			resExprs = append(resExprs, fi.json)
		}
		if len(fieldExprs) > 1 {
			// Related field (e.g. User.Profile.Age)
			if fi.relatedModel != nil {
				subFieldName := strings.Join(fieldExprs[1:], ExprSep)
				subFieldRes := filterOnDBFields(fi.relatedModel, []string{subFieldName})
				if len(subFieldRes) > 0 {
					resExprs = append(resExprs, subFieldRes[0])
				}
			} else {
				tools.LogAndPanic(log, "Field is not a relation in model", "field", fieldExprs[0], "model", mi.name)
			}
		}
		if len(resExprs) > 0 {
			res = append(res, strings.Join(resExprs, ExprSep))
		}
	}
	// Check we have "id" else add it to our list
	var idPresent bool
	for _, r := range res {
		if r == "id" {
			idPresent = true
			break
		}
	}
	if !idPresent {
		res = append(res, "id")
	}
	return res
}

// convertInterfaceToFielMap converts the given data which can be of type:
// - FieldMap
// - map[string]interface{}
// - struct pointer
// to a FieldMap
func convertInterfaceToFieldMap(data interface{}) FieldMap {
	var fMap FieldMap
	switch d := data.(type) {
	case FieldMap:
		fMap = d
	case map[string]interface{}:
		fMap = FieldMap(d)
	default:
		if err := checkStructPtr(data); err != nil {
			tools.LogAndPanic(log, err.Error(), "data", data)
		}
		fMap = structToMap(data)
	}
	return fMap
}
