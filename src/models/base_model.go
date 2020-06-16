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

	"github.com/google/uuid"
	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
)

const (
	// TransientModel means that the records of this model will be automatically
	// removed periodically. Transient models are mainly used for wizards.
	TransientModel Option = 1 << iota
	// MixinModel means that this model will not be accessible like a regular model
	// but is meant to be mixed in other models.
	MixinModel
	// Many2ManyLinkModel is a model that abstracts the link
	// table of a many2many relationship
	Many2ManyLinkModel
	// ContextsModel is a model for holding fields values that depend on contexts
	ContextsModel
	// ManualModel is a model whose table is not automatically generated in the
	// database. Such models include SQL views and materialized SQL views.
	ManualModel
	// SystemModel is a model that is used internally by the Hexya Framework
	SystemModel
)

//  declareCommonMixin creates the common mixin that is needed for all models
func declareCommonMixin() {
	commonMixin := NewMixinModel("CommonMixin")
	commonMixin.addMethod("New", commonMixinNew)
	commonMixin.addMethod("Create", commonMixinCreate)
	commonMixin.addMethod("Read", commonMixinRead)
	commonMixin.addMethod("Load", commonMixinLoad)
	commonMixin.addMethod("Write", commonMixinWrite)
	commonMixin.addMethod("Unlink", commonMixinUnlink)
	commonMixin.addMethod("CopyData", commonMixinCopyData)
	commonMixin.addMethod("Copy", commonMixinCopy)
	commonMixin.addMethod("NameGet", commonMixinNameGet)
	commonMixin.addMethod("SearchByName", commonMixinSearchByName)
	commonMixin.addMethod("FieldsGet", commonMixinFieldsGet)
	commonMixin.addMethod("FieldGet", commonMixinFieldGet)
	commonMixin.addMethod("DefaultGet", commonMixinDefaultGet)
	commonMixin.addMethod("CheckRecursion", commonMixinCheckRecursion)
	commonMixin.addMethod("Onchange", commonMixinOnChange)
	commonMixin.addMethod("Search", commonMixinSearch)
	commonMixin.addMethod("Browse", commonMixinBrowse)
	commonMixin.addMethod("BrowseOne", commonMixinBrowseOne)
	commonMixin.addMethod("SearchCount", commonMixinSearchCount)
	commonMixin.addMethod("Fetch", commonMixinFetch)
	commonMixin.addMethod("SearchAll", commonMixinSearchAll)
	commonMixin.addMethod("GroupBy", commonMixinGroupBy)
	commonMixin.addMethod("Limit", commonMixinLimit)
	commonMixin.addMethod("Offset", commonMixinOffset)
	commonMixin.addMethod("OrderBy", commonMixinOrderBy)
	commonMixin.addMethod("Union", commonMixinUnion)
	commonMixin.addMethod("Subtract", commonMixinSubtract)
	commonMixin.addMethod("Intersect", commonMixinIntersect)
	commonMixin.addMethod("CartesianProduct", commonMixinCartesianProduct)
	commonMixin.addMethod("Equals", commonMixinEquals)
	commonMixin.addMethod("Sorted", commonMixinSorted)
	commonMixin.addMethod("SortedDefault", commonMixinSortedDefault)
	commonMixin.addMethod("SortedByField", commonMixinSortedByField)
	commonMixin.addMethod("Filtered", commonMixinFiltered)
	commonMixin.addMethod("GetRecord", commonMixinGetRecord)
	commonMixin.addMethod("CheckExecutionPermission", commonMixinCheckExecutionPermission)
	commonMixin.addMethod("SQLFromCondition", commonMixinSQLFromCondition)
	commonMixin.addMethod("WithEnv", commonMixinWithEnv)
	commonMixin.addMethod("WithContext", commonMixinWithContext)
	commonMixin.addMethod("WithNewContext", commonMixinWithNewContext)
	commonMixin.addMethod("Sudo", commonMixinSudo)
}

// New creates a memory only record from the given data.
// Such a record has a negative ID and cannot be loaded from database.
//
// Note that New does not work with embedded records.
func commonMixinNew(rc *RecordCollection, data RecordData) *RecordCollection {
	return rc.new(data)
}

