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
	"time"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/tools/strutils"
)

// A modelCouple holds a model and one of its mixin
type modelCouple struct {
	model *Model
	mixIn *Model
}

const workloopPeriod = 1 * time.Minute

var (
	mixed      = map[modelCouple]bool{}
	workerStop chan bool
)

// BootStrap freezes model, fields and method caches and syncs the database structure
// with the declared data.
func BootStrap() {
	log.Info("Bootstrapping models")
	if Registry.bootstrapped == true {
		log.Panic("Trying to bootstrap models twice !")
	}
	Registry.Lock()
	defer Registry.Unlock()

	inflateMixIns()
	createModelLinks()
	inflateEmbeddings()
	processUpdates()
	updateFieldDefs()
	syncRelatedFieldInfo()
	inflateContexts()
	bootStrapMethods()
	processDepends()
	checkFieldMethodsExist()
	checkComputeMethodsSignature()
	setupSecurity()
	workloop(workloopPeriod)

	Registry.bootstrapped = true
}

// BootStrapped returns true if the models have been bootstrapped
func BootStrapped() bool {
	return Registry.bootstrapped
}

// processUpdates applies all the directives of the update map to the fields
func processUpdates() {
	for _, model := range Registry.registryByName {
		for _, fi := range model.fields.registryByName {
			for _, update := range fi.updates {
				for property, value := range update {
					switch property {
					case "selection_add":
						for k, v := range value.(types.Selection) {
							fi.selection[k] = v
						}
					case "contexts_add":
						for k, v := range value.(FieldContexts) {
							fi.contexts[k] = v
						}
					default:
						fi.setProperty(property, value)
					}
				}
			}
			fi.updates = nil
		}
	}
}

// updateFieldDefs updates fields definitions if necessary
func updateFieldDefs() {
	for _, model := range Registry.registryByName {
		for _, fi := range model.fields.registryByName {
			switch fi.fieldType {
			case fieldtype.Boolean:
				if fi.defaultFunc != nil && fi.isSettable() {
					fi.required = true
				}
			case fieldtype.Selection:
				if fi.selectionFunc != nil {
					fi.selection = fi.selectionFunc()
				}
			}
		}
	}
}

// createModelLinks create links with related Model
// where applicable. Also populates jsonReverseFK field
func createModelLinks() {
	for _, mi := range Registry.registryByName {
		for _, fi := range mi.fields.registryByName {
			var (
				relatedMI *Model
				ok        bool
			)
			if !fi.fieldType.IsRelationType() {
				continue
			}
			relatedMI, ok = Registry.Get(fi.relatedModelName)
			if !ok {
				log.Panic("Unknown related model in field declaration", "model", mi.name, "field", fi.name, "relatedName", fi.relatedModelName)
			}
			if fi.fieldType.IsReverseRelationType() {
				fi.jsonReverseFK = relatedMI.fields.MustGet(fi.reverseFK).json
			}
			fi.relatedModel = relatedMI
		}
		mi.fields.bootstrapped = true
	}
}

// inflateMixIns inserts fields and methods of mixed in models.
func inflateMixIns() {
	for _, mi := range Registry.registryByName {
		if mi.isM2MLink() {
			// We don"t mix in M2M link
			continue
		}
		for _, mixInMI := range mi.mixins {
			injectMixInModel(mixInMI, mi)
		}
	}
}

// injectMixInModel injects fields and methods of mixInMI in model
func injectMixInModel(mixInMI, mi *Model) {
	for _, mmm := range mixInMI.mixins {
		injectMixInModel(mmm, mixInMI)
	}
	if mixed[modelCouple{model: mi, mixIn: mixInMI}] {
		return
	}
	// Add mixIn fields
	addMixinFields(mixInMI, mi)
	// Add mixIn methods
	addMixinMethods(mixInMI, mi)
	mixed[modelCouple{model: mi, mixIn: mixInMI}] = true
}

