// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"reflect"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	"github.com/hexya-erp/hexya/src/tools/strutils"
)

// A FieldDefinition is a struct that declares a new field in a fields collection;
type FieldDefinition interface {
	// DeclareField creates a field for the given FieldsCollection with the given name and returns the created field.
	DeclareField(*FieldsCollection, string) *Field
}

// DummyField is used internally to inflate mixins. It should not be used.
type DummyField struct{}

// DeclareField creates a dummy field for the given FieldsCollection with the given name.
func (df DummyField) DeclareField(fc FieldsCollection, name string) *Field {
	json := SnakeCaseFieldName(name, fieldtype.NoType)
	fInfo := &Field{
		model: fc.model,
		name:  name,
		json:  json,
		structField: reflect.StructField{
			Name: name,
			Type: reflect.TypeOf(*new(bool)),
		},
		fieldType: fieldtype.NoType,
	}
	return fInfo
}

// CreateFieldFromStruct creates a generic Field with the data from the given fStruct
//
// fStruct must be a pointer to a struct and goType a pointer to a type instance
func CreateFieldFromStruct(fc *FieldsCollection, fStruct interface{}, name string, fieldType fieldtype.Type, goType interface{}) *Field {
	val := reflect.ValueOf(fStruct).Elem()
	typ := reflect.TypeOf(goType).Elem()
	if val.FieldByName("GoType").IsValid() && !val.FieldByName("GoType").IsNil() {
		typ = reflect.Indirect(val.FieldByName("GoType").Elem()).Type()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, fieldType, val.FieldByName("JSON").String(), val.FieldByName("String").String())

	comp, _ := val.FieldByName("Compute").Interface().(Methoder)
	inv, _ := val.FieldByName("Inverse").Interface().(Methoder)
	onc, _ := val.FieldByName("OnChange").Interface().(Methoder)
	onw, _ := val.FieldByName("OnChangeWarning").Interface().(Methoder)
	onf, _ := val.FieldByName("OnChangeFilters").Interface().(Methoder)
	cons, _ := val.FieldByName("Constraint").Interface().(Methoder)
	compute, inverse, onchange, onchangeWarning, onchangeFilters, constraint := getFuncNames(comp, inv, onc, onw, onf, cons)

	var unique bool
	if uni := val.FieldByName("Unique"); uni.IsValid() {
		unique = uni.Bool()
	}

	var contexts FieldContexts
	if cont := val.FieldByName("Contexts"); cont.IsValid() {
		contexts = cont.Interface().(FieldContexts)
	}
	if trans := val.FieldByName("Translate"); trans.IsValid() && trans.Bool() {
		if contexts == nil {
			contexts = make(FieldContexts)
		}
		contexts["lang"] = func(rs RecordSet) string {
			res := rs.Env().Context().GetString("lang")
			return res
		}
	}
	var noCopy bool
	if noc := val.FieldByName("NoCopy"); noc.IsValid() {
		noCopy = noc.Bool()
	}
	fInfo := &Field{
		model:           fc.model,
		name:            name,
		json:            json,
		description:     str,
		help:            val.FieldByName("Help").String(),
		stored:          val.FieldByName("Stored").Bool(),
		required:        val.FieldByName("Required").Bool(),
		readOnly:        val.FieldByName("ReadOnly").Bool(),
		readOnlyFunc:    val.FieldByName("ReadOnlyFunc").Interface().(func(Environment) (bool, Conditioner)),
		requiredFunc:    val.FieldByName("RequiredFunc").Interface().(func(Environment) (bool, Conditioner)),
		invisibleFunc:   val.FieldByName("InvisibleFunc").Interface().(func(Environment) (bool, Conditioner)),
		unique:          unique,
		index:           val.FieldByName("Index").Bool(),
		compute:         compute,
		inverse:         inverse,
		depends:         val.FieldByName("Depends").Interface().([]string),
		relatedPathStr:  val.FieldByName("Related").String(),
		noCopy:          noCopy,
		structField:     structField,
		fieldType:       fieldType,
		defaultFunc:     val.FieldByName("Default").Interface().(func(Environment) interface{}),
		onChange:        onchange,
		onChangeWarning: onchangeWarning,
		onChangeFilters: onchangeFilters,
		constraint:      constraint,
		contexts:        contexts,
	}
	return fInfo
}