// Create inserts a record in the database from the given data.
// Returns the created RecordCollection.
func commonMixinCreate(rc *RecordCollection, data RecordData) *RecordCollection {
	return rc.create(data)
}

// Read reads the database and returns a slice of FieldMap of the given model.
func commonMixinRead(rc *RecordCollection, fields FieldNames) []RecordData {
	var res []RecordData
	// Check if we have id in fields, and add it otherwise
	fields = addIDIfNotPresent(fields)
	// Do the actual reading
	for _, rec := range rc.Records() {
		fData := NewModelData(rc.model)
		for _, fName := range fields {
			fData.Underlying().Set(fName, rec.Get(fName))
		}
		res = append(res, fData)
	}
	return res
}

// Load looks up cache for fields of the RecordCollection and
// query database for missing values.
// fields are the fields to retrieve in the expression format,
// i.e. "User.Profile.Age" or "user_id.profile_id.age".
// If no fields are given, all DB columns of the RecordCollection's
// model are retrieved.
func commonMixinLoad(rc *RecordCollection, fields ...FieldName) *RecordCollection {
	return rc.Load(fields...)
}

// Write is the base implementation of the 'Write' method which updates
// records in the database with the given data.
// Data can be either a struct pointer or a FieldMap.`,
func commonMixinWrite(rc *RecordCollection, data RecordData) bool {
	return rc.update(data)
}

// Unlink deletes the given records in the database.
func commonMixinUnlink(rc *RecordCollection) int64 {
	return rc.unlink()
}

// CopyData copies given record's data with all its fields values.
//
// overrides contains field values to override in the original values of the copied record.
func commonMixinCopyData(rc *RecordCollection, overrides RecordData) *ModelData {
	rc.EnsureOne()
	// Handle case when overrides is nil
	oVal := reflect.ValueOf(overrides)
	if !oVal.IsValid() || (oVal.Kind() != reflect.Struct && oVal.IsNil()) {
		overrides = NewModelDataFromRS(rc)
	}

	// Create the RecordData
	res := NewModelDataFromRS(rc)
	for _, fi := range rc.model.fields.registryByName {
		fName := rc.model.FieldName(fi.name)
		if overrides.Underlying().Has(fName) {
			// Overrides are applied below
			continue
		}
		if fi.noCopy || fi.isComputedField() {
			continue
		}
		switch fi.fieldType {
		case fieldtype.One2One:
			// One2one related records must be copied to avoid duplicate keys on FK
			res = res.Create(fName, rc.Get(fName).(RecordSet).Collection().Call("CopyData", nil).(RecordData).Underlying())
		case fieldtype.One2Many, fieldtype.Rev2One:
			for _, rec := range rc.Get(fName).(RecordSet).Collection().Records() {
				res = res.Create(fName, rec.Call("CopyData", nil).(RecordData).Underlying().Unset(fi.relatedModel.FieldName(fi.reverseFK)))
			}
		default:
			res.Set(fName, rc.Get(fName))
		}
	}
	// Apply overrides
	res.RemovePK()
	res.MergeWith(overrides.Underlying())
	return res
}

// Copy duplicates the given records.
//
// overrides contains field values to override in the original values of the copied record.`,
func commonMixinCopy(rc *RecordCollection, overrides RecordData) *RecordCollection {
	rc.EnsureOne()
	data := rc.Call("CopyData", overrides).(RecordData).Underlying()
	newRs := rc.Call("Create", data).(RecordSet).Collection()
	return newRs
}

// NameGet retrieves the human readable name of this record.`,
func commonMixinNameGet(rc *RecordCollection) string {
	if _, nameExists := rc.model.fields.Get("Name"); nameExists {
		switch name := rc.Get(rc.model.FieldName("Name")).(type) {
		case string:
			return name
		case fmt.Stringer:
			return name.String()
		default:
			log.Panic("Name field is neither a string nor a fmt.Stringer", "model", rc.model)
		}
	}
	return rc.String()
}

