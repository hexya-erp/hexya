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

	"github.com/npiganeau/yep/yep/models/fieldtype"
)

var (
	// Testing is true if we are testing the framework
	Testing bool
)

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
// Computation is made relatively to the given Model
// e.g. [User Profile Name] -> [user_id profile_id name]
func jsonizeExpr(mi *Model, exprs []string) []string {
	if len(exprs) == 0 {
		return []string{}
	}
	var res []string
	fi := mi.fields.MustGet(exprs[0])
	res = append(res, fi.json)
	if len(exprs) > 1 {
		if fi.relatedModel != nil {
			res = append(res, jsonizeExpr(fi.relatedModel, exprs[1:])...)
		} else {
			log.Panic("Field is not a relation in model", "field", exprs[0], "model", mi.name)
		}
	}
	return res
}

// inflate2ManyConditions modifies the query to handle searches on
// one2many and many2many fields.
func inflate2ManyConditions(mi *Model, cond *Condition) {
	for i, cv := range cond.predicates {
		if cv.cond != nil {
			inflate2ManyConditions(mi, cv.cond)
			continue
		}
		path := strings.Join(cv.exprs, ExprSep)
		fi := mi.getRelatedFieldInfo(path)
		switch fi.fieldType {
		case fieldtype.One2Many:
			cond.predicates[i].exprs = append(cv.exprs, "id")
		case fieldtype.Many2Many:
		case fieldtype.Rev2One:
		}
	}
}

// addNameSearchesToCondition recursively modifies the given condition to search
// on the name of the related records if they point to a relation field.
func addNameSearchesToCondition(mi *Model, cond *Condition) {
	for i, p := range cond.predicates {
		if len(p.exprs) == 0 {
			continue
		}
		fi := mi.getRelatedFieldInfo(strings.Join(p.exprs, ExprSep))
		if !fi.isRelationField() {
			continue
		}
		switch p.arg.(type) {
		case bool:
			cond.predicates[i].arg = int64(0)
		case string:
			cond.predicates[i].exprs = addNameSearchToExprs(mi, fi, p.exprs)
		}
		if p.cond != nil {
			addNameSearchesToCondition(mi, p.cond)
		}
	}
}

// addNameSearchToExprs modifies the given exprs to search on the name of the related record
// if it points to a relation field.
func addNameSearchToExprs(mi *Model, fi *Field, exprs []string) []string {
	_, exists := fi.relatedModel.fields.get("name")
	if exists {
		exprs = append(exprs, "name")
	}
	return exprs
}

// jsonizePath returns a path with field names changed to the field json names
// Computation is made relatively to the given Model
// e.g. User.Profile.Name -> user_id.profile_id.name
func jsonizePath(mi *Model, path string) string {
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
		log.Panic("structPtr must be a pointer to a struct", "structPtr", structPtr)
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
		log.Panic("structPtr must be a pointer to a struct", "structPtr", structPtr)
	}
	for i := 0; i < ind.NumField(); i++ {
		fVal := ind.Field(i)
		sf := ind.Type().Field(i)
		fi, ok := rc.model.fields.get(sf.Name)
		if !ok {
			log.Panic("Unregistered field in model", "field", sf.Name, "model", rc.ModelName())
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

// filterOnDBFields returns the given fields slice with only stored fields
// This function also adds the "id" field to the list if not present unless dontAddID is true
func filterOnDBFields(mi *Model, fields []string, dontAddID ...bool) []string {
	var res []string
	// Check if fields are stored
	for _, field := range fields {
		fieldExprs := jsonizeExpr(mi, strings.Split(field, ExprSep))
		fi := mi.fields.MustGet(fieldExprs[0])
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
				log.Panic("Field is not a relation in model", "field", fieldExprs[0], "model", mi.name)
			}
		}
		if len(resExprs) > 0 {
			res = append(res, strings.Join(resExprs, ExprSep))
		}
	}
	if len(dontAddID) == 0 || !dontAddID[0] {
		res = addIDIfNotPresent(res)
	}
	return res
}

