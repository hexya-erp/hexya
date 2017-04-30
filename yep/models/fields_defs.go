// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/tools"
	"github.com/npiganeau/yep/yep/tools/logging"
)

// A SimpleFieldParams holds all the possible options for a simple (not relational) field
type SimpleFieldParams struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       string
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	GoType        interface{}
	Translate     bool
	override      bool
}

// A FloatFieldParams holds all the possible options for a float field
type FloatFieldParams struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       string
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	Digits        types.Digits
	GoType        interface{}
	Translate     bool
	override      bool
}

// A StringFieldParams holds all the possible options for a string field
type StringFieldParams struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Unique        bool
	Index         bool
	Compute       string
	Depends       []string
	Related       string
	GroupOperator string
	NoCopy        bool
	Size          int
	GoType        interface{}
	Translate     bool
	override      bool
}

// A SelectionFieldParams holds all the possible options for a selection field
type SelectionFieldParams struct {
	JSON      string
	String    string
	Help      string
	Stored    bool
	Required  bool
	Unique    bool
	Index     bool
	Compute   string
	Depends   []string
	Related   string
	NoCopy    bool
	Selection Selection
	Translate bool
	override  bool
}

// A ForeignKeyFieldParams holds all the possible options for a many2one or one2one field
type ForeignKeyFieldParams struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Index         bool
	Compute       string
	Depends       []string
	Related       string
	NoCopy        bool
	RelationModel string
	Embed         bool
	Translate     bool
	OnDelete      OnDeleteAction
	override      bool
}

// A ReverseFieldParams holds all the possible options for a one2many or rev2one field
type ReverseFieldParams struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	Index         bool
	Compute       string
	Depends       []string
	Related       string
	NoCopy        bool
	RelationModel string
	ReverseFK     string
	Translate     bool
	override      bool
}

// A Many2ManyFieldParams holds all the possible options for a many2many field
type Many2ManyFieldParams struct {
	JSON             string
	String           string
	Help             string
	Stored           bool
	Required         bool
	Index            bool
	Compute          string
	Depends          []string
	Related          string
	NoCopy           bool
	RelationModel    string
	M2MLinkModelName string
	M2MOurField      string
	M2MTheirField    string
	Translate        bool
	override         bool
}

// getJSONAndString computes the default json and description fields for the
// given name. It returns this default value unless given json or str are not
// empty strings, in which case the latters are returned.
func getJSONAndString(name string, typ types.FieldType, json, str string) (string, string) {
	if json == "" {
		json = snakeCaseFieldName(name, typ)
	}
	if str == "" {
		str = tools.TitleString(name)
	}
	return json, str
}

// addSimpleField adds or overrides a new simple field with the given data and returns the fieldInfo
func (m *Model) addSimpleField(name string, params SimpleFieldParams, fieldType types.FieldType, typ reflect.Type) *fieldInfo {
	if params.GoType != nil {
		typ = reflect.TypeOf(params.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, fieldType, params.JSON, params.String)
	fInfo := &fieldInfo{
		model:         m,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          params.Help,
		stored:        params.Stored,
		required:      params.Required,
		unique:        params.Unique,
		index:         params.Index,
		compute:       params.Compute,
		depends:       params.Depends,
		relatedPath:   params.Related,
		groupOperator: tools.GetDefaultString(params.GroupOperator, "sum"),
		noCopy:        params.NoCopy,
		structField:   structField,
		fieldType:     fieldType,
	}
	if params.override {
		m.fields.override(fInfo)
	} else {
		m.fields.add(fInfo)
	}
	return fInfo
}

// addStringField adds or overrides a new string field with the given data and returns the fieldInfo
func (m *Model) addStringField(name string, params StringFieldParams, fieldType types.FieldType, typ reflect.Type) *fieldInfo {
	if params.GoType != nil {
		typ = reflect.TypeOf(params.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, fieldType, params.JSON, params.String)
	fInfo := &fieldInfo{
		model:         m,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          params.Help,
		stored:        params.Stored,
		required:      params.Required,
		unique:        params.Unique,
		index:         params.Index,
		compute:       params.Compute,
		depends:       params.Depends,
		relatedPath:   params.Related,
		groupOperator: tools.GetDefaultString(params.GroupOperator, "sum"),
		noCopy:        params.NoCopy,
		structField:   structField,
		size:          params.Size,
		fieldType:     fieldType,
	}
	if params.override {
		m.fields.override(fInfo)
	} else {
		m.fields.add(fInfo)
	}
	return fInfo
}

// addForeignKeyField adds or overrides a new FK field with the given data and returns the fieldInfo
func (m *Model) addForeignKeyField(name string, params ForeignKeyFieldParams, fieldType types.FieldType, typ reflect.Type) *fieldInfo {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new(int64)),
	}
	json, str := getJSONAndString(name, fieldType, params.JSON, params.String)
	onDelete := SetNull
	if params.OnDelete != "" {
		onDelete = params.OnDelete
	}
	fInfo := &fieldInfo{
		model:            m,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             params.Help,
		stored:           params.Stored,
		required:         params.Required,
		index:            params.Index,
		compute:          params.Compute,
		depends:          params.Depends,
		relatedPath:      params.Related,
		noCopy:           params.NoCopy,
		structField:      structField,
		embed:            params.Embed,
		relatedModelName: params.RelationModel,
		fieldType:        fieldType,
		onDelete:         onDelete,
	}
	if params.override {
		m.fields.override(fInfo)
	} else {
		m.fields.add(fInfo)
	}
	return fInfo
}

