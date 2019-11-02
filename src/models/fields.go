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
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	"github.com/hexya-erp/hexya/src/tools/strutils"
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

type ctxType int

const (
	ctxNone = iota
	ctxValue
	ctxContext
	ctxFK
)

// computeData holds data to recompute another field.
// - model is a pointer to the Model instance to recompute
// - fieldName is the name of the field to recompute in model.
// - compute is the name of the method to call on model
// - path is the search string that will be used to find records to update
// (e.g. path = "Profile.BestPost").
// - stored is true if the computed field is stored
type computeData struct {
	model     *Model
	stored    bool
	fieldName string
	compute   string
	path      string
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

// Get returns the Field of the field with the given name.
// name can be either the name of the field or its JSON name.
func (fc *FieldsCollection) Get(name string) (fi *Field, ok bool) {
	fi, ok = fc.registryByName[name]
	if !ok {
		fi, ok = fc.registryByJSON[name]
	}
	return
}

// MustGet returns the Field of the field with the given name or panics
// name can be either the name of the field or its JSON name.
func (fc *FieldsCollection) MustGet(name string) *Field {
	fi, ok := fc.Get(name)
	if !ok {
		log.Panic("Unknown field in model", "model", fc.model.name, "field", name)
	}
	return fi
}

// storedFieldNames returns a slice with the names of all the stored fields
// If fields are given, return only names in the list
func (fc *FieldsCollection) storedFieldNames(fieldNames ...FieldName) []FieldName {
	var res []FieldName
	for fName, fi := range fc.registryByName {
		var keepField bool
		if len(fieldNames) == 0 {
			keepField = true
		} else {
			for _, f := range fieldNames {
				if fName == f.Name() {
					keepField = true
					break
				}
			}
		}
		if (fi.isStored() || fi.isRelatedField()) && keepField {
			res = append(res, fc.model.FieldName(fName))
		}
	}
	return res
}

// allFieldNames returns a slice with the name of all field's JSON names of this collection
func (fc *FieldsCollection) allFieldNames() []FieldName {
	res := make([]FieldName, len(fc.registryByJSON))
	var i int
	for f := range fc.registryByName {
		res[i] = fc.model.FieldName(f)
		i++
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
	name             string
	json             string
	description      string
	help             string
	stored           bool
	required         bool
	readOnly         bool
	requiredFunc     func(Environment) (bool, Conditioner)
	readOnlyFunc     func(Environment) (bool, Conditioner)
	invisibleFunc    func(Environment) (bool, Conditioner)
	unique           bool
	index            bool
	compute          string
	depends          []string
	relatedModelName string
	relatedModel     *Model
	reverseFK        string
	jsonReverseFK    string
	m2mRelModel      *Model
	m2mOurField      *Field
	m2mTheirField    *Field
	selection        types.Selection
	selectionFunc    func() types.Selection
	fieldType        fieldtype.Type
	groupOperator    string
	size             int
	digits           nbutils.Digits
	structField      reflect.StructField
	relatedPathStr   string
	relatedPath      FieldName
	dependencies     []computeData
	embed            bool
	noCopy           bool
	defaultFunc      func(Environment) interface{}
	onDelete         OnDeleteAction
	onChange         string
	onChangeWarning  string
	onChangeFilters  string
	constraint       string
	inverse          string
	filter           *Condition
	contexts         FieldContexts
	ctxType          ctxType
	updates          []map[string]interface{}
}

// isComputedField returns true if this field is computed
func (f *Field) isComputedField() bool {
	return f.compute != ""
}

// isComputedField returns true if this field is related
func (f *Field) isRelatedField() bool {
	return f.relatedPath != nil
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

// isSettable returns true if the given field can be set directly
func (f *Field) isSettable() bool {
	if f.isComputedField() && f.inverse == "" {
		return false
	}
	return true
}

// isReadOnly returns true if this field must not be set directly
// by the user.
func (f *Field) isReadOnly() bool {
	if f.readOnly {
		return true
	}
	fInfo := f
	if fInfo.isRelatedField() {
		fInfo = f.model.getRelatedFieldInfo(fInfo.relatedPath)
	}
	if fInfo.compute != "" && fInfo.inverse == "" {
		return true
	}
	return false
}

// isContextedField returns true if the value of this field depends on contexts
func (f *Field) isContextedField() bool {
	if f.contexts != nil && len(f.contexts) > 0 {
		return true
	}
	return false
}

// JSON returns this field name as FieldName type
func (f *Field) JSON() string {
	return f.json
}

// String method for the Field type. Returns the field's name.
func (f *Field) Name() string {
	return f.name
}

var _ FieldName = new(Field)

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
	res := strutils.SnakeCase(fName)
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
//
// If mixin is true, the created M2M model is created as a mixin model.
func createM2MRelModelInfo(relModelName, model1, model2, field1, field2 string, mixin bool) (*Model, *Field, *Field) {
	if relMI, exists := Registry.Get(relModelName); exists {
		var m1, m2 *Field
		for fName, fi := range relMI.fields.registryByName {
			if fName == field1 {
				m1 = fi
			} else if fName == field2 {
				m2 = fi
			}
		}
		return relMI, m1, m2
	}

	newMI := &Model{
		name:         relModelName,
		tableName:    strutils.SnakeCase(relModelName),
		fields:       newFieldsCollection(),
		methods:      newMethodsCollection(),
		options:      Many2ManyLinkModel | SystemModel,
		sqlErrors:    make(map[string]string),
		defaultOrder: []orderPredicate{{field: ID}},
	}
	if mixin {
		newMI.options |= MixinModel
	}
	ourField := &Field{
		name:             field1,
		json:             strutils.SnakeCase(field1) + "_id",
		model:            newMI,
		required:         true,
		noCopy:           true,
		fieldType:        fieldtype.Many2One,
		relatedModelName: model1,
		index:            true,
		onDelete:         Cascade,
		structField: reflect.StructField{
			Name: field1,
			Type: reflect.TypeOf(int64(0)),
		},
	}
	newMI.fields.add(ourField)

	theirField := &Field{
		name:             field2,
		json:             strutils.SnakeCase(field2) + "_id",
		model:            newMI,
		required:         true,
		noCopy:           true,
		fieldType:        fieldtype.Many2One,
		relatedModelName: model2,
		index:            true,
		onDelete:         Cascade,
		structField: reflect.StructField{
			Name: field2,
			Type: reflect.TypeOf(int64(0)),
		},
	}
	newMI.fields.add(theirField)
	Registry.add(newMI)
	return newMI, ourField, theirField
}

// createContextsModel creates a new contexts model for holding field values that depends on contexts
func createContextsModel(fi *Field, contexts FieldContexts) *Model {
	if !fi.isStored() {
		log.Panic("You cannot add contexts to non stored fields", "model", fi.model.name, "field", fi.name)
	}
	name := fmt.Sprintf("%sHexya%s", fi.model.name, fi.name)
	newModel := Model{
		name:          name,
		rulesRegistry: newRecordRuleRegistry(),
		tableName:     strutils.SnakeCase(name),
		fields:        newFieldsCollection(),
		methods:       newMethodsCollection(),
		options:       ContextsModel | SystemModel,
		sqlErrors:     make(map[string]string),
		defaultOrder:  []orderPredicate{{field: ID}},
	}
	pkField := &Field{
		name:      "ID",
		json:      "id",
		model:     &newModel,
		required:  true,
		noCopy:    true,
		fieldType: fieldtype.Integer,
		structField: reflect.TypeOf(
			struct {
				ID int64
			}{},
		).Field(0),
	}
	newModel.fields.add(pkField)
	fkField := &Field{
		name:             "Record",
		json:             "record_id",
		model:            &newModel,
		required:         true,
		noCopy:           true,
		fieldType:        fieldtype.Many2One,
		relatedModelName: fi.model.name,
		relatedModel:     fi.model,
		index:            true,
		onDelete:         Cascade,
		ctxType:          ctxFK,
		structField: reflect.StructField{
			Name: "Record",
			Type: reflect.TypeOf(int64(0)),
		},
	}
	newModel.fields.add(fkField)
	valueField := *fi
	valueField.model = &newModel
	valueField.compute = ""
	valueField.embed = false
	valueField.stored = false
	valueField.onChange = ""
	valueField.constraint = ""
	valueField.contexts = nil
	valueField.ctxType = ctxValue
	if valueField.defaultFunc == nil && valueField.required {
		valueField.defaultFunc = DefaultValue(reflect.Zero(valueField.structField.Type).Interface())
	}
	newModel.fields.add(&valueField)

	for ctName := range contexts {
		ctField := &Field{
			name:      ctName,
			json:      strutils.SnakeCase(ctName),
			model:     &newModel,
			noCopy:    true,
			fieldType: fieldtype.Char,
			index:     true,
			ctxType:   ctxContext,
			structField: reflect.StructField{
				Name: ctName,
				Type: reflect.TypeOf(""),
			},
		}
		newModel.fields.add(ctField)
	}
	Registry.add(&newModel)
	injectMixInModel(Registry.MustGet("BaseMixin"), &newModel)
	return &newModel
}

// processDepends populates the dependencies of each Field from the depends strings of
// each Field instances.
func processDepends() {
	for _, mi := range Registry.registryByTableName {
		for _, fInfo := range mi.fields.registryByJSON {
			var refName string
			for _, depString := range fInfo.depends {
				if depString == "" {
					continue
				}
				tokens := jsonizeExpr(mi, strings.Split(depString, ExprSep))
				refName = tokens[len(tokens)-1]
				path := strings.Join(tokens[:len(tokens)-1], ExprSep)
				targetComputeData := computeData{
					model:     mi,
					stored:    fInfo.stored,
					fieldName: fInfo.name,
					compute:   fInfo.compute,
					path:      path,
				}
				refModelInfo := mi.getRelatedModelInfo(mi.FieldName(path))
				refField := refModelInfo.fields.MustGet(refName)
				refField.dependencies = append(refField.dependencies, targetComputeData)
			}
		}
	}
}

// checkComputeMethodsSignature check the signature of all methods used
// in computed fields and for OnChange methods.
// It panics if it is not the case.
func checkComputeMethodsSignature() {
	for _, model := range Registry.registryByName {
		for _, fi := range model.fields.computedFields {
			method := fi.model.methods.MustGet(fi.compute)
			if err := checkMethType(method, "Compute methods"); err != nil {
				log.Panic(err.Error(), "model", method.model.name, "method", method.name, "field", fi.name)
			}
		}
		for _, fi := range model.fields.computedStoredFields {
			method := fi.model.methods.MustGet(fi.compute)
			if err := checkMethType(method, "Compute method for stored fields"); err != nil {
				log.Panic(err.Error(), "model", method.model.name, "method", method.name, "field", fi.name)
			}
		}
		for _, fi := range model.fields.registryByName {
			if fi.onChange == "" {
				continue
			}
			method := fi.model.methods.MustGet(fi.onChange)
			if err := checkMethType(method, "OnchangeMethods"); err != nil {
				log.Panic(err.Error(), "model", method.model.name, "method", method.name, "field", fi.name)
			}
		}
		for _, fi := range model.fields.registryByName {
			if fi.onChangeWarning == "" {
				continue
			}
			method := fi.model.methods.MustGet(fi.onChangeWarning)
			if err := checkOnChangeWarningType(method); err != nil {
				log.Panic(err.Error(), "model", method.model.name, "method", method.name, "field", fi.name)
			}
		}
		for _, fi := range model.fields.registryByName {
			if fi.onChangeFilters == "" {
				continue
			}
			method := fi.model.methods.MustGet(fi.onChangeFilters)
			if err := checkOnChangeFiltersType(method); err != nil {
				log.Panic(err.Error(), "model", method.model.name, "method", method.name, "field", fi.name)
			}
		}
		for _, fi := range model.fields.registryByName {
			if fi.inverse == "" {
				continue
			}
			method := model.methods.MustGet(fi.inverse)
			methType := method.methodType
			if methType.NumIn() != 2 {
				log.Panic("Inverse methods should have 2 arguments", "model", model.name, "field", fi.name, "method", method.name)
			}
			if methType.NumOut() != 0 {
				log.Panic("Inverse methods should not return any value", "model", model.name, "field", fi.name, "method", method.name)
			}
		}
	}
}

// checkMethType panics if the given method does not have
// the correct number and type of arguments and returns for a compute/onChange method
func checkMethType(method *Method, label string) error {
	methType := method.methodType
	var msg string
	switch {
	case methType.NumIn() != 1:
		msg = fmt.Sprintf("%s should have no arguments", label)
	case methType.NumOut() == 0:
		msg = fmt.Sprintf("%s should return a value", label)
	case methType.NumOut() > 1:
		msg = fmt.Sprintf("Too many return values for %s", label)
	case !methType.Out(0).Implements(reflect.TypeOf((*RecordData)(nil)).Elem()):
		msg = fmt.Sprintf("%s returned value must implement models.RecordData", label)
	}
	if msg != "" {
		return errors.New(msg)
	}
	return nil
}

// checkOnChangeWarningType panics if the given method does not have
// the correct number and type of arguments and returns for a onChangeWarning method
func checkOnChangeWarningType(method *Method) error {
	methType := method.methodType
	var msg string
	switch {
	case methType.NumIn() != 1:
		msg = "OnChangeWarning methods should have no arguments"
	case methType.NumOut() == 0:
		msg = "OnChangeWarning methods should return a value"
	case methType.NumOut() > 1:
		msg = "Too many return values for OnChangeWarning method"
	case methType.Out(0) != reflect.TypeOf("string"):
		msg = "OnChangeWarning methods returned value must be of type string"
	}
	if msg != "" {
		return errors.New(msg)
	}
	return nil
}

// checkOnChangeFiltersType panics if the given method does not have
// the correct number and type of arguments and returns for a onChangeFilters method
func checkOnChangeFiltersType(method *Method) error {
	methType := method.methodType
	var msg string
	switch {
	case methType.NumIn() != 1:
		msg = "OnChangeFilters methods should have no arguments"
	case methType.NumOut() == 0:
		msg = "OnChangeFilters methods should return a value"
	case methType.NumOut() > 1:
		msg = "Too many return values for OnChangeFilters method"
	case methType.Out(0) != reflect.TypeOf(map[FieldName]Conditioner{}):
		msg = "OnChangeFilters methods returned value must be of type map[models.FieldName]models.Conditioner"
	}
	if msg != "" {
		return errors.New(msg)
	}
	return nil
}
