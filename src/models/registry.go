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
	"time"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/strutils"
	"github.com/hexya-erp/hexya/src/tools/typesutils"
	"github.com/jmoiron/sqlx"
)

// transientModelTimeout is the timeout after which transient model
// records can be removed from the database
var transientModelTimeout = 30 * time.Minute

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
		jsonBoot := strutils.SnakeCase(nameOrJSON) + "_bootseq"
		s, ok = mc.sequences[jsonBoot]
		if !ok {
			jsonMan := strutils.SnakeCase(nameOrJSON) + "_manseq"
			s, ok = mc.sequences[jsonMan]
		}
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

// add the given Model to the modelCollection
func (mc *modelCollection) addSequence(s *Sequence) {
	if _, exists := mc.GetSequence(s.JSON); exists {
		log.Panic("Trying to add already existing sequence", "sequence", s.JSON)
	}
	mc.Lock()
	defer mc.Unlock()
	mc.sequences[s.JSON] = s
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
	name            string
	options         Option
	rulesRegistry   *recordRuleRegistry
	tableName       string
	fields          *FieldsCollection
	methods         *MethodsCollection
	mixins          []*Model
	sqlConstraints  map[string]sqlConstraint
	sqlErrors       map[string]string
	defaultOrderStr []string
	defaultOrder    []orderPredicate
}

// An sqlConstraint holds the data needed to create a table constraint in the database
type sqlConstraint struct {
	name        string
	sql         string
	errorString string
}

// Name returns the name of this model
func (m *Model) Name() string {
	return m.name
}

// getRelatedModelInfo returns the Model of the related model when
// following path.
// - If skipLast is true, getRelatedModelInfo does not follow the last part of the path
// - If the last part of path is a non relational field, it is simply ignored, whatever
// the value of skipLast.
func (m *Model) getRelatedModelInfo(path FieldName, skipLast ...bool) *Model {
	if path == nil {
		return m
	}
	var skip bool
	if len(skipLast) > 0 {
		skip = skipLast[0]
	}

	exprs := splitFieldNames(path, ExprSep)
	fi := m.fields.MustGet(exprs[0].JSON())
	if fi.relatedModel == nil || (len(exprs) == 1 && skip) {
		// The field is a non relational field, so we are already
		// on the related Model. Or we have only 1 exprs and we skip the last one.
		return m
	}
	if len(exprs) > 1 {
		return fi.relatedModel.getRelatedModelInfo(joinFieldNames(exprs[1:], ExprSep), skipLast...)
	}
	return fi.relatedModel
}

// getRelatedFieldIfo returns the Field of the related field when
// following path. Path can be formed from field names or JSON names.
func (m *Model) getRelatedFieldInfo(path FieldName) *Field {
	colExprs := splitFieldNames(path, ExprSep)
	var rmi *Model
	num := len(colExprs)
	if len(colExprs) > 1 {
		rmi = m.getRelatedModelInfo(path, true)
	} else {
		rmi = m
	}
	fi := rmi.fields.MustGet(colExprs[num-1].JSON())
	return fi
}

// scanToFieldMap scans the db query result r into the given FieldMap.
// Unlike slqx.MapScan, the returned interface{} values are of the type
// of the Model fields instead of the database types.
//
// substs is a map for substituting field names in the ColScanner if necessary (typically if length is over 64 chars).
// Keys are the alias used in the query, and values are '__' separated paths such as "user_id__profile_id__age"
func (m *Model) scanToFieldMap(r sqlx.ColScanner, dest *FieldMap, substs map[string]string) error {
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
		colName := columns[i]
		if s, ok := substs[colName]; ok {
			colName = s
		}
		colName = strings.Replace(colName, sqlSep, ExprSep, -1)
		dbVal := reflect.ValueOf(dbValue).Elem().Interface()
		(*dest)[colName] = dbVal
	}

	// Step 3: We convert values with the type of the corresponding Field
	// if the value is not nil.
	m.convertValuesToFieldType(dest, false)
	return r.Err()
}

