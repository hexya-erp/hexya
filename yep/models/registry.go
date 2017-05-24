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
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/npiganeau/yep/yep/models/fieldtype"
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/tools"
)

// Registry is the registry of all Model instances.
var Registry *modelCollection

// Option describes a optional feature of a model
type Option int

type modelCollection struct {
	sync.RWMutex
	bootstrapped        bool
	registryByName      map[string]*Model
	registryByTableName map[string]*Model
	sequences           map[string]*Sequence
}

// Get the given Model by name or by table name
func (mc *modelCollection) Get(nameOrJSON string) (mi *Model, ok bool) {
	mi, ok = mc.registryByName[nameOrJSON]
	if !ok {
		mi, ok = mc.registryByTableName[nameOrJSON]
	}
	return
}

// MustGet the given Model by name or by table name.
// It panics if the Model does not exist
func (mc *modelCollection) MustGet(nameOrJSON string) *Model {
	mi, ok := mc.Get(nameOrJSON)
	if !ok {
		log.Panic("Unknown model", "model", nameOrJSON)
	}
	return mi
}

// GetSequence the given Sequence by name or by db name
func (mc *modelCollection) GetSequence(nameOrJSON string) (s *Sequence, ok bool) {
	s, ok = mc.sequences[nameOrJSON]
	if !ok {
		s, ok = mc.sequences[nameOrJSON]
	}
	return
}

// MustGet the given sequence by name or by db name.
// It panics if the Sequence does not exist
func (mc *modelCollection) MustGetSequence(nameOrJSON string) *Sequence {
	s, ok := mc.GetSequence(nameOrJSON)
	if !ok {
		log.Panic("Unknown sequence", "sequence", nameOrJSON)
	}
	return s
}

// mustGetMixInModel returns the Model of the given mixin name.
// It panics if the given name is not the name of a registered mixin
func (mc *modelCollection) mustGetMixInModel(name string) *Model {
	mixInMI := mc.MustGet(name)
	if !mixInMI.isMixin() {
		log.Panic("Model is not a mixin model", "model", name)
	}
	return mixInMI
}

// add the given Model to the modelCollection
func (mc *modelCollection) add(mi *Model) {
	if _, exists := mc.Get(mi.name); exists {
		log.Panic("Trying to add already existing model", "model", mi.name)
	}
	mc.registryByName[mi.name] = mi
	mc.registryByTableName[mi.tableName] = mi
	mi.methods.model = mi
	mi.fields.model = mi
}

// newModelCollection returns a pointer to a new modelCollection
func newModelCollection() *modelCollection {
	return &modelCollection{
		registryByName:      make(map[string]*Model),
		registryByTableName: make(map[string]*Model),
		sequences:           make(map[string]*Sequence),
	}
}

// A Model is the definition of a business object (e.g. a partner, a sale order, etc.)
// including fields and methods.
type Model struct {
	name          string
	options       Option
	acl           *security.AccessControlList
	rulesRegistry *recordRuleRegistry
	tableName     string
	fields        *FieldsCollection
	methods       *MethodsCollection
	mixins        []*Model
}

// getRelatedModelInfo returns the Model of the related model when
// following path.
// - If skipLast is true, getRelatedModelInfo does not follow the last part of the path
// - If the last part of path is a non relational field, it is simply ignored, whatever
// the value of skipLast.
//
// Paths can be formed from field names or JSON names.
func (m *Model) getRelatedModelInfo(path string, skipLast ...bool) *Model {
	if path == "" {
		return m
	}
	var skip bool
	if len(skipLast) > 0 {
		skip = skipLast[0]
	}

	exprs := strings.Split(path, ExprSep)
	jsonizeExpr(m, exprs)
	fi := m.fields.MustGet(exprs[0])
	if fi.relatedModel == nil || (len(exprs) == 1 && skip) {
		// The field is a non relational field, so we are already
		// on the related Model. Or we have only 1 exprs and we skip the last one.
		return m
	}
	if len(exprs) > 1 {
		return fi.relatedModel.getRelatedModelInfo(strings.Join(exprs[1:], ExprSep), skipLast...)
	}
	return fi.relatedModel
}

