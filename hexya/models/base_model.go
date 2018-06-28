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

	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/operator"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
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
	NewMixinModel("CommonMixin")
	declareCRUDMethods()
	declareRecordSetMethods()
	declareRecordSetSpecificMethods()
	declareSearchMethods()
	declareEnvironmentMethods()
}

// declareBaseMixin creates the mixin that implements all the necessary base methods of a model
func declareBaseMixin() {
	baseMixin := NewMixinModel("BaseMixin")
	declareBaseComputeMethods()
	baseMixin.AddFields(map[string]FieldDefinition{
		"CreateDate": DateTimeField{NoCopy: true},
		"CreateUID":  IntegerField{NoCopy: true},
		"WriteDate":  DateTimeField{NoCopy: true},
		"WriteUID":   IntegerField{NoCopy: true},
		"LastUpdate": DateTimeField{JSON: "__last_update", Compute: baseMixin.Methods().MustGet("ComputeLastUpdate"),
			Depends: []string{"WriteDate", "CreateDate"}},
		"DisplayName": CharField{Compute: baseMixin.Methods().MustGet("ComputeDisplayName"), Depends: []string{""}},
	})
	baseMixin.InheritModel(Registry.MustGet("CommonMixin"))
}

func declareModelMixin() {
	idSeq := NewSequence("HexyaExternalID")

	modelMixin := NewMixinModel("ModelMixin")
	modelMixin.AddFields(map[string]FieldDefinition{
		"HexyaExternalID": CharField{Unique: true, Index: true, NoCopy: true, Required: true,
			Default: func(env Environment) interface{} {
				return fmt.Sprintf("__hexya_external_id__%d", idSeq.NextValue())
			},
		},
		"HexyaVersion": IntegerField{GoType: new(int)},
	})
	modelMixin.InheritModel(Registry.MustGet("BaseMixin"))
}

// declareComputeMethods declares methods used to compute fields
func declareBaseComputeMethods() {
	model := Registry.MustGet("BaseMixin")

	model.AddMethod("ComputeLastUpdate",
		`ComputeLastUpdate returns the last datetime at which the record has been updated.`,
		func(rc *RecordCollection) FieldMap {
			if !rc.Get("WriteDate").(dates.DateTime).IsZero() {
				return FieldMap{"LastUpdate": rc.Get("WriteDate").(dates.DateTime)}
			}
			if !rc.Get("CreateDate").(dates.DateTime).IsZero() {
				return FieldMap{"LastUpdate": rc.Get("CreateDate").(dates.DateTime)}
			}
			return FieldMap{"LastUpdate": dates.Now()}
		})

	model.AddMethod("ComputeDisplayName",
		`ComputeDisplayName updates the DisplayName field with the result of NameGet.`,
		func(rc *RecordCollection) FieldMap {
			return FieldMap{"DisplayName": rc.Call("NameGet")}
		})

}