// convertValuesToFieldType converts all values of the given FieldMap to
// their type in the Model.
//
// If this method is used to convert values before writing to DB, you
// should set writeDB to true.
func (m *Model) convertValuesToFieldType(fMap *FieldMap, writeDB bool) {
	destVals := reflect.ValueOf(fMap).Elem()
	for colName, fMapValue := range *fMap {
		if val, ok := fMapValue.(bool); ok && !val {
			// Hack to manage client returning false instead of nil
			fMapValue = nil
		}
		fi := m.getRelatedFieldInfo(m.FieldName(colName))
		fType := fi.structField.Type
		typedValue := reflect.New(fType).Interface()
		err := typesutils.Convert(fMapValue, typedValue, fi.isRelationField())
		if err != nil {
			log.Panic(err.Error(), "model", m.name, "field", colName, "type", fType, "value", fMapValue)
		}
		destVals.SetMapIndex(reflect.ValueOf(colName), reflect.ValueOf(typedValue).Elem())
	}
	if writeDB {
		// Change zero values to NULL if writing to DB when applicable
		for colName, fMapValue := range *fMap {
			fi := m.getRelatedFieldInfo(m.FieldName(colName))
			val := reflect.ValueOf(fMapValue)
			switch {
			case fi.fieldType.IsFKRelationType() && val.Kind() == reflect.Int64 && val.Int() == 0:
				val = reflect.ValueOf((*interface{})(nil))
				destVals.SetMapIndex(reflect.ValueOf(colName), val)
			}
		}
	}
}

// AddFields adds the given fields to the model.
func (m *Model) AddFields(fields map[string]FieldDefinition) {
	for name, field := range fields {
		newField := field.DeclareField(m.fields, name)
		if _, exists := m.fields.Get(name); exists {
			log.Panic("models.Field already exists", "model", m.name, "field", name)
		}
		m.fields.add(newField)
	}
}

// IsMixin returns true if this is a mixin model.
func (m *Model) IsMixin() bool {
	if m.options&MixinModel > 0 {
		return true
	}
	return false
}

