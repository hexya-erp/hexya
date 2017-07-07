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

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
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
	declareSearchMethods()
	declareEnvironmentMethods()
}

// declareBaseMixin creates the mixin that implements all the necessary base methods of a model
func declareBaseMixin() {
	baseMixin := NewMixinModel("BaseMixin")
	baseMixin.AddDateTimeField("CreateDate", SimpleFieldParams{NoCopy: true})
	baseMixin.AddIntegerField("CreateUID", SimpleFieldParams{NoCopy: true})
	baseMixin.AddDateTimeField("WriteDate", SimpleFieldParams{NoCopy: true})
	baseMixin.AddIntegerField("WriteUID", SimpleFieldParams{NoCopy: true})
	baseMixin.AddDateTimeField("LastUpdate", SimpleFieldParams{JSON: "__last_update", Compute: "ComputeLastUpdate"})
	baseMixin.InheritModel(Registry.MustGet("CommonMixin"))
	baseMixin.AddCharField("DisplayName", StringFieldParams{Compute: "ComputeNameGet"})
	declareBaseComputeMethods()
}

func declareModelMixin() {
	idSeq := NewSequence("HexyaExternalID")

	modelMixin := NewMixinModel("ModelMixin")
	modelMixin.AddCharField("HexyaExternalID", StringFieldParams{Unique: true, Index: true, NoCopy: true,
		Default: func(env Environment, values FieldMap) interface{} {
			return fmt.Sprintf("__hexya_external_id__%d", idSeq.NextValue())
		},
	})
	modelMixin.AddIntegerField("HexyaVersion", SimpleFieldParams{GoType: new(int)})
	modelMixin.InheritModel(Registry.MustGet("BaseMixin"))
}

// declareComputeMethods declares methods used to compute fields
func declareBaseComputeMethods() {
	model := Registry.MustGet("BaseMixin")

	model.AddMethod("ComputeLastUpdate",
		`ComputeLastUpdate returns the last datetime at which the record has been updated.`,
		func(rc RecordCollection) (FieldMap, []FieldNamer) {
			if !rc.Get("WriteDate").(types.DateTime).IsZero() {
				return FieldMap{"LastUpdate": rc.Get("WriteDate").(types.DateTime)}, []FieldNamer{FieldName("LastUpdate")}
			}
			if !rc.Get("CreateDate").(types.DateTime).IsZero() {
				return FieldMap{"LastUpdate": rc.Get("CreateDate").(types.DateTime)}, []FieldNamer{FieldName("LastUpdate")}
			}
			return FieldMap{"LastUpdate": types.Now()}, []FieldNamer{FieldName("LastUpdate")}
		}).AllowGroup(security.GroupEveryone)

	model.AddMethod("ComputeNameGet",
		`ComputeNameGet updates the DisplayName field with the result of NameGet.`,
		func(rc RecordCollection) (FieldMap, []FieldNamer) {
			return FieldMap{"DisplayName": rc.Call("NameGet").(string)}, []FieldNamer{FieldName("LastUpdate")}
		}).AllowGroup(security.GroupEveryone)

}

// declareCRUDMethods declares RecordSet CRUD methods
func declareCRUDMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("Create",
		`Create inserts a record in the database from the given data.
		Returns the created RecordCollection.`,
		func(rc RecordCollection, data FieldMapper) RecordCollection {
			return rc.create(data)
		})

	commonMixin.AddMethod("Read",
		`Read reads the database and returns a slice of FieldMap of the given model`,
		func(rc RecordCollection, fields []string) []FieldMap {
			res := make([]FieldMap, rc.Len())
			// Check if we have id in fields, and add it otherwise
			fields = addIDIfNotPresent(fields)
			// Do the actual reading
			for i, rec := range rc.Records() {
				res[i] = make(FieldMap)
				for _, fName := range fields {
					res[i][fName] = rec.Get(fName)
				}
			}
			return res
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Load",
		`Load query all data of the RecordCollection and store in cache.
		fields are the fields to retrieve in the expression format,
		i.e. "User.Profile.Age" or "user_id.profile_id.age".
		If no fields are given, all DB columns of the RecordCollection's
		model are retrieved.`,
		func(rc RecordCollection, fields ...string) RecordCollection {
			return rc.Load(fields...)
		})

	commonMixin.AddMethod("Write",
		`Write is the base implementation of the 'Write' method which updates
		records in the database with the given data.
		Data can be either a struct pointer or a FieldMap.`,
		func(rc RecordCollection, data FieldMapper, fieldsToUnset ...FieldNamer) bool {
			return rc.update(data, fieldsToUnset...)
		})

	commonMixin.AddMethod("Unlink",
		`Unlink deletes the given records in the database.`,
		func(rc RecordCollection) int64 {
			return rc.unlink()
		})

	commonMixin.AddMethod("Copy",
		`Copy duplicates the given record
		It panics if rs is not a singleton`,
		func(rc RecordCollection) RecordCollection {
			rc.EnsureOne()

			var fields []string
			for _, fi := range rc.model.fields.registryByName {
				if fi.noCopy {
					continue
				}
				if fi.fieldType.IsReverseRelationType() {
					continue
				}
				fields = append(fields, fi.json)
			}

			// We invalidate the cache in case we have noCopy fields that have
			// been loaded previously
			rc.env.cache.invalidateRecord(rc.model, rc.Get("id").(int64))
			rc.Load(fields...)

			fMap := rc.env.cache.getRecord(rc.ModelName(), rc.Get("id").(int64))
			fMap.RemovePK()
			newRs := rc.Call("Create", fMap).(RecordSet).Collection()
			return newRs
		})

}

