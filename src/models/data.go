// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"encoding/base64"
	"encoding/csv"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/security"
)

// LoadCSVDataFile loads the data of the given file into the database.
func LoadCSVDataFile(fileName string) {
	log.Info("Importing data file", "fileName", fileName)
	csvFile, err := os.Open(fileName)
	if err != nil {
		log.Panic("Unable to open CSV data file", "error", err, "fileName", fileName)
	}
	defer csvFile.Close()

	elements := strings.Split(filepath.Base(fileName), "_")
	modelName := strings.Split(elements[0], ".")[0]
	modelName = strings.TrimLeft(modelName, "01234567890-")
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

			values := getRecordValuesMap(headers, modelName, record, env, line, fileName)

			externalID := values["id"]
			delete(values, "id")
			values["hexya_external_id"] = externalID
			values["hexya_version"] = version
			// We deliberately call Search directly without Call so as not to be polluted by Search overrides
			// such as "Active test".
			rec := rc.Search(rc.Model().Field("HexyaExternalID").Equals(externalID)).Limit(1)
			switch {
			case rec.Len() == 0:
				rc.applyDefaults(&values, true)
				rc.Call("Create", NewModelData(rc.model, values))
			case rec.Len() == 1:
				if version > rec.Get("HexyaVersion").(int) || update {
					rec.Call("Write", NewModelData(rc.model, values))
				}
			}
			line++
		}
	})
	if err != nil {
		panic(err)
	}
	log.Debug("Data file imported successfully", "fileName", fileName)
}

func getRecordValuesMap(headers []string, modelName string, record []string, env Environment, line int, fileName string) FieldMap {
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
				log.Panic("Error while converting integer", "fileName", fileName, "line", line, "field", headers[i], "value", record[i], "error", err)
			}
		case fi.fieldType == fieldtype.Float:
			val, err = strconv.ParseFloat(record[i], 64)
			if err != nil {
				log.Panic("Error while converting float", "fileName", fileName, "line", line, "field", headers[i], "value", record[i], "error", err)
			}
		case fi.fieldType.IsFKRelationType():
			val = env.Pool(fi.relatedModelName)
			if record[i] != "" {
				relRC := env.Pool(fi.relatedModelName).Search(fi.relatedModel.Field("HexyaExternalID").Equals(record[i]))
				if relRC.Len() != 1 {
					log.Panic("Unable to find related record from external ID", "fileName", fileName, "line", line, "field", headers[i], "value", record[i])
				}
				val = relRC
			}
		case fi.fieldType == fieldtype.Many2Many:
			ids := strings.Split(record[i], "|")
			relRC := env.Pool(fi.relatedModelName).Search(fi.relatedModel.Field("HexyaExternalID").In(ids))
			val = relRC
		case fi.fieldType == fieldtype.Binary:
			if record[i] == "" {
				continue
			}
			dir := filepath.Dir(fileName)
			bFileName := filepath.Join(dir, record[i])
			fileContent, err := ioutil.ReadFile(bFileName)
			if err != nil {
				log.Panic("Unable to open file with binary data", "error", err, "line", line, "field", headers[i], "value", record[i])
			}
			val = base64.StdEncoding.EncodeToString(fileContent)
		case fi.fieldType == fieldtype.Boolean:
			val = false
			if res, _ := strconv.ParseBool(record[i]); res {
				val = true
			}
		default:
			val = record[i]
		}
		values[headers[i]] = val
	}
	return values
}
