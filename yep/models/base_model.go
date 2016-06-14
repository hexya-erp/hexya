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

import "time"

const (
	TRANSIENT_MODEL Option = 1 << iota
)

type BaseModel struct {
	ID          int64     `orm:"column(id)"`
	CreateDate  time.Time `orm:"auto_now_add"`
	CreateUid   int64
	WriteDate   time.Time `yep:"compute(ComputeWriteDate);store;depends(ID)" orm:"null"`
	WriteUid    int64
	DisplayName string `orm:"-" yep:"compute(ComputeNameGet)"`
}

type BaseTransientModel struct {
	ID int64 `orm:"column(id)"`
}

func declareBaseMethods(name string) {
	//DeclareMethod(name, "ComputeWriteDate", ComputeWriteDate)
	//DeclareMethod(name, "ComputeNameGet", ComputeNameGet)
	//DeclareMethod(name, "Create", Create)
	//DeclareMethod(name, "Read", ReadModel)
	//DeclareMethod(name, "Write", Write)
	//DeclareMethod(name, "Unlink", Unlink)
	//DeclareMethod(name, "NameGet", NameGet)
	//DeclareMethod(name, "FieldsViewGet", FieldsViewGet)
	//DeclareMethod(name, "FieldsGet", FieldsGet)
	//DeclareMethod(name, "ProcessView", ProcessView)
	//DeclareMethod(name, "AddModifiers", AddModifiers)
	//DeclareMethod(name, "UpdateFieldNames", UpdateFieldNames)
	//DeclareMethod(name, "SearchRead", SearchRead)
	//DeclareMethod(name, "DefaultGet", DefaultGet)
	//DeclareMethod(name, "Onchange", Onchange)
}