// getJSONAndString computes the default json and description fields for the
// given name. It returns this default value unless given json or str are not
// empty strings, in which case the latters are returned.
func getJSONAndString(name string, typ fieldtype.Type, json, str string) (string, string) {
	if json == "" {
		json = SnakeCaseFieldName(name, typ)
	}
	if str == "" {
		str = strutils.Title(name)
	}
	return json, str
}

// getFuncNames returns the methods names of the given Methoder instances in the same order.
// Returns "" if the Methoder is nil
func getFuncNames(compute, inverse, onchange, onchangeWarning, onchangeFilters, constraint Methoder) (string, string, string, string, string, string) {
	var com, inv, onc, onw, onf, con string
	if compute != nil {
		com = compute.Underlying().name
	}
	if inverse != nil {
		inv = inverse.Underlying().name
	}
	if onchange != nil {
		onc = onchange.Underlying().name
	}
	if onchangeWarning != nil {
		onw = onchangeWarning.Underlying().name
	}
	if onchangeFilters != nil {
		onf = onchangeFilters.Underlying().name
	}
	if constraint != nil {
		con = constraint.Underlying().name
	}
	return com, inv, onc, onw, onf, con
}

// addUpdate adds an update entry for for this field with the given property and the given value
func (f *Field) addUpdate(property string, value interface{}) {
	if Registry.bootstrapped {
		log.Panic("Fields must not be modified after bootstrap", "model", f.model.name, "field", f.name, "property", property, "value", value)
	}
	update := map[string]interface{}{property: value}
	f.updates = append(f.updates, update)
}

// SetProperty sets the given property value in this field
// This method uses switch as they are unexported struct fields
func (f *Field) SetProperty(property string, value interface{}) {
	switch property {
	case "fieldType":
		f.fieldType = value.(fieldtype.Type)
	case "description":
		f.description = value.(string)
	case "help":
		f.help = value.(string)
	case "stored":
		f.stored = value.(bool)
	case "required":
		f.required = value.(bool)
	case "readOnly":
		f.readOnly = value.(bool)
	case "requiredFunc":
		f.requiredFunc = value.(func(Environment) (bool, Conditioner))
	case "readOnlyFunc":
		f.readOnlyFunc = value.(func(Environment) (bool, Conditioner))
	case "invisibleFunc":
		f.invisibleFunc = value.(func(Environment) (bool, Conditioner))
	case "unique":
		f.unique = value.(bool)
	case "index":
		f.index = value.(bool)
	case "compute":
		f.compute = value.(string)
	case "depends":
		f.depends = value.([]string)
	case "selection":
		f.selection = value.(types.Selection)
	case "selectionFunc":
		f.selectionFunc = value.(func() types.Selection)
	case "groupOperator":
		f.groupOperator = value.(string)
	case "size":
		f.size = value.(int)
	case "digits":
		f.digits = value.(nbutils.Digits)
	case "relatedPathStr":
		f.relatedPathStr = value.(string)
	case "embed":
		f.embed = value.(bool)
	case "noCopy":
		f.noCopy = value.(bool)
	case "defaultFunc":
		f.defaultFunc = value.(func(Environment) interface{})
	case "onDelete":
		f.onDelete = value.(OnDeleteAction)
	case "onChange":
		f.onChange = value.(string)
	case "onChangeWarning":
		f.onChangeWarning = value.(string)
	case "onChangeFilters":
		f.onChangeFilters = value.(string)
	case "constraint":
		f.constraint = value.(string)
	case "inverse":
		f.inverse = value.(string)
	case "filter":
		f.filter = value.(*Condition)
	case "relationModel":
		f.relatedModelName = value.(*Model).Name()
	case "m2mRelModel":
		f.m2mRelModel = value.(*Model)
	case "m2mOurField":
		f.m2mOurField = value.(*Field)
	case "m2mTheirField":
		f.m2mTheirField = value.(*Field)
	case "reverseFK":
		f.reverseFK = value.(string)
	case "translate":
		switch value.(bool) {
		case true:
			if f.contexts == nil {
				f.contexts = make(FieldContexts)
			}
			f.contexts["lang"] = func(rs RecordSet) string {
				res := rs.Env().Context().GetString("lang")
				return res
			}
		case false:
			if f.contexts == nil {
				return
			}
			delete(f.contexts, "lang")
		}
	case "contexts":
		f.contexts = value.(FieldContexts)
	default:
		log.Panic("Unknown property", "property", property, "value", value)
	}
}

