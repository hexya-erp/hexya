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
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/npiganeau/yep/yep/ir"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/tools/logging"
)

const (
	// TransientModel means that the records of this model will be automatically
	// removed periodically. Transient models are mainly used for wizards.
	TransientModel Option = 1 << iota
)

// BaseModel is the base implementation  of all non transient models
type BaseModel struct {
	ID          int64
	CreateDate  DateTime `yep:"nocopy"`
	CreateUID   int64    `yep:"nocopy"`
	WriteDate   DateTime `yep:"nocopy"`
	WriteUID    int64    `yep:"nocopy"`
	LastUpdate  DateTime `yep:"compute(ComputeLastUpdate);json(__last_update)"`
	DisplayName string   `yep:"compute(ComputeNameGet)"`
}

// BaseTransientModel is the base implementation of all transient models
type BaseTransientModel struct {
	ID int64 `orm:"column(id)"`
}

// declareBaseMethods creates all the necessary base methods of a model
func declareBaseMethods(name string) {
	CreateMethod(name, "ComputeWriteDate", ComputeWriteDate)
	CreateMethod(name, "ComputeLastUpdate", ComputeLastUpdate)
	CreateMethod(name, "ComputeNameGet", ComputeNameGet)
	CreateMethod(name, "Create", Create)
	CreateMethod(name, "Read", Read)
	CreateMethod(name, "Load", Load)
	CreateMethod(name, "Write", Write)
	CreateMethod(name, "Unlink", Unlink)
	CreateMethod(name, "Copy", Copy)
	CreateMethod(name, "NameGet", NameGet)
	CreateMethod(name, "NameSearch", NameSearch)
	CreateMethod(name, "GetFormviewId", GetFormviewId)
	CreateMethod(name, "GetFormviewAction", GetFormviewAction)
	CreateMethod(name, "FieldsViewGet", FieldsViewGet)
	CreateMethod(name, "FieldsGet", FieldsGet)
	CreateMethod(name, "ProcessView", ProcessView)
	CreateMethod(name, "AddModifiers", AddModifiers)
	CreateMethod(name, "UpdateFieldNames", UpdateFieldNames)
	CreateMethod(name, "SearchRead", SearchRead)
	CreateMethod(name, "DefaultGet", DefaultGet)
	CreateMethod(name, "Onchange", Onchange)
	CreateMethod(name, "Search", Search)
	CreateMethod(name, "Filter", Filter)
	CreateMethod(name, "Exclude", Exclude)
	CreateMethod(name, "Distinct", Distinct)
	CreateMethod(name, "Fetch", Fetch)
	CreateMethod(name, "GroupBy", GroupBy)
	CreateMethod(name, "Limit", Limit)
	CreateMethod(name, "Offset", Offset)
	CreateMethod(name, "OrderBy", OrderBy)
	CreateMethod(name, "Union", Union)
}

// Search returns a new RecordSet filtering on the current one with the
// additional given Condition
func Search(rc RecordCollection, cond *Condition) RecordCollection {
	return rc.Search(cond)
}

// Filter returns a new RecordSet filtered on records matching the given additional condition.
func Filter(rc RecordCollection, fieldName, op string, data interface{}) RecordCollection {
	return rc.Filter(fieldName, op, data)
}

// Exclude returns a new RecordSet filtered on records NOT matching the given additional condition.
func Exclude(rc RecordCollection, fieldName, op string, data interface{}) RecordCollection {
	return rc.Exclude(fieldName, op, data)
}

// Distinct returns a new RecordSet without duplicates
func Distinct(rc RecordCollection) RecordCollection {
	return rc.Distinct()
}

// Fetch query the database with the current filter and returns a RecordSet
// with the queries ids. Fetch is lazy and only return ids. Use Load() instead
// if you want to fetch all fields.
func Fetch(rc RecordCollection) RecordCollection {
	return rc.Fetch()
}