///*
//ComputeWriteDate updates the WriteDate field with the current datetime.
//*/
//func ComputeWriteDate(rs RecordSet) orm.Params {
//	return orm.Params{"WriteDate": time.Now()}
//}
//
///*
//ComputeNameGet updates the DisplayName field with the result of NameGet.
//*/
//func ComputeNameGet(rs RecordSet) orm.Params {
//	return orm.Params{"DisplayName": rs.Call("NameGet").(string)}
//}
//
///*
//Create is the base implementation of the 'Create' method which creates
//a record in the database from the given structPtr.
//Returns the id of the created record and the associated RecordSet.
//*/
//func Create(rs RecordSet, data interface{}) RecordSet {
//	if md, ok := data.(map[string]interface{}); ok {
//		params := orm.Params(md)
//		id, err := rs.Env().Cr().InsertValues(rs.ModelName(), params)
//		if err != nil {
//			panic(fmt.Errorf("Create error: %s", err))
//		}
//		rStruct := rs.Filter("id", id).Search().(recordStruct)
//		rStruct.updateStoredFields(params)
//		rStruct.computeFieldValues(&params)
//		return rStruct
//	} else {
//		return rs.Env().Create(data)
//	}
//}
//
///*
//ReadModel is the base implementation of the 'Read' method.
//It reads the database and returns a list of maps[string]interface{}
//of the given model
//*/
//func ReadModel(rs RecordSet, fields []string) []orm.Params {
//	var res []orm.Params
//	// Add id field to the list
//	fList := []string{"id"}
//	if fields != nil {
//		fList = append(fList, fields...)
//	}
//	// Get the values
//	rs.Values(&res, fList...)
//	// Postprocessing results
//	for _, line := range res {
//		for k, v := range line {
//			if strings.HasSuffix(k, orm.ExprSep) {
//				// Add display name to rel/reverse fields
//				path := strings.TrimRight(k, orm.ExprSep)
//				id := v.(int64)
//				relModelName := getRelatedModelName(rs.ModelName(), fmt.Sprintf("%s%sid", path, orm.ExprSep))
//				relRS := NewRecordSet(rs.Env(), relModelName).Filter("id", id).Search()
//				delete(line, k)
//				line[path] = [2]interface{}{id, relRS.Call("NameGet").(string)}
//			}
//		}
//	}
//	return res
//}
//
///*
//Write is the base implementation of the 'Write' method which updates
//records in the database with the given params.
//*/
//func Write(rs RecordSet, data orm.Params) bool {
//	return rs.Write(data)
//}
//
///*
//Unlink is the base implementation of the 'Unlink' method which deletes
//records in the database.
//*/
//func Unlink(rs RecordSet) int64 {
//	return rs.Unlink()
//}
//
///*
//NameGet is the base implementation of the 'NameGet' method which retrieves the
//human readable name of an object.
//*/
//func NameGet(rs RecordSet) string {
//	rs.EnsureOne()
//	var idParams orm.ParamsList
//	ref := fieldRef{modelName: rs.ModelName(), name: "name"}
//	_, exists := fieldsCache.get(ref)
//	if !exists {
//		return rs.String()
//	}
//	num := rs.ValuesFlat(&idParams, "name")
//	if num == 0 {
//		return rs.String()
//	}
//	return idParams[0].(string)
//}
//
//// Args struct for the FieldsViewGet function
//type FieldsViewGetParams struct {
//	ViewID   string `json:"view_id"`
//	ViewType string `json:"view_type"`
//	Toolbar  bool   `json:"toolbar"`
//}
//
//// Return type string for the FieldsViewGet function
//type FieldsViewData struct {
//	Name        string                `json:"name"`
//	Arch        string                `json:"arch"`
//	ViewID      string                `json:"view_id"`
//	Model       string                `json:"model"`
//	Type        ir.ViewType           `json:"type"`
//	Fields      map[string]*FieldInfo `json:"fields"`
//	Toolbar     ir.Toolbar            `json:"toolbar"`
//	FieldParent string                `json:"field_parent"`
//}
//
//// Exportable field information struct
//type FieldInfo struct {
//	ChangeDefault    bool                   `json:"change_default"`
//	Help             string                 `json:"help"`
//	Searchable       bool                   `json:"searchable"`
//	Views            map[string]interface{} `json:"views"`
//	Required         bool                   `json:"required"`
//	Manual           bool                   `json:"manual"`
//	ReadOnly         bool                   `json:"readonly"`
//	Depends          []string               `json:"depends"`
//	CompanyDependent bool                   `json:"company_dependent"`
//	Sortable         bool                   `json:"sortable"`
//	Translate        bool                   `json:"translate"`
//	Type             tools.FieldType        `json:"type"`
//	Store            bool                   `json:"store"`
//	String           string                 `json:"string"`
//}
//
///*
//FieldsViewGet is the base implementation of the 'FieldsViewGet' method which
//gets the detailed composition of the requested view like fields, model,
//view architecture.
//*/
//func FieldsViewGet(rs RecordSet, args FieldsViewGetParams) *FieldsViewData {
//	view := ir.ViewsRegistry.GetViewById(args.ViewID)
//	if view == nil {
//		view = ir.ViewsRegistry.GetFirstViewForModel(rs.ModelName(), ir.ViewType(args.ViewType))
//	}
//	cols := make([]string, len(view.Fields))
//	for i, f := range view.Fields {
//		cols[i] = getFieldColumn(rs.ModelName(), f)
//	}
//	fInfos := rs.Call("FieldsGet", FieldsGetArgs{AllFields: cols}).(map[string]*FieldInfo)
//	arch := rs.Call("ProcessView", view.Arch, fInfos).(string)
//	res := FieldsViewData{
//		Name:   view.Name,
//		Arch:   arch,
//		ViewID: args.ViewID,
//		Model:  view.Model,
//		Type:   view.Type,
//		Fields: fInfos,
//	}
//	return &res
//}
//
///*
//Process view makes all the necessary modifications to the view
//arch and returns the new xml string.
//*/
//func ProcessView(rs RecordSet, arch string, fieldInfos map[string]*FieldInfo) string {
//	// Load arch as etree
//	doc := etree.NewDocument()
//	if err := doc.ReadFromString(arch); err != nil {
//		panic(fmt.Errorf("<ProcessView>%s", err))
//	}
//	// Apply changes
//	rs.Call("UpdateFieldNames", doc)
//	rs.Call("AddModifiers", doc, fieldInfos)
//	// Dump xml to string and return
//	res, err := doc.WriteToString()
//	if err != nil {
//		panic(fmt.Errorf("<ProcessView>Unable to render XML: %s", err))
//	}
//	return res
//}
//
///*
//AddModifiers adds the modifiers attribute nodes to given xml doc.
//*/
//func AddModifiers(rs RecordSet, doc *etree.Document, fieldInfos map[string]*FieldInfo) {
//	for _, fieldTag := range doc.FindElements("//field") {
//		fieldName := fieldTag.SelectAttr("name").Value
//		var mods []string
//		if fieldInfos[fieldName].ReadOnly {
//			mods = append(mods, "&quot;readonly&quot;: true")
//		}
//		modStr := fmt.Sprintf("{%s}", strings.Join(mods, ","))
//		fieldTag.CreateAttr("modifiers", modStr)
//	}
//}
//
///*
//UpdateFieldNames changes the field names in the view to the column names.
//If a field name is already column names then it does nothing.
//*/
//func UpdateFieldNames(rs RecordSet, doc *etree.Document) {
//	for _, fieldTag := range doc.FindElements("//field") {
//		fieldName := fieldTag.SelectAttr("name").Value
//		jsonName := GetFieldJSON(rs.ModelName(), fieldName)
//		fieldTag.RemoveAttr("name")
//		fieldTag.CreateAttr("name", jsonName)
//	}
//}
//
//// Args for the FieldsGet function
//type FieldsGetArgs struct {
//	// list of fields to document, all if empty or not provided
//	AllFields []string `json:"allfields"`
//}
//
///*
//FieldsGet returns the definition of each field.
//The _inherits'd fields are included. The string, help,
//and selection (if present) attributes are translated. (TODO)
//*/
//func FieldsGet(rs RecordSet, args FieldsGetArgs) map[string]*FieldInfo {
//	res := make(map[string]*FieldInfo)
//	fields := args.AllFields
//	if len(args.AllFields) == 0 {
//		ormFields := orm.FieldsGet(rs.ModelName())
//		for fi, v := range ormFields {
//			if v.FieldType != orm.RelReverseMany {
//				// We don't want ReverseMany as it points to the link table
//				fields = append(fields, fi)
//			}
//		}
//	}
//	for _, f := range fields {
//		key := fieldRef{modelName: rs.ModelName(), name: f}
//		fInfo, ok := fieldsCache.get(key)
//		if !ok {
//			panic(fmt.Errorf("Unknown field: %+v", key))
//		}
//		res[fInfo.json] = &FieldInfo{
//			Help:       fInfo.help,
//			Searchable: true,
//			Depends:    fInfo.depends,
//			Sortable:   true,
//			Type:       fInfo.fieldType,
//			Store:      fInfo.stored,
//			String:     fInfo.description,
//		}
//	}
//	return res
//}
//
type SearchParams struct {
	Domain Domain      `json:"domain"`
	Fields []string    `json:"fields"`
	Offset int         `json:"offset"`
	Limit  interface{} `json:"limit"`
	Order  string      `json:"order"`
}