// SearchByName searches for records that have a display name matching the given
// "name" pattern when compared with the given "op" operator, while also
// matching the optional search condition ("additionalCond").
//
// This is used for example to provide suggestions based on a partial
// value for a relational field. Sometimes be seen as the inverse
// function of NameGet but it is not guaranteed to be.
func commonMixinSearchByName(rc *RecordCollection, name string, op operator.Operator, additionalCond Conditioner, limit int) *RecordCollection {
	if op == "" {
		op = operator.IContains
	}
	cond := rc.Model().Field(rc.model.FieldName("Name")).AddOperator(op, name)
	if !additionalCond.Underlying().IsEmpty() {
		cond = cond.AndCond(additionalCond.Underlying())
	}
	return rc.Model().Search(rc.Env(), cond).Limit(limit)
}

// FieldsGet returns the definition of each field.
// The embedded fields are included.
// The string, help, and selection (if present) attributes are translated.
//
// The result map is indexed by the fields JSON names.
func commonMixinFieldsGet(rc *RecordCollection, args FieldsGetArgs) map[string]*FieldInfo {
	// Get the field informations
	res := rc.model.FieldsGet(args.Fields...)

	// Translate attributes when required
	lang := rc.Env().Context().GetString("lang")
	for fName, fInfo := range res {
		res[fName].Help = i18n.Registry.TranslateFieldHelp(lang, rc.model.name, fInfo.Name, fInfo.Help)
		res[fName].String = i18n.Registry.TranslateFieldDescription(lang, rc.model.name, fInfo.Name, fInfo.String)
		res[fName].Selection = i18n.Registry.TranslateFieldSelection(lang, rc.model.name, fInfo.Name, fInfo.Selection)
	}
	return res
}

// FieldGet returns the definition of the given field.
// The string, help, and selection (if present) attributes are translated.
func commonMixinFieldGet(rc *RecordCollection, field FieldName) *FieldInfo {
	args := FieldsGetArgs{
		Fields: []FieldName{field},
	}
	return rc.Call("FieldsGet", args).(map[string]*FieldInfo)[field.JSON()]
}

// DefaultGet returns a Params map with the default values for the model.
func commonMixinDefaultGet(rc *RecordCollection) *ModelData {
	res := rc.getDefaults(rc.Env().Context().GetBool("hexya_ignore_computed_defaults"))
	return res
}

// CheckRecursion verifies that there is no loop in a hierarchical structure of records,
// by following the parent relationship using the 'Parent' field until a loop is detected or
// until a top-level record is found.
//
// It returns true if no loop was found, false otherwise`,
func commonMixinCheckRecursion(rc *RecordCollection) bool {
	if _, exists := rc.model.fields.Get("Parent"); !exists {
		// No Parent field in model, so no loop
		return true
	}
	if rc.hasNegIds {
		// We have a negative id, so we can't have a loop
		return true
	}
	// We use direct SQL query to bypass access control
	query := fmt.Sprintf(`SELECT parent_id FROM %s WHERE id = ?`, adapters[db.DriverName()].quoteTableName(rc.model.tableName))
	rc.Load(rc.model.FieldName("Parent"))
	for _, record := range rc.Records() {
		currentID := record.ids[0]
		for {
			var parentID sql.NullInt64
			rc.env.cr.Get(&parentID, query, currentID)
			if !parentID.Valid {
				break
			}
			currentID = parentID.Int64
			if currentID == record.ids[0] {
				return false
			}
		}
	}
	return true
}

