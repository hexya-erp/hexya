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

	"github.com/npiganeau/yep/yep/ir"
	"github.com/npiganeau/yep/yep/orm"
)

const (
	TRANSIENT_MODEL Option = 1 << iota
)

type BaseModel struct {
	ID         int64     `orm:"column(id)"`
	CreateDate time.Time `orm:"auto_now_add"`
	CreateUid  int64
	WriteDate  time.Time `yep:"compute(ComputeWriteDate),store,depends(ID)" orm:"null"`
	WriteUid   int64
}

type BaseTransientModel struct {
	ID int64 `orm:"column(id)"`
}

func declareBaseMethods(name string) {
	DeclareMethod(name, "ComputeWriteDate", ComputeWriteDate)
	DeclareMethod(name, "Create", Create)
	DeclareMethod(name, "Read", ReadModel)
	DeclareMethod(name, "NameGet", NameGet)
	DeclareMethod(name, "FieldsViewGet", FieldsViewGet)
}

/*
ComputeWriteDate updates the WriteDate field with the current datetime.
*/
func ComputeWriteDate(rs RecordSet) orm.Params {
	return orm.Params{"WriteDate": time.Now()}
}

/*
Create is the base implementation of the 'Create' method which creates
a record in the database from the given structPtr
*/
func Create(rs RecordSet, structPtr interface{}) RecordSet {
	return rs.Env().Create(structPtr)
}

/*
ReadModel is the base implementation of the 'Read' method.
It reads the database and returns a list of maps[string]interface{}
of the given
*/
func ReadModel(rs RecordSet, fields []string) []orm.Params {
	var res []orm.Params
	// Add id field to the list
	fList := []string{"id"}
	if fields != nil {
		fList = append(fList, fields...)
	}
	// Get the values
	rs.Values(&res, fList...)
	// Postprocessing results
	for _, line := range res {
		for k, v := range line {
			if strings.HasSuffix(k, orm.ExprSep) {
				// Add display name to rel/reverse fields
				path := strings.TrimRight(k, orm.ExprSep)
				id := v.(int64)
				relModelName := getRelatedModelName(rs.ModelName(), fmt.Sprintf("%s%sid", path, orm.ExprSep))
				relRS := NewRecordSet(rs.Env(), relModelName).Filter("id", id).Search()
				delete(line, k)
				line[path] = [2]interface{}{id, relRS.Call("NameGet").(string)}
			}
		}
	}
	return res
}

/*
NameGet is the base implementation of the 'NameGet' method which retrieves the
human readable name of an object.
*/
func NameGet(rs RecordSet) string {
	rs.EnsureOne()
	var idParams orm.ParamsList
	num := rs.ValuesFlat(&idParams, "name")
	if num == 0 {
		return rs.String()
	}
	return idParams[0].(string)
}

// Args struct for the FieldsViewGet function
type FieldsViewGetParams struct {
	ViewID   string `json:"view_id"`
	ViewType string `json:"view_type"`
	Toolbar  bool   `json:"toolbar"`
}

// Return type string for the FieldsViewGet function
type FieldsViewData struct {
	Name        string                  `json:"name"`
	Arch        string                  `json:"arch"`
	ViewID      string                  `json:"view_id"`
	Model       string                  `json:"model"`
	Type        ir.ViewType             `json:"type"`
	Fields      map[string]ir.FieldInfo `json:"fields"`
	Toolbar     ir.Toolbar              `json:"toolbar"`
	FieldParent string                  `json:"field_parent"`
}

/*
FieldsViewGet is the base implementation of the 'FieldsViewGet' method which
gets the detailed composition of the requested view like fields, model,
view architecture.
*/
func FieldsViewGet(rs RecordSet, args FieldsViewGetParams) *FieldsViewData {
	view := ir.ViewsRegistry.GetViewById(args.ViewID)
	res := FieldsViewData{
		Name:   view.Name,
		Arch:   view.Arch,
		ViewID: args.ViewID,
		Model:  view.Model,
		Type:   view.Type,
		Fields: view.Fields,
	}
	return &res
}
