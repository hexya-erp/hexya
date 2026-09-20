// Copyright 2019 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
)

var (
	companyDependent = FieldContexts{
		"company": func(rs RecordSet) string {
			companyID := rs.Env().Context().GetInteger("force_company")
			if companyID == 0 {
				companyID = rs.Env().Context().GetInteger("company_id")
			}
			if companyID == 0 {
				return ""
			}
			return fmt.Sprintf("%d", companyID)
		},
	}
	userDependent = FieldContexts{
		"user": func(rs RecordSet) string {
			return "user"
		},
	}
)

type TestFieldMap FieldMap

func (f TestFieldMap) Underlying() FieldMap {
	return FieldMap(f)
}

func UnBootStrap() {
	Registry.bootstrapped = false
	for _, mi := range Registry.registryByName {
		if mi.options&ContextsModel > 0 {
			delete(Registry.registryByName, mi.name)
			delete(Registry.registryByTableName, mi.tableName)
			continue
		}
		for _, fi := range mi.fields.registryByName {
			if fi.contexts != nil && len(fi.contexts) > 0 {
				fi.relatedPathStr = ""
				fi.relatedPath = nil
				continue
			}
			if strings.HasSuffix(fi.name, "HexyaContexts") {
				delete(mi.fields.registryByName, fi.name)
				delete(mi.fields.registryByJSON, fi.json)
			}
		}
	}
	for seqName, seq := range Registry.sequences {
		if strings.HasSuffix(seq.JSON, "_manseq") {
			delete(Registry.sequences, seqName)
		}
	}
}

func checkUpdates(t *testing.T, f *Field, property string, value any) {
	assert.Greater(t, len(f.updates), 0)
	assert.Contains(t, f.updates[len(f.updates)-1], property)
	assert.EqualValues(t, f.updates[len(f.updates)-1][property], value)
}

func lastUpdateShouldResemble(t *testing.T, f *Field, key string, s any) {
	assert.Contains(t, f.updates[len(f.updates)-1], key)
	assert.Equal(t, f.updates[len(f.updates)-1][key], s)
}

func lastUpdateDefFuncShouldEqual(t *testing.T, f *Field, key string, res string) {
	assert.Contains(t, f.updates[len(f.updates)-1], key)
	assert.EqualValues(t, f.updates[len(f.updates)-1][key].(func(env Environment) any)(Environment{}), res)
}