// Onchange returns the values that must be modified according to each field's Onchange
// method in the pseudo-record given as params.Values`,
func commonMixinOnChange(rc *RecordCollection, params OnchangeParams) OnchangeResult {
	var retValues *ModelData
	var warnings []string
	filters := make(map[FieldName]Conditioner)

	err := SimulateInNewEnvironment(rc.Env().Uid(), func(env Environment) {
		values := params.Values.Underlying().FieldMap
		data := NewModelDataFromRS(rc.WithEnv(env), values)
		if rc.IsNotEmpty() {
			data.Set(ID, rc.ids[0])
		}
		retValues = NewModelDataFromRS(rc.WithEnv(env))
		var rs *RecordCollection
		if id, _ := nbutils.CastToInteger(data.Get(ID)); id != 0 {
			rs = rc.WithEnv(env).withIds([]int64{id})
			rs = rs.WithContext("hexya_onchange_origin", rs.First().Wrap())
			rs.WithContext("hexya_force_compute_write", true).update(data)
		} else {
			rs = rc.WithEnv(env).WithContext("hexya_force_compute_write", true).create(data)
		}
		// Set inverse fields
		for field := range values {
			fName := rs.model.FieldName(field)
			fi := rs.model.getRelatedFieldInfo(fName)
			if fi.inverse != "" {
				fVal := data.Get(fName)
				rs.Call(fi.inverse, fVal)
			}
		}
		todo := params.Fields
		done := make(map[string]bool)
		// Apply onchanges or compute
		for len(todo) > 0 {
			field := todo[0]
			todo = todo[1:]
			if done[field.JSON()] {
				continue
			}
			done[field.JSON()] = true
			if params.Onchange[field.Name()] == "" && params.Onchange[field.JSON()] == "" {
				continue
			}
			fi := rs.model.getRelatedFieldInfo(field)
			fnct := fi.onChange
			if fnct == "" {
				fnct = fi.compute
			}
			rrs := rs
			toks := splitFieldNames(field, ExprSep)
			if len(toks) > 1 {
				rrs = rs.Get(joinFieldNames(toks[:len(toks)-1], ExprSep)).(RecordSet).Collection()
			}
			// Values
			if fnct != "" {
				vals := rrs.Call(fnct).(RecordData)
				for _, f := range vals.Underlying().FieldNames() {
					if !done[f.JSON()] {
						todo = append(todo, f)
					}
				}
				rrs.WithContext("hexya_force_compute_write", true).Call("Write", vals)
			}
			// Warning
			if fi.onChangeWarning != "" {
				w := rrs.Call(fi.onChangeWarning).(string)
				if w != "" {
					warnings = append(warnings, w)
				}
			}
			// Filters
			if fi.onChangeFilters != "" {
				ff := rrs.Call(fi.onChangeFilters).(map[FieldName]Conditioner)
				for k, v := range ff {
					filters[k] = v
				}
			}
		}
		// Collect modified values
		for field, val := range values {
			fName := rs.model.FieldName(field)
			if fName.JSON() == "__last_update" {
				continue
			}
			fi := rs.Collection().Model().getRelatedFieldInfo(fName)
			newVal := rs.Get(fName)
			switch {
			case fi.fieldType.IsRelationType():
				v := rs.convertToRecordSet(val, fi.relatedModelName)
				nv := rs.convertToRecordSet(newVal, fi.relatedModelName)
				if !v.Equals(nv) {
					retValues.Set(fName, newVal)
				}
			default:
				if val != newVal {
					retValues.Set(fName, newVal)
				}
			}
		}
	})
	if err != nil {
		panic(err)
	}
	retValues.Unset(ID)
	return OnchangeResult{
		Value:   retValues,
		Warning: strings.Join(warnings, "\n\n"),
		Filters: filters,
	}
}

// Search returns a new RecordSet filtering on the current one with the
// additional given Condition.
func commonMixinSearch(rc *RecordCollection, cond Conditioner) *RecordCollection {
	return rc.Search(cond.Underlying())
}

// Browse returns a new RecordSet with only the records with the given ids.
// Note that this function is just a shorcut for Search on a list of ids.
func commonMixinBrowse(rc *RecordCollection, ids []int64) *RecordCollection {
	return rc.Call("Search", rc.Model().Field(ID).In(ids)).(RecordSet).Collection()
}

// BrowseOne returns a new RecordSet with only the record with the given id.
// Note that this function is just a shorcut for Search on a given id.
func commonMixinBrowseOne(rc *RecordCollection, id int64) *RecordCollection {
	return rc.Call("Search", rc.Model().Field(ID).Equals(id)).(RecordSet).Collection()
}

// SearchCount fetch from the database the number of records that match the RecordSet conditions.
func commonMixinSearchCount(rc *RecordCollection) int {
	return rc.SearchCount()
}

// Fetch query the database with the current filter and returns a RecordSet
// with the queries ids.
//
// Fetch is lazy and only return ids. Use Load() instead if you want to fetch all fields.
func commonMixinFetch(rc *RecordCollection) *RecordCollection {
	return rc.Fetch()
}

