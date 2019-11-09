// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/hexya-erp/hexya/src/models/field"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	"github.com/hexya-erp/hexya/src/tools/strutils"
)

// A FieldDefinition is a struct that declares a new field in a fields collection;
type FieldDefinition interface {
	// DeclareField creates a field for the given FieldsCollection with the given name and returns the created field.
	DeclareField(*FieldsCollection, string) *Field
}

// A BinaryField is a field for storing binary data, such as images.
//
// Clients are expected to handle binary fields as file uploads.
//
// Binary fields are stored in the database. Consider other disk based
// alternatives if you have a large amount of data to store.
type BinaryField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	GoType          interface{}
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a binary field for the given FieldsCollection with the given name.
func (bf BinaryField) DeclareField(fc *FieldsCollection, name string) *Field {
	return genericDeclareField(fc, &bf, name, field.Binary, new(string))
}

// A BooleanField is a field for storing true/false values.
//
// Clients are expected to handle boolean fields as checkboxes.
type BooleanField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	GoType          interface{}
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a boolean field for the given FieldsCollection with the given name.
func (bf BooleanField) DeclareField(fc *FieldsCollection, name string) *Field {
	if bf.Default == nil {
		bf.Default = DefaultValue(false)
	}
	return genericDeclareField(fc, &bf, name, field.Boolean, new(bool))
}

// A CharField is a field for storing short text. There is no
// default max size, but it can be forced by setting the Size value.
//
// Clients are expected to handle Char fields as single line inputs.
type CharField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	Size            int
	GoType          interface{}
	Translate       bool
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a char field for the given FieldsCollection with the given name.
func (cf CharField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &cf, name, field.Char, new(string))
	fInfo.size = cf.Size
	return fInfo
}

// A DateField is a field for storing dates without time.
//
// Clients are expected to handle Date fields with a date picker.
type DateField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	GroupOperator   string
	NoCopy          bool
	GoType          interface{}
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a date field for the given FieldsCollection with the given name.
func (df DateField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &df, name, field.Date, new(dates.Date))
	fInfo.groupOperator = strutils.GetDefaultString(df.GroupOperator, "sum")
	return fInfo
}

// A DateTimeField is a field for storing dates with time.
//
// Clients are expected to handle DateTime fields with a date and time picker.
type DateTimeField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	GroupOperator   string
	NoCopy          bool
	GoType          interface{}
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a datetime field for the given FieldsCollection with the given name.
func (df DateTimeField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &df, name, field.DateTime, new(dates.DateTime))
	fInfo.groupOperator = strutils.GetDefaultString(df.GroupOperator, "sum")
	return fInfo
}

// A FloatField is a field for storing decimal numbers.
type FloatField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	GroupOperator   string
	NoCopy          bool
	Digits          nbutils.Digits
	GoType          interface{}
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField adds this datetime field for the given FieldsCollection with the given name.
func (ff FloatField) DeclareField(fc *FieldsCollection, name string) *Field {
	if ff.Default == nil {
		ff.Default = DefaultValue(0)
	}
	fInfo := genericDeclareField(fc, &ff, name, field.Float, new(float64))
	fInfo.groupOperator = strutils.GetDefaultString(ff.GroupOperator, "sum")
	fInfo.digits = ff.Digits
	return fInfo
}

// An HTMLField is a field for storing HTML formatted strings.
//
// Clients are expected to handle HTML fields with multi-line HTML editors.
type HTMLField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	Size            int
	GoType          interface{}
	Translate       bool
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a html field for the given FieldsCollection with the given name.
func (tf HTMLField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &tf, name, field.HTML, new(string))
	fInfo.size = tf.Size
	return fInfo
}

// An IntegerField is a field for storing non decimal numbers.
type IntegerField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	GroupOperator   string
	NoCopy          bool
	GoType          interface{}
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a datetime field for the given FieldsCollection with the given name.
func (i IntegerField) DeclareField(fc *FieldsCollection, name string) *Field {
	if i.Default == nil {
		i.Default = DefaultValue(0)
	}
	fInfo := genericDeclareField(fc, &i, name, field.Integer, new(int64))
	fInfo.groupOperator = strutils.GetDefaultString(i.GroupOperator, "sum")
	return fInfo
}

