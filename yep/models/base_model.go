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
	"time"

	"github.com/npiganeau/yep/yep/models/operator"
	"github.com/npiganeau/yep/yep/models/types"
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
)

// declareBaseMixin creates the mixin that implements all the necessary base methods of a model
func declareBaseMixin() {
	model := NewMixinModel("BaseMixin")

	model.AddDateTimeField("CreateDate", SimpleFieldParams{NoCopy: true})
	model.AddIntegerField("CreateUID", SimpleFieldParams{NoCopy: true})
	model.AddDateTimeField("WriteDate", SimpleFieldParams{NoCopy: true})
	model.AddIntegerField("WriteUID", SimpleFieldParams{NoCopy: true})
	model.AddDateTimeField("LastUpdate", SimpleFieldParams{JSON: "__last_update", Compute: "ComputeLastUpdate"})
	model.AddCharField("DisplayName", StringFieldParams{Compute: "ComputeNameGet"})

	declareComputeMethods(model)
	declareCRUDMethods(model)
	declareClientHelperMethods(model)
	declareSearchMethods(model)

	MixInAllModels(model)
}

func declareComputeMethods(model *Model) {

	model.AddMethod("ComputeWriteDate",
		`ComputeWriteDate updates the WriteDate field with the current datetime.`,
		func(rc RecordCollection) FieldMap {
			return FieldMap{"WriteDate": DateTime(time.Now())}
		})

	model.AddMethod("ComputeLastUpdate",
		`ComputeLastUpdate returns the last datetime at which the record has been updated.`,
		func(rc RecordCollection) FieldMap {
			lastUpdate := DateTime(time.Now())
			if !rc.Get("WriteDate").(DateTime).IsNull() {
				lastUpdate = rc.Get("WriteDate").(DateTime)
			}
			if !rc.Get("CreateDate").(DateTime).IsNull() {
				lastUpdate = rc.Get("CreateDate").(DateTime)
			}
			return FieldMap{"LastUpdate": lastUpdate}
		})

	model.AddMethod("ComputeNameGet",
		`ComputeNameGet updates the DisplayName field with the result of NameGet.`,
		func(rc RecordCollection) FieldMap {
			return FieldMap{"DisplayName": rc.Call("NameGet").(string)}
		})
}

func declareCRUDMethods(model *Model) {

	model.AddMethod("Create",
		`Create inserts a record in the database from the given data.
		Returns the created RecordCollection.`,
		func(rc RecordCollection, data interface{}) RecordCollection {
			return rc.create(data)
		})

	model.AddMethod("Read",
		`Read reads the database and returns a slice of FieldMap of the given model`,
		func(rc RecordCollection, fields []string) []FieldMap {
			res := make([]FieldMap, rc.Len())
			// Check if we have id in fields, and add it otherwise
			fields = addIDIfNotPresent(fields)
			// Do the actual reading
			for i, rec := range rc.Records() {
				res[i] = make(FieldMap)
				for _, fName := range fields {
					value := rec.Get(fName)
					if relRC, ok := value.(RecordCollection); ok {
						relRC = relRC.Fetch()
						fi := rc.model.fields.mustGet(fName)
						switch {
						case fi.fieldType.Is2OneRelationType():
							if rcId := relRC.Get("id"); rcId != 0 {
								value = [2]interface{}{rcId, relRC.Call("NameGet").(string)}
							} else {
								value = nil
							}
						case fi.fieldType.Is2ManyRelationType():
							value = relRC.Ids()
						}
					}
					res[i][fName] = value
				}
			}
			return res
		})

	model.AddMethod("Load",
		`Load query all data of the RecordCollection and store in cache.
		fields are the fields to retrieve in the expression format,
		i.e. "User.Profile.Age" or "user_id.profile_id.age".
		If no fields are given, all DB columns of the RecordCollection's
		model are retrieved.`,
		func(rc RecordCollection, fields ...string) RecordCollection {
			return rc.Load(fields...)
		})

	model.AddMethod("Write",
		`Write is the base implementation of the 'Write' method which updates
		records in the database with the given data.
		Data can be either a struct pointer or a FieldMap.`,
		func(rc RecordCollection, data interface{}, fieldsToUnset ...FieldNamer) bool {
			return rc.update(data, fieldsToUnset...)
		})

	model.AddMethod("Unlink",
		`Unlink deletes the given records in the database.`,
		func(rc RecordCollection) int64 {
			return rc.delete()
		})

	model.AddMethod("Copy",
		`Copy duplicates the given record
		It panics if rs is not a singleton`,
		func(rc RecordCollection) RecordCollection {
			rc.EnsureOne()

			var fields []string
			for _, fi := range rc.model.fields.registryByName {
				if !fi.noCopy {
					fields = append(fields, fi.json)
				}
			}

			rc.Load(fields...)

			fMap := rc.env.cache.getRecord(rc.ModelName(), rc.Get("id").(int64))
			delete(fMap, "ID")
			delete(fMap, "id")
			newRs := rc.Call("Create", fMap).(RecordCollection)
			return newRs
		})
}