// declareCRUDMethods declares RecordSet CRUD methods
func declareCRUDMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("Create",
		`Create inserts a record in the database from the given data.
		Returns the created RecordCollection.`,
		func(rc *RecordCollection, data FieldMapper) *RecordCollection {
			return rc.create(data)
		})

	commonMixin.AddMethod("Read",
		`Read reads the database and returns a slice of FieldMap of the given model`,
		func(rc *RecordCollection, fields []string) []FieldMap {
			var res []FieldMap
			// Check if we have id in fields, and add it otherwise
			fields = addIDIfNotPresent(fields)
			// Do the actual reading
			for _, rec := range rc.Records() {
				fData := make(FieldMap)
				for _, fName := range fields {
					fData[fName] = rec.Get(fName)
				}
				res = append(res, fData)
			}
			return res
		})

	commonMixin.AddMethod("Load",
		`Load query all data of the RecordCollection and store in cache.
		fields are the fields to retrieve in the expression format,
		i.e. "User.Profile.Age" or "user_id.profile_id.age".
		If no fields are given, all DB columns of the RecordCollection's
		model are retrieved.`,
		func(rc *RecordCollection, fields ...string) *RecordCollection {
			return rc.Load(fields...)
		})

	commonMixin.AddMethod("Write",
		`Write is the base implementation of the 'Write' method which updates
		records in the database with the given data.
		Data can be either a struct pointer or a FieldMap.`,
		func(rc *RecordCollection, data FieldMapper, fieldsToUnset ...FieldNamer) bool {
			return rc.update(data, fieldsToUnset...)
		})

	commonMixin.AddMethod("Unlink",
		`Unlink deletes the given records in the database.`,
		func(rc *RecordCollection) int64 {
			return rc.unlink()
		})

	commonMixin.AddMethod("Copy",
		`Copy duplicates the given record
		It panics if rs is not a singleton`,
		func(rc *RecordCollection, overrides FieldMapper, fieldsToUnset ...FieldNamer) *RecordCollection {
			rc.EnsureOne()
			oVal := reflect.ValueOf(overrides)
			if oVal.IsNil() {
				overr := make(FieldMap)
				for _, f := range fieldsToUnset {
					overr[f.String()] = nil
				}
				overrides = overr
			}

			// Prevent infinite recursion if we have circular references
			if !rc.Env().Context().HasKey("__copy_data_seen") {
				rc = rc.WithContext("__copy_data_seen", map[string]bool{})
			}
			seenMap := rc.Env().Context().Get("__copy_data_seen").(map[string]bool)
			if seenMap[fmt.Sprintf("%s,%d", rc.ModelName(), rc.Ids()[0])] {
				return rc.env.Pool(rc.ModelName())
			}
			seenMap[fmt.Sprintf("%s,%d", rc.ModelName(), rc.Ids()[0])] = true
			rc = rc.WithContext("__copy_data_seen", seenMap)

			// Create a slice of fields to copy directly
			var fields []string
			for _, fi := range rc.model.fields.registryByName {
				if fi.noCopy || fi.isComputedField() {
					continue
				}
				if fi.fieldType == fieldtype.One2One || fi.fieldType.IsReverseRelationType() {
					// These fields will be copied after duplication
					continue
				}
				fields = append(fields, fi.json)
			}

			// We invalidate the cache in case we have noCopy fields that have
			// been loaded previously
			rc.env.cache.invalidateRecord(rc.model, rc.Get("id").(int64))
			rc.Load(fields...)

			fMap := rc.env.cache.getRecord(rc.Model(), rc.Get("id").(int64), rc.query.ctxArgsSlug())
			fMap.RemovePK()
			fMap.MergeWith(overrides.FieldMap(fieldsToUnset...), rc.model)
			// Reload original record to prevent cache discrepancies
			rc.Load()

			// Create our duplicated record
			newRs := rc.WithContext("hexya_force_compute_write", true).Call("Create", fMap).(RecordSet).Collection()

			// Duplicate one2one and one2many records
			for _, fi := range rc.model.fields.registryByName {
				if _, inOverrides := overrides.FieldMap(fieldsToUnset...).Get(fi.name, rc.model); inOverrides {
					continue
				}
				switch fi.fieldType {
				case fieldtype.One2One:
					newRs.Set(fi.name, rc.Get(fi.json).(RecordSet).Collection().Call("Copy", nil))
				case fieldtype.One2Many:
					for _, rec := range rc.Get(fi.json).(RecordSet).Collection().Records() {
						rec.Call("Copy", FieldMap{fi.reverseFK: newRs.Ids()[0]})
					}
				}
			}
			return newRs
		})
}

