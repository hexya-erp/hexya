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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/tools/nbutils"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
	"github.com/jmoiron/sqlx"
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

// MustGetSequence gets the given sequence by name or by db name.
// It panics if the Sequence does not exist
func (mc *modelCollection) MustGetSequence(nameOrJSON string) *Sequence {
	s, ok := mc.GetSequence(nameOrJSON)
	if !ok {
		log.Panic("Unknown sequence", "sequence", nameOrJSON)
	}
	return s
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
	name           string
	options        Option
	acl            *security.AccessControlList
	rulesRegistry  *recordRuleRegistry
	tableName      string
	fields         *FieldsCollection
	methods        *MethodsCollection
	mixins         []*Model
	sqlConstraints map[string]sqlConstraint
	sqlErrors      map[string]string
	defaultOrder   []string
}

// An sqlConstraint holds the data needed to create a table constraint in the database
type sqlConstraint struct {
	name        string
	sql         string
	errorString string
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
	// and we scan our DB row into it. This enables us to Get null values
	// without panic, since null values will map to nil.
	dbValues := make([]interface{}, len(columns))
	for i := range dbValues {
		dbValues[i] = new(interface{})
	}

	err = r.Scan(dbValues...)
	if err != nil {
		return err
	}

	// Step 2: We populate our dest FieldMap with these values
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
			// except if we have a nullable FK relation field
			if fi.fieldType.IsFKRelationType() && !fi.required {
				val = reflect.ValueOf((*interface{})(nil))
			} else {
				val = reflect.Zero(fType)
			}
		case reflect.PtrTo(fType).Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()):
			// the type implements sql.Scanner, so we call Scan
			valPtr := reflect.New(fType)
			scanFunc := valPtr.MethodByName("Scan")
			inArgs := []reflect.Value{reflect.ValueOf(fMapValue)}
			res := scanFunc.Call(inArgs)
			if res[0].Interface() != nil {
				log.Panic("Unable to scan into target Type", "error", res[0].Interface())
			}
			val = valPtr.Elem()
		default:
			var err error
			if fi.isRelationField() {
				val, err = getRelationFieldValue(fMapValue, fType)
			} else {
				val, err = getSimpleTypeValue(fMapValue, fType)
			}
			if err != nil {
				log.Panic(err.Error(), "model", m.name, "field", colName, "type", fType, "value", fMapValue)
			}
		}
		destVals.SetMapIndex(reflect.ValueOf(colName), val)
	}
}

// getSimpleTypeValue returns value as a reflect.Value with type of targetType
// It returns an error if the value cannot be converted to the target type
func getSimpleTypeValue(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	val := reflect.ValueOf(value)
	if val.IsValid() {
		typ := val.Type()
		switch {
		case typ.ConvertibleTo(targetType):
			val = val.Convert(targetType)
		case targetType.Kind() == reflect.Bool:
			val = reflect.ValueOf(!reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface()))
		case typ == reflect.TypeOf([]byte{}) && targetType.Kind() == reflect.Float32:
			// backend may return floats as []byte when stored as numeric
			fval, err := strconv.ParseFloat(string(value.([]byte)), 32)
			if err != nil {
				return reflect.Value{}, err
			}
			val = reflect.ValueOf(float32(fval))
		case typ == reflect.TypeOf([]byte{}) && targetType.Kind() == reflect.Float64:
			// backend may return floats as []byte when stored as numeric
			fval, err := strconv.ParseFloat(string(value.([]byte)), 64)
			if err != nil {
				return reflect.Value{}, err
			}
			val = reflect.ValueOf(fval)
		}
	}
	return val, nil
}