// addReverseField adds or overrides a new reverse field with the given data and returns the fieldInfo
func (m *Model) addReverseField(name string, params ReverseFieldParams, fieldType types.FieldType, typ reflect.Type) *fieldInfo {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new([]int64)),
	}
	json, str := getJSONAndString(name, fieldType, params.JSON, params.String)
	fInfo := &fieldInfo{
		model:            m,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             params.Help,
		stored:           params.Stored,
		required:         params.Required,
		index:            params.Index,
		compute:          params.Compute,
		depends:          params.Depends,
		relatedPath:      params.Related,
		noCopy:           params.NoCopy,
		structField:      structField,
		relatedModelName: params.RelationModel,
		reverseFK:        params.ReverseFK,
		fieldType:        fieldType,
	}
	if params.override {
		m.fields.override(fInfo)
	} else {
		m.fields.add(fInfo)
	}
	return fInfo
}

// AddBinaryField adds a database stored binary field with the given name to this Model.
// Binary fields are mapped to string type in go.
func (m *Model) AddBinaryField(name string, params SimpleFieldParams) {
	m.addSimpleField(name, params, types.Binary, reflect.TypeOf(*new(string)))
}

// AddBooleanField adds a boolean field with the given name to this Model.
func (m *Model) AddBooleanField(name string, params SimpleFieldParams) {
	m.addSimpleField(name, params, types.Boolean, reflect.TypeOf(true))
}

// AddCharField adds a single line text field with the given name to this Model.
// Char fields are mapped to strings in go. There is no limitation in the size
// of the string, unless specified in the parameters.
func (m *Model) AddCharField(name string, params StringFieldParams) {
	m.addStringField(name, params, types.Char, reflect.TypeOf(*new(string)))
}

// AddDateField adds a date field with the given name to this Model.
// Date fields are mapped to Date type.
func (m *Model) AddDateField(name string, params SimpleFieldParams) {
	m.addSimpleField(name, params, types.Date, reflect.TypeOf(*new(Date)))
}

// AddDateTimeField adds a datetime field with the given name to this Model.
// DateTime fields are mapped to DateTime type.
func (m *Model) AddDateTimeField(name string, params SimpleFieldParams) {
	m.addSimpleField(name, params, types.DateTime, reflect.TypeOf(*new(DateTime)))
}