// getRelatedFieldIfo returns the Field of the related field when
// following path. Path can be formed from field names or JSON names.
func (m *Model) getRelatedFieldInfo(path string) *Field {
	colExprs := strings.Split(path, ExprSep)
	var rmi *Model
	num := len(colExprs)
	if len(colExprs) > 1 {
		rmi = m.getRelatedModelInfo(path, true)
	} else {
		rmi = m
	}
	fi := rmi.fields.MustGet(colExprs[num-1])
	return fi
}

// scanToFieldMap scans the db query result r into the given FieldMap.
// Unlike slqx.MapScan, the returned interface{} values are of the type
// of the Model fields instead of the database types.
func (m *Model) scanToFieldMap(r sqlx.ColScanner, dest *FieldMap) error {
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	// Step 1: We create a []interface{} which is in fact a []*interface{}
	// and we scan our DB row into it. This enables us to get null values
	// without panic, since null values will map to nil.
	dbValues := make([]interface{}, len(columns))
	for i := range dbValues {
		dbValues[i] = new(interface{})
	}

	err = r.Scan(dbValues...)
	if err != nil {
		return err
	}

	// Step 2: We populate our FieldMap with these values
	for i, dbValue := range dbValues {
		colName := strings.Replace(columns[i], sqlSep, ExprSep, -1)
		dbVal := reflect.ValueOf(dbValue).Elem().Interface()
		(*dest)[colName] = dbVal
	}

	// Step 3: We convert values with the type of the corresponding Field
	// if the value is not nil.
	m.convertValuesToFieldType(dest)
	return r.Err()
}

// convertValuesToFieldType converts all values of the given FieldMap to
// their type in the Model.
func (m *Model) convertValuesToFieldType(fMap *FieldMap) {
	destVals := reflect.ValueOf(fMap).Elem()
	for colName, fMapValue := range *fMap {
		if val, ok := fMapValue.(bool); ok && !val {
			// Hack to manage client returning false instead of nil
			fMapValue = nil
		}
		fi := m.getRelatedFieldInfo(colName)
		fType := fi.structField.Type
		if fType == reflect.TypeOf(fMapValue) {
			// If we already have the good type, don't do anything
			continue
		}
		var val reflect.Value
		switch {
		case fMapValue == nil:
			// dbValue is null, we put the type zero value instead
			val = reflect.Zero(fType)
		case reflect.PtrTo(fType).Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()):
			// the type implements sql.Scanner, so we call Scan
			val = reflect.New(fType)
			scanFunc := val.MethodByName("Scan")
			inArgs := []reflect.Value{reflect.ValueOf(fMapValue)}
			scanFunc.Call(inArgs)
		default:
			rVal := reflect.ValueOf(fMapValue)
			if rVal.Type().Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
				// Our field is a related field
				ids := fMapValue.(RecordSet).Ids()
				if fType == reflect.TypeOf(int64(0)) {
					if len(ids) > 0 {
						val = reflect.ValueOf(ids[0])
					} else {
						val = reflect.ValueOf(nil)
					}
				} else if fType == reflect.TypeOf([]int64{}) {
					val = reflect.ValueOf(ids)
				} else {
					log.Panic("Non consistent type", "model", m.name, "field", colName, "type", fType, "value", fMapValue)
				}
			} else {
				val = reflect.ValueOf(fMapValue)
				if val.IsValid() {
					if typ := val.Type(); typ.ConvertibleTo(fType) {
						val = val.Convert(fType)
					}
				}
			}
		}
		destVals.SetMapIndex(reflect.ValueOf(colName), val)
	}
}

// isMixin returns true if this is a mixin model.
func (m *Model) isMixin() bool {
	if m.options&MixinModel > 0 {
		return true
	}
	return false
}

// isManual returns true if this is a manual model.
func (m *Model) isManual() bool {
	if m.options&ManualModel > 0 {
		return true
	}
	return false
}

// isSystem returns true if this is a system model.
func (m *Model) isSystem() bool {
	if m.options&SystemModel > 0 {
		return true
	}
	return false
}

// isSystem returns true if this is a n M2M Link model.
func (m *Model) isM2MLink() bool {
	if m.options&Many2ManyLinkModel > 0 {
		return true
	}
	return false
}

// Fields returns the fields collection of this model
func (m *Model) Fields() *FieldsCollection {
	return m.fields
}