// SearchAll returns a RecordSet with all items of the table, regardless of the
// current RecordSet query. It is mainly meant to be used on an empty RecordSet.
func commonMixinSearchAll(rc *RecordCollection) *RecordCollection {
	return rc.SearchAll()
}

// GroupBy returns a new RecordSet grouped with the given GROUP BY expressions.
func commonMixinGroupBy(rc *RecordCollection, exprs ...FieldName) *RecordCollection {
	return rc.GroupBy(exprs...)
}

// Limit returns a new RecordSet with only the first 'limit' records.
func commonMixinLimit(rc *RecordCollection, limit int) *RecordCollection {
	return rc.Limit(limit)
}

// Offset returns a new RecordSet with only the records starting at offset
func commonMixinOffset(rc *RecordCollection, offset int) *RecordCollection {
	return rc.Offset(offset)
}

// OrderBy returns a new RecordSet ordered by the given ORDER BY expressions.
// Each expression contains a field name and optionally one of "asc" or "desc", such as:
//
// rs.OrderBy("Company", "Name desc")
func commonMixinOrderBy(rc *RecordCollection, exprs ...string) *RecordCollection {
	return rc.OrderBy(exprs...)
}

// Union returns a new RecordSet that is the union of this RecordSet and the given
// "other" RecordSet. The result is guaranteed to be a set of unique records.
func commonMixinUnion(rc *RecordCollection, other RecordSet) *RecordCollection {
	return rc.Union(other)
}

// Subtract returns a RecordSet with the Records that are in this
// RecordCollection but not in the given 'other' one.
// The result is guaranteed to be a set of unique records.
func commonMixinSubtract(rc *RecordCollection, other RecordSet) *RecordCollection {
	return rc.Subtract(other)
}

// Intersect returns a new RecordCollection with only the records that are both
// in this RecordCollection and in the other RecordSet.
func commonMixinIntersect(rc *RecordCollection, other RecordSet) *RecordCollection {
	return rc.Intersect(other)
}

// CartesianProduct returns the cartesian product of this RecordCollection with others.
func commonMixinCartesianProduct(rc *RecordCollection, other ...RecordSet) []*RecordCollection {
	return rc.CartesianProduct(other...)
}

// Equals returns true if this RecordSet is the same as other
// i.e. they are of the same model and have the same ids
func commonMixinEquals(rc *RecordCollection, other RecordSet) bool {
	return rc.Equals(other)
}

// Sorted returns a new RecordCollection sorted according to the given less function.
//
// The less function should return true if rs1 < rs2`,
func commonMixinSorted(rc *RecordCollection, less func(rs1 RecordSet, rs2 RecordSet) bool) *RecordCollection {
	return rc.Sorted(less)
}

// SortedDefault returns a new record set with the same records as rc but sorted according
// to the default order of this model
func commonMixinSortedDefault(rc *RecordCollection) *RecordCollection {
	return rc.SortedDefault()
}

// SortedByField returns a new record set with the same records as rc but sorted by the given field.
// If reverse is true, the sort is done in reversed order
func commonMixinSortedByField(rc *RecordCollection, namer FieldName, reverse bool) *RecordCollection {
	return rc.SortedByField(namer, reverse)
}

// Filtered returns a new record set with only the elements of this record set
// for which test is true.
//
// Note that if this record set is not fully loaded, this function will call the database
// to load the fields before doing the filtering. In this case, it might be more efficient
// to search the database directly with the filter condition.
func commonMixinFiltered(rc *RecordCollection, test func(rs RecordSet) bool) *RecordCollection {
	return rc.Filtered(test)
}

// GetRecord returns the Recordset with the given externalID. It panics if the externalID does not exist.
func commonMixinGetRecord(rc *RecordCollection, externalID string) *RecordCollection {
	return rc.GetRecord(externalID)
}

// CheckExecutionPermission panics if the current user is not allowed to execute the given method.
//
// If dontPanic is false, this function will panic, otherwise it returns true
// if the user has the execution permission and false otherwise.
func commonMixinCheckExecutionPermission(rc *RecordCollection, method *Method, dontPanic ...bool) bool {
	return rc.CheckExecutionPermission(method, dontPanic...)
}