// A Many2ManyField is a field for storing many-to-many relations.
//
// Clients are expected to handle many2many fields with a table or with tags.
type Many2ManyField struct {
	JSON             string
	String           string
	Help             string
	Stored           bool
	Required         bool
	ReadOnly         bool
	RequiredFunc     func(Environment) (bool, Conditioner)
	ReadOnlyFunc     func(Environment) (bool, Conditioner)
	InvisibleFunc    func(Environment) (bool, Conditioner)
	Index            bool
	Compute          Methoder
	Depends          []string
	Related          string
	NoCopy           bool
	RelationModel    Modeler
	M2MLinkModelName string
	M2MOurField      string
	M2MTheirField    string
	OnChange         Methoder
	OnChangeWarning  Methoder
	OnChangeFilters  Methoder
	Constraint       Methoder
	Filter           Conditioner
	Inverse          Methoder
	Default          func(Environment) interface{}
}

// DeclareField creates a many2many field for the given FieldsCollection with the given name.
func (mf Many2ManyField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &mf, name, field.Many2Many, new([]int64))
	our := mf.M2MOurField
	if our == "" {
		our = fc.model.name
	}
	their := mf.M2MTheirField
	if their == "" {
		their = mf.RelationModel.Underlying().name
	}
	if our == their {
		log.Panic("Many2many relation must have different 'M2MOurField' and 'M2MTheirField'",
			"model", fc.model.name, "field", name, "ours", our, "theirs", their)
	}

	modelNames := []string{fc.model.name, mf.RelationModel.Underlying().name}
	sort.Strings(modelNames)
	m2mRelModName := mf.M2MLinkModelName
	if m2mRelModName == "" {
		m2mRelModName = fmt.Sprintf("%s%sRel", modelNames[0], modelNames[1])
	}
	m2mRelModel, m2mOurField, m2mTheirField := createM2MRelModelInfo(m2mRelModName, fc.model.name, mf.RelationModel.Underlying().name, our, their, fc.model.isMixin())

	var filter *Condition
	if mf.Filter != nil {
		filter = mf.Filter.Underlying()
	}
	fInfo.relatedModelName = mf.RelationModel.Underlying().name
	fInfo.m2mRelModel = m2mRelModel
	fInfo.m2mOurField = m2mOurField
	fInfo.m2mTheirField = m2mTheirField
	fInfo.filter = filter
	return fInfo
}

// A Many2OneField is a field for storing many-to-one relations,
// i.e. the FK to another model.
//
// Clients are expected to handle many2one fields with a combo-box.
type Many2OneField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	RelationModel   Modeler
	Embed           bool
	OnDelete        OnDeleteAction
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Filter          Conditioner
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a many2one field for the given FieldsCollection with the given name.
func (mf Many2OneField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &mf, name, field.Many2One, new(int64))
	onDelete := SetNull
	if mf.OnDelete != "" {
		onDelete = mf.OnDelete
	}
	noCopy := mf.NoCopy
	required := mf.Required
	if mf.Embed {
		onDelete = Cascade
		noCopy = true
		required = false
	}
	var filter *Condition
	if mf.Filter != nil {
		filter = mf.Filter.Underlying()
	}
	fInfo.filter = filter
	fInfo.relatedModelName = mf.RelationModel.Underlying().name
	fInfo.onDelete = onDelete
	fInfo.noCopy = noCopy
	fInfo.required = required
	fInfo.embed = mf.Embed
	return fInfo
}

// A One2ManyField is a field for storing one-to-many relations.
//
// Clients are expected to handle one2many fields with a table.
type One2ManyField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	Copy            bool
	RelationModel   Modeler
	ReverseFK       string
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Filter          Conditioner
	Inverse         Methoder
	Default         func(Environment) interface{}
}

// DeclareField creates a one2many field for the given FieldsCollection with the given name.
func (of One2ManyField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &of, name, field.One2Many, new([]int64))
	var filter *Condition
	if of.Filter != nil {
		filter = of.Filter.Underlying()
	}
	fInfo.filter = filter
	fInfo.relatedModelName = of.RelationModel.Underlying().name
	fInfo.reverseFK = of.ReverseFK
	if !of.Copy {
		fInfo.noCopy = true
	}
	return fInfo
}

// A One2OneField is a field for storing one-to-one relations,
// i.e. the FK to another model with a unique constraint.
//
// Clients are expected to handle one2one fields with a combo-box.
type One2OneField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	RelationModel   Modeler
	Embed           bool
	OnDelete        OnDeleteAction
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Filter          Conditioner
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a one2one field for the given FieldsCollection with the given name.
func (of One2OneField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &of, name, field.One2One, new(int64))
	onDelete := SetNull
	if of.OnDelete != "" {
		onDelete = of.OnDelete
	}
	noCopy := of.NoCopy
	required := of.Required
	if of.Embed {
		onDelete = Cascade
		required = true
		noCopy = true
	}
	var filter *Condition
	if of.Filter != nil {
		filter = of.Filter.Underlying()
	}
	fInfo.filter = filter
	fInfo.relatedModelName = of.RelationModel.Underlying().name
	fInfo.onDelete = onDelete
	fInfo.noCopy = noCopy
	fInfo.required = required
	fInfo.embed = of.Embed
	return fInfo
}