// addMixinMethods adds the methodss of mixinModel to model
func addMixinMethods(mixinModel, model *Model) {
	for methName, methInfo := range mixinModel.methods.registry {
		// Extract all method layers functions by inverse order
		layersInv := methInfo.invertedLayers()
		if emi, exists := model.methods.registry[methName]; exists {
			// The method already exists in our target model
			// We insert our new method layers above previous mixins layers
			// but below the target model implementations.
			lastImplLayer := emi.topLayer
			firstMixedLayer := emi.getNextLayer(lastImplLayer)
			for firstMixedLayer != nil {
				if firstMixedLayer.mixedIn {
					break
				}
				lastImplLayer = firstMixedLayer
				firstMixedLayer = emi.getNextLayer(lastImplLayer)
			}
			for _, lf := range layersInv {
				ml := methodLayer{
					funcValue: wrapFunctionForMethodLayer(lf.funcValue),
					mixedIn:   true,
					method:    emi,
				}
				emi.nextLayer[&ml] = firstMixedLayer
				firstMixedLayer = &ml
			}
			if emi.topLayer == nil {
				// The existing method was empty
				emi.topLayer = firstMixedLayer
				emi.methodType = methInfo.methodType
			} else {
				emi.nextLayer[lastImplLayer] = firstMixedLayer
			}
		} else {
			// The method does not exist
			newMethInfo := copyMethod(model, methInfo)
			for i := 0; i < len(layersInv); i++ {
				newMethInfo.addMethodLayer(layersInv[i].funcValue, layersInv[i].doc)
			}
			model.methods.set(methName, newMethInfo)
		}
		// Copy groups to our methods in the target model
		for group := range methInfo.groups {
			model.methods.MustGet(methName).groups[group] = true
		}
	}
}

// addMixinFields adds the fields of mixinModel into model
func addMixinFields(mixinModel, model *Model) {
	for fName, fi := range mixinModel.fields.registryByName {
		existingFI, exists := model.fields.registryByName[fName]
		newFI := *fi
		if exists {
			if existingFI.fieldType != fieldtype.NoType {
				// We do not add fields that already exist in the targetModel
				// since the target model should always override mixins.
				continue
			}
			// We extract updates from our DummyField and remove it from the registry
			newFI.updates = append(newFI.updates, existingFI.updates...)
			delete(model.fields.registryByJSON, existingFI.json)
			delete(model.fields.registryByName, existingFI.name)
		}
		newFI.model = model
		newFI.acl = security.NewAccessControlList()
		if newFI.fieldType == fieldtype.Many2Many {
			m2mRelModel, m2mOurField, m2mTheirField := createM2MRelModelInfo(newFI.m2mRelModel.name, model.name,
				newFI.relatedModelName, newFI.m2mOurField.name, newFI.m2mTheirField.name, false)
			newFI.m2mRelModel = m2mRelModel
			newFI.m2mOurField = m2mOurField
			newFI.m2mTheirField = m2mTheirField
		}
		model.fields.add(&newFI)
		// We add the permissions of the mixin to the target model
		for group, perm := range fi.acl.Permissions() {
			newFI.acl.AddPermission(group, perm)
		}
	}
}

// inflateEmbeddings creates related fields for all fields of embedded models.
func inflateEmbeddings() {
	for _, model := range Registry.registryByName {
		for _, fi := range model.fields.registryByName {
			if !fi.embed {
				continue
			}
			for relName, relFI := range fi.relatedModel.fields.registryByName {
				if relFI.relatedModelName == model.name && relFI.jsonReverseFK != "" && relFI.jsonReverseFK == fi.name {
					// We do not add reverse fields to our own model
					continue
				}
				newFI := Field{
					name:        relName,
					json:        relFI.json,
					acl:         security.NewAccessControlList(),
					model:       model,
					stored:      fi.stored,
					structField: relFI.structField,
					noCopy:      true,
					relatedPath: fmt.Sprintf("%s%s%s", fi.name, ExprSep, relName),
				}
				if existingFI, ok := model.fields.Get(relName); ok {
					if existingFI.fieldType != fieldtype.NoType {
						// We do not add fields that already exist in the targetModel
						// since the target model should always override embedded fields.
						continue
					}
					// We extract updates from our DummyField and remove it from the registry
					newFI.updates = append(newFI.updates, existingFI.updates...)
					delete(model.fields.registryByJSON, existingFI.json)
					delete(model.fields.registryByName, existingFI.name)
				}
				model.fields.add(&newFI)
			}
		}
	}
}