// SQLFromCondition returns the WHERE clause sql and arguments corresponding to
// the given condition.`,
func commonMixinSQLFromCondition(rc *RecordCollection, c *Condition) (string, SQLParams) {
	return rc.SQLFromCondition(c)
}

// WithEnv returns a copy of the current RecordSet with the given Environment.
func commonMixinWithEnv(rc *RecordCollection, env Environment) *RecordCollection {
	return rc.WithEnv(env)
}

// WithContext returns a copy of the current RecordSet with
// its context extended by the given key and value.
func commonMixinWithContext(rc *RecordCollection, key string, value interface{}) *RecordCollection {
	return rc.WithContext(key, value)
}

// WithNewContext returns a copy of the current RecordSet with its context
// replaced by the given one.
func commonMixinWithNewContext(rc *RecordCollection, context *types.Context) *RecordCollection {
	return rc.WithNewContext(context)
}

// Sudo returns a new RecordSet with the given userID
// or the superuser ID if not specified
func commonMixinSudo(rc *RecordCollection, userID ...int64) *RecordCollection {
	return rc.Sudo(userID...)
}

// declareBaseMixin creates the mixin that implements all the necessary base methods of a model
func declareBaseMixin() {
	baseMixin := NewMixinModel("BaseMixin")
	baseMixin.InheritModel(Registry.MustGet("CommonMixin"))
	baseMixin.addMethod("ComputeLastUpdate", baseMixinComputeLastUpdate)
	baseMixin.addMethod("ComputeDisplayName", baseMixinComputeDisplayName)
	baseMixin.fields.add(&Field{
		model:       baseMixin,
		name:        "CreateDate",
		description: "Created On",
		json:        "create_date",
		fieldType:   fieldtype.DateTime,
		structField: reflect.StructField{Type: reflect.TypeOf(dates.DateTime{})},
		noCopy:      true,
	})
	baseMixin.fields.add(&Field{
		model:       baseMixin,
		name:        "CreateUID",
		description: "Created By",
		json:        "create_uid",
		fieldType:   fieldtype.Integer,
		structField: reflect.StructField{Type: reflect.TypeOf(int64(0))},
		noCopy:      true,
		defaultFunc: func(env Environment) interface{} {
			return env.uid
		},
	})
	baseMixin.fields.add(&Field{
		model:       baseMixin,
		name:        "WriteDate",
		description: "Updated On",
		json:        "write_date",
		fieldType:   fieldtype.DateTime,
		structField: reflect.StructField{Type: reflect.TypeOf(dates.DateTime{})},
		noCopy:      true,
	})
	baseMixin.fields.add(&Field{
		model:       baseMixin,
		name:        "WriteUID",
		description: "UpdatedBy",
		json:        "write_uid",
		fieldType:   fieldtype.Integer,
		structField: reflect.StructField{Type: reflect.TypeOf(int64(0))},
		noCopy:      true,
		defaultFunc: func(env Environment) interface{} {
			return env.uid
		},
	})
	baseMixin.fields.add(&Field{
		model:       baseMixin,
		name:        "LastUpdate",
		description: "Last Updated On",
		json:        "__last_update",
		fieldType:   fieldtype.DateTime,
		structField: reflect.StructField{Type: reflect.TypeOf(dates.DateTime{})},
		compute:     "ComputeLastUpdate",
		depends:     []string{"WriteDate", "CreateDate"},
	})
	baseMixin.fields.add(&Field{
		model:       baseMixin,
		name:        "DisplayName",
		description: "Display Name",
		json:        "display_name",
		fieldType:   fieldtype.Char,
		structField: reflect.StructField{Type: reflect.TypeOf("")},
		compute:     "ComputeDisplayName",
		depends:     []string{""},
	})
}

// ComputeLastUpdate returns the last datetime at which the record has been updated.
func baseMixinComputeLastUpdate(rc *RecordCollection) *ModelData {
	res := NewModelData(rc.model)
	if !rc.Get(rc.model.FieldName("WriteDate")).(dates.DateTime).IsZero() {
		res.Set(rc.model.FieldName("LastUpdate"), rc.Get(rc.model.FieldName("WriteDate")).(dates.DateTime))
		return res
	}
	if !rc.Get(rc.model.FieldName("CreateDate")).(dates.DateTime).IsZero() {
		res.Set(rc.model.FieldName("LastUpdate"), rc.Get(rc.model.FieldName("CreateDate")).(dates.DateTime))
		return res
	}
	res.Set(rc.model.FieldName("LastUpdate"), dates.Now())
	return res
}

