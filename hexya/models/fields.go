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
	"sync"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/tools/nbutils"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
)

// An OnDeleteAction defines what to be done with this record when
// the target record is deleted.
type OnDeleteAction string

const (
	// SetNull sets the foreign key to null in referencing records. This is the default
	SetNull OnDeleteAction = "set null"
	// Restrict throws an error if there are record referencing the deleted one.
	Restrict OnDeleteAction = "restrict"
	// Cascade deletes all referencing records.
	Cascade OnDeleteAction = "cascade"
)

// computeData holds data to recompute another field.
// - model is a pointer to the Model instance to recompute
// - compute is the name of the method to call on model
// - path is the search string that will be used to find records to update
// (e.g. path = "Profile.BestPost").
type computeData struct {
	model   *Model
	compute string
	path    string
}

// FieldsCollection is a collection of Field instances in a model.
type FieldsCollection struct {
	sync.RWMutex
	model                *Model
	registryByName       map[string]*Field
	registryByJSON       map[string]*Field
	computedFields       []*Field
	computedStoredFields []*Field
	relatedFields        []*Field
	bootstrapped         bool
}

// get returns the Field of the field with the given name.
// name can be either the name of the field or its JSON name.
func (fc *FieldsCollection) get(name string) (fi *Field, ok bool) {
	fi, ok = fc.registryByName[name]
	if !ok {
		fi, ok = fc.registryByJSON[name]
	}
	return
}

// MustGet returns the Field of the field with the given name or panics
// name can be either the name of the field or its JSON name.
func (fc *FieldsCollection) MustGet(name string) *Field {
	fi, ok := fc.get(name)
	if !ok {
		log.Panic("Unknown field in model", "model", fc.model.name, "field", name)
	}
	return fi
}

// storedFieldNames returns a slice with the names of all the stored fields
// If fields are given, return only names in the list
func (fc *FieldsCollection) storedFieldNames(fieldNames ...string) []string {
	var res []string
	for fName, fi := range fc.registryByName {
		var keepField bool
		if len(fieldNames) == 0 {
			keepField = true
		} else {
			for _, f := range fieldNames {
				if fName == f {
					keepField = true
					break
				}
			}
		}
		if fi.isStored() && keepField {
			res = append(res, fName)
		}
	}
	return res
}

// getComputedFields returns the slice of Field of the computed, but not
// stored fields of the given modelName.
// If fields are given, return only Field instances in the list
func (fc *FieldsCollection) getComputedFields(fields ...string) (fil []*Field) {
	fInfos := fc.computedFields
	if len(fields) > 0 {
		for _, f := range fields {
			for _, fInfo := range fInfos {
				if f == fInfo.name || f == fInfo.json {
					fil = append(fil, fInfo)
					continue
				}
			}
		}
	} else {
		fil = fInfos
	}
	return
}

// newFieldsCollection returns a pointer to a new empty FieldsCollection with
// all maps initialized.
func newFieldsCollection() *FieldsCollection {
	return &FieldsCollection{
		registryByName: make(map[string]*Field),
		registryByJSON: make(map[string]*Field),
	}
}

// add the given Field to the FieldsCollection.
func (fc *FieldsCollection) add(fInfo *Field) {
	if _, exists := fc.registryByName[fInfo.name]; exists {
		log.Panic("Trying to add already existing field", "model", fInfo.model.name, "field", fInfo.name)
	}
	fc.register(fInfo)
}

// register adds the given fInfo in the collection.
func (fc *FieldsCollection) register(fInfo *Field) {
	fc.Lock()
	defer fc.Unlock()

	checkFieldInfo(fInfo)
	name := fInfo.name
	jsonName := fInfo.json
	fc.registryByName[name] = fInfo
	fc.registryByJSON[jsonName] = fInfo
	if fInfo.isComputedField() {
		if fInfo.stored {
			fc.computedStoredFields = append(fc.computedStoredFields, fInfo)
		} else {
			fc.computedFields = append(fc.computedFields, fInfo)
		}
	}
	if fInfo.isRelatedField() {
		fc.relatedFields = append(fc.relatedFields, fInfo)
	}
}