// Methods returns the methods collection of this model
func (m *Model) Methods() *MethodsCollection {
	return m.methods
}

// NewModel creates a new model with the given name and
// extends it with the given struct pointer.
func NewModel(name string) *Model {
	model := createModel(name, Option(0))
	model.InheritModel(Registry.MustGet("ModelMixin"))
	return model
}

// NewMixinModel creates a new mixin model with the given name and
// extends it with the given struct pointer.
func NewMixinModel(name string) *Model {
	model := createModel(name, MixinModel)
	return model
}

// NewTransientModel creates a new mixin model with the given name and
// extends it with the given struct pointers.
func NewTransientModel(name string) *Model {
	model := createModel(name, TransientModel)
	model.InheritModel(Registry.MustGet("BaseMixin"))
	return model
}

// NewManualModel creates a model whose table is not automatically generated
// in the database. This is particularly useful for SQL view models.
func NewManualModel(name string) *Model {
	model := createModel(name, ManualModel)
	model.InheritModel(Registry.MustGet("CommonMixin"))
	return model
}

// InheritModel extends this Model by importing all fields and methods of mixInModel.
// MixIn methods and fields have a lower priority than those of the model and are
// overridden by the them when applicable.
func (m *Model) InheritModel(mixInModel *Model) {
	m.mixins = append(m.mixins, mixInModel)
}

// createModel creates and populates a new Model with the given name
// by parsing the given struct pointer.
func createModel(name string, options Option) *Model {
	mi := &Model{
		name:          name,
		options:       options,
		acl:           security.NewAccessControlList(),
		rulesRegistry: newRecordRuleRegistry(),
		tableName:     tools.SnakeCaseString(name),
		fields:        newFieldsCollection(),
		methods:       newMethodsCollection(),
	}
	pk := &Field{
		name:      "ID",
		json:      "id",
		acl:       security.NewAccessControlList(),
		model:     mi,
		required:  true,
		noCopy:    true,
		fieldType: fieldtype.Integer,
		structField: reflect.TypeOf(
			struct {
				ID int64
			}{},
		).Field(0),
	}
	mi.fields.add(pk)
	Registry.add(mi)
	return mi
}

// JSONizeFieldName returns the json name of the given fieldName
// If fieldName is already the json name, returns it without modifying it.
// fieldName may be a dot separated path from this model.
// It panics if the path is invalid.
func (m *Model) JSONizeFieldName(fieldName string) string {
	return jsonizePath(m, string(fieldName))
}

// Field starts a condition on this model
func (m *Model) Field(name string) *ConditionField {
	newExprs := strings.Split(name, ExprSep)
	cp := ConditionField{}
	cp.exprs = append(cp.exprs, newExprs...)
	return &cp
}

// FilteredOn adds a condition with a table join on the given field and
// filters the result with the given condition
func (m *Model) FilteredOn(field string, condition *Condition) *Condition {
	res := Condition{predicates: make([]predicate, len(condition.predicates))}
	i := 0
	for _, p := range condition.predicates {
		p.exprs = append([]string{field}, p.exprs...)
		res.predicates[i] = p
		i++
	}
	return &res
}

// Create creates a new record in this model with the given data.
func (m *Model) Create(env Environment, data interface{}) RecordCollection {
	return env.Pool(m.name).Call("Create", data).(RecordSet).Collection()
}

// Search searches the database and returns records matching the given condition.
func (m *Model) Search(env Environment, cond *Condition) RecordCollection {
	return env.Pool(m.name).Call("Search", cond).(RecordSet).Collection()
}

// A Sequence holds the metadata of a DB sequence
type Sequence struct {
	Name string
	JSON string
}

// NewSequence creates a new Sequence and returns a pointer to it
func NewSequence(name string) *Sequence {
	json := fmt.Sprintf("%s_manseq", tools.SnakeCaseString(name))
	seq := &Sequence{
		Name: name,
		JSON: json,
	}
	Registry.sequences[name] = seq
	return seq
}

// NextValue returns the next value of this Sequence
func (s *Sequence) NextValue() int64 {
	adapter := adapters[db.DriverName()]
	return adapter.nextSequenceValue(s.JSON)
}