// GroupBy returns a new RecordSet grouped with the given GROUP BY expressions
func GroupBy(rc RecordCollection, exprs ...string) RecordCollection {
	return rc.GroupBy(exprs...)
}

// Limit returns a new RecordSet with only the first 'limit' records.
func Limit(rc RecordCollection, limit int) RecordCollection {
	return rc.Limit(limit)
}

// Offset returns a new RecordSet with only the records starting at offset
func Offset(rc RecordCollection, offset int) RecordCollection {
	return rc.Offset(offset)
}

// OrderBy returns a new RecordSet ordered by the given ORDER BY expressions
func OrderBy(rc RecordCollection, exprs ...string) RecordCollection {
	return rc.OrderBy(exprs...)
}

// ComputeWriteDate updates the WriteDate field with the current datetime.
func ComputeWriteDate(rc RecordCollection) FieldMap {
	return FieldMap{"WriteDate": DateTime(time.Now())}
}

// ComputeLastUpdate returns the last datetime at which the record has been updated.
func ComputeLastUpdate(rc RecordCollection) FieldMap {
	lastUpdate := DateTime(time.Now())
	if !rc.Get("WriteDate").(DateTime).IsNull() {
		lastUpdate = rc.Get("WriteDate").(DateTime)
	}
	if !rc.Get("CreateDate").(DateTime).IsNull() {
		lastUpdate = rc.Get("CreateDate").(DateTime)
	}
	fmt.Println("last_update", rc.Get("WriteDate").(DateTime), rc.Get("CreateDate").(DateTime), lastUpdate)
	return FieldMap{"LastUpdate": lastUpdate}
}

// ComputeNameGet updates the DisplayName field with the result of NameGet.
func ComputeNameGet(rc RecordCollection) FieldMap {
	return FieldMap{"DisplayName": rc.Call("NameGet").(string)}
}

// Create inserts a record in the database from the given data.
// Returns the created RecordCollection.
func Create(rc RecordCollection, data interface{}) RecordCollection {
	return rc.create(data)
}

// Load query all data of the RecordCollection and store in cache.
// fields are the fields to retrieve in the expression format,
// i.e. "User.Profile.Age" or "user_id.profile_id.age".
// If no fields are given, all DB columns of the RecordCollection's
// model are retrieved.
func Load(rc RecordCollection, fields ...string) RecordCollection {
	return rc.Load(fields...)
}

// Read reads the database and returns a slice of FieldMap of the given model
func Read(rc RecordCollection, fields []string) []FieldMap {
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
				fi := rc.mi.fields.mustGet(fName)
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
}

// Write is the base implementation of the 'Write' method which updates
// records in the database with the given data.
// Data can be either a struct pointer or a FieldMap.
func Write(rc RecordCollection, data interface{}, fieldsToUnset ...string) bool {
	return rc.update(data, fieldsToUnset...)
}

// Unlink deletes the given records in the database.
func Unlink(rc RecordCollection) int64 {
	return rc.delete()
}