// Field holds the meta information about a field
type Field struct {
	model            *Model
	acl              *security.AccessControlList
	name             string
	json             string
	description      string
	help             string
	stored           bool
	required         bool
	unique           bool
	index            bool
	compute          string
	depends          []string
	relatedModelName string
	relatedModel     *Model
	reverseFK        string
	m2mRelModel      *Model
	m2mOurField      *Field
	m2mTheirField    *Field
	selection        types.Selection
	fieldType        fieldtype.Type
	groupOperator    string
	size             int
	digits           nbutils.Digits
	structField      reflect.StructField
	relatedPath      string
	dependencies     []computeData
	embed            bool
	noCopy           bool
	defaultFunc      func(Environment, FieldMap) interface{}
	onDelete         OnDeleteAction
	onChange         string
	constraint       string
	inverse          string
	filter           *Condition
	translate        bool
}

// isComputedField returns true if this field is computed
func (f *Field) isComputedField() bool {
	return f.compute != ""
}

// isComputedField returns true if this field is related
func (f *Field) isRelatedField() bool {
	return f.relatedPath != ""
}

// isRelationField returns true if this field points to another model
func (f *Field) isRelationField() bool {
	// We check on relatedModelName and not relatedModel to be able
	// to use this method even if the models have not been bootstrapped yet.
	return f.relatedModelName != ""
}

// isStored returns true if this field is stored in database
func (f *Field) isStored() bool {
	if f.fieldType.IsNonStoredRelationType() {
		// reverse fields are not stored
		return false
	}
	if (f.isComputedField() || f.isRelatedField()) && !f.stored {
		// Computed and related non stored fields are not stored
		return false
	}
	return true
}

// isReadOnly returns true if this field must not be set directly
// by the user.
func (f *Field) isReadOnly() bool {
	fInfo := f
	if fInfo.isRelatedField() {
		fInfo = f.model.getRelatedFieldInfo(fInfo.relatedPath)
	}
	if fInfo.compute != "" && fInfo.inverse == "" {
		return true
	}
	return false
}

// checkFieldInfo makes sanity checks on the given Field.
// It panics in case of severe error and logs recoverable errors.
func checkFieldInfo(fi *Field) {
	if fi.fieldType.IsReverseRelationType() && fi.reverseFK == "" {
		log.Panic("'one2many' and 'rev2one' fields must define an 'ReverseFK' parameter", "model",
			fi.model.name, "field", fi.name, "type", fi.fieldType)
	}

	if fi.embed && !fi.fieldType.IsFKRelationType() {
		log.Warn("'Embed' should be set only on many2one or one2one fields", "model", fi.model.name, "field", fi.name,
			"type", fi.fieldType)
		fi.embed = false
	}

	if fi.structField.Type == reflect.TypeOf(RecordCollection{}) && fi.relatedModel.name == "" {
		log.Panic("Undefined relation model on related field", "model", fi.model.name, "field", fi.name,
			"type", fi.fieldType)
	}

	if fi.stored && !fi.isComputedField() {
		log.Warn("'stored' should be set only on computed fields", "model", fi.model.name, "field", fi.name,
			"type", fi.fieldType)
		fi.stored = false
	}
}

// jsonizeFieldName returns a snake cased field name, adding '_id' on x2one
// relation fields and '_ids' to x2many relation fields.
func snakeCaseFieldName(fName string, typ fieldtype.Type) string {
	res := strutils.SnakeCaseString(fName)
	if typ.Is2OneRelationType() {
		res += "_id"
	} else if typ.Is2ManyRelationType() {
		res += "_ids"
	}
	return res
}