// AddFloatField adds a float field with the given name to this Model.
// Float fields are mapped to go float64 type and stored as numeric in database.
func (m *Model) AddFloatField(name string, params FloatFieldParams) {
	typ := reflect.TypeOf(*new(float64))
	if params.GoType != nil {
		typ = reflect.TypeOf(params.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, types.Float, params.JSON, params.String)
	fInfo := &fieldInfo{
		model:         m,
		acl:           security.NewAccessControlList(),
		name:          name,
		json:          json,
		description:   str,
		help:          params.Help,
		stored:        params.Stored,
		required:      params.Required,
		unique:        params.Unique,
		index:         params.Index,
		compute:       params.Compute,
		depends:       params.Depends,
		relatedPath:   params.Related,
		groupOperator: tools.GetDefaultString(params.GroupOperator, "sum"),
		noCopy:        params.NoCopy,
		structField:   structField,
		digits:        params.Digits,
		fieldType:     types.Float,
	}
	if params.override {
		m.fields.override(fInfo)
	} else {
		m.fields.add(fInfo)
	}
}

// AddHTMLField adds an html field with the given name to this Model.
// HTML fields are mapped to string type in go.
func (m *Model) AddHTMLField(name string, params StringFieldParams) {
	m.addStringField(name, params, types.HTML, reflect.TypeOf(*new(string)))
}

// AddIntegerField adds an integer field with the given name to this Model.
// Integer fields are mapped to int64 type in go.
func (m *Model) AddIntegerField(name string, params SimpleFieldParams) {
	m.addSimpleField(name, params, types.Integer, reflect.TypeOf(*new(int64)))
}

// AddMany2ManyField adds a many2many field with the given name to this Model.
func (m *Model) AddMany2ManyField(name string, params Many2ManyFieldParams) {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new([]int64)),
	}
	our := params.M2MOurField
	if our == "" {
		our = m.name
	}
	their := params.M2MTheirField
	if their == "" {
		their = params.RelationModel
	}
	if our == their {
		logging.LogAndPanic(log, "Many2many relation must have different 'm2m_ours' and 'm2m_theirs'",
			"model", m.name, "field", name, "ours", our, "theirs", their)
	}

	modelNames := []string{our, their}
	sort.Strings(modelNames)
	m2mRelModName := params.M2MLinkModelName
	if m2mRelModName == "" {
		m2mRelModName = fmt.Sprintf("%s%sRel", modelNames[0], modelNames[1])
	}
	m2mRelModel, m2mOurField, m2mTheirField := createM2MRelModelInfo(m2mRelModName, our, their)

	json, str := getJSONAndString(name, types.Float, params.JSON, params.String)
	fInfo := &fieldInfo{
		model:            m,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             params.Help,
		stored:           params.Stored,
		required:         params.Required,
		index:            params.Index,
		compute:          params.Compute,
		depends:          params.Depends,
		relatedPath:      params.Related,
		noCopy:           params.NoCopy,
		structField:      structField,
		relatedModelName: params.RelationModel,
		m2mRelModel:      m2mRelModel,
		m2mOurField:      m2mOurField,
		m2mTheirField:    m2mTheirField,
		fieldType:        types.Many2Many,
	}
	if params.override {
		m.fields.override(fInfo)
	} else {
		m.fields.add(fInfo)
	}
}

// AddMany2OneField adds a many2one field with the given name to this Model.
func (m *Model) AddMany2OneField(name string, params ForeignKeyFieldParams) {
	m.addForeignKeyField(name, params, types.Many2One, reflect.TypeOf(*new(int64)))
}

// AddOne2ManyField adds a one2many field with the given name to this Model.
func (m *Model) AddOne2ManyField(name string, params ReverseFieldParams) {
	m.addReverseField(name, params, types.One2Many, reflect.TypeOf(*new(int64)))
}

// AddOne2OneField adds a one2one field with the given name to this Model.
func (m *Model) AddOne2OneField(name string, params ForeignKeyFieldParams) {
	fInfo := m.addForeignKeyField(name, params, types.One2One, reflect.TypeOf(*new(int64)))
	fInfo.unique = true
}

// AddRev2OneField adds a rev2one field with the given name to this Model.
func (m *Model) AddRev2OneField(name string, params ReverseFieldParams) {
	m.addReverseField(name, params, types.Rev2One, reflect.TypeOf(*new(int64)))
}

// AddSelectionField adds a selection field with the given name to this Model.
func (m *Model) AddSelectionField(name string, params SelectionFieldParams) {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new(Selection)),
	}
	json, str := getJSONAndString(name, types.Float, params.JSON, params.String)
	fInfo := &fieldInfo{
		model:       m,
		acl:         security.NewAccessControlList(),
		name:        name,
		json:        json,
		description: str,
		help:        params.Help,
		stored:      params.Stored,
		required:    params.Required,
		unique:      params.Unique,
		index:       params.Index,
		compute:     params.Compute,
		depends:     params.Depends,
		relatedPath: params.Related,
		noCopy:      params.NoCopy,
		structField: structField,
		selection:   params.Selection,
		fieldType:   types.Selection,
	}
	if params.override {
		m.fields.override(fInfo)
	} else {
		m.fields.add(fInfo)
	}
}

