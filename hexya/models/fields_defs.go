// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
	"github.com/hexya-erp/hexya/hexya/tools/nbutils"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
)

// A FieldDefinition is a struct that declares a new field in a fields collection;
type FieldDefinition interface {
	// DeclareField adds this field to the given FieldsCollection with the given name.
	DeclareField(*FieldsCollection, string)
}

// A BinaryField is a field for storing binary data, such as images.
//
// Clients are expected to handle binary fields as file uploads.
//
// Binary fields are stored in the database. Consider other disk based
// alternatives if you have a large amount of data to store.
type BinaryField struct {
	JSON       string
	String     string
	Help       string
	Stored     bool
	Required   bool
	Unique     bool
	Index      bool
	Compute    Methoder
	Depends    []string
	Related    string
	NoCopy     bool
	GoType     interface{}
	Translate  bool
	OnChange   Methoder
	Constraint Methoder
	Inverse    Methoder
	Default    func(Environment, FieldMap) interface{}
}

// DeclareField adds this binary field to the given FieldsCollection with the given name.
func (bf BinaryField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(string))
	if bf.GoType != nil {
		typ = reflect.TypeOf(bf.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	fieldType := fieldtype.Binary
	json, str := getJSONAndString(name, fieldType, bf.JSON, bf.String)
	compute, inverse, onchange, constraint := getFuncNames(bf.Compute, bf.Inverse, bf.OnChange, bf.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          bf.Help,
		stored:        bf.Stored,
		required:      bf.Required,
		unique:        bf.Unique,
		index:         bf.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       bf.Depends,
		relatedPath:   bf.Related,
		groupOperator: "sum",
		noCopy:        bf.NoCopy,
		structField:   structField,
		fieldType:     fieldType,
		defaultFunc:   bf.Default,
		translate:     bf.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
}

// A BooleanField is a field for storing true/false values.
//
// Clients are expected to handle boolean fields as checkboxes.
type BooleanField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this boolean field to the given FieldsCollection with the given name.
func (bf BooleanField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(bool))
	if bf.GoType != nil {
		typ = reflect.TypeOf(bf.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	fieldType := fieldtype.Boolean
	json, str := getJSONAndString(name, fieldType, bf.JSON, bf.String)
	compute, inverse, onchange, constraint := getFuncNames(bf.Compute, bf.Inverse, bf.OnChange, bf.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          bf.Help,
		stored:        bf.Stored,
		required:      bf.Required,
		unique:        bf.Unique,
		index:         bf.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       bf.Depends,
		relatedPath:   bf.Related,
		groupOperator: strutils.GetDefaultString(bf.GroupOperator, "sum"),
		noCopy:        bf.NoCopy,
		structField:   structField,
		fieldType:     fieldType,
		defaultFunc:   bf.Default,
		translate:     bf.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
}

// A CharField is a field for storing short text. There is no
// default max size, but it can be forced by setting the Size value.
//
// Clients are expected to handle Char fields as single line inputs.
type CharField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	Size          int
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this char field to the given FieldsCollection with the given name.
func (cf CharField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(string))
	if cf.GoType != nil {
		typ = reflect.TypeOf(cf.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	fieldType := fieldtype.Char
	json, str := getJSONAndString(name, fieldType, cf.JSON, cf.String)
	compute, inverse, onchange, constraint := getFuncNames(cf.Compute, cf.Inverse, cf.OnChange, cf.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          cf.Help,
		stored:        cf.Stored,
		required:      cf.Required,
		unique:        cf.Unique,
		index:         cf.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       cf.Depends,
		relatedPath:   cf.Related,
		groupOperator: strutils.GetDefaultString(cf.GroupOperator, "sum"),
		noCopy:        cf.NoCopy,
		structField:   structField,
		size:          cf.Size,
		fieldType:     fieldType,
		defaultFunc:   cf.Default,
		translate:     cf.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
}

// A DateField is a field for storing dates without time.
//
// Clients are expected to handle Date fields with a date picker.
type DateField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this date field to the given FieldsCollection with the given name.
func (df DateField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(dates.Date))
	if df.GoType != nil {
		typ = reflect.TypeOf(df.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	fieldType := fieldtype.Date
	json, str := getJSONAndString(name, fieldType, df.JSON, df.String)
	compute, inverse, onchange, constraint := getFuncNames(df.Compute, df.Inverse, df.OnChange, df.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          df.Help,
		stored:        df.Stored,
		required:      df.Required,
		unique:        df.Unique,
		index:         df.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       df.Depends,
		relatedPath:   df.Related,
		groupOperator: strutils.GetDefaultString(df.GroupOperator, "sum"),
		noCopy:        df.NoCopy,
		structField:   structField,
		fieldType:     fieldType,
		defaultFunc:   df.Default,
		translate:     df.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
}

// A DateTimeField is a field for storing dates with time.
//
// Clients are expected to handle DateTime fields with a date and time picker.
type DateTimeField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this datetime field to the given FieldsCollection with the given name.
func (df DateTimeField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(dates.DateTime))
	if df.GoType != nil {
		typ = reflect.TypeOf(df.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	fieldType := fieldtype.DateTime
	json, str := getJSONAndString(name, fieldType, df.JSON, df.String)
	compute, inverse, onchange, constraint := getFuncNames(df.Compute, df.Inverse, df.OnChange, df.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          df.Help,
		stored:        df.Stored,
		required:      df.Required,
		unique:        df.Unique,
		index:         df.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       df.Depends,
		relatedPath:   df.Related,
		groupOperator: strutils.GetDefaultString(df.GroupOperator, "sum"),
		noCopy:        df.NoCopy,
		structField:   structField,
		fieldType:     fieldType,
		defaultFunc:   df.Default,
		translate:     df.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
}

// A FloatField is a field for storing decimal numbers.
type FloatField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	Digits        nbutils.Digits
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this datetime field to the given FieldsCollection with the given name.
func (ff FloatField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(float64))
	if ff.GoType != nil {
		typ = reflect.TypeOf(ff.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, fieldtype.Float, ff.JSON, ff.String)
	compute, inverse, onchange, constraint := getFuncNames(ff.Compute, ff.Inverse, ff.OnChange, ff.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          ff.Help,
		stored:        ff.Stored,
		required:      ff.Required,
		unique:        ff.Unique,
		index:         ff.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       ff.Depends,
		relatedPath:   ff.Related,
		groupOperator: strutils.GetDefaultString(ff.GroupOperator, "sum"),
		noCopy:        ff.NoCopy,
		structField:   structField,
		digits:        ff.Digits,
		fieldType:     fieldtype.Float,
		defaultFunc:   ff.Default,
		translate:     ff.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
}

// An HTMLField is a field for storing HTML formatted strings.
//
// Clients are expected to handle HTML fields with multi-line HTML editors.
type HTMLField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	Size          int
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this html field to the given FieldsCollection with the given name.
func (tf HTMLField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(string))
	if tf.GoType != nil {
		typ = reflect.TypeOf(tf.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	fieldType := fieldtype.HTML
	json, str := getJSONAndString(name, fieldType, tf.JSON, tf.String)
	compute, inverse, onchange, constraint := getFuncNames(tf.Compute, tf.Inverse, tf.OnChange, tf.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          tf.Help,
		stored:        tf.Stored,
		required:      tf.Required,
		unique:        tf.Unique,
		index:         tf.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       tf.Depends,
		relatedPath:   tf.Related,
		groupOperator: strutils.GetDefaultString(tf.GroupOperator, "sum"),
		noCopy:        tf.NoCopy,
		structField:   structField,
		size:          tf.Size,
		fieldType:     fieldType,
		defaultFunc:   tf.Default,
		translate:     tf.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
}

// An IntegerField is a field for storing non decimal numbers.
type IntegerField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this datetime field to the given FieldsCollection with the given name.
func (i IntegerField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(int64))
	if i.GoType != nil {
		typ = reflect.TypeOf(i.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	fieldType := fieldtype.Integer
	json, str := getJSONAndString(name, fieldType, i.JSON, i.String)
	compute, inverse, onchange, constraint := getFuncNames(i.Compute, i.Inverse, i.OnChange, i.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          i.Help,
		stored:        i.Stored,
		required:      i.Required,
		unique:        i.Unique,
		index:         i.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       i.Depends,
		relatedPath:   i.Related,
		groupOperator: strutils.GetDefaultString(i.GroupOperator, "sum"),
		noCopy:        i.NoCopy,
		structField:   structField,
		fieldType:     fieldType,
		defaultFunc:   i.Default,
		translate:     i.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
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
	Index            bool
	Compute          Methoder
	Depends          []string
	Related          string
	NoCopy           bool
	RelationModel    Modeler
	M2MLinkModelName string
	M2MOurField      string
	M2MTheirField    string
	Translate        bool
	OnChange         Methoder
	Constraint       Methoder
	Filter           Conditioner
	Inverse          Methoder
	Default          func(Environment, FieldMap) interface{}
}

// DeclareField adds this many2many field to the given FieldsCollection with the given name.
func (mf Many2ManyField) DeclareField(fc *FieldsCollection, name string) {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new([]int64)),
	}
	our := mf.M2MOurField
	if our == "" {
		our = fc.model.name
	}
	their := mf.M2MTheirField
	if their == "" {
		their = mf.RelationModel.Underlying().name
	}
	if our == their {
		log.Panic("Many2many relation must have different 'm2m_ours' and 'm2m_theirs'",
			"model", fc.model.name, "field", name, "ours", our, "theirs", their)
	}

	modelNames := []string{our, their}
	sort.Strings(modelNames)
	m2mRelModName := mf.M2MLinkModelName
	if m2mRelModName == "" {
		m2mRelModName = fmt.Sprintf("%s%sRel", modelNames[0], modelNames[1])
	}
	m2mRelModel, m2mOurField, m2mTheirField := createM2MRelModelInfo(m2mRelModName, our, their)

	json, str := getJSONAndString(name, fieldtype.Float, mf.JSON, mf.String)
	compute, inverse, onchange, constraint := getFuncNames(mf.Compute, mf.Inverse, mf.OnChange, mf.Constraint)
	var filter *Condition
	if mf.Filter != nil {
		filter = mf.Filter.Underlying()
	}
	fInfo := &Field{
		model:            fc.model,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             mf.Help,
		stored:           mf.Stored,
		required:         mf.Required,
		index:            mf.Index,
		compute:          compute,
		inverse:          inverse,
		depends:          mf.Depends,
		relatedPath:      mf.Related,
		noCopy:           mf.NoCopy,
		structField:      structField,
		relatedModelName: mf.RelationModel.Underlying().name,
		m2mRelModel:      m2mRelModel,
		m2mOurField:      m2mOurField,
		m2mTheirField:    m2mTheirField,
		fieldType:        fieldtype.Many2Many,
		defaultFunc:      mf.Default,
		translate:        mf.Translate,
		filter:           filter,
		onChange:         onchange,
		constraint:       constraint,
	}
	fc.add(fInfo)
}

// A Many2OneField is a field for storing many-to-one relations,
// i.e. the FK to another model.
//
// Clients are expected to handle many2one fields with a combo-box.
type Many2OneField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	NoCopy        bool
	RelationModel Modeler
	Embed         bool
	Translate     bool
	OnDelete      OnDeleteAction
	OnChange      Methoder
	Constraint    Methoder
	Filter        Conditioner
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this many2one field to the given FieldsCollection with the given name.
func (mf Many2OneField) DeclareField(fc *FieldsCollection, name string) {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new(int64)),
	}
	fieldType := fieldtype.Many2One
	json, str := getJSONAndString(name, fieldType, mf.JSON, mf.String)
	onDelete := SetNull
	if mf.OnDelete != "" {
		onDelete = mf.OnDelete
	}
	noCopy := mf.NoCopy
	required := mf.Required
	if mf.Embed {
		onDelete = Cascade
		required = true
		noCopy = true
	}
	compute, inverse, onchange, constraint := getFuncNames(mf.Compute, mf.Inverse, mf.OnChange, mf.Constraint)
	var filter *Condition
	if mf.Filter != nil {
		filter = mf.Filter.Underlying()
	}
	fInfo := &Field{
		model:            fc.model,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             mf.Help,
		stored:           mf.Stored,
		required:         required,
		index:            mf.Index,
		compute:          compute,
		inverse:          inverse,
		depends:          mf.Depends,
		relatedPath:      mf.Related,
		noCopy:           noCopy,
		structField:      structField,
		embed:            mf.Embed,
		relatedModelName: mf.RelationModel.Underlying().name,
		fieldType:        fieldType,
		onDelete:         onDelete,
		defaultFunc:      mf.Default,
		translate:        mf.Translate,
		onChange:         onchange,
		filter:           filter,
		constraint:       constraint,
	}
	fc.add(fInfo)
}

// A One2ManyField is a field for storing one-to-many relations.
//
// Clients are expected to handle one2many fields with a table.
type One2ManyField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	NoCopy        bool
	RelationModel Modeler
	ReverseFK     string
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Filter        Conditioner
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this one2many field to the given FieldsCollection with the given name.
func (of One2ManyField) DeclareField(fc *FieldsCollection, name string) {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new([]int64)),
	}
	fieldType := fieldtype.One2Many
	json, str := getJSONAndString(name, fieldType, of.JSON, of.String)
	compute, inverse, onchange, constraint := getFuncNames(of.Compute, of.Inverse, of.OnChange, of.Constraint)
	var filter *Condition
	if of.Filter != nil {
		filter = of.Filter.Underlying()
	}
	fInfo := &Field{
		model:            fc.model,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             of.Help,
		stored:           of.Stored,
		required:         of.Required,
		index:            of.Index,
		compute:          compute,
		inverse:          inverse,
		depends:          of.Depends,
		relatedPath:      of.Related,
		noCopy:           of.NoCopy,
		structField:      structField,
		relatedModelName: of.RelationModel.Underlying().name,
		reverseFK:        of.ReverseFK,
		fieldType:        fieldType,
		defaultFunc:      of.Default,
		translate:        of.Translate,
		filter:           filter,
		onChange:         onchange,
		constraint:       constraint,
	}
	fc.add(fInfo)
}

// A One2OneField is a field for storing one-to-one relations,
// i.e. the FK to another model with a unique constraint.
//
// Clients are expected to handle one2one fields with a combo-box.
type One2OneField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	NoCopy        bool
	RelationModel Modeler
	Embed         bool
	Translate     bool
	OnDelete      OnDeleteAction
	OnChange      Methoder
	Constraint    Methoder
	Filter        Conditioner
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this one2one field to the given FieldsCollection with the given name.
func (of One2OneField) DeclareField(fc *FieldsCollection, name string) {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new(int64)),
	}
	fieldType := fieldtype.One2One
	json, str := getJSONAndString(name, fieldType, of.JSON, of.String)
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
	compute, inverse, onchange, constraint := getFuncNames(of.Compute, of.Inverse, of.OnChange, of.Constraint)
	var filter *Condition
	if of.Filter != nil {
		filter = of.Filter.Underlying()
	}
	fInfo := &Field{
		model:            fc.model,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             of.Help,
		stored:           of.Stored,
		required:         required,
		index:            of.Index,
		compute:          compute,
		inverse:          inverse,
		depends:          of.Depends,
		relatedPath:      of.Related,
		noCopy:           noCopy,
		structField:      structField,
		embed:            of.Embed,
		relatedModelName: of.RelationModel.Underlying().name,
		fieldType:        fieldType,
		onDelete:         onDelete,
		defaultFunc:      of.Default,
		translate:        of.Translate,
		onChange:         onchange,
		filter:           filter,
		constraint:       constraint,
	}
	fc.add(fInfo)
}

// A Rev2OneField is a field for storing reverse one-to-one relations,
// i.e. the relation on the model without FK.
//
// Clients are expected to handle rev2one fields with a combo-box.
type Rev2OneField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	NoCopy        bool
	RelationModel Modeler
	ReverseFK     string
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Filter        Conditioner
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this rev2one field to the given FieldsCollection with the given name.
func (rf Rev2OneField) DeclareField(fc *FieldsCollection, name string) {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new(int64)),
	}
	fieldType := fieldtype.Rev2One
	json, str := getJSONAndString(name, fieldType, rf.JSON, rf.String)
	compute, inverse, onchange, constraint := getFuncNames(rf.Compute, rf.Inverse, rf.OnChange, rf.Constraint)
	var filter *Condition
	if rf.Filter != nil {
		filter = rf.Filter.Underlying()
	}
	fInfo := &Field{
		model:            fc.model,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             rf.Help,
		stored:           rf.Stored,
		required:         rf.Required,
		index:            rf.Index,
		compute:          compute,
		inverse:          inverse,
		depends:          rf.Depends,
		relatedPath:      rf.Related,
		noCopy:           rf.NoCopy,
		structField:      structField,
		relatedModelName: rf.RelationModel.Underlying().name,
		reverseFK:        rf.ReverseFK,
		fieldType:        fieldType,
		defaultFunc:      rf.Default,
		translate:        rf.Translate,
		filter:           filter,
		onChange:         onchange,
		constraint:       constraint,
	}
	fc.add(fInfo)
}

// A SelectionField is a field for storing a value from a preset list.
//
// Clients are expected to handle selection fields with a combo-box or radio buttons.
type SelectionField struct {
	JSON       string
	String     string
	Help       string
	Stored     bool
	Required   bool
	Unique     bool
	Index      bool
	Compute    Methoder
	Depends    []string
	Related    string
	NoCopy     bool
	Selection  types.Selection
	Translate  bool
	OnChange   Methoder
	Constraint Methoder
	Inverse    Methoder
	Default    func(Environment, FieldMap) interface{}
}

// DeclareField adds this selection field to the given FieldsCollection with the given name.
func (sf SelectionField) DeclareField(fc *FieldsCollection, name string) {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new(string)),
	}
	json, str := getJSONAndString(name, fieldtype.Selection, sf.JSON, sf.String)
	compute, inverse, onchange, constraint := getFuncNames(sf.Compute, sf.Inverse, sf.OnChange, sf.Constraint)
	fInfo := &Field{
		model:       fc.model,
		acl:         security.NewAccessControlList(),
		name:        name,
		json:        json,
		description: str,
		help:        sf.Help,
		stored:      sf.Stored,
		required:    sf.Required,
		unique:      sf.Unique,
		index:       sf.Index,
		compute:     compute,
		inverse:     inverse,
		depends:     sf.Depends,
		relatedPath: sf.Related,
		noCopy:      sf.NoCopy,
		structField: structField,
		selection:   sf.Selection,
		fieldType:   fieldtype.Selection,
		defaultFunc: sf.Default,
		translate:   sf.Translate,
		onChange:    onchange,
		constraint:  constraint,
	}
	fc.add(fInfo)
}

// A TextField is a field for storing long text. There is no
// default max size, but it can be forced by setting the Size value.
//
// Clients are expected to handle text fields as multi-line inputs.
type TextField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	Size          int
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Default       func(Environment, FieldMap) interface{}
}

// DeclareField adds this text field to the given FieldsCollection with the given name.
func (tf TextField) DeclareField(fc *FieldsCollection, name string) {
	typ := reflect.TypeOf(*new(string))
	if tf.GoType != nil {
		typ = reflect.TypeOf(tf.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	fieldType := fieldtype.Text
	json, str := getJSONAndString(name, fieldType, tf.JSON, tf.String)
	compute, inverse, onchange, constraint := getFuncNames(tf.Compute, tf.Inverse, tf.OnChange, tf.Constraint)
	fInfo := &Field{
		model:         fc.model,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          tf.Help,
		stored:        tf.Stored,
		required:      tf.Required,
		unique:        tf.Unique,
		index:         tf.Index,
		compute:       compute,
		inverse:       inverse,
		depends:       tf.Depends,
		relatedPath:   tf.Related,
		groupOperator: strutils.GetDefaultString(tf.GroupOperator, "sum"),
		noCopy:        tf.NoCopy,
		structField:   structField,
		size:          tf.Size,
		fieldType:     fieldType,
		defaultFunc:   tf.Default,
		translate:     tf.Translate,
		onChange:      onchange,
		constraint:    constraint,
	}
	fc.add(fInfo)
}

// getJSONAndString computes the default json and description fields for the
// given name. It returns this default value unless given json or str are not
// empty strings, in which case the latters are returned.
func getJSONAndString(name string, typ fieldtype.Type, json, str string) (string, string) {
	if json == "" {
		json = snakeCaseFieldName(name, typ)
	}
	if str == "" {
		str = strutils.TitleString(name)
	}
	return json, str
}

// getFuncNames returns the methods names of the given Methoder instances in the same order.
// Returns "" if the Methoder is nil
func getFuncNames(compute, inverse, onchange, constraint Methoder) (string, string, string, string) {
	var com, inv, onc, con string
	if compute != nil {
		com = compute.Underlying().name
	}
	if inverse != nil {
		inv = inverse.Underlying().name
	}
	if onchange != nil {
		onc = onchange.Underlying().name
	}
	if constraint != nil {
		con = constraint.Underlying().name
	}
	return com, inv, onc, con
}

// AddFields adds the given fields to the model.
func (m *Model) AddFields(fields map[string]FieldDefinition) {
	for name, field := range fields {
		field.DeclareField(m.fields, name)
	}
}

// SetString overrides the value of the String parameter of this Field
func (f *Field) SetString(value string) *Field {
	f.description = value
	return f
}

// SetHelp overrides the value of the Help parameter of this Field
func (f *Field) SetHelp(value string) *Field {
	f.help = value
	return f
}

// SetGroupOperator overrides the value of the GroupOperator parameter of this Field
func (f *Field) SetGroupOperator(value string) *Field {
	f.groupOperator = value
	return f
}

// SetRelated overrides the value of the Related parameter of this Field
func (f *Field) SetRelated(value string) *Field {
	f.relatedPath = value
	return f
}

// SetCompute overrides the value of the Compute parameter of this Field
func (f *Field) SetCompute(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.compute = methName
	return f
}

// SetDepends overrides the value of the Depends parameter of this Field
func (f *Field) SetDepends(value []string) *Field {
	f.depends = value
	return f
}

// SetStored overrides the value of the Stored parameter of this Field
func (f *Field) SetStored(value bool) *Field {
	f.stored = value
	return f
}

// SetRequired overrides the value of the Required parameter of this Field
func (f *Field) SetRequired(value bool) *Field {
	f.required = value
	return f
}

// SetUnique overrides the value of the Unique parameter of this Field
func (f *Field) SetUnique(value bool) *Field {
	f.unique = value
	return f
}

// SetIndex overrides the value of the Index parameter of this Field
func (f *Field) SetIndex(value bool) *Field {
	f.index = value
	return f
}

// SetNoCopy overrides the value of the NoCopy parameter of this Field
func (f *Field) SetNoCopy(value bool) *Field {
	f.noCopy = value
	return f
}

// SetTranslate overrides the value of the Translate parameter of this Field
func (f *Field) SetTranslate(value bool) *Field {
	f.translate = value
	return f
}

// SetDefault overrides the value of the Default parameter of this Field
func (f *Field) SetDefault(value func(Environment, FieldMap) interface{}) *Field {
	f.defaultFunc = value
	return f
}

// SetSelection overrides the value of the Selection parameter of this Field
func (f *Field) SetSelection(value types.Selection) *Field {
	f.selection = value
	return f
}

// UpdateSelection updates the value of the Selection parameter of this Field
// with the given value. Existing keys are overridden.
func (f *Field) UpdateSelection(value types.Selection) *Field {
	for k, v := range value {
		f.selection[k] = v
	}
	return f
}

// SetOnchange overrides the value of the Onchange parameter of this Field
func (f *Field) SetOnchange(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.onChange = methName
	return f
}

// SetConstraint overrides the value of the Constraint parameter of this Field
func (f *Field) SetConstraint(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.constraint = methName
	return f
}

// SetInverse overrides the value of the Inverse parameter of this Field
func (f *Field) SetInverse(value Methoder) *Field {
	var methName string
	if value != nil {
		methName = value.Underlying().name
	}
	f.inverse = methName
	return f
}