// declareRecordSetMethods declares general RecordSet methods
func declareRecordSetMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("NameGet",
		`NameGet retrieves the human readable name of this record.`,
		func(rc *RecordCollection) string {
			if _, nameExists := rc.model.fields.Get("Name"); nameExists {
				if !rc.env.cache.checkIfInCache(rc.model, rc.ids, []string{"Name"}, rc.query.ctxArgsSlug(), false) {
					rc.Load("Name")
				}
				switch name := rc.Get("Name").(type) {
				case string:
					return name
				case fmt.Stringer:
					return name.String()
				default:
					log.Panic("Name field is neither a string nor a fmt.Stringer", "model", rc.model)
				}
			}
			return rc.String()
		})

	commonMixin.AddMethod("SearchByName",
		`SearchByName searches for records that have a display name matching the given
		"name" pattern when compared with the given "op" operator, while also
		matching the optional search condition ("additionalCond").

		This is used for example to provide suggestions based on a partial
		value for a relational field. Sometimes be seen as the inverse
		function of NameGet but it is not guaranteed to be.`,
		func(rc *RecordCollection, name string, op operator.Operator, additionalCond Conditioner, limit int) *RecordCollection {
			if op == "" {
				op = operator.IContains
			}
			cond := rc.Model().Field("Name").AddOperator(op, name)
			if !additionalCond.Underlying().IsEmpty() {
				cond = cond.AndCond(additionalCond.Underlying())
			}
			return rc.Model().Search(rc.Env(), cond).Limit(limit)

		})

	commonMixin.AddMethod("FieldsGet",
		`FieldsGet returns the definition of each field.
		The embedded fields are included.
		The string, help, and selection (if present) attributes are translated.

		The result map is indexed by the fields JSON names.`,
		func(rc *RecordCollection, args FieldsGetArgs) map[string]*FieldInfo {
			fields := convertToFieldNamerSlice(args.Fields)
			// Get the field informations
			res := rc.model.FieldsGet(fields...)

			// Translate attributes when required
			lang := rc.Env().Context().GetString("lang")
			for fieldName, fInfo := range res {
				res[fieldName].Help = i18n.Registry.TranslateFieldHelp(lang, rc.model.name, fieldName, fInfo.Help)
				res[fieldName].String = i18n.Registry.TranslateFieldDescription(lang, rc.model.name, fieldName, fInfo.String)
				res[fieldName].Selection = i18n.Registry.TranslateFieldSelection(lang, rc.model.name, fieldName, fInfo.Selection)
			}
			return res
		})

	commonMixin.AddMethod("FieldGet",
		`FieldGet returns the definition of the given field.
		The string, help, and selection (if present) attributes are translated.`,
		func(rc *RecordCollection, field FieldNamer) *FieldInfo {
			args := FieldsGetArgs{
				Fields: []FieldName{field.FieldName()},
			}
			fJSON := rc.model.JSONizeFieldName(field.String())
			return rc.Call("FieldsGet", args).(map[string]*FieldInfo)[fJSON]
		})

	commonMixin.AddMethod("DefaultGet",
		`DefaultGet returns a Params map with the default values for the model.`,
		func(rc *RecordCollection) FieldMap {
			res := make(FieldMap)
			rc.applyDefaults(&res, false)
			rc.model.convertValuesToFieldType(&res)
			return res
		})
}

func declareRecordSetSpecificMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("CheckRecursion",
		`CheckRecursion verifies that there is no loop in a hierarchical structure of records,
        by following the parent relationship using the 'Parent' field until a loop is detected or
        until a top-level record is found.

        It returns true if no loop was found, false otherwise`,
		func(rc *RecordCollection) bool {
			if _, exists := rc.model.fields.Get("Parent"); !exists {
				// No Parent field in model, so no loop
				return true
			}
			// We use direct SQL query to bypass access control
			query := fmt.Sprintf(`SELECT parent_id FROM %s WHERE id = ?`, adapters[db.DriverName()].quoteTableName(rc.model.tableName))
			rc.Load("Parent")
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
		})

	commonMixin.AddMethod("Onchange",
		`Onchange returns the values that must be modified according to each field's Onchange
		method in the pseudo-record given as params.Values`,
		func(rc *RecordCollection, params OnchangeParams) OnchangeResult {
			var fields []FieldNamer
			values := params.Values
			rc.model.convertValuesToFieldType(&values)
			retValues := make(FieldMap)

			SimulateInNewEnvironment(rc.Env().Uid(), func(env Environment) {
				rs := env.Pool(rc.ModelName())
				// Tweaks for Onchange to work on creation with empty
				// RecordSet with ID = 0
				var rsID int64
				if !rc.IsEmpty() {
					rsID = rc.Ids()[0]
				}
				values.MergeWith(FieldMap{"ID": rsID}, rc.model)
				rs.ids = []int64{rsID}

				for _, field := range params.Fields {
					fi := rs.Model().Fields().MustGet(field)
					if fi.onChange == "" {
						continue
					}
					env.cache.invalidateRecord(rs.model, rsID)
					env.cache.addRecord(rs.model, rsID, values, rs.query.ctxArgsSlug())
					res := rs.CallMulti(fi.onChange)
					fields = res[1].([]FieldNamer)
					resMap := res[0].(FieldMapper).FieldMap(fields...)
					val := resMap.JSONized(rs.Model())
					values.MergeWith(val, rs.model)
					retValues.MergeWith(val, rs.model)
				}
			})
			retValues.RemovePK()
			return OnchangeResult{
				Value: retValues,
			}
		})
}

func declareSearchMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("Search",
		`Search returns a new RecordSet filtering on the current one with the
		additional given Condition`,
		func(rc *RecordCollection, cond Conditioner) *RecordCollection {
			return rc.Search(cond.Underlying())
		})

	commonMixin.AddMethod("Browse",
		`Browse returns a new RecordSet with only the records with the given ids.
		Note that this function is just a shorcut for Search on a list of ids.`,
		func(rc *RecordCollection, ids []int64) *RecordCollection {
			return rc.Call("Search", rc.Model().Field("ID").In(ids)).(RecordSet).Collection()
		})

	commonMixin.AddMethod("SearchCount",
		`SearchCount fetch from the database the number of records that match the RecordSet conditions`,
		func(rc *RecordCollection) int {
			return rc.SearchCount()
		})

	commonMixin.AddMethod("Fetch",
		`Fetch query the database with the current filter and returns a RecordSet
		with the queries ids.

		Fetch is lazy and only return ids. Use Load() instead if you want to fetch all fields.`,
		func(rc *RecordCollection) *RecordCollection {
			return rc.Fetch()
		})

	commonMixin.AddMethod("SearchAll",
		`SearchAll returns a RecordSet with all items of the table, regardless of the
		current RecordSet query. It is mainly meant to be used on an empty RecordSet`,
		func(rc *RecordCollection) *RecordCollection {
			return rc.SearchAll()
		})

	commonMixin.AddMethod("GroupBy",
		`GroupBy returns a new RecordSet grouped with the given GROUP BY expressions`,
		func(rc *RecordCollection, exprs ...FieldNamer) *RecordCollection {
			return rc.GroupBy(exprs...)
		})

	commonMixin.AddMethod("Aggregates",
		`Aggregates returns the result of this RecordSet query, which must by a grouped query.`,
		func(rc *RecordCollection, exprs ...FieldNamer) []GroupAggregateRow {
			return rc.Aggregates(exprs...)
		})

	commonMixin.AddMethod("Limit",
		`Limit returns a new RecordSet with only the first 'limit' records.`,
		func(rc *RecordCollection, limit int) *RecordCollection {
			return rc.Limit(limit)
		})

	commonMixin.AddMethod("Offset",
		`Offset returns a new RecordSet with only the records starting at offset`,
		func(rc *RecordCollection, offset int) *RecordCollection {
			return rc.Offset(offset)
		})

	commonMixin.AddMethod("OrderBy",
		`OrderBy returns a new RecordSet ordered by the given ORDER BY expressions.
		Each expression contains a field name and optionally one of "asc" or "desc", such as:

		rs.OrderBy("Company", "Name desc")`,
		func(rc *RecordCollection, exprs ...string) *RecordCollection {
			return rc.OrderBy(exprs...)
		})

	commonMixin.AddMethod("Union",
		`Union returns a new RecordSet that is the union of this RecordSet and the given
		"other" RecordSet. The result is guaranteed to be a set of unique records.`,
		func(rc *RecordCollection, other RecordSet) *RecordCollection {
			return rc.Union(other)
		})

	commonMixin.AddMethod("Subtract",
		`Subtract returns a RecordSet with the Records that are in this
		RecordCollection but not in the given 'other' one.
		The result is guaranteed to be a set of unique records.`,
		func(rc *RecordCollection, other RecordSet) *RecordCollection {
			return rc.Subtract(other)
		})

	commonMixin.AddMethod("Intersect",
		`Intersect returns a new RecordCollection with only the records that are both
		in this RecordCollection and in the other RecordSet.`,
		func(rc *RecordCollection, other RecordSet) *RecordCollection {
			return rc.Intersect(other)
		})

	commonMixin.AddMethod("CartesianProduct",
		`CartesianProduct returns the cartesian product of this RecordCollection with others.`,
		func(rc *RecordCollection, other ...RecordSet) []*RecordCollection {
			return rc.CartesianProduct(other...)
		})

	commonMixin.AddMethod("Equals",
		`Equals returns true if this RecordSet is the same as other
		i.e. they are of the same model and have the same ids`,
		func(rc *RecordCollection, other RecordSet) bool {
			return rc.Equals(other)
		})

	commonMixin.AddMethod("Sorted",
		`Sorted returns a new RecordCollection sorted according to the given less function.

		The less function should return true if rs1 < rs2`,
		func(rc *RecordCollection, less func(rs1 RecordSet, rs2 RecordSet) bool) *RecordCollection {
			return rc.Sorted(less)
		})

	commonMixin.AddMethod("SortedDefault",
		`SortedDefault returns a new record set with the same records as rc but sorted according
		to the default order of this model`,
		func(rc *RecordCollection) *RecordCollection {
			return rc.SortedDefault()
		})

	commonMixin.AddMethod("SortedByField",
		`SortedByField returns a new record set with the same records as rc but sorted by the given field.
		If reverse is true, the sort is done in reversed order`,
		func(rc *RecordCollection, namer FieldNamer, reverse bool) *RecordCollection {
			return rc.SortedByField(namer, reverse)
		})

	commonMixin.AddMethod("Filtered",
		`Filtered returns a new record set with only the elements of this record set
		for which test is true.
		
		Note that if this record set is not fully loaded, this function will call the database
		to load the fields before doing the filtering. In this case, it might be more efficient
		to search the database directly with the filter condition.`,
		func(rc *RecordCollection, test func(rs RecordSet) bool) *RecordCollection {
			return rc.Filtered(test)
		})
}

func declareEnvironmentMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("WithEnv",
		`WithEnv returns a copy of the current RecordSet with the given Environment.`,
		func(rc *RecordCollection, env Environment) *RecordCollection {
			return rc.WithEnv(env)
		})

	commonMixin.AddMethod("WithContext",
		`WithContext returns a copy of the current RecordSet with
		its context extended by the given key and value.`,
		func(rc *RecordCollection, key string, value interface{}) *RecordCollection {
			// Because this method returns an env with the same callstack as inside this layer,
			// we need to remove ourselves from the callstack.
			return rc.WithContext(key, value)
		})

	commonMixin.AddMethod("WithNewContext",
		`WithNewContext returns a copy of the current RecordSet with its context
	 	replaced by the given one.`,
		func(rc *RecordCollection, context *types.Context) *RecordCollection {
			// Because this method returns an env with the same callstack as inside this layer,
			// we need to remove ourselves from the callstack.
			return rc.WithNewContext(context)
		})

	commonMixin.AddMethod("Sudo",
		`Sudo returns a new RecordSet with the given userID
	 	or the superuser ID if not specified`,
		func(rc *RecordCollection, userID ...int64) *RecordCollection {
			// Because this method returns an env with the same callstack as inside this layer,
			// we need to remove ourselves from the callstack.
			return rc.Sudo(userID...)
		})
}

// ConvertLimitToInt converts the given limit as interface{} to an int
func ConvertLimitToInt(limit interface{}) int {
	var lim int
	switch limit.(type) {
	case bool:
		lim = -1
	case int:
		lim = limit.(int)
	default:
		lim = 80
	}
	return lim
}

// FieldInfo is the exportable field information struct
type FieldInfo struct {
	ChangeDefault    bool                   `json:"change_default"`
	Help             string                 `json:"help"`
	Searchable       bool                   `json:"searchable"`
	Views            map[string]interface{} `json:"views"`
	Required         bool                   `json:"required"`
	Manual           bool                   `json:"manual"`
	ReadOnly         bool                   `json:"readonly"`
	Depends          []string               `json:"depends"`
	CompanyDependent bool                   `json:"company_dependent"`
	Sortable         bool                   `json:"sortable"`
	Translate        bool                   `json:"translate"`
	Type             fieldtype.Type         `json:"type"`
	Store            bool                   `json:"store"`
	String           string                 `json:"string"`
	Relation         string                 `json:"relation"`
	Selection        types.Selection        `json:"selection"`
	Domain           interface{}            `json:"domain"`
	OnChange         bool                   `json:"-"`
	ReverseFK        string                 `json:"-"`
}

// FieldsGetArgs is the args struct for the FieldsGet method
type FieldsGetArgs struct {
	// Fields is a list of fields to document, all if empty or not provided
	Fields []FieldName `json:"allfields"`
}

// OnchangeParams is the args struct of the Onchange function
type OnchangeParams struct {
	Values   FieldMap          `json:"values"`
	Fields   []string          `json:"field_name"`
	Onchange map[string]string `json:"field_onchange"`
}

// OnchangeResult is the result struct type of the Onchange function
type OnchangeResult struct {
	Value FieldMapper `json:"value"`
}