// Copy duplicates the given record
// It panics if rs is not a singleton
func Copy(rc RecordCollection) RecordCollection {
	rc.EnsureOne()

	var fields []string
	for _, fi := range rc.mi.fields.registryByName {
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
}

// Union returns a new RecordSet that is the union of this RecordSet and the given
// `other` RecordSet. The result is guaranteed to be a set of unique records.
func Union(rc RecordCollection, other RecordCollection) RecordCollection {
	return rc.Union(other)
}

// NameGet retrieves the human readable name of this record.
func NameGet(rc RecordCollection) string {
	if _, nameExists := rc.mi.fields.get("name"); nameExists {
		if !rc.env.cache.checkIfInCache(rc.mi, rc.ids, []string{"name"}) {
			rc.Load("name")
		}
		return rc.Get("name").(string)
	}
	return rc.String()
}

// convertLimitToInt converts the given limit as interface{} to an int
func convertLimitToInt(limit interface{}) int {
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
	Args     Domain      `json:"args"`
	Name     string      `json:"name"`
	Operator string      `json:"operator"`
	Limit    interface{} `json:"limit"`
}

// NameSearch searches for records that have a display name matching the given
// `name` pattern when compared with the given `operator`, while also
// matching the optional search domain (`args`).
//
// This is used for example to provide suggestions based on a partial
// value for a relational field. Sometimes be seen as the inverse
// function of NameGet but it is not guaranteed to be.
func NameSearch(rc RecordCollection, params NameSearchParams) []RecordIDWithName {
	searchRs := rc.Filter("Name", params.Operator, params.Name).Limit(convertLimitToInt(params.Limit))
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
}

// GetFormviewId returns an view id to open the document with.
// This method is meant to be overridden in addons that want
// to give specific view ids for example.
func GetFormviewId(rc RecordCollection) string {
	return ""
}

// GetFormviewAction returns an action to open the document.
// This method is meant to be overridden in addons that want
// to give specific view ids for example.
func GetFormviewAction(rc RecordCollection) *ir.BaseAction {
	viewID := rc.Call("GetFormviewId").(string)
	return &ir.BaseAction{
		Type:        ir.ActionActWindow,
		Model:       rc.ModelName(),
		ActViewType: ir.ActionViewTypeForm,
		ViewMode:    "form",
		Views:       []ir.ViewRef{{viewID, string(ir.VIEW_TYPE_FORM)}},
		Target:      "current",
		ResID:       rc.Get("id").(int64),
		Context:     rc.Env().Context(),
	}
}

// FieldsViewGetParams is the args struct for the FieldsViewGet function
type FieldsViewGetParams struct {
	ViewID   string `json:"view_id"`
	ViewType string `json:"view_type"`
	Toolbar  bool   `json:"toolbar"`
}

// FieldsViewData is the return type string for the FieldsViewGet function
type FieldsViewData struct {
	Name        string                `json:"name"`
	Arch        string                `json:"arch"`
	ViewID      string                `json:"view_id"`
	Model       string                `json:"model"`
	Type        ir.ViewType           `json:"type"`
	Fields      map[string]*FieldInfo `json:"fields"`
	Toolbar     ir.Toolbar            `json:"toolbar"`
	FieldParent string                `json:"field_parent"`
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

// FieldsViewGet is the base implementation of the 'FieldsViewGet' method which
// gets the detailed composition of the requested view like fields, model,
// view architecture.
func FieldsViewGet(rc RecordCollection, args FieldsViewGetParams) *FieldsViewData {
	view := ir.ViewsRegistry.GetViewById(args.ViewID)
	if view == nil {
		view = ir.ViewsRegistry.GetFirstViewForModel(rc.ModelName(), ir.ViewType(args.ViewType))
	}
	cols := make([]string, len(view.Fields))
	for i, f := range view.Fields {
		fi := rc.mi.fields.mustGet(f)
		cols[i] = fi.json
	}
	fInfos := rc.Call("FieldsGet", FieldsGetArgs{AllFields: cols}).(map[string]*FieldInfo)
	arch := rc.Call("ProcessView", view.Arch, fInfos).(string)
	res := FieldsViewData{
		Name:   view.Name,
		Arch:   arch,
		ViewID: args.ViewID,
		Model:  view.Model,
		Type:   view.Type,
		Fields: fInfos,
	}
	return &res
}

// ProcessView makes all the necessary modifications to the view
// arch and returns the new xml string.
func ProcessView(rc RecordCollection, arch string, fieldInfos map[string]*FieldInfo) string {
	// Load arch as etree
	doc := etree.NewDocument()
	if err := doc.ReadFromString(arch); err != nil {
		logging.LogAndPanic(log, "Unable to parse view arch", "arch", arch, "error", err)
	}
	// Apply changes
	rc.Call("UpdateFieldNames", doc)
	rc.Call("AddModifiers", doc, fieldInfos)
	// Dump xml to string and return
	res, err := doc.WriteToString()
	if err != nil {
		logging.LogAndPanic(log, "Unable to render XML", "error", err)
	}
	return res
}

// AddModifiers adds the modifiers attribute nodes to given xml doc.
func AddModifiers(rc RecordCollection, doc *etree.Document, fieldInfos map[string]*FieldInfo) {
	for _, fieldTag := range doc.FindElements("//field") {
		fieldName := fieldTag.SelectAttr("name").Value
		var mods []string
		if fieldInfos[fieldName].ReadOnly {
			mods = append(mods, "&quot;readonly&quot;: true")
		}
		modStr := fmt.Sprintf("{%s}", strings.Join(mods, ","))
		fieldTag.CreateAttr("modifiers", modStr)
	}
}

// UpdateFieldNames changes the field names in the view to the column names.
// If a field name is already column names then it does nothing.
func UpdateFieldNames(rc RecordCollection, doc *etree.Document) {
	for _, fieldTag := range doc.FindElements("//field") {
		fieldName := fieldTag.SelectAttr("name").Value
		fi := rc.mi.fields.mustGet(fieldName)
		fieldTag.RemoveAttr("name")
		fieldTag.CreateAttr("name", fi.json)
	}
	for _, labelTag := range doc.FindElements("//label") {
		fieldName := labelTag.SelectAttr("for").Value
		fi := rc.mi.fields.mustGet(fieldName)
		labelTag.RemoveAttr("for")
		labelTag.CreateAttr("for", fi.json)
	}
}

// FieldsGetArgs is the args struct for the FieldsGet method
type FieldsGetArgs struct {
	// list of fields to document, all if empty or not provided
	AllFields []string `json:"allfields"`
}

// FieldsGet returns the definition of each field.
// The embedded fields are included.
// TODO The string, help, and selection (if present) attributes are translated.
func FieldsGet(rc RecordCollection, args FieldsGetArgs) map[string]*FieldInfo {
	res := make(map[string]*FieldInfo)
	fields := args.AllFields
	if len(args.AllFields) == 0 {
		for jName := range rc.mi.fields.registryByJSON {
			//if fi.fieldType != tools.MANY2MANY {
			// We don't want Many2Many as it points to the link table
			fields = append(fields, jName)
			//}
		}
	}
	for _, f := range fields {
		fInfo := rc.mi.fields.mustGet(f)
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
}

// SearchParams is the args struct for the SearchRead method
type SearchParams struct {
	Domain Domain      `json:"domain"`
	Fields []string    `json:"fields"`
	Offset int         `json:"offset"`
	Limit  interface{} `json:"limit"`
	Order  string      `json:"order"`
}

// SearchRead retrieves database records according to the filters defined in params.
func SearchRead(rc RecordCollection, params SearchParams) []FieldMap {
	if searchCond := ParseDomain(params.Domain); searchCond != nil {
		rc = rc.Search(searchCond)
	}
	// Limit
	rc = rc.Limit(convertLimitToInt(params.Limit))

	// Offset
	if params.Offset != 0 {
		rc = rc.Offset(params.Offset)
	}

	// Order
	if params.Order != "" {
		rc = rc.OrderBy(strings.Split(params.Order, ",")...)
	}

	rSet := rc.Fetch()
	return rSet.Call("Read", params.Fields).([]FieldMap)
}

// DefaultGet returns a Params map with the default values for the model.
func DefaultGet(rc RecordCollection) FieldMap {
	// TODO Implement DefaultGet
	return make(FieldMap)
}

// OnchangeParams is the args struct of the Onchange function
type OnchangeParams struct {
	Values   FieldMap          `json:"values"`
	Fields   []string          `json:"field_name"`
	Onchange map[string]string `json:"field_onchange"`
}

// Onchange returns the values that must be modified in the pseudo-record
// given as params.Values
func Onchange(rc RecordCollection, params OnchangeParams) FieldMap {
	// TODO Implement Onchange
	return make(FieldMap)
}