// filterMapOnStoredFields returns a new FieldMap from fMap
// with only stored fields keys.
func filterMapOnStoredFields(mi *Model, fMap FieldMap) FieldMap {
	newFMap := make(FieldMap)
	for field, value := range fMap {
		if fi := mi.getRelatedFieldInfo(field); fi.isStored() {
			newFMap[field] = value
		}
	}
	return newFMap
}

// ConvertInterfaceToFieldMap converts the given data which can be of type:
// - FieldMap
// - map[string]interface{}
// - struct pointer
// to a FieldMap
func ConvertInterfaceToFieldMap(data interface{}) FieldMap {
	var fMap FieldMap
	switch d := data.(type) {
	case FieldMap:
		fMap = d
	case map[string]interface{}:
		fMap = FieldMap(d)
	default:
		if err := checkStructPtr(data); err != nil {
			log.Panic(err.Error(), "data", data)
		}
		fMap = structToMap(data)
	}
	return fMap
}

// addIDIfNotPresent returns a new fields slice including ID if it
// is not already present. Otherwise returns the original slice.
func addIDIfNotPresent(fields []string) []string {
	var hadID bool
	for _, fName := range fields {
		if fName == "id" || fName == "ID" {
			hadID = true
		}
	}
	if !hadID {
		fields = append(fields, "id")
	}
	return fields
}

// convertToStringSlice converts the given FieldNamer slice into a slice of strings
func convertToStringSlice(fieldNames []FieldNamer) []string {
	res := make([]string, len(fieldNames))
	for i, v := range fieldNames {
		res[i] = string(v.FieldName())
	}
	return res
}

// ConvertToFieldNameSlice converts the given string fields slice into a slice of FieldNames
func ConvertToFieldNameSlice(fields []string) []FieldNamer {
	res := make([]FieldNamer, len(fields))
	for i, v := range fields {
		res[i] = FieldName(v)
	}
	return res
}

// getGroupCondition returns the condition to retrieve the individual aggregated rows in vals
// knowing that they were grouped by groups
func getGroupCondition(groups []string, vals map[string]interface{}) *Condition {
	res := newCondition()
	for _, group := range groups {
		res = res.And().Field(group).Equals(vals[group])
	}
	return res
}

// serializePredicates returns a list that mimics Odoo domains from the given
// condition values.
func serializePredicates(predicates []predicate) []interface{} {
	var res []interface{}
	i := 0
	for i < len(predicates) {
		if predicates[i].isOr {
			subRes := []interface{}{"|"}
			subRes = appendPredicateToSerial(subRes, predicates[i])
			subRes, i = consumeAndPredicates(i+1, predicates, subRes)
			res = append(subRes, res...)
		} else {
			res, i = consumeAndPredicates(i, predicates, res)
		}
	}
	return res
}

// consumeAndPredicates appends res with all successive AND predicates
// starting from position i and returns the next position as second argument.
func consumeAndPredicates(i int, predicates []predicate, res []interface{}) ([]interface{}, int) {
	if i >= len(predicates) || predicates[i].isOr {
		return res, i
	}
	j := i
	for j < len(predicates)-1 {
		if predicates[j+1].isOr {
			break
		}
		j++
	}
	for k := i; k < j; k++ {
		res = append(res, "&")
		res = appendPredicateToSerial(res, predicates[k])
	}
	res = appendPredicateToSerial(res, predicates[j])
	return res, j + 1
}

// appendPredicateToSerial appends the given predicate to the given serialized
// predicate list and returns the result.
func appendPredicateToSerial(res []interface{}, predicate predicate) []interface{} {
	if predicate.isCond {
		res = append(res, serializePredicates(predicate.cond.predicates)...)
	} else {
		res = append(res, []interface{}{strings.Join(predicate.exprs, ExprSep), predicate.operator, predicate.arg})
	}
	return res
}