func TestFieldModification(t *testing.T) {
	t.Run("Testing field modification", func(t *testing.T) {
		numsField := Registry.MustGet("User").Fields().MustGet("Nums")
		numsField.SetString("Nums Reloaded")
		checkUpdates(t, numsField, "description", "Nums Reloaded")
		numsField.SetHelp("Num's Help")
		checkUpdates(t, numsField, "help", "Num's Help")
		numsField.SetCompute(Registry.MustGet("User").Methods().MustGet("ComputeNum"))
		checkUpdates(t, numsField, "compute", "ComputeNum")
		numsField.SetCompute(nil)
		checkUpdates(t, numsField, "compute", "")
		numsField.SetDefault(DefaultValue("DV"))
		lastUpdateDefFuncShouldEqual(t, numsField, "defaultFunc", "DV")
		numsField.SetDepends([]string{"Dep1", "Dep2"})
		lastUpdateShouldResemble(t, numsField, "depends", []string{"Dep1", "Dep2"})
		numsField.SetDepends(nil)
		lastUpdateShouldResemble(t, numsField, "depends", []string(nil))
		numsField.SetGroupOperator("avg")
		checkUpdates(t, numsField, "groupOperator", "avg")
		numsField.SetGroupOperator("sum")
		checkUpdates(t, numsField, "groupOperator", "sum")
		numsField.SetIndex(true)
		checkUpdates(t, numsField, "index", true)
		numsField.SetNoCopy(true)
		checkUpdates(t, numsField, "noCopy", true)
		numsField.SetNoCopy(false)
		checkUpdates(t, numsField, "noCopy", false)
		numsField.SetRelated("Profile.Money")
		checkUpdates(t, numsField, "relatedPathStr", "Profile.Money")
		numsField.SetRelated("")
		checkUpdates(t, numsField, "relatedPathStr", "")
		numsField.SetRequired(true)
		checkUpdates(t, numsField, "required", true)
		numsField.SetRequired(false)
		checkUpdates(t, numsField, "required", false)
		numsField.SetStored(true)
		checkUpdates(t, numsField, "stored", true)
		numsField.SetStored(false)
		checkUpdates(t, numsField, "stored", false)
		numsField.SetUnique(true)
		checkUpdates(t, numsField, "unique", true)
		numsField.SetUnique(false)
		checkUpdates(t, numsField, "unique", false)
		nameField := Registry.MustGet("User").Fields().MustGet("Name")
		nameField.SetSize(127)
		checkUpdates(t, nameField, "size", 127)
		nameField.SetTranslate(true)
		checkUpdates(t, nameField, "translate", true)
		nameField.SetTranslate(false)
		checkUpdates(t, nameField, "translate", false)
		nameField.SetContexts(companyDependent)
		lastUpdateShouldResemble(t, nameField, "contexts", companyDependent)
		nameField.AddContexts(userDependent)
		lastUpdateShouldResemble(t, nameField, "contexts_add", userDependent)
		nameField.SetContexts(nil)
		nameField.SetOnchange(nil)
		nameField.SetOnchange(Registry.MustGet("User").Methods().MustGet("OnChangeName"))
		nameField.SetOnchangeWarning(Registry.MustGet("User").Methods().MustGet("OnChangeNameWarning"))
		nameField.SetOnchangeFilters(Registry.MustGet("User").Methods().MustGet("OnChangeNameFilters"))
		nameField.SetConstraint(Registry.MustGet("User").Methods().MustGet("UpdateCity"))
		nameField.SetConstraint(nil)
		nameField.SetInverse(Registry.MustGet("User").Methods().MustGet("InverseSetAge"))
		nameField.SetInverse(nil)
		sizeField := Registry.MustGet("User").Fields().MustGet("Size")
		sizeField.SetDigits(nbutils.Digits{Precision: 6, Scale: 2})
		lastUpdateShouldResemble(t, sizeField, "digits", nbutils.Digits{Precision: 6, Scale: 2})
		userField := Registry.MustGet("Post").Fields().MustGet("User")
		userField.SetOnDelete(Cascade)
		checkUpdates(t, userField, "onDelete", Cascade)
		userField.SetOnDelete(SetNull)
		checkUpdates(t, userField, "onDelete", SetNull)
		userField.SetEmbed(true)
		checkUpdates(t, userField, "embed", true)
		userField.SetEmbed(false)
		checkUpdates(t, userField, "embed", false)
		userField.SetFilter(Registry.MustGet("User").Field(NewFieldName("SetActive", "set_active")).Equals(true))
		userField.SetFilter(Condition{})
		userField.SetRelationModel(Registry.MustGet("Tag"))
		checkUpdates(t, userField, "relationModel", Registry.MustGet("Tag").Underlying())
		userField.SetRelationModel(Registry.MustGet("User"))
		visibilityField := Registry.MustGet("Post").Fields().MustGet("Visibility")
		visibilityField.UpdateSelection(types.Selection{"logged_in": "Logged in users"})
		lastUpdateShouldResemble(t, visibilityField, "selection_add", types.Selection{"logged_in": "Logged in users"})
		genderField := Registry.MustGet("Profile").Fields().MustGet("Gender")
		genderField.SetSelection(types.Selection{"m": "Male", "f": "Female"})
		lastUpdateShouldResemble(t, genderField, "selection", types.Selection{"m": "Male", "f": "Female"})
		statusField := Registry.MustGet("User").Fields().MustGet("Status")
		statusField.SetReadOnly(false)
		checkUpdates(t, statusField, "readOnly", false)
		nFunc := func(env Environment) (b bool, conditioner Conditioner) { return }
		statusField.SetReadOnlyFunc(nFunc)
		statusField.SetReadOnlyFunc(nil)
		statusField.SetInvisibleFunc(nFunc)
		statusField.SetInvisibleFunc(nil)
		statusField.SetRequiredFunc(nFunc)
		statusField.SetRequiredFunc(nil)
		tagsField := Registry.MustGet("Post").Fields().MustGet("Tags")
		tagsField.SetM2MRelModel(Registry.MustGet("Resume"))
		checkUpdates(t, tagsField, "m2mRelModel", Registry.MustGet("Resume"))
		tagsField.SetM2MOurField(Registry.MustGet("Resume").Fields().MustGet("Education"))
		checkUpdates(t, tagsField, "m2mOurField", Registry.MustGet("Resume").Fields().MustGet("Education"))
		tagsField.SetM2MTheirField(Registry.MustGet("Resume").Fields().MustGet("Experience"))
		checkUpdates(t, tagsField, "m2mTheirField", Registry.MustGet("Resume").Fields().MustGet("Experience"))
		tagsField.SetM2MRelModel(Registry.MustGet("PostTagRel"))
		tagsField.SetM2MOurField(Registry.MustGet("PostTagRel").Fields().MustGet("Post"))
		tagsField.SetM2MTheirField(Registry.MustGet("PostTagRel").Fields().MustGet("Tag"))
		commentsField := Registry.MustGet("Post").Fields().MustGet("Comments")
		commentsField.SetReverseFK("ReverseFK")
		checkUpdates(t, commentsField, "reverseFK", "ReverseFK")
		commentsField.SetReverseFK("Post")
		visibilityField.SetSelectionFunc(func() types.Selection {
			return types.Selection{"1": "Yes", "2": "No"}
		})
		visibilityField.SetSelectionFunc(nil)
	})
}

