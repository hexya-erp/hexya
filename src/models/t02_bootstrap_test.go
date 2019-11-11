// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	. "github.com/smartystreets/goconvey/convey"
)

type TestFieldMap FieldMap

func (f TestFieldMap) Underlying() FieldMap {
	return FieldMap(f)
}
func unBootStrap() {
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
func TestFieldModification(t *testing.T) {
	Convey("Testing field modification", t, func() {
		numsField := Registry.MustGet("User").Fields().MustGet("Nums")
		numsField.SetString("Nums Reloaded")
		CheckUpdates(numsField, "description", "Nums Reloaded")
		numsField.SetHelp("Num's Help")
		CheckUpdates(numsField, "help", "Num's Help")
		numsField.SetCompute(Registry.MustGet("User").Methods().MustGet("ComputeNum"))
		CheckUpdates(numsField, "compute", "ComputeNum")
		numsField.SetCompute(nil)
		CheckUpdates(numsField, "compute", "")
		numsField.SetDefault(DefaultValue("DV"))
		LastUpdateDefFuncShouldEqual(numsField, "defaultFunc", "DV")
		numsField.SetDepends([]string{"Dep1", "Dep2"})
		LastUpdateShouldResemble(numsField, "depends", []string{"Dep1", "Dep2"})
		numsField.SetDepends(nil)
		LastUpdateShouldResemble(numsField, "depends", []string(nil))
		numsField.SetGroupOperator("avg")
		CheckUpdates(numsField, "groupOperator", "avg")
		numsField.SetGroupOperator("sum")
		CheckUpdates(numsField, "groupOperator", "sum")
		numsField.SetIndex(true)
		CheckUpdates(numsField, "index", true)
		numsField.SetNoCopy(true)
		CheckUpdates(numsField, "noCopy", true)
		numsField.SetNoCopy(false)
		CheckUpdates(numsField, "noCopy", false)
		numsField.SetRelated("Profile.Money")
		CheckUpdates(numsField, "relatedPathStr", "Profile.Money")
		numsField.SetRelated("")
		CheckUpdates(numsField, "relatedPathStr", "")
		numsField.SetRequired(true)
		CheckUpdates(numsField, "required", true)
		numsField.SetRequired(false)
		CheckUpdates(numsField, "required", false)
		numsField.SetStored(true)
		CheckUpdates(numsField, "stored", true)
		numsField.SetStored(false)
		CheckUpdates(numsField, "stored", false)
		numsField.SetUnique(true)
		CheckUpdates(numsField, "unique", true)
		numsField.SetUnique(false)
		CheckUpdates(numsField, "unique", false)
		nameField := Registry.MustGet("User").Fields().MustGet("Name")
		nameField.SetSize(127)
		CheckUpdates(nameField, "size", 127)
		nameField.SetTranslate(true)
		CheckUpdates(nameField, "translate", true)
		nameField.SetTranslate(false)
		CheckUpdates(nameField, "translate", false)
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
		LastUpdateShouldResemble(sizeField, "digits", nbutils.Digits{Precision: 6, Scale: 2})
		userField := Registry.MustGet("Post").Fields().MustGet("User")
		userField.SetOnDelete(Cascade)
		CheckUpdates(userField, "onDelete", Cascade)
		userField.SetOnDelete(SetNull)
		CheckUpdates(userField, "onDelete", SetNull)
		userField.SetEmbed(true)
		CheckUpdates(userField, "embed", true)
		userField.SetEmbed(false)
		CheckUpdates(userField, "embed", false)
		userField.SetFilter(Registry.MustGet("User").Field(NewFieldName("SetActive", "set_active")).Equals(true))
		userField.SetFilter(Condition{})
		visibilityField := Registry.MustGet("Post").Fields().MustGet("Visibility")
		visibilityField.UpdateSelection(types.Selection{"logged_in": "Logged in users"})
		LastUpdateShouldResemble(visibilityField, "selection_add", types.Selection{"logged_in": "Logged in users"})
		genderField := Registry.MustGet("Profile").Fields().MustGet("Gender")
		genderField.SetSelection(types.Selection{"m": "Male", "f": "Female"})
		LastUpdateShouldResemble(genderField, "selection", types.Selection{"m": "Male", "f": "Female"})
		statusField := Registry.MustGet("User").Fields().MustGet("Status")
		statusField.SetReadOnly(false)
		CheckUpdates(statusField, "readOnly", false)
		nFunc := func(env Environment) (b bool, conditioner Conditioner) { return }
		statusField.SetReadOnlyFunc(nFunc)
		CheckUpdates(statusField, "readOnlyFunc", nFunc)
		statusField.SetInvisibleFunc(nFunc)
		CheckUpdates(statusField, "invisibleFunc", nFunc)
		statusField.SetRequiredFunc(nFunc)
		CheckUpdates(statusField, "requiredFunc", nFunc)
	})
}

func TestMiscellaneous(t *testing.T) {
	Convey("Check that Field instances are FieldNamers", t, func() {
		So(Registry.MustGet("User").Fields().MustGet("Name").JSON(), ShouldEqual, "name")
		So(Registry.MustGet("User").Fields().MustGet("Name").Name(), ShouldEqual, "Name")
	})
}

func TestSequences(t *testing.T) {
	Convey("Testing sequences before bootstrap", t, func() {
		testSeq := CreateSequence("TestSequence", 5, 13)
		_, ok := Registry.GetSequence("TestSequence")
		So(ok, ShouldBeTrue)
		So(testSeq.Increment, ShouldEqual, 5)
		So(testSeq.Start, ShouldEqual, 13)
		testSeq.Alter(3, 14)
		So(testSeq.Increment, ShouldEqual, 3)
		So(testSeq.Start, ShouldEqual, 14)
		testSeq.Drop()
		So(func() { Registry.MustGetSequence("TestSequence") }, ShouldPanic)
		CreateSequence("TestSequence", 5, 13)
	})
}

func TestIllegalMethods(t *testing.T) {
	Convey("Checking that invalid data leads to panic", t, func() {
		So(func() { Registry.MustGet("NonExistentModel") }, ShouldPanic)

		userModel := Registry.MustGet("User")
		So(func() { userModel.Fields().MustGet("NonExistentField") }, ShouldPanic)
		So(func() { userModel.Methods().MustGet("NonExistentMethod") }, ShouldPanic)

		So(func() { userModel.AddMethod("WrongType", "Test with int instead of func literal", 12) }, ShouldPanic)
		So(func() {
			userModel.AddMethod("ComputeAge", "Trying to add existing method", func(rc *RecordCollection) {})
		}, ShouldPanic)
		So(func() {
			userModel.AddMethod("Create", "Trying to add existing method", func(rc *RecordCollection) {})
		}, ShouldPanic)
		So(func() { userModel.AddEmptyMethod("ComputeAge") }, ShouldPanic)
		So(func() { userModel.Methods().MustGet("ComputeAge").Extend("Test with int instead of func literal", 12) }, ShouldPanic)
		So(func() {
			userModel.Methods().MustGet("ComputeAge").Extend("Test with wrong signature", func(rc string) (int, bool) { return 0, true })
		}, ShouldPanic)
		So(func() {
			userModel.Methods().MustGet("ComputeAge").Extend("Test with wrong signature", func(rc *RecordCollection, x string) (int, bool) { return 0, true })
		}, ShouldPanic)
		So(func() {
			userModel.Methods().MustGet("ComputeAge").Extend("Test with wrong signature", func(rc *RecordCollection) (int, int, bool) { return 0, 0, true })
		}, ShouldPanic)
		So(func() {
			userModel.Methods().MustGet("ComputeAge").Extend("Test with wrong signature", func(rc *RecordCollection) (int, bool) { return 0, true })
		}, ShouldPanic)
		So(func() {
			userModel.Methods().MustGet("DecorateEmail").Extend("Test with wrong signature", func(rc *RecordCollection, email []byte) string { return "" })
		}, ShouldPanic)
	})
	Convey("Test checkTypesMatch", t, func() {
		type TestRecordSet struct {
			*RecordCollection
		}

		var _ FieldMapper = TestFieldMap{}

		So(checkTypesMatch(reflect.TypeOf("bar"), reflect.TypeOf("bar")), ShouldBeTrue)
		So(checkTypesMatch(reflect.TypeOf(0), reflect.TypeOf("bar")), ShouldBeFalse)
		So(checkTypesMatch(reflect.TypeOf(new(RecordCollection)), reflect.TypeOf(TestRecordSet{})), ShouldBeTrue)
		So(checkTypesMatch(reflect.TypeOf(TestRecordSet{}), reflect.TypeOf(new(RecordCollection))), ShouldBeTrue)
		So(checkTypesMatch(reflect.TypeOf(TestFieldMap{}), reflect.TypeOf(FieldMap{})), ShouldBeTrue)
		So(checkTypesMatch(reflect.TypeOf(FieldMap{}), reflect.TypeOf(TestFieldMap{})), ShouldBeTrue)
	})
	Convey("Test compute and onChange method signature", t, func() {
		userModel := Registry.MustGet("User")
		nameField := userModel.Fields().MustGet("Name")
		nameField.SetOnchange(userModel.Methods().MustGet("SubSetSuper"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchange(userModel.Methods().MustGet("OnChangeName"))
		processUpdates()

		nameField.SetOnchangeWarning(userModel.Methods().MustGet("OnChangeName"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchangeWarning(userModel.Methods().MustGet("UpdateCity"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchangeWarning(userModel.Methods().MustGet("NoReturnValue"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchangeWarning(userModel.Methods().MustGet("TwoReturnValues"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchangeWarning(userModel.Methods().MustGet("OnChangeNameWarning"))
		processUpdates()

		nameField.SetOnchangeFilters(userModel.Methods().MustGet("OnChangeName"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchangeFilters(userModel.Methods().MustGet("UpdateCity"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchangeFilters(userModel.Methods().MustGet("NoReturnValue"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchangeFilters(userModel.Methods().MustGet("TwoReturnValues"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		nameField.SetOnchangeFilters(userModel.Methods().MustGet("OnChangeNameFilters"))
		processUpdates()

		ageField := userModel.Fields().MustGet("Age")
		ageField.SetCompute(userModel.Methods().MustGet("SubSetSuper"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		ageField.SetCompute(userModel.Methods().MustGet("ComputeAge"))
		processUpdates()

		ageField.SetInverse(userModel.Methods().MustGet("SubSetSuper"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		ageField.SetInverse(userModel.Methods().MustGet("WrongInverseSetAge"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		ageField.SetInverse(userModel.Methods().MustGet("InverseSetAge"))
		processUpdates()

		dnField := userModel.Fields().MustGet("DecoratedName")
		dnField.SetCompute(userModel.Methods().MustGet("TwoReturnValues"))
		processUpdates()
		So(checkComputeMethodsSignature, ShouldPanic)
		dnField.SetCompute(userModel.Methods().MustGet("ComputeDecoratedName"))
		processUpdates()
	})
	Convey("Test methods signature check", t, func() {
		userModel := Registry.MustGet("User")
		Convey("Onchange/compute method should have no arguments", func() {
			meth := userModel.Methods().MustGet("InverseSetAge")
			So(checkMethType(meth, "Onchange"), ShouldNotBeNil)
		})
		Convey("Onchange/compute method should return a value", func() {
			meth := userModel.Methods().MustGet("NoReturnValue")
			So(checkMethType(meth, "Onchange"), ShouldNotBeNil)
		})
		Convey("Onchange/compute method returned value must be a FieldMapper", func() {
			meth := userModel.Methods().MustGet("SubSetSuper")
			So(checkMethType(meth, "Onchange"), ShouldNotBeNil)
		})
		Convey("Onchange/compute method should not return more than one value", func() {
			meth := userModel.Methods().MustGet("TwoReturnValues")
			So(checkMethType(meth, "Onchange"), ShouldNotBeNil)
		})
	})
}

func TestBootStrap(t *testing.T) {
	// Creating a dummy table to check that it is correctly removed by Bootstrap
	dbExecuteNoTx("CREATE TABLE IF NOT EXISTS shouldbedeleted (id serial NOT NULL PRIMARY KEY)")

	// Creating a manual sequence that must be loaded in the registry
	dbExecuteNoTx(`CREATE SEQUENCE test_manseq INCREMENT BY 5 START WITH 1`)

	Convey("Database creation should run fine", t, func() {
		Convey("Dummy table should exist", func() {
			So(TestAdapter.tables(), ShouldContainKey, "shouldbedeleted")
		})
		Convey("Bootstrap should not panic", func() {
			BootStrap()
			SyncDatabase()
		})
		Convey("Boostrapping twice should panic", func() {
			So(BootStrapped(), ShouldBeTrue)
			So(BootStrap, ShouldPanic)
		})
		Convey("Creating methods after bootstrap should panic", func() {
			So(func() {
				Registry.MustGet("User").AddMethod("NewMethod", "Method after boostrap", func(rc *RecordCollection) {})
			}, ShouldPanic)
		})
		Convey("Creating SQL view should run fine", func() {
			So(func() {
				dbExecuteNoTx(`DROP VIEW IF EXISTS user_view;
					CREATE VIEW user_view AS (
						SELECT u.id, u.name, p.city, u.active
						FROM "user" u
							LEFT JOIN "profile" p ON p.id = u.profile_id
					)`)
			}, ShouldNotPanic)
		})
		Convey("All models should have a DB table", func() {
			dbTables := TestAdapter.tables()
			for tableName, mi := range Registry.registryByTableName {
				if mi.IsMixin() || mi.IsManual() {
					continue
				}
				So(dbTables[tableName], ShouldBeTrue)
			}
		})
		Convey("All DB tables should have a model", func() {
			for dbTable := range TestAdapter.tables() {
				So(Registry.registryByTableName, ShouldContainKey, dbTable)
			}
		})
		Convey("Table constraints should have been created", func() {
			So(TestAdapter.constraints("%_mancon"), ShouldHaveLength, 1)
			So(TestAdapter.constraints("%_mancon")[0], ShouldEqual, "nums_premium_user_mancon")
		})
		Convey("Boot Sequence should be created", func() {
			So(TestAdapter.sequences("%_bootseq"), ShouldHaveLength, 1)
			So(TestAdapter.sequences("%_bootseq")[0].Name, ShouldEqual, "test_sequence_bootseq")
		})
		Convey("Manual sequences should be loaded in registry", func() {
			So(TestAdapter.sequences("%_manseq"), ShouldHaveLength, 1)
			So(TestAdapter.sequences("%_manseq")[0].Name, ShouldEqual, "test_manseq")
			seq, ok := Registry.GetSequence("Test")
			So(ok, ShouldBeTrue)
			So(seq.JSON, ShouldEqual, "test_manseq")
			So(seq.Increment, ShouldEqual, 5)
			So(seq.Start, ShouldEqual, 1)
		})
		Convey("Applying DB modifications", func() {
			unBootStrap()
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
				structField: reflect.StructField{Type: reflect.TypeOf(dates.Date{})},
				defaultFunc: func(env Environment) interface{} {
					return dates.Today()
				},
			})
			textField := Registry.MustGet("Comment").Fields().MustGet("Text")
			textField.SetFieldType(fieldtype.Text)
			BootStrap()
			//So(BootStrap, ShouldNotPanic)
			So(contentField.required, ShouldBeFalse)
			So(profileField.required, ShouldBeFalse)
			So(numsField.index, ShouldBeFalse)
			So(SyncDatabase, ShouldNotPanic)
		})
	})

	Convey("Post testing models modifications", t, func() {
		visibilityField := Registry.MustGet("Post").Fields().MustGet("Visibility")
		So(visibilityField.selection, ShouldHaveLength, 3)
		So(visibilityField.selection, ShouldContainKey, "visible")
		So(visibilityField.selection, ShouldContainKey, "invisible")
		So(visibilityField.selection, ShouldContainKey, "logged_in")
		genderField := Registry.MustGet("Profile").Fields().MustGet("Gender")
		So(genderField.selection, ShouldHaveLength, 2)
		So(genderField.selection, ShouldContainKey, "m")
		So(genderField.selection, ShouldContainKey, "f")
	})

	Convey("Truncating all tables...", t, func() {
		for tn, mi := range Registry.registryByTableName {
			if mi.IsMixin() || mi.IsManual() {
				continue
			}
			dbExecuteNoTx(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, tn))
		}
	})
}