// syncRelatedFieldInfo overwrites the Field data of the related fields
// with the data of the Field of the target.
func syncRelatedFieldInfo() {
	for _, mi := range Registry.registryByName {
		for _, fi := range mi.fields.registryByName {
			if !fi.isRelatedField() {
				continue
			}
			newFI := *mi.getRelatedFieldInfo(fi.relatedPath)
			newFI.name = fi.name
			newFI.json = fi.json
			newFI.relatedPath = fi.relatedPath
			newFI.stored = fi.stored
			newFI.model = mi
			newFI.noCopy = true
			newFI.onChange = ""
			newFI.index = false
			newFI.compute = ""
			newFI.constraint = ""
			newFI.inverse = ""
			newFI.depends = nil
			newFI.contexts = nil
			*fi = newFI
		}
	}
}

// inflateContexts creates the field value tables for fields with contexts.
func inflateContexts() {
	for _, mi := range Registry.registryByName {
		for _, fi := range mi.fields.registryByName {
			if !fi.isContextedField() {
				continue
			}
			contextsModel := createContextsModel(fi, fi.contexts)
			createContextsTreeView(fi, fi.contexts)
			// We copy execution permission on CRUD methods to the context model
			fieldName := fmt.Sprintf("%sHexyaContexts", fi.name)
			o2mField := &Field{
				name:             fieldName,
				json:             strutils.SnakeCase(fieldName),
				acl:              security.NewAccessControlList(),
				model:            mi,
				fieldType:        fieldtype.One2Many,
				relatedModelName: contextsModel.name,
				relatedModel:     contextsModel,
				reverseFK:        "Record",
				jsonReverseFK:    "record_id",
				structField: reflect.StructField{
					Name: fieldName,
					Type: reflect.TypeOf([]int64{}),
				},
			}
			mi.fields.add(o2mField)
			fi.relatedPath = fmt.Sprintf("%s%s%s", fieldName, ExprSep, fi.name)
			fi.index = false
			fi.unique = false
		}
	}
}

// createContextsTreeView creates an editable tree view for the given context model.
// The created view is added to the Views map which will be processed by the views package at bootstrap.
func createContextsTreeView(fi *Field, contexts FieldContexts) {
	arch := strings.Builder{}
	arch.WriteString("<tree editable=\"bottom\" create=\"false\" delete=\"false\">\n")
	for ctx := range contexts {
		arch.WriteString("	<field name=\"")
		arch.WriteString(ctx)
		arch.WriteString("\" readonly=\"1\"/>\n")
	}
	arch.WriteString(" <field name=\"")
	arch.WriteString(fi.json)
	arch.WriteString("\"/>\n")
	arch.WriteString("</tree>")

	modelName := fmt.Sprintf("%sHexya%s", fi.model.name, fi.name)
	view := fmt.Sprintf(`<view id="%s_hexya_contexts_tree" model="%s">
%s
</view>`, strutils.SnakeCase(modelName), modelName, arch.String())
	Views[fi.model] = append(Views[fi.model], view)
}

// runInit runs the Init function of the given model if it exists
func runInit(model *Model) {
	if _, exists := model.methods.Get("Init"); exists {
		ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			env.Pool(model.name).Call("Init")
		})
	}
}

// SyncDatabase creates or updates database tables with the data in the model registry
func SyncDatabase() {
	log.Info("Updating database schema")
	adapter := adapters[db.DriverName()]
	dbTables := adapter.tables()
	// Create or update sequences
	updateDBSequences()
	// Create or update existing tables
	for tableName, model := range Registry.registryByTableName {
		if model.isMixin() {
			// Don't create table for mixin models
			continue
		}
		if model.isManual() {
			// Don't create table for manual models
			continue
		}
		if _, ok := dbTables[tableName]; !ok {
			createDBTable(model.tableName)
		}
		updateDBColumns(model)
		updateDBIndexes(model)
	}
	// Setup constraints
	for _, model := range Registry.registryByTableName {
		if model.isMixin() {
			continue
		}
		if model.isManual() {
			continue
		}
		buildSQLErrorSubstitutionMap(model)
		updateDBForeignKeyConstraints(model)
		updateDBConstraints(model)
	}
	// Run init method on each model
	for _, model := range Registry.registryByTableName {
		if model.isMixin() {
			continue
		}
		runInit(model)
	}

	// Drop DB tables that are not in the models
	for dbTable := range adapter.tables() {
		var modelExists bool
		for tableName, model := range Registry.registryByTableName {
			if dbTable != tableName {
				continue
			}
			if model.isMixin() {
				// We don't want a table for mixin models
				continue
			}
			modelExists = true
			break
		}
		if !modelExists {
			dropDBTable(dbTable)
		}
	}
}