// declareRecordSetMethods declares general RecordSet methods
func declareRecordSetMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("NameGet",
		`NameGet retrieves the human readable name of this record.`,
		func(rc RecordCollection) string {
			if _, nameExists := rc.model.fields.get("Name"); nameExists {
				if !rc.env.cache.checkIfInCache(rc.model, rc.ids, []string{"Name"}) {
					rc.Load("Name")
				}
				return rc.Get("Name").(string)
			}
			return rc.String()
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("FieldsGet",
		`FieldsGet returns the definition of each field.
		The embedded fields are included.
		The string, help, and selection (if present) attributes are translated.`,
		func(rc RecordCollection, args FieldsGetArgs) map[string]*FieldInfo {
			//TODO The string, help, and selection (if present) attributes are translated.
			res := make(map[string]*FieldInfo)
			fields := args.Fields
			if len(args.Fields) == 0 {
				for jName := range rc.model.fields.registryByJSON {
					fields = append(fields, FieldName(jName))
				}
			}
			for _, f := range fields {
				fInfo := rc.model.fields.MustGet(string(f))
				var relation string
				if fInfo.relatedModel != nil {
					relation = fInfo.relatedModel.name
				}
				var filter interface{}
				if fInfo.filter != nil {
					filter = fInfo.filter.Serialize()
				}
				res[fInfo.json] = &FieldInfo{
					Help:       fInfo.help,
					Searchable: true,
					Depends:    fInfo.depends,
					Sortable:   true,
					Type:       fInfo.fieldType,
					Store:      fInfo.isStored(),
					String:     fInfo.description,
					Relation:   relation,
					Required:   fInfo.required,
					Selection:  fInfo.selection,
					Domain:     filter,
				}
			}
			return res
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("FieldGet",
		`FieldGet returns the definition of the given field.
		The string, help, and selection (if present) attributes are translated.`,
		func(rc RecordCollection, field FieldNamer) *FieldInfo {
			args := FieldsGetArgs{
				Fields: []FieldName{field.FieldName()},
			}
			fJSON := rc.model.JSONizeFieldName(string(field.FieldName()))
			return rc.Call("FieldsGet", args).(map[string]*FieldInfo)[fJSON]
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("DefaultGet",
		`DefaultGet returns a Params map with the default values for the model.`,
		func(rc RecordCollection) FieldMap {
			res := make(FieldMap)
			rc.applyDefaults(&res)
			rc.model.convertValuesToFieldType(&res)
			return res
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Onchange",
		`Onchange returns the values that must be modified according to each field's Onchange
		method in the pseudo-record given as params.Values`,
		func(rc RecordCollection, params OnchangeParams) OnchangeResult {
			var fields []FieldNamer
			values := params.Values
			retValues := make(FieldMap)

			SimulateInNewEnvironment(rc.Env().Uid(), func(env Environment) {
				rs := env.Pool(rc.ModelName())
				// Tweaks for Onchange to work on creation with empty
				// RecordSet with ID = 0
				var rsID int64
				if !rc.IsEmpty() {
					rsID = rc.Ids()[0]
				}
				rc.model.MergeFieldMaps(values, FieldMap{"ID": rsID})
				rs.ids = []int64{rsID}

				for _, field := range params.Fields {
					fi := rs.Model().Fields().MustGet(field)
					if fi.onChange == "" {
						continue
					}
					env.cache.invalidateRecord(rs.model, rsID)
					env.cache.addRecord(rs.model, rsID, values)
					res := rs.CallMulti(fi.onChange)
					fields = res[1].([]FieldNamer)
					val := rs.model.JSONizeFieldMap(res[0].(FieldMapper).FieldMap(fields...))
					rs.model.MergeFieldMaps(values, val)
					rs.model.MergeFieldMaps(retValues, val)
				}
			})
			retValues.RemovePK()
			return OnchangeResult{
				Value: retValues,
			}
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("CheckRecursion",
		`CheckRecursion verifies that there is no loop in a hierarchical structure of records,
        by following the parent relationship using the 'Parent' field until a loop is detected or
        until a top-level record is found.

        It returns true if no loop was found, false otherwise`,
		func(rc RecordCollection) bool {
			if _, exists := rc.model.fields.get("Parent"); !exists {
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
		}).AllowGroup(security.GroupEveryone)
}

func declareSearchMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("Search",
		`Search returns a new RecordSet filtering on the current one with the
		additional given Condition`,
		func(rc RecordCollection, cond *Condition) RecordCollection {
			return rc.Search(cond)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Browse",
		`Browse returns a new RecordSet with the records with the given ids.
		Note that this function is just a shorcut for Search on a list of ids.`,
		func(rc RecordCollection, ids []int64) RecordCollection {
			return rc.Call("Search", rc.Model().Field("ID").In(ids)).(RecordSet).Collection()
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Fetch",
		`Fetch query the database with the current filter and returns a RecordSet
		with the queries ids. Fetch is lazy and only return ids. Use Load() instead
		if you want to fetch all fields.`,
		func(rc RecordCollection) RecordCollection {
			return rc.Fetch()
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("FetchAll",
		`FetchAll returns a RecordSet with all items of the table, regardless of the
		current RecordSet query. It is mainly meant to be used on an empty RecordSet`,
		func(rc RecordCollection) RecordCollection {
			return rc.FetchAll()
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("GroupBy",
		`GroupBy returns a new RecordSet grouped with the given GROUP BY expressions`,
		func(rc RecordCollection, exprs ...FieldNamer) RecordCollection {
			return rc.GroupBy(exprs...)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Aggregates",
		`Aggregates returns the result of this RecordSet query, which must by a grouped query.`,
		func(rc RecordCollection, exprs ...FieldNamer) []GroupAggregateRow {
			return rc.Aggregates(exprs...)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Limit",
		`Limit returns a new RecordSet with only the first 'limit' records.`,
		func(rc RecordCollection, limit int) RecordCollection {
			return rc.Limit(limit)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Offset",
		`Offset returns a new RecordSet with only the records starting at offset`,
		func(rc RecordCollection, offset int) RecordCollection {
			return rc.Offset(offset)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("OrderBy",
		`OrderBy returns a new RecordSet ordered by the given ORDER BY expressions`,
		func(rc RecordCollection, exprs ...string) RecordCollection {
			return rc.OrderBy(exprs...)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Union",
		`Union returns a new RecordSet that is the union of this RecordSet and the given
		"other" RecordSet. The result is guaranteed to be a set of unique records.`,
		func(rc RecordCollection, other RecordCollection) RecordCollection {
			return rc.Union(other)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Subtract",
		`Subtract returns a RecordSet with the Records that are in this
		RecordCollection but not in the given 'other' one.
		The result is guaranteed to be a set of unique records.`,
		func(rc RecordCollection, other RecordCollection) RecordCollection {
			return rc.Subtract(other)
		}).AllowGroup(security.GroupEveryone)
}

func declareEnvironmentMethods() {
	commonMixin := Registry.MustGet("CommonMixin")

	commonMixin.AddMethod("WithEnv",
		`WithEnv returns a copy of the current RecordSet with the given Environment.`,
		func(rc RecordCollection, env Environment) RecordCollection {
			return rc.WithEnv(env)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("WithContext",
		`WithContext returns a copy of the current RecordSet with
		its context extended by the given key and value.`,
		func(rc RecordCollection, key string, value interface{}) RecordCollection {
			return rc.WithContext(key, value)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("WithNewContext",
		`WithNewContext returns a copy of the current RecordSet with its context
	 	replaced by the given one.`,
		func(rc RecordCollection, context *types.Context) RecordCollection {
			return rc.WithNewContext(context)
		}).AllowGroup(security.GroupEveryone)

	commonMixin.AddMethod("Sudo",
		`Sudo returns a new RecordSet with the given userID
	 	or the superuser ID if not specified`,
		func(rc RecordCollection, userID ...int64) RecordCollection {
			return rc.Sudo(userID...)
		}).AllowGroup(security.GroupEveryone)
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
}

// FieldsGetArgs is the args struct for the FieldsGet method
type FieldsGetArgs struct {
	// list of fields to document, all if empty or not provided
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
