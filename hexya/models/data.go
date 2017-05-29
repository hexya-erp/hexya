// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"encoding/csv"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
)

// LoadCSVDataFile loads the data of the given file into the database.
func LoadCSVDataFile(fileName string) {
	csvFile, err := os.Open(fileName)
	defer csvFile.Close()
	if err != nil {
		log.Panic("Unable to open CSV data file", "error", err, "fileName", fileName)
	}

	elements := strings.Split(path.Base(fileName), "_")
	modelName := strings.Split(elements[0], ".")[0]
	var (
		update  bool
		version int
	)
	if len(elements) == 2 {
		mod := strings.Split(elements[1], ".")[0]
		ver, err := strconv.Atoi(mod)
		switch {
		case strings.ToLower(mod) == "update":
			update = true
		case err == nil:
			version = ver
		}
	}

	r := csv.NewReader(csvFile)
	headers, err := r.Read()
	if err != nil {
		log.Panic("Unable to read CSV headers in data file", "error", err, "fileName", fileName)
	}

	err = ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
		rc := env.Pool(modelName)
		// JSONize all field names
		for i, header := range headers {
			headers[i] = rc.Model().JSONizeFieldName(header)
		}
		line := 1
		// Load records
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}

			values := getRecordValuesMap(headers, modelName, record, env, line)

			externalID := values["id"]
			delete(values, "id")
			values["hexya_external_id"] = externalID
			values["hexya_version"] = version
			rec := rc.Call("Search", rc.Model().Field("HexyaExternalID").Equals(externalID)).(RecordCollection).Limit(1)
			switch {
			case rec.Len() == 0:
				rc.Call("Create", values)
			case rec.Len() == 1:
				if version > rec.Get("HexyaVersion").(int) || update {
					rec.Call("Write", values)
				}
			}
			line++
		}
	})
	if err != nil {
		log.Panic("Error while loading data", "error", err)
	}
}

func getRecordValuesMap(headers []string, modelName string, record []string, env Environment, line int) FieldMap {
	values := make(map[string]interface{})
	for i := 0; i < len(headers); i++ {
		fi := Registry.MustGet(modelName).getRelatedFieldInfo(headers[i])
		var (
			val interface{}
			err error
		)
		switch {
		case headers[i] == "id":
			val = record[i]
		case fi.fieldType == fieldtype.Integer:
			val, err = strconv.ParseInt(record[i], 0, 64)
			if err != nil {
				log.Panic("Error while converting integer", "line", line, "field", headers[i], "value", record[i], "error", err)
			}
		case fi.fieldType == fieldtype.Float:
			val, err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				log.Panic("Error while converting float", "line", line, "field", headers[i], "value", record[i], "error", err)
			}
		case fi.fieldType.IsFKRelationType():
			relRC := env.Pool(fi.relatedModelName).Search(fi.relatedModel.Field("HexyaExternalID").Equals(record[i]))
			if relRC.Len() != 1 {
				log.Panic("Unable to find related record from external ID", "line", line, "field", headers[i], "value", record[i])
			}
			val = relRC.Ids()[0]
		case fi.fieldType == fieldtype.Many2Many:
			ids := strings.Split(record[i], "|")
			relRC := env.Pool(fi.relatedModelName).Search(fi.relatedModel.Field("HexyaExternalID").In(ids))
			val = relRC.Ids()
		default:
			val = record[i]
		}
		values[headers[i]] = val
	}
	return values
}