// IsManual returns true if this is a manual model.
func (m *Model) IsManual() bool {
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

// IsM2MLink returns true if this is an M2M Link model.
func (m *Model) IsM2MLink() bool {
	if m.options&Many2ManyLinkModel > 0 {
		return true
	}
	return false
}

// IsTransient returns true if this Model is transient
func (m *Model) IsTransient() bool {
	return m.options == TransientModel
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
	m.defaultOrderStr = orders
}

// ordersFromStrings returns the given order by exprs as a slice of order structs
func (m *Model) ordersFromStrings(exprs []string) []orderPredicate {
	res := make([]orderPredicate, len(exprs))
	for i, o := range exprs {
		toks := strings.Split(o, " ")
		var desc bool
		if len(toks) > 1 && strings.ToLower(toks[1]) == "desc" {
			desc = true
		}
		res[i] = orderPredicate{field: m.FieldName(toks[0]), desc: desc}
	}
	return res
}

// JSONizeFieldName returns the json name of the given fieldName
// If fieldName is already the json name, returns it without modifying it.
// fieldName may be a dot separated path from this model.
// It panics if the path is invalid.
func (m *Model) JSONizeFieldName(fieldName string) string {
	return jsonizePath(m, fieldName)
}

// FieldName returns a FieldName for the field with the given name.
// name may be a dot separated path from this model.
// It returns nil if the name is empty and panics if the path is invalid.
func (m *Model) FieldName(name string) FieldName {
	if name == "" {
		return nil
	}
	jsonName := jsonizePath(m, name)
	return fieldName{name: name, json: jsonName}
}

// Field starts a condition on this model
func (m *Model) Field(name FieldName) *ConditionField {
	newExprs := splitFieldNames(name, ExprSep)
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
func (m *Model) FieldsGet(fields ...FieldName) map[string]*FieldInfo {
	if len(fields) == 0 {
		for n := range m.fields.registryByName {
			fields = append(fields, m.FieldName(n))
		}
	}
	res := make(map[string]*FieldInfo)
	for _, f := range fields {
		fInfo := m.fields.MustGet(f.Name())
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
			Name:          fInfo.name,
			JSON:          fInfo.json,
			Help:          fInfo.help,
			Searchable:    true,
			Depends:       fInfo.depends,
			Sortable:      true,
			Type:          fInfo.fieldType,
			Store:         fInfo.isSettable(),
			String:        fInfo.description,
			Relation:      relation,
			Selection:     fInfo.selection,
			Domain:        filter,
			ReverseFK:     fInfo.jsonReverseFK,
			OnChange:      fInfo.onChange != "",
			Translate:     translate,
			InvisibleFunc: fInfo.invisibleFunc,
			ReadOnly:      fInfo.isReadOnly(),
			ReadOnlyFunc:  fInfo.readOnlyFunc,
			Required:      fInfo.required,
			RequiredFunc:  fInfo.requiredFunc,
			GoType:        fInfo.structField.Type,
		}
	}
	return res
}

// FilteredOn adds a condition with a table join on the given field and
// filters the result with the given condition
func (m *Model) FilteredOn(field FieldName, condition *Condition) *Condition {
	res := Condition{predicates: make([]predicate, len(condition.predicates))}
	i := 0
	for _, p := range condition.predicates {
		p.exprs = append([]FieldName{field}, p.exprs...)
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

// BrowseOne returns a new RecordSet with the record with the given id.
// Note that this function is just a shorcut for Search the given id.
func (m *Model) BrowseOne(env Environment, id int64) *RecordCollection {
	return env.Pool(m.name).Call("BrowseOne", id).(RecordSet).Collection()
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

// TableName return the db table name
func (m *Model) TableName() string {
	return m.tableName
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
		name:            name,
		options:         options,
		rulesRegistry:   newRecordRuleRegistry(),
		tableName:       strutils.SnakeCase(name),
		fields:          newFieldsCollection(),
		methods:         newMethodsCollection(),
		sqlConstraints:  make(map[string]sqlConstraint),
		sqlErrors:       make(map[string]string),
		defaultOrderStr: []string{"ID"},
	}
	pk := &Field{
		name:      "ID",
		json:      "id",
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
	JSON      string
	Increment int64
	Start     int64
	boot      bool
}

// CreateSequence creates a new Sequence in the database and returns a pointer to it
func CreateSequence(name string, increment, start int64) *Sequence {
	var boot bool
	suffix := "manseq"
	if !Registry.bootstrapped {
		boot = true
		suffix = "bootseq"
	}
	json := fmt.Sprintf("%s_%s", strutils.SnakeCase(name), suffix)
	seq := &Sequence{
		JSON:      json,
		Increment: increment,
		Start:     start,
		boot:      boot,
	}
	if !boot {
		// Create the sequence on the fly if we already bootstrapped.
		// Otherwise, this will be done in Bootstrap
		adapters[db.DriverName()].createSequence(seq.JSON, seq.Increment, seq.Start)
	}
	Registry.addSequence(seq)
	return seq
}

// Drop this sequence and removes it from the database
func (s *Sequence) Drop() {
	Registry.Lock()
	defer Registry.Unlock()
	delete(Registry.sequences, s.JSON)
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

// FreeTransientModels remove transient models records from database which are
// older than the given timeout.
func FreeTransientModels() {
	for _, model := range Registry.registryByName {
		if model.IsTransient() {
			ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				createDate := model.FieldName("CreateDate")
				model.Search(env, model.Field(createDate).Lower(dates.Now().Add(-transientModelTimeout))).Call("Unlink")
			})
		}
	}
}