// buildSQLErrorSubstitutionMap populates the sqlErrors map of the
// model with the appropriate error message substitution
func buildSQLErrorSubstitutionMap(model *Model) {
	for sqlConstrName, sqlConstr := range model.sqlConstraints {
		model.sqlErrors[sqlConstrName] = sqlConstr.errorString
	}
	for _, field := range model.fields.registryByJSON {
		if field.unique {
			cName := fmt.Sprintf("%s_%s_key", model.tableName, field.json)
			model.sqlErrors[cName] = fmt.Sprintf("%s must be unique", field.name)
		}
		if field.fieldType.IsFKRelationType() {
			cName := fmt.Sprintf("%s_%s_fkey", model.tableName, field.json)
			model.sqlErrors[cName] = fmt.Sprintf("%s must reference an existing %s record", field.name, field.relatedModelName)
		}
	}
}

// updateDBSequences creates sequences in the DB from data in the registry.
func updateDBSequences() {
	adapter := adapters[db.DriverName()]
	// Create or alter boot sequences
	for _, sequence := range Registry.sequences {
		if !sequence.boot {
			continue
		}
		exists := false
		for _, dbSeq := range adapter.sequences("%_manseq") {
			if sequence.JSON == dbSeq {
				exists = true
			}
		}
		if !exists {
			adapter.createSequence(sequence.JSON, sequence.Increment, sequence.Start)
			continue
		}
		adapter.alterSequence(sequence.JSON, sequence.Increment, sequence.Start)
	}
	// Drop unused sequences
	for _, dbSeq := range adapter.sequences("%_manseq") {
		var sequenceExists bool
		for _, sequence := range Registry.sequences {
			if sequence.JSON != dbSeq || !sequence.boot {
				continue
			}
			sequenceExists = true
			break
		}
		if !sequenceExists {
			adapter.dropSequence(dbSeq)
		}
	}
}

// createDBTable creates a table in the database from the given Model
// It only creates the primary key. Call updateDBColumns to create columns.
func createDBTable(tableName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
	CREATE TABLE %s (
		id serial NOT NULL PRIMARY KEY
	)
	`, adapter.quoteTableName(tableName))
	dbExecuteNoTx(query)
}

// dropDBTable drops the given table in the database
func dropDBTable(tableName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`DROP TABLE %s`, adapter.quoteTableName(tableName))
	dbExecuteNoTx(query)
}

// updateDBColumns synchronizes the colums of the database with the
// given Model.
func updateDBColumns(mi *Model) {
	adapter := adapters[db.DriverName()]
	dbColumns := adapter.columns(mi.tableName)
	// create or update columns from registry data
	for colName, fi := range mi.fields.registryByJSON {
		if colName == "id" || !fi.isStored() {
			continue
		}
		dbColData, ok := dbColumns[colName]
		if !ok {
			createDBColumn(fi)
		}
		if dbColData.DataType != adapter.typeSQL(fi) {
			updateDBColumnDataType(fi)
		}
		if (dbColData.IsNullable == "NO" && !adapter.fieldIsNotNull(fi)) ||
			(dbColData.IsNullable == "YES" && adapter.fieldIsNotNull(fi)) {
			updateDBColumnNullable(fi)
		}
	}
	// drop columns that no longer exist
	for colName := range dbColumns {
		if _, ok := mi.fields.registryByJSON[colName]; !ok {
			dropDBColumn(mi.tableName, colName)
		}
	}
}

// createDBColumn insert the column described by Field in the database
func createDBColumn(fi *Field) {
	if !fi.isStored() {
		log.Panic("createDBColumn should not be called on non stored fields", "model", fi.model.name, "field", fi.json)
	}
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ADD COLUMN %s %s
	`, adapter.quoteTableName(fi.model.tableName), fi.json, adapter.columnSQLDefinition(fi))
	dbExecuteNoTx(query)
}