// SetFieldType overrides the type of Field.
// This may fail at database sync if the table already has values and
// the old type cannot be casted into the new type by the database.
func (f *Field) SetFieldType(value fieldtype.Type) *Field {
	f.addUpdate("fieldType", value)
	return f
}

// SetString overrides the value of the String parameter of this Field
func (f *Field) SetString(value string) *Field {
	f.addUpdate("description", value)
	return f
}

// SetHelp overrides the value of the Help parameter of this Field
func (f *Field) SetHelp(value string) *Field {
	f.addUpdate("help", value)
	return f
}

// SetGroupOperator overrides the value of the GroupOperator parameter of this Field
func (f *Field) SetGroupOperator(value string) *Field {
	f.addUpdate("groupOperator", value)
	return f
}

// SetRelated overrides the value of the Related parameter of this Field
func (f *Field) SetRelated(value string) *Field {
	f.addUpdate("relatedPathStr", value)
	return f
}

// SetOnDelete overrides the value of the OnDelete parameter of this Field
func (f *Field) SetOnDelete(value OnDeleteAction) *Field {
	f.addUpdate("onDelete", value)
	return f
}

// SetCompute overrides the value of the Compute parameter of this Field
func (f *Field) SetCompute(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.addUpdate("compute", methName)
	return f
}

// SetDepends overrides the value of the Depends parameter of this Field
func (f *Field) SetDepends(value []string) *Field {
	f.addUpdate("depends", value)
	return f
}

// SetStored overrides the value of the Stored parameter of this Field
func (f *Field) SetStored(value bool) *Field {
	f.addUpdate("stored", value)
	return f
}

// SetRequired overrides the value of the Required parameter of this Field
func (f *Field) SetRequired(value bool) *Field {
	f.addUpdate("required", value)
	return f
}

// SetReadOnly overrides the value of the ReadOnly parameter of this Field
func (f *Field) SetReadOnly(value bool) *Field {
	f.addUpdate("readOnly", value)
	return f
}

// SetReadOnlyFunc overrides the value of the ReadOnlyFunc parameter of this Field
func (f *Field) SetReadOnlyFunc(value func(Environment) (bool, Conditioner)) *Field {
	f.addUpdate("readOnlyFunc", value)
	return f
}

// SetRequiredFunc overrides the value of the RequiredFunc parameter of this Field
func (f *Field) SetRequiredFunc(value func(Environment) (bool, Conditioner)) *Field {
	f.addUpdate("requiredFunc", value)
	return f
}

// SetInvisibleFunc overrides the value of the InvisibleFunc parameter of this Field
func (f *Field) SetInvisibleFunc(value func(Environment) (bool, Conditioner)) *Field {
	f.addUpdate("invisibleFunc", value)
	return f
}

// SetUnique overrides the value of the Unique parameter of this Field
func (f *Field) SetUnique(value bool) *Field {
	f.addUpdate("unique", value)
	return f
}

// SetIndex overrides the value of the Index parameter of this Field
func (f *Field) SetIndex(value bool) *Field {
	f.addUpdate("index", value)
	return f
}

// SetEmbed overrides the value of the Embed parameter of this Field
func (f *Field) SetEmbed(value bool) *Field {
	f.addUpdate("embed", value)
	return f
}