// ComputeDisplayName updates the DisplayName field with the result of NameGet
func baseMixinComputeDisplayName(rc *RecordCollection) *ModelData {
	res := NewModelData(rc.model).Set(rc.model.FieldName("DisplayName"), rc.Call("NameGet"))
	return res
}

func declareModelMixin() {
	modelMixin := NewMixinModel("ModelMixin")
	modelMixin.InheritModel(Registry.MustGet("BaseMixin"))
	modelMixin.fields.add(&Field{
		model:       modelMixin,
		name:        "HexyaExternalID",
		description: "Record External ID",
		json:        "hexya_external_id",
		fieldType:   fieldtype.Char,
		structField: reflect.StructField{Type: reflect.TypeOf("")},
		noCopy:      true,
		unique:      true,
		index:       true,
		required:    true,
		readOnly:    true,
		defaultFunc: func(env Environment) interface{} {
			return uuid.New().String()
		},
	})
	modelMixin.fields.add(&Field{
		model:       modelMixin,
		name:        "HexyaVersion",
		description: "Data Version",
		json:        "hexya_version",
		fieldType:   fieldtype.Integer,
		structField: reflect.StructField{Type: reflect.TypeOf(0)},
		noCopy:      true,
		defaultFunc: DefaultValue(0),
	})
}

// ConvertLimitToInt converts the given limit as interface{} to an int
func ConvertLimitToInt(limit interface{}) int {
	if l, ok := limit.(bool); ok && !l {
		return -1
	}
	val, err := nbutils.CastToInteger(limit)
	if err != nil {
		return 80
	}
	return int(val)
}

// FieldInfo is the exportable field information struct
type FieldInfo struct {
	ChangeDefault    bool                                  `json:"change_default"`
	Help             string                                `json:"help"`
	Searchable       bool                                  `json:"searchable"`
	Views            map[string]interface{}                `json:"views"`
	Required         bool                                  `json:"required"`
	Manual           bool                                  `json:"manual"`
	ReadOnly         bool                                  `json:"readonly"`
	Depends          []string                              `json:"depends"`
	CompanyDependent bool                                  `json:"company_dependent"`
	Sortable         bool                                  `json:"sortable"`
	Translate        bool                                  `json:"translate"`
	Type             fieldtype.Type                        `json:"type"`
	Store            bool                                  `json:"store"`
	String           string                                `json:"string"`
	Relation         string                                `json:"relation"`
	Selection        types.Selection                       `json:"selection"`
	Domain           interface{}                           `json:"domain"`
	OnChange         bool                                  `json:"-"`
	ReverseFK        string                                `json:"-"`
	Name             string                                `json:"-"`
	JSON             string                                `json:"-"`
	ReadOnlyFunc     func(Environment) (bool, Conditioner) `json:"-"`
	RequiredFunc     func(Environment) (bool, Conditioner) `json:"-"`
	InvisibleFunc    func(Environment) (bool, Conditioner) `json:"-"`
	DefaultFunc      func(Environment) interface{}         `json:"-"`
	GoType           reflect.Type                          `json:"-"`
	Index            bool                                  `json:"-"`
}

// FieldsGetArgs is the args struct for the FieldsGet method
type FieldsGetArgs struct {
	// Fields is a list of fields to document, all if empty or not provided
	Fields FieldNames `json:"allfields"`
}

// OnchangeParams is the args struct of the Onchange function
type OnchangeParams struct {
	Values   RecordData        `json:"values"`
	Fields   FieldNames        `json:"field_name"`
	Onchange map[string]string `json:"field_onchange"`
}

// OnchangeResult is the result struct type of the Onchange function
type OnchangeResult struct {
	Value   RecordData                `json:"value"`
	Warning string                    `json:"warning"`
	Filters map[FieldName]Conditioner `json:"domain"`
}