// A Rev2OneField is a field for storing reverse one-to-one relations,
// i.e. the relation on the model without FK.
//
// Clients are expected to handle rev2one fields with a combo-box.
type Rev2OneField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	Copy            bool
	RelationModel   Modeler
	ReverseFK       string
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Filter          Conditioner
	Inverse         Methoder
	Default         func(Environment) interface{}
}

// DeclareField creates a rev2one field for the given FieldsCollection with the given name.
func (rf Rev2OneField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &rf, name, field.Rev2One, new(int64))
	var filter *Condition
	if rf.Filter != nil {
		filter = rf.Filter.Underlying()
	}
	fInfo.filter = filter
	fInfo.relatedModelName = rf.RelationModel.Underlying().name
	fInfo.reverseFK = rf.ReverseFK
	if !rf.Copy {
		fInfo.noCopy = true
	}
	return fInfo
}

// A SelectionField is a field for storing a value from a preset list.
//
// Clients are expected to handle selection fields with a combo-box or radio buttons.
type SelectionField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	Selection       types.Selection
	SelectionFunc   func() types.Selection
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a selection field for the given FieldsCollection with the given name.
func (sf SelectionField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &sf, name, field.Selection, new(string))
	fInfo.selection = sf.Selection
	fInfo.selectionFunc = sf.SelectionFunc
	return fInfo
}

// A TextField is a field for storing long text. There is no
// default max size, but it can be forced by setting the Size value.
//
// Clients are expected to handle text fields as multi-line inputs.
type TextField struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(Environment) (bool, Conditioner)
	ReadOnlyFunc    func(Environment) (bool, Conditioner)
	InvisibleFunc   func(Environment) (bool, Conditioner)
	Unique          bool
	Index           bool
	Compute         Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	Size            int
	GoType          interface{}
	Translate       bool
	OnChange        Methoder
	OnChangeWarning Methoder
	OnChangeFilters Methoder
	Constraint      Methoder
	Inverse         Methoder
	Contexts        FieldContexts
	Default         func(Environment) interface{}
}

// DeclareField creates a text field for the given FieldsCollection with the given name.
func (tf TextField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &tf, name, field.Text, new(string))
	fInfo.size = tf.Size
	return fInfo
}

// DummyField is used internally to inflate mixins. It should not be used.
type DummyField struct{}

// DeclareField creates a dummy field for the given FieldsCollection with the given name.
func (df DummyField) DeclareField(fc *FieldsCollection, name string) *Field {
	json, _ := getJSONAndString(name, field.NoType, "", "")
	fInfo := &Field{
		model: fc.model,
		name:  name,
		json:  json,
		structField: reflect.StructField{
			Name: name,
			Type: reflect.TypeOf(*new(bool)),
		},
		fieldType: field.NoType,
	}
	return fInfo
}

// genericDeclareField creates a generic Field with the data from the given fStruct
//
// fStruct must be a pointer to a struct and goType a pointer to a type instance
func genericDeclareField(fc *FieldsCollection, fStruct interface{}, name string, fieldType field.Type, goType interface{}) *Field {
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
func getJSONAndString(name string, typ field.Type, json, str string) (string, string) {
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

// AddFields adds the given fields to the model.
func (m *Model) AddFields(fields map[string]FieldDefinition) {
	for name, field := range fields {
		newField := field.DeclareField(m.fields, name)
		if _, exists := m.fields.Get(name); exists {
			log.Panic("Field already exists", "model", m.name, "field", name)
		}
		m.fields.add(newField)
	}
}

// addUpdate adds an update entry for for this field with the given property and the given value
func (f *Field) addUpdate(property string, value interface{}) {
	if Registry.bootstrapped {
		log.Panic("Fields must not be modified after bootstrap", "model", f.model.name, "field", f.name, "property", property, "value", value)
	}
	update := map[string]interface{}{property: value}
	f.updates = append(f.updates, update)
}

// setProperty sets the given property value in this field
// This method uses switch as they are unexported struct fields
func (f *Field) setProperty(property string, value interface{}) {
	switch property {
	case "fieldType":
		f.fieldType = value.(field.Type)
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
	}
}

// SetFieldType overrides the type of Field.
// This may fail at database sync if the table already has values and
// the old type cannot be casted into the new type by the database.
func (f *Field) SetFieldType(value field.Type) *Field {
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
