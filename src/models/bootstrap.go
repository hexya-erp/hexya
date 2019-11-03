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

const freeTransientPeriod = 1 * time.Minute

var mixed = map[modelCouple]bool{}

// BootStrap freezes model, fields and method caches and syncs the database structure
// with the declared data.
func BootStrap() {
	log.Info("Bootstrapping models")
	if Registry.bootstrapped == true {
		log.Panic("Trying to bootstrap models twice !")
	}
	// loadManualSequencesFromDB locks registry, so we call it first
	loadManualSequencesFromDB()

	Registry.Lock()
	defer Registry.Unlock()

	inflateMixIns()
	createModelLinks()
	inflateEmbeddings()
	processUpdates()
	updateFieldDefs()
	updateRelatedPaths()
	syncRelatedFieldInfo()
	inflateContexts()
	updateRelatedPaths()
	updateDefaultOrder()
	bootStrapMethods()
	processDepends()
	checkFieldMethodsExist()
	checkComputeMethodsSignature()
	setupSecurity()
	RegisterWorker(NewWorkerFunction(FreeTransientModels, freeTransientPeriod))

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
		if newFI.fieldType == fieldtype.Many2Many {
			m2mRelModel, m2mOurField, m2mTheirField := createM2MRelModelInfo(newFI.m2mRelModel.name, model.name,
				newFI.relatedModelName, newFI.m2mOurField.name, newFI.m2mTheirField.name, false)
			newFI.m2mRelModel = m2mRelModel
			newFI.m2mOurField = m2mOurField
			newFI.m2mTheirField = m2mTheirField
		}
		model.fields.add(&newFI)
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
					name:           relName,
					json:           relFI.json,
					model:          model,
					stored:         fi.stored,
					structField:    relFI.structField,
					noCopy:         true,
					relatedPathStr: fmt.Sprintf("%s%s%s", fi.name, ExprSep, relName),
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
			newFI.relatedPathStr = fi.relatedPathStr
			newFI.stored = fi.stored
			newFI.model = mi
			newFI.noCopy = true
			newFI.onChange = ""
			newFI.onChangeWarning = ""
			newFI.onChangeFilters = ""
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
			fName := fmt.Sprintf("%sHexyaContexts", fi.name)
			o2mField := &Field{
				name:             fName,
				json:             strutils.SnakeCase(fName),
				model:            mi,
				fieldType:        fieldtype.One2Many,
				relatedModelName: contextsModel.name,
				relatedModel:     contextsModel,
				reverseFK:        "Record",
				jsonReverseFK:    "record_id",
				structField: reflect.StructField{
					Name: fName,
					Type: reflect.TypeOf([]int64{}),
				},
			}
			mi.fields.add(o2mField)
			relPath := fmt.Sprintf("%s%s%s", fName, ExprSep, fi.name)
			fi.relatedPathStr = relPath
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
	}
	updateContextModelsSecurity()
}

// updateContextModelsSecurity synchronizes the methods permissions of context models with their base model.
func updateContextModelsSecurity() {
	for _, model := range Registry.registryByName {
		if !model.isContext() {
			continue
		}
		baseModel := model.fields.MustGet("Record").relatedModel
		for _, methName := range []string{"Create", "Load", "Write", "Unlink"} {
			method := model.methods.MustGet(methName)
			method.AllowGroup(security.GroupEveryone, baseModel.methods.MustGet(methName))
			for grp := range baseModel.methods.MustGet(methName).groups {
				method.AllowGroup(grp)
			}
			for cGroup := range baseModel.methods.MustGet(methName).groupsCallers {
				method.AllowGroup(cGroup.group, cGroup.caller)
			}
		}
		model.methods.MustGet("Load").AllowGroup(security.GroupEveryone, baseModel.methods.MustGet("Create"))
	}
}

// updateRelatedPaths sets relatedPath from relatedPathStr
func updateRelatedPaths() {
	for _, model := range Registry.registryByName {
		for _, field := range model.fields.registryByName {
			if field.relatedPathStr != "" {
				field.relatedPath = model.FieldName(field.relatedPathStr)
			}
		}
	}
}

// updateDefaultOrder sets defaultOrder from defaultOrderStr
func updateDefaultOrder() {
	for _, model := range Registry.registryByName {
		if model.isM2MLink() {
			continue
		}
		model.defaultOrder = model.ordersFromStrings(model.defaultOrderStr)
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
			if field.onChangeWarning != "" {
				model.methods.MustGet(field.onChangeWarning)
			}
			if field.onChangeFilters != "" {
				model.methods.MustGet(field.onChangeFilters)
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

// loadManualSequencesFromDB fetches manual sequences from DB and updates registry
func loadManualSequencesFromDB() {
	if db == nil {
		// Happens when bootstrapping models without DB for tests
		return
	}
	adapter := adapters[db.DriverName()]
	for _, dbSeq := range adapter.sequences("%_manseq") {
		seq := &Sequence{
			JSON:      dbSeq.Name,
			Start:     dbSeq.StartValue,
			Increment: dbSeq.Increment,
		}
		Registry.addSequence(seq)
	}
}