//
///*
//SearchRead retrieves database records according to the filters defined in params.
//*/
//func SearchRead(rs RecordSet, params SearchParams) []orm.Params {
//	var fields []string
//	for _, f := range params.Fields {
//		fields = append(fields, GetFieldName(rs.ModelName(), f))
//	}
//	searchCond := ParseDomain(params.Domain)
//	rs = rs.SetCond(searchCond)
//	// TODO Add support for ordering & offset
//	var limit int
//	switch params.Limit.(type) {
//	case bool:
//		limit = -1
//	case int:
//		limit = params.Limit.(int)
//	default:
//		limit = 80
//	}
//	rs = rs.Limit(limit).Search()
//	return rs.Call("Read", fields).([]orm.Params)
//}
//
///*
//DefaultGet returns a Params map with the default values for the model.
//*/
//func DefaultGet(rs RecordSet) orm.Params {
//	// TODO Implement DefaultGet
//	return make(orm.Params)
//}
//
//type OnchangeParams struct {
//	Values   orm.Params        `json:"values"`
//	Fields   []string          `json:"field_name"`
//	Onchange map[string]string `json:"field_onchange"`
//}
//
///*
//Onchange returns the values that must be modified in the pseudo-record given as params.Values
//*/
//func Onchange(rs RecordSet, params OnchangeParams) orm.Params {
//	// TODO Implement Onchange
//	return make(orm.Params)
//}