// createM2MRelModelInfo creates a Model relModelName (if it does not exist)
// for the m2m relation defined between model1 and model2.
// It returns the Model of the intermediate model, the Field of that model
// pointing to our model, and the Field pointing to the other model.
func createM2MRelModelInfo(relModelName, model1, model2 string) (*Model, *Field, *Field) {
	if relMI, exists := Registry.Get(relModelName); exists {
		var m1, m2 *Field
		for fName, fi := range relMI.fields.registryByName {
			if fName == model1 {
				m1 = fi
			} else if fName == model2 {
				m2 = fi
			}
		}
		return relMI, m1, m2
	}

	newMI := &Model{
		name:      relModelName,
		acl:       security.NewAccessControlList(),
		tableName: strutils.SnakeCaseString(relModelName),
		fields:    newFieldsCollection(),
		methods:   newMethodsCollection(),
		options:   Many2ManyLinkModel,
		sqlErrors: make(map[string]string),
	}
	ourField := &Field{
		name:             model1,
		json:             strutils.SnakeCaseString(model1) + "_id",
		acl:              security.NewAccessControlList(),
		model:            newMI,
		required:         true,
		noCopy:           true,
		fieldType:        fieldtype.Many2One,
		relatedModelName: model1,
		index:            true,
		onDelete:         Cascade,
		structField: reflect.StructField{
			Name: model1,
			Type: reflect.TypeOf(int64(0)),
		},
	}
	newMI.fields.add(ourField)

	theirField := &Field{
		name:             model2,
		json:             strutils.SnakeCaseString(model2) + "_id",
		acl:              security.NewAccessControlList(),
		model:            newMI,
		required:         true,
		noCopy:           true,
		fieldType:        fieldtype.Many2One,
		relatedModelName: model2,
		index:            true,
		onDelete:         Cascade,
		structField: reflect.StructField{
			Name: model2,
			Type: reflect.TypeOf(int64(0)),
		},
	}
	newMI.fields.add(theirField)
	Registry.add(newMI)
	return newMI, ourField, theirField
}

// processDepends populates the dependencies of each Field from the depends strings of
// each Field instances.
func processDepends() {
	for _, mi := range Registry.registryByTableName {
		for _, fInfo := range mi.fields.registryByJSON {
			var refName string
			for _, depString := range fInfo.depends {
				if depString != "" {
					tokens := jsonizeExpr(mi, strings.Split(depString, ExprSep))
					refName = tokens[len(tokens)-1]
					path := strings.Join(tokens[:len(tokens)-1], ExprSep)
					targetComputeData := computeData{
						model:   mi,
						compute: fInfo.compute,
						path:    path,
					}
					refModelInfo := mi.getRelatedModelInfo(path)
					refField := refModelInfo.fields.MustGet(refName)
					refField.dependencies = append(refField.dependencies, targetComputeData)
				}
			}
		}
	}
}

// checkComputeMethodsSignature check the signature of all methods used
// in computed fields and for OnChange methods.
// It panics if it is not the case.
func checkComputeMethodsSignature() {
	for _, mi := range Registry.registryByName {
		for _, fi := range mi.fields.computedFields {
			method := mi.methods.MustGet(fi.compute)
			checkMethType(method, "Compute methods")
		}
		for _, fi := range mi.fields.computedStoredFields {
			method := mi.methods.MustGet(fi.compute)
			checkMethType(method, "Compute method for stored fields")
		}
		for _, fi := range mi.fields.registryByName {
			if fi.onChange == "" {
				continue
			}
			method := mi.methods.MustGet(fi.onChange)
			checkMethType(method, "OnChange methods")
		}
		for _, fi := range mi.fields.registryByName {
			if fi.inverse == "" {
				continue
			}
			method := mi.methods.MustGet(fi.inverse)
			methType := method.methodType
			switch {
			case methType.NumIn() != 2:
				log.Panic("Inverse methods should have 2 arguments", "model", mi.name, "field", fi.name, "method", method.name)
			case !methType.In(1).Implements(reflect.TypeOf((*FieldMapper)(nil)).Elem()):
				log.Panic("Inverse method second argument must be FieldMapper", "model", mi.name, "field", fi.name, "method", method.name)
			}
		}
	}
}

// checkMethType panics if the given method does not have
// the correct number and type for its arguments and returns
func checkMethType(method *Method, label string) {
	methType := method.methodType
	var msg string
	switch {
	case methType.NumIn() != 1:
		msg = fmt.Sprintf("%s should have no arguments", label)
	case methType.NumOut() == 0:
		msg = fmt.Sprintf("%s should return a value", label)
	case !methType.Out(0).Implements(reflect.TypeOf((*FieldMapper)(nil)).Elem()):
		msg = "First return argument must implement models.FieldMapper"
	case methType.NumOut() < 2:
		msg = fmt.Sprintf("%s must return fields to unset as second value", label)
	case methType.NumOut() == 2 && methType.Out(1) != reflect.TypeOf([]FieldNamer{}):
		msg = fmt.Sprintf("Second return value of %s must be []models.FieldNamer", label)
	case methType.NumOut() > 2:
		msg = fmt.Sprintf("Too many return values for %s", label)
	}
	if msg != "" {
		log.Panic(msg, "model", method.model.name, "method", method.name)
	}
}