// AddTextField adds a multi line text field with the given name to this Model.
// Text fields are mapped to strings in go. There is no limitation in the size
// of the string, unless specified in the parameters.
func (m *Model) AddTextField(name string, params StringFieldParams) {
	m.addStringField(name, params, types.Text, reflect.TypeOf(*new(string)))
}

// OverrideBinaryField overrides the database stored binary field with the given name of this Model.
// Binary fields are mapped to '[]byte' type in go.
func (m *Model) OverrideBinaryField(name string, params SimpleFieldParams) {
	params.override = true
	m.AddBinaryField(name, params)
}

// OverrideBooleanField overrides the boolean field with the given name of this Model.
func (m *Model) OverrideBooleanField(name string, params SimpleFieldParams) {
	params.override = true
	m.AddBooleanField(name, params)
}

// OverrideCharField overrides the single line text field with the given name of this Model.
// Char fields are mapped to strings in go. There is no limitation in the size
// of the string, unless specified in the parameters.
func (m *Model) OverrideCharField(name string, params StringFieldParams) {
	params.override = true
	m.AddCharField(name, params)
}

// OverrideDateField overrides the date field with the given name of this Model.
// Date fields are mapped to Date type.
func (m *Model) OverrideDateField(name string, params SimpleFieldParams) {
	params.override = true
	m.AddDateField(name, params)
}

// OverrideDateTimeField overrides the datetime field with the given name of this Model.
// DateTime fields are mapped to DateTime type.
func (m *Model) OverrideDateTimeField(name string, params SimpleFieldParams) {
	params.override = true
	m.AddDateTimeField(name, params)
}

// OverrideFloatField overrides the float field with the given name of this Model.
// Float fields are mapped to go float64 type and stored as numeric in database.
func (m *Model) OverrideFloatField(name string, params FloatFieldParams) {
	params.override = true
	m.AddFloatField(name, params)
}

// OverrideHTMLField overrides then html field with the given name of this Model.
// HTML fields are mapped to string type in go.
func (m *Model) OverrideHTMLField(name string, params StringFieldParams) {
	params.override = true
	m.AddHTMLField(name, params)
}

// OverrideIntegerField overrides then integer field with the given name of this Model.
// Integer fields are mapped to int64 type in go.
func (m *Model) OverrideIntegerField(name string, params SimpleFieldParams) {
	params.override = true
	m.AddIntegerField(name, params)
}

// OverrideMany2ManyField overrides the many2many field with the given name of this Model.
func (m *Model) OverrideMany2ManyField(name string, params Many2ManyFieldParams) {
	params.override = true
	m.AddMany2ManyField(name, params)
}

// OverrideMany2OneField overrides the many2one field with the given name of this Model.
func (m *Model) OverrideMany2OneField(name string, params ForeignKeyFieldParams) {
	params.override = true
	m.AddMany2OneField(name, params)
}

// OverrideOne2ManyField overrides the one2many field with the given name of this Model.
func (m *Model) OverrideOne2ManyField(name string, params ReverseFieldParams) {
	params.override = true
	m.AddOne2ManyField(name, params)
}

// OverrideOne2OneField overrides the one2one field with the given name of this Model.
func (m *Model) OverrideOne2OneField(name string, params ForeignKeyFieldParams) {
	params.override = true
	m.AddOne2OneField(name, params)
}

// OverrideRev2OneField overrides the rev2one field with the given name of this Model.
func (m *Model) OverrideRev2OneField(name string, params ReverseFieldParams) {
	params.override = true
	m.AddRev2OneField(name, params)
}

// OverrideSelectionField overrides the selection field with the given name of this Model.
func (m *Model) OverrideSelectionField(name string, params SelectionFieldParams) {
	params.override = true
	m.AddSelectionField(name, params)
}

// OverrideTextField overrides the multi line text field with the given name of this Model.
// Text fields are mapped to strings in go. There is no limitation in the size
// of the string, unless specified in the parameters.
func (m *Model) OverrideTextField(name string, params StringFieldParams) {
	params.override = true
	m.AddTextField(name, params)
}