func declareClientHelperMethods(model *Model) {

	model.AddMethod("NameGet",
		`NameGet retrieves the human readable name of this record.`,
		func(rc RecordCollection) string {
			if _, nameExists := rc.model.fields.get("name"); nameExists {
				if !rc.env.cache.checkIfInCache(rc.model, rc.ids, []string{"name"}) {
					rc.Load("name")
				}
				return rc.Get("name").(string)
			}
			return rc.String()
		})

	model.AddMethod("NameSearch",
		`NameSearch searches for records that have a display name matching the given
		"name" pattern when compared with the given "operator", while also
		matching the optional search domain ("args").

		This is used for example to provide suggestions based on a partial
		value for a relational field. Sometimes be seen as the inverse
		function of NameGet but it is not guaranteed to be.`,
		func(rc RecordCollection, params NameSearchParams) []RecordIDWithName {
			searchRs := rc.Search(rc.Model().Field("Name").addOperator(params.Operator, params.Name)).Limit(ConvertLimitToInt(params.Limit))
			if extraCondition := ParseDomain(params.Args); extraCondition != nil {
				searchRs = searchRs.Search(extraCondition)
			}

			searchRs.Load("ID", "DisplayName")

			res := make([]RecordIDWithName, searchRs.Len())
			for i, rec := range searchRs.Records() {
				res[i].ID = rec.Get("id").(int64)
				res[i].Name = rec.Get("display_name").(string)
			}
			return res
		})

	model.AddMethod("FieldsGet",
		`FieldsGet returns the definition of each field.
		The embedded fields are included.
		The string, help, and selection (if present) attributes are translated.`,
		func(rc RecordCollection, args FieldsGetArgs) map[string]*FieldInfo {
			//TODO The string, help, and selection (if present) attributes are translated.
			res := make(map[string]*FieldInfo)
			fields := args.AllFields
			if len(args.AllFields) == 0 {
				for jName := range rc.model.fields.registryByJSON {
					//if fi.fieldType != tools.MANY2MANY {
					// We don't want Many2Many as it points to the link table
					fields = append(fields, FieldName(jName))
					//}
				}
			}
			for _, f := range fields {
				fInfo := rc.model.fields.mustGet(string(f))
				var relation string
				if fInfo.relatedModel != nil {
					relation = fInfo.relatedModel.name
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
				}
			}
			return res
		})

	model.AddMethod("DefaultGet",
		`DefaultGet returns a Params map with the default values for the model.`,
		func(rc RecordCollection) FieldMap {
			// TODO Implement DefaultGet
			return make(FieldMap)
		})

	model.AddMethod("Onchange",
		`Onchange returns the values that must be modified in the pseudo-record
		given as params.Values`,
		func(rc RecordCollection, params OnchangeParams) FieldMap {
			// TODO Implement Onchange
			return make(FieldMap)
		})
}

func declareSearchMethods(model *Model) {
	model.AddMethod("Search",
		`Search returns a new RecordSet filtering on the current one with the
		additional given Condition`,
		func(rc RecordCollection, cond *Condition) RecordCollection {
			return rc.Search(cond)
		})

	model.AddMethod("Fetch",
		`Fetch query the database with the current filter and returns a RecordSet
		with the queries ids. Fetch is lazy and only return ids. Use Load() instead
		if you want to fetch all fields.`,
		func(rc RecordCollection) RecordCollection {
			return rc.Fetch()
		})

	model.AddMethod("GroupBy",
		`GroupBy returns a new RecordSet grouped with the given GROUP BY expressions`,
		func(rc RecordCollection, exprs ...string) RecordCollection {
			return rc.GroupBy(exprs...)
		})

	model.AddMethod("Limit",
		`Limit returns a new RecordSet with only the first 'limit' records.`,
		func(rc RecordCollection, limit int) RecordCollection {
			return rc.Limit(limit)
		})

	model.AddMethod("Offset",
		`Offset returns a new RecordSet with only the records starting at offset`,
		func(rc RecordCollection, offset int) RecordCollection {
			return rc.Offset(offset)
		})

	model.AddMethod("OrderBy",
		`OrderBy returns a new RecordSet ordered by the given ORDER BY expressions`,
		func(rc RecordCollection, exprs ...string) RecordCollection {
			return rc.OrderBy(exprs...)
		})

	model.AddMethod("Union",
		`Union returns a new RecordSet that is the union of this RecordSet and the given
		"other" RecordSet. The result is guaranteed to be a set of unique records.`,
		func(rc RecordCollection, other RecordCollection) RecordCollection {
			return rc.Union(other)
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

// NameSearchParams is the args struct for the NameSearch function
type NameSearchParams struct {
	Args     Domain            `json:"args"`
	Name     string            `json:"name"`
	Operator operator.Operator `json:"operator"`
	Limit    interface{}       `json:"limit"`
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
	Type             types.FieldType        `json:"type"`
	Store            bool                   `json:"store"`
	String           string                 `json:"string"`
	Domain           Domain                 `json:"domain"`
	Relation         string                 `json:"relation"`
}

// FieldsGetArgs is the args struct for the FieldsGet method
type FieldsGetArgs struct {
	// list of fields to document, all if empty or not provided
	AllFields []FieldName `json:"allfields"`
}

// OnchangeParams is the args struct of the Onchange function
type OnchangeParams struct {
	Values   FieldMap          `json:"values"`
	Fields   []string          `json:"field_name"`
	Onchange map[string]string `json:"field_onchange"`
}