// getRelationFieldValue returns value as a reflect.Value with type of targetType
// It returns an error if the value is not consistent with a relation field value
// (i.e. is not of type RecordSet or int64 or []int64)
func getRelationFieldValue(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	var (
		val reflect.Value
		err error
	)
	switch tValue := value.(type) {
	case RecordSet:
		ids := tValue.Ids()
		if targetType == reflect.TypeOf(int64(0)) {
			if len(ids) > 0 {
				val = reflect.ValueOf(ids[0])
			} else {
				val = reflect.ValueOf((*interface{})(nil))
			}
		} else if targetType == reflect.TypeOf([]int64{}) {
			val = reflect.ValueOf(ids)
		} else {
			err = errors.New("non consistent type")
		}
	case []interface{}:
		if len(tValue) == 0 {
			val = reflect.ValueOf((*interface{})(nil))
			break
		}
		err = errors.New("non empty []interface{} given")
	case []int64, *interface{}:
		val = reflect.ValueOf(value)
	default:
		nbValue, nbErr := nbutils.CastToInteger(tValue)
		val = reflect.ValueOf(nbValue)
		err = nbErr
	}
	return val, err
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

// isContext returns true if this is a context model.
func (m *Model) isContext() bool {
	if m.options&ContextsModel > 0 {
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

// hasParentField returns true if this model is recursive and has a Parent field.
func (m *Model) hasParentField() bool {
	_, parentExists := m.fields.Get("Parent")
	return parentExists
}

// Fields returns the fields collection of this model
func (m *Model) Fields() *FieldsCollection {
	return m.fields
}

// Methods returns the methods collection of this model
func (m *Model) Methods() *MethodsCollection {
	return m.methods
}

// SetDefaultOrder sets the default order used by this model
// when no OrderBy() is specified in a query. When unspecified,
// default order is 'id asc'.
//
// Give the order fields in separate strings, such as
// model.SetDefaultOrder("Name desc", "date asc", "id")
func (m *Model) SetDefaultOrder(orders ...string) {
	m.defaultOrder = orders
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

// FieldsGet returns the definition of each field.
// The embedded fields are included.
//
// If no fields are given, then all fields are returned.
//
// The result map is indexed by the fields JSON names.
func (m *Model) FieldsGet(fields ...FieldNamer) map[string]*FieldInfo {
	if len(fields) == 0 {
		for jName := range m.fields.registryByJSON {
			fields = append(fields, FieldName(jName))
		}
	}
	res := make(map[string]*FieldInfo)
	for _, f := range fields {
		fInfo := m.fields.MustGet(f.String())
		var relation string
		if fInfo.relatedModel != nil {
			relation = fInfo.relatedModel.name
		}
		var filter interface{}
		if fInfo.filter != nil {
			filter = fInfo.filter.Serialize()
		}
		_, translate := fInfo.contexts["lang"]
		res[fInfo.json] = &FieldInfo{
			Name:       fInfo.name,
			JSON:       fInfo.json,
			Help:       fInfo.help,
			Searchable: true,
			Depends:    fInfo.depends,
			Sortable:   true,
			Type:       fInfo.fieldType,
			Store:      fInfo.isSettable(),
			String:     fInfo.description,
			Relation:   relation,
			Required:   fInfo.required,
			Selection:  fInfo.selection,
			Domain:     filter,
			ReadOnly:   fInfo.isReadOnly(),
			ReverseFK:  fInfo.jsonReverseFK,
			OnChange:   fInfo.onChange != "",
			Translate:  translate,
		}
	}
	return res
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
func (m *Model) Create(env Environment, data interface{}) *RecordCollection {
	return env.Pool(m.name).Call("Create", data).(RecordSet).Collection()
}

// Search searches the database and returns records matching the given condition.
func (m *Model) Search(env Environment, cond Conditioner) *RecordCollection {
	return env.Pool(m.name).Call("Search", cond).(RecordSet).Collection()
}

// Browse returns a new RecordSet with the records with the given ids.
// Note that this function is just a shorcut for Search on a list of ids.
func (m *Model) Browse(env Environment, ids []int64) *RecordCollection {
	return env.Pool(m.name).Call("Browse", ids).(RecordSet).Collection()
}

// AddSQLConstraint adds a table constraint in the database.
//    - name is an arbitrary name to reference this constraint. It will be appended by
//      the table name in the database, so there is only need to ensure that it is unique
//      in this model.
//    - sql is constraint definition to pass to the database.
//    - errorString is the text to display to the user when the constraint is violated
func (m *Model) AddSQLConstraint(name, sql, errorString string) {
	constraintName := fmt.Sprintf("%s_%s_mancon", name, m.tableName)
	m.sqlConstraints[constraintName] = sqlConstraint{
		name:        constraintName,
		sql:         sql,
		errorString: errorString,
	}
}

// RemoveSQLConstraint removes the sql constraint with the given name from the database.
func (m *Model) RemoveSQLConstraint(name string) {
	delete(m.sqlConstraints, fmt.Sprintf("%s_mancon", name))
}

// Underlying returns the underlying Model data object, i.e. itself
func (m *Model) Underlying() *Model {
	return m
}

var _ Modeler = new(Model)

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
func (m *Model) InheritModel(mixInModel Modeler) {
	m.mixins = append(m.mixins, mixInModel.Underlying())
}

// createModel creates and populates a new Model with the given name
// by parsing the given struct pointer.
func createModel(name string, options Option) *Model {
	mi := &Model{
		name:           name,
		options:        options,
		acl:            security.NewAccessControlList(),
		rulesRegistry:  newRecordRuleRegistry(),
		tableName:      strutils.SnakeCase(name),
		fields:         newFieldsCollection(),
		methods:        newMethodsCollection(),
		sqlConstraints: make(map[string]sqlConstraint),
		sqlErrors:      make(map[string]string),
		defaultOrder:   []string{"id"},
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

// A Sequence holds the metadata of a DB sequence
//
// There are two types of sequences: those created before bootstrap
// and those created after. The former will be created and updated at
// bootstrap and cannot be modified afterwards. The latter will be
// created, updated or dropped immediately.
type Sequence struct {
	Name      string
	JSON      string
	Increment int64
	Start     int64
	boot      bool
}

// CreateSequence creates a new Sequence in the database and returns a pointer to it
func CreateSequence(name string, increment, start int64) *Sequence {
	var boot bool
	if !Registry.bootstrapped {
		boot = true
	}
	json := fmt.Sprintf("%s_manseq", strutils.SnakeCase(name))
	seq := &Sequence{
		Name:      name,
		JSON:      json,
		Increment: increment,
		Start:     start,
		boot:      boot,
	}
	Registry.Lock()
	defer Registry.Unlock()
	Registry.sequences[name] = seq
	if !boot {
		// Create the sequence on the fly if we already bootstrapped.
		// Otherwise, this will be done in Bootstrap
		adapters[db.DriverName()].createSequence(seq.JSON, seq.Increment, seq.Start)
	}
	return seq
}

// DropSequence drops this sequence and removes it from the database
func (s *Sequence) Drop() {
	Registry.Lock()
	defer Registry.Unlock()
	delete(Registry.sequences, s.Name)
	if Registry.bootstrapped {
		// Drop the sequence on the fly if we already bootstrapped.
		// Otherwise, this will be done in Bootstrap
		if s.boot {
			log.Panic("Boot Sequences cannot be dropped after bootstrap")
		}
		adapters[db.DriverName()].dropSequence(s.JSON)
	}
}

// Alter alters this sequence by changing next number and/or increment.
// Set a parameter to 0 to leave it unchanged.
func (s *Sequence) Alter(increment, restart int64) {
	var boot bool
	if !Registry.bootstrapped {
		boot = true
	}
	if s.boot && !boot {
		log.Panic("Boot Sequences cannot be modified after bootstrap")
	}
	if restart > 0 {
		s.Start = restart
	}
	if increment > 0 {
		s.Increment = increment
	}
	if !boot {
		adapters[db.DriverName()].alterSequence(s.JSON, increment, restart)
	}
}

// NextValue returns the next value of this Sequence
func (s *Sequence) NextValue() int64 {
	adapter := adapters[db.DriverName()]
	return adapter.nextSequenceValue(s.JSON)
}