func TestMiscellaneous(t *testing.T) {
	t.Run("Check that Field instances are FieldNamers", func(t *testing.T) {
		assert.EqualValues(t, Registry.MustGet("User").Fields().MustGet("Name").JSON(), "name")
		assert.EqualValues(t, Registry.MustGet("User").Fields().MustGet("Name").Name(), "Name")
	})
}

func TestSequences(t *testing.T) {
	t.Run("Testing sequences before bootstrap", func(t *testing.T) {
		testSeq := CreateSequence("TestSequence", 5, 13)
		_, ok := Registry.GetSequence("TestSequence")
		assert.True(t, ok)
		assert.EqualValues(t, testSeq.Increment, 5)
		assert.EqualValues(t, testSeq.Start, 13)
		testSeq.Alter(3, 14)
		assert.EqualValues(t, testSeq.Increment, 3)
		assert.EqualValues(t, testSeq.Start, 14)
		testSeq.Drop()
		assert.Panics(t, func() { Registry.MustGetSequence("TestSequence") })
		CreateSequence("TestSequence", 5, 13)
	})
}

func TestIllegalMethods(t *testing.T) {
	t.Run("Checking that invalid data leads to panic", func(t *testing.T) {
		assert.Panics(t, func() { Registry.MustGet("NonExistentModel") })

		userModel := Registry.MustGet("User")
		assert.Panics(t, func() { userModel.Fields().MustGet("NonExistentField") })
		assert.Panics(t, func() { userModel.Methods().MustGet("NonExistentMethod") })

		assert.Panics(t, func() { userModel.NewMethod("WrongType", 12) })
		assert.Panics(t, func() {
			userModel.NewMethod("ComputeAge", func(rc *RecordCollection) {})
		})
		assert.Panics(t, func() {
			userModel.NewMethod("Create", func(rc *RecordCollection) {})
		})
		assert.Panics(t, func() { userModel.AddEmptyMethod("ComputeAge") })
		assert.Panics(t, func() { userModel.Methods().MustGet("ComputeAge").Extend(12) })
		assert.Panics(t, func() {
			userModel.Methods().MustGet("Copy").Extend(func(rc bool, overrides RecordData) *RecordCollection { return &RecordCollection{} })
		})
		assert.Panics(t, func() {
			userModel.Methods().MustGet("Copy").Extend(func(rc *RecordCollection) *RecordCollection { return &RecordCollection{} })
		})
		assert.Panics(t, func() {
			userModel.Methods().MustGet("Copy").Extend(func(rc *RecordCollection, overrides bool) *RecordCollection { return &RecordCollection{} })
		})
		assert.Panics(t, func() {
			userModel.Methods().MustGet("Copy").Extend(func(rc *RecordCollection, overrides RecordData) (*RecordCollection, bool) {
				return &RecordCollection{}, false
			})
		})
		assert.Panics(t, func() {
			userModel.Methods().MustGet("Copy").Extend(func(rc *RecordCollection, overrides RecordData) bool { return false })
		})
		assert.Panics(t, func() {
			userModel.Methods().MustGet("OrderBy").Extend(func(rc *RecordCollection, exprs []string) *RecordCollection { return &RecordCollection{} })
		})
	})
	t.Run("Test checkTypesMatch", func(t *testing.T) {
		type TestRecordSet struct {
			*RecordCollection
		}

		var _ FieldMapper = TestFieldMap{}

		assert.True(t, checkTypesMatch(reflect.TypeFor[string](), reflect.TypeFor[string]()))
		assert.False(t, checkTypesMatch(reflect.TypeFor[int](), reflect.TypeFor[string]()))
		assert.True(t, checkTypesMatch(reflect.TypeFor[*RecordCollection](), reflect.TypeFor[TestRecordSet]()))
		assert.True(t, checkTypesMatch(reflect.TypeFor[TestRecordSet](), reflect.TypeFor[*RecordCollection]()))
		assert.True(t, checkTypesMatch(reflect.TypeFor[TestFieldMap](), reflect.TypeFor[FieldMap]()))
		assert.True(t, checkTypesMatch(reflect.TypeFor[FieldMap](), reflect.TypeFor[TestFieldMap]()))
	})
	t.Run("Test compute and onChange method signature", func(t *testing.T) {
		userModel := Registry.MustGet("User")
		nameField := userModel.Fields().MustGet("Name")
		nameField.SetOnchange(userModel.Methods().MustGet("SubSetSuper"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchange(userModel.Methods().MustGet("OnChangeName"))
		processUpdates()

		nameField.SetOnchangeWarning(userModel.Methods().MustGet("OnChangeName"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchangeWarning(userModel.Methods().MustGet("UpdateCity"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchangeWarning(userModel.Methods().MustGet("NoReturnValue"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchangeWarning(userModel.Methods().MustGet("TwoReturnValues"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchangeWarning(userModel.Methods().MustGet("OnChangeNameWarning"))
		processUpdates()

		nameField.SetOnchangeFilters(userModel.Methods().MustGet("OnChangeName"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchangeFilters(userModel.Methods().MustGet("UpdateCity"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchangeFilters(userModel.Methods().MustGet("NoReturnValue"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchangeFilters(userModel.Methods().MustGet("TwoReturnValues"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		nameField.SetOnchangeFilters(userModel.Methods().MustGet("OnChangeNameFilters"))
		processUpdates()

		ageField := userModel.Fields().MustGet("Age")
		ageField.SetCompute(userModel.Methods().MustGet("SubSetSuper"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		ageField.SetCompute(userModel.Methods().MustGet("ComputeAge"))
		processUpdates()

		ageField.SetInverse(userModel.Methods().MustGet("SubSetSuper"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		ageField.SetInverse(userModel.Methods().MustGet("WrongInverseSetAge"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		ageField.SetInverse(userModel.Methods().MustGet("InverseSetAge"))
		processUpdates()

		dnField := userModel.Fields().MustGet("DecoratedName")
		dnField.SetCompute(userModel.Methods().MustGet("TwoReturnValues"))
		processUpdates()
		assert.Panics(t, checkComputeMethodsSignature)
		dnField.SetCompute(userModel.Methods().MustGet("ComputeDecoratedName"))
		processUpdates()
	})
	t.Run("Test methods signature check", func(t *testing.T) {
		userModel := Registry.MustGet("User")
		t.Run("Onchange/compute method should have no arguments", func(t *testing.T) {
			meth := userModel.Methods().MustGet("InverseSetAge")
			assert.NotNil(t, checkMethType(meth, "Onchange"))
		})
		t.Run("Onchange/compute method should return a value", func(t *testing.T) {
			meth := userModel.Methods().MustGet("NoReturnValue")
			assert.NotNil(t, checkMethType(meth, "Onchange"))
		})
		t.Run("Onchange/compute method returned value must be a FieldMapper", func(t *testing.T) {
			meth := userModel.Methods().MustGet("SubSetSuper")
			assert.NotNil(t, checkMethType(meth, "Onchange"))
		})
		t.Run("Onchange/compute method should not return more than one value", func(t *testing.T) {
			meth := userModel.Methods().MustGet("TwoReturnValues")
			assert.NotNil(t, checkMethType(meth, "Onchange"))
		})
	})
}

func TestBootStrap(t *testing.T) {
	// Creating a dummy table to check that it is correctly removed by Bootstrap
	dbExecuteNoTx("CREATE TABLE IF NOT EXISTS shouldbedeleted (id serial NOT NULL PRIMARY KEY)")

	// Creating a manual sequence that must be loaded in the registry
	dbExecuteNoTx(`CREATE SEQUENCE test_manseq INCREMENT BY 5 START WITH 1`)

	t.Run("Database creation should run fine", func(t *testing.T) {
		t.Run("Dummy table should exist", func(t *testing.T) {
			assert.Contains(t, TestAdapter.tables(), "shouldbedeleted")
		})
		t.Run("Bootstrap should not panic", func(t *testing.T) {
			BootStrap()
			SyncDatabase()
		})
		t.Run("Boostrapping twice should panic", func(t *testing.T) {
			assert.True(t, BootStrapped())
			assert.Panics(t, BootStrap)
		})
		t.Run("Creating methods after bootstrap should panic", func(t *testing.T) {
			assert.Panics(t, func() {
				Registry.MustGet("User").NewMethod("NewMethod", func(rc *RecordCollection) {})
			})
		})
		t.Run("Creating SQL view should run fine", func(t *testing.T) {
			assert.NotPanics(t, func() {
				dbExecuteNoTx(`DROP VIEW IF EXISTS user_view;
					CREATE VIEW user_view AS (
						SELECT u.id, u.name, p.city, u.active
						FROM "user" u
							LEFT JOIN "profile" p ON p.id = u.profile_id
					)`)
			})
		})
		t.Run("All models should have a DB table", func(t *testing.T) {
			dbTables := TestAdapter.tables()
			for tableName, mi := range Registry.registryByTableName {
				if mi.IsMixin() || mi.IsManual() {
					continue
				}
				assert.True(t, dbTables[tableName])
			}
		})
		t.Run("All DB tables should have a model", func(t *testing.T) {
			for dbTable := range TestAdapter.tables() {
				assert.Contains(t, Registry.registryByTableName, dbTable)
			}
		})
		t.Run("Table constraints should have been created", func(t *testing.T) {
			assert.Len(t, TestAdapter.constraints("%_mancon"), 1)
			assert.EqualValues(t, TestAdapter.constraints("%_mancon")[0], "nums_premium_user_mancon")
		})
		t.Run("Boot Sequence should be created", func(t *testing.T) {
			assert.Len(t, TestAdapter.sequences("%_bootseq"), 1)
			assert.EqualValues(t, TestAdapter.sequences("%_bootseq")[0].Name, "test_sequence_bootseq")
		})
		t.Run("Manual sequences should be loaded in registry", func(t *testing.T) {
			assert.Len(t, TestAdapter.sequences("%_manseq"), 1)
			assert.EqualValues(t, TestAdapter.sequences("%_manseq")[0].Name, "test_manseq")
			seq, ok := Registry.GetSequence("Test")
			assert.True(t, ok)
			assert.EqualValues(t, seq.JSON, "test_manseq")
			assert.EqualValues(t, seq.Increment, 5)
			assert.EqualValues(t, seq.Start, 1)
		})
		t.Run("Applying DB modifications", func(t *testing.T) {
			UnBootStrap()
			contentField := Registry.MustGet("Post").Fields().MustGet("Content")
			contentField.SetRequired(false)
			profileField := Registry.MustGet("User").Fields().MustGet("Profile")
			profileField.SetRequired(false)
			numsField := Registry.MustGet("User").Fields().MustGet("Nums")
			numsField.SetDefault(nil).SetIndex(false)
			Registry.MustGet("Comment").fields.add(&Field{
				model:       Registry.MustGet("Comment"),
				name:        "Date",
				json:        "date",
				fieldType:   fieldtype.Date,
				structField: reflect.StructField{Type: reflect.TypeFor[dates.Date]()},
				defaultFunc: func(env Environment) any {
					return dates.Today()
				},
			})
			textField := Registry.MustGet("Comment").Fields().MustGet("Text")
			textField.SetFieldType(fieldtype.Text)
			assert.NotPanics(t, BootStrap)
			assert.False(t, contentField.required)
			assert.False(t, profileField.required)
			assert.False(t, numsField.index)
			assert.NotPanics(t, SyncDatabase)
		})
	})

	t.Run("Post testing models modifications", func(t *testing.T) {
		visibilityField := Registry.MustGet("Post").Fields().MustGet("Visibility")
		assert.Len(t, visibilityField.selection, 3)
		assert.Contains(t, visibilityField.selection, "visible")
		assert.Contains(t, visibilityField.selection, "invisible")
		assert.Contains(t, visibilityField.selection, "logged_in")
		genderField := Registry.MustGet("Profile").Fields().MustGet("Gender")
		assert.Len(t, genderField.selection, 2)
		assert.Contains(t, genderField.selection, "m")
		assert.Contains(t, genderField.selection, "f")
	})

	t.Run("Truncating all tables...", func(t *testing.T) {
		for tn, mi := range Registry.registryByTableName {
			if mi.IsMixin() || mi.IsManual() {
				continue
			}
			dbExecuteNoTx(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, tn))
		}
	})
}