// updateDBColumnDataType updates the data type in database for the given Field
func updateDBColumnDataType(fi *Field) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ALTER COLUMN %s SET DATA TYPE %s
	`, adapter.quoteTableName(fi.model.tableName), fi.json, adapter.typeSQL(fi))
	dbExecuteNoTx(query)
}

// updateDBColumnNullable updates the NULL/NOT NULL data in database for the given Field
func updateDBColumnNullable(fi *Field) {
	adapter := adapters[db.DriverName()]
	var verb string
	if adapter.fieldIsNotNull(fi) {
		verb = "SET"
	} else {
		verb = "DROP"
	}
	query := fmt.Sprintf(`
		ALTER TABLE %s
		ALTER COLUMN %s %s NOT NULL
	`, adapter.quoteTableName(fi.model.tableName), fi.json, verb)
	dbExecuteNoTx(query)
}

// dropDBColumn drops the column colName from table tableName in database
func dropDBColumn(tableName, colName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s
		DROP COLUMN %s
	`, adapter.quoteTableName(tableName), colName)
	dbExecuteNoTx(query)
}

// updateDBForeignKeyConstraints creates or updates fk constraints
// based on the data of the given Model
func updateDBForeignKeyConstraints(m *Model) {
	adapter := adapters[db.DriverName()]
	for colName, fi := range m.fields.registryByJSON {
		fkContraintInDB := adapter.constraintExists(fmt.Sprintf("%s_%s_fkey", m.tableName, colName))
		fieldIsFK := fi.fieldType.IsFKRelationType() && fi.isStored()
		switch {
		case fieldIsFK && !fkContraintInDB:
			createFKConstraint(m.tableName, colName, fi.relatedModel.tableName, string(fi.onDelete))
		case !fieldIsFK && fkContraintInDB:
			dropFKConstraint(m.tableName, colName)
		}
	}
}

// updateDBConstraints creates or updates sql constraints
// based on the data of the given Model
func updateDBConstraints(m *Model) {
	adapter := adapters[db.DriverName()]
	for constraintName, constraint := range m.sqlConstraints {
		if !adapter.constraintExists(constraintName) {
			createConstraint(m.tableName, constraintName, constraint.sql)
		}
	}
dbConLoop:
	for _, dbConstraintName := range adapter.constraints(fmt.Sprintf("%%_%s_mancon", m.tableName)) {
		for constraintName := range m.sqlConstraints {
			if constraintName == dbConstraintName {
				continue dbConLoop
			}
		}
		dropConstraint(m.tableName, dbConstraintName)
	}
}

// createFKConstraint creates an FK constraint for the given column that references the given targetTable
func createFKConstraint(tableName, colName, targetTable, ondelete string) {
	adapter := adapters[db.DriverName()]
	constraint := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s ON DELETE %s", colName, adapter.quoteTableName(targetTable), ondelete)
	createConstraint(tableName, fmt.Sprintf("%s_%s_fkey", tableName, colName), constraint)
}

// dropFKConstraint drops an FK constraint for colName in the given table
func dropFKConstraint(tableName, colName string) {
	dropConstraint(tableName, fmt.Sprintf("%s_%s_fkey", tableName, colName))
}

// createConstraint creates a constraint in the given table
func createConstraint(tableName, constraintName, sql string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s ADD CONSTRAINT %s %s
	`, adapter.quoteTableName(tableName), constraintName, sql)
	dbExecuteNoTx(query)
}

// dropConstraint drops a constraint with the given name
func dropConstraint(tableName, constraintName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s
	`, adapter.quoteTableName(tableName), constraintName)
	dbExecuteNoTx(query)
}