// SetSize overrides the value of the Size parameter of this Field
func (f *Field) SetSize(value int) *Field {
	f.addUpdate("size", value)
	return f
}

// SetDigits overrides the value of the Digits parameter of this Field
func (f *Field) SetDigits(value nbutils.Digits) *Field {
	f.addUpdate("digits", value)
	return f
}

// SetNoCopy overrides the value of the NoCopy parameter of this Field
func (f *Field) SetNoCopy(value bool) *Field {
	f.addUpdate("noCopy", value)
	return f
}

// SetTranslate overrides the value of the Translate parameter of this Field
func (f *Field) SetTranslate(value bool) *Field {
	f.addUpdate("translate", value)
	return f
}

// SetContexts overrides the value of the Contexts parameter of this Field
func (f *Field) SetContexts(value FieldContexts) *Field {
	f.addUpdate("contexts", value)
	return f
}

// AddContexts adds the given contexts to the Contexts parameter of this Field
func (f *Field) AddContexts(value FieldContexts) *Field {
	f.addUpdate("contexts_add", value)
	return f
}

// SetDefault overrides the value of the Default parameter of this Field
func (f *Field) SetDefault(value func(Environment) interface{}) *Field {
	f.addUpdate("defaultFunc", value)
	return f
}

// SetSelection overrides the value of the Selection parameter of this Field
func (f *Field) SetSelection(value types.Selection) *Field {
	f.addUpdate("selection", value)
	return f
}

// UpdateSelection updates the value of the Selection parameter of this Field
// with the given value. Existing keys are overridden.
func (f *Field) UpdateSelection(value types.Selection) *Field {
	f.addUpdate("selection_add", value)
	return f
}

// SetOnchange overrides the value of the Onchange parameter of this Field
func (f *Field) SetOnchange(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.addUpdate("onChange", methName)
	return f
}

// SetOnchangeWarning overrides the value of the OnChangeWarning parameter of this Field
func (f *Field) SetOnchangeWarning(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.addUpdate("onChangeWarning", methName)
	return f
}

// SetOnchangeFilters overrides the value of the OnChangeFilters parameter of this Field
func (f *Field) SetOnchangeFilters(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.addUpdate("onChangeFilters", methName)
	return f
}

// SetConstraint overrides the value of the Constraint parameter of this Field
func (f *Field) SetConstraint(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.addUpdate("constraint", methName)
	return f
}

// SetInverse overrides the value of the Inverse parameter of this Field
func (f *Field) SetInverse(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.addUpdate("inverse", methName)
	return f
}

// SetFilter overrides the value of the Filter parameter of this Field
func (f *Field) SetFilter(value Conditioner) *Field {
	f.addUpdate("filter", value.Underlying())
	return f
}

// SetRelationModel overrides the value of the Filter parameter of this Field
func (f *Field) SetRelationModel(value Modeler) *Field {
	f.addUpdate("relationModel", value.Underlying())
	return f
}

// SetM2MRelModel sets the relation model between this model and
// the target model.
func (f *Field) SetM2MRelModel(value Modeler) *Field {
	f.addUpdate("m2mRelModel", value.Underlying())
	return f
}

// SetM2MOurField sets the field of the M2MRelModel pointing to this model.
func (f *Field) SetM2MOurField(value *Field) *Field {
	f.addUpdate("m2mOurField", value)
	return f
}

// SetM2MTheirField sets the field of the M2MRelModel pointing to the other model.
func (f *Field) SetM2MTheirField(value *Field) *Field {
	f.addUpdate("m2mTheirField", value)
	return f
}

// SetReverseFK sets the name of the FK pointing to this model in a O2M or R2O relation
func (f *Field) SetReverseFK(value string) *Field {
	f.addUpdate("reverseFK", value)
	return f
}

// SetSelectionFunc defines the function that will return the selection of this field
func (f *Field) SetSelectionFunc(value func() types.Selection) *Field {
	f.addUpdate("selectionFunc", value)
	return f
}