// updateDBIndexes creates or updates indexes based on the data of
// the given Model
func updateDBIndexes(m *Model) {
	adapter := adapters[db.DriverName()]
	for colName, fi := range m.fields.registryByJSON {
		indexInDB := adapter.indexExists(m.tableName, fmt.Sprintf("%s_%s_index", m.tableName, colName))
		switch {
		case fi.index && !indexInDB:
			createColumnIndex(m.tableName, colName)
		case indexInDB && !fi.index:
			dropColumnIndex(m.tableName, colName)
		}
	}
}

// createColumnIndex creates an column index for colName in the given table
func createColumnIndex(tableName, colName string) {
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`
		CREATE INDEX %s ON %s (%s)
	`, fmt.Sprintf("%s_%s_index", tableName, colName), adapter.quoteTableName(tableName), colName)
	dbExecuteNoTx(query)
}

// dropColumnIndex drops a column index for colName in the given table
func dropColumnIndex(tableName, colName string) {
	query := fmt.Sprintf(`
		DROP INDEX IF EXISTS %s
	`, fmt.Sprintf("%s_%s_index", tableName, colName))
	dbExecuteNoTx(query)
}

// bootStrapMethods freezes the methods of the models.
func bootStrapMethods() {
	for _, model := range Registry.registryByName {
		model.methods.bootstrapped = true
	}
}

// setupSecurity adds execution permission to:
// - the admin group for all methods
// - to CRUD methods to call "Load"
// - to "Create" method to call "Write"
// - to execute CRUD on context models
func setupSecurity() {
	for _, model := range Registry.registryByName {
		loadMeth, loadExists := model.methods.Get("Load")
		fetchMeth, fetchExists := model.methods.Get("Fetch")
		writeMeth, writeExists := model.methods.Get("Write")
		for _, meth := range model.methods.registry {
			meth.AllowGroup(security.GroupAdmin)
			if loadExists && unauthorizedMethods[meth.name] {
				loadMeth.AllowGroup(security.GroupEveryone, meth)
			}
			if writeExists && meth.name == "Create" {
				writeMeth.AllowGroup(security.GroupEveryone, meth)
			}
		}
		if fetchExists {
			loadMeth.AllowGroup(security.GroupEveryone, fetchMeth)
		}

		if model.isContext() {
			baseModel := model.fields.MustGet("Record").relatedModel
			model.methods.MustGet("Create").AllowGroup(security.GroupEveryone, baseModel.methods.MustGet("Create"))
			model.methods.MustGet("Load").AllowGroup(security.GroupEveryone, baseModel.methods.MustGet("Create"))
			model.methods.MustGet("Load").AllowGroup(security.GroupEveryone, baseModel.methods.MustGet("Load"))
			model.methods.MustGet("Write").AllowGroup(security.GroupEveryone, baseModel.methods.MustGet("Write"))
			model.methods.MustGet("Unlink").AllowGroup(security.GroupEveryone, baseModel.methods.MustGet("Unlink"))
		}
	}
}

// checkFieldMethodsExist checks that all methods referenced by fields,
// such as Compute, Constraint or Onchange exist.
func checkFieldMethodsExist() {
	for _, model := range Registry.registryByName {
		for _, field := range model.fields.registryByName {
			if field.onChange != "" {
				model.methods.MustGet(field.onChange)
			}
			if field.constraint != "" {
				model.methods.MustGet(field.constraint)
			}
			if field.compute != "" && field.stored {
				model.methods.MustGet(field.compute)
				if len(field.depends) == 0 {
					log.Warn("Computed fields should have a 'Depends' parameter set", "model", model.name, "field", field.name)
				}
			}
			if field.inverse != "" {
				if _, ok := model.methods.Get(field.compute); !ok {
					log.Panic("Inverse method must only be set on computed fields", "model", model.name, "field", field.name, "method", field.inverse)
				}
				model.methods.MustGet(field.inverse)
			}
		}
	}
}

// workloopMethods executes all methods that must be run regularly.
func workloopMethods() {
	go FreeTransientModels()
}

// workloop launches the hexya core worker loop.
func workloop(period time.Duration) {
	if workerStop == nil {
		workerStop = make(chan bool)
	}
	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				workloopMethods()
			case <-workerStop:
				return
			}
		}
	}()
}
