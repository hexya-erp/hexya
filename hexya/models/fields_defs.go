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
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
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
	Default       func(Environment, FieldMap) interface{}
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
	Default       func(Environment, FieldMap) interface{}
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
	Default       func(Environment, FieldMap) interface{}
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
	Selection types.Selection
	Translate bool
	Default   func(Environment, FieldMap) interface{}
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
	Default       func(Environment, FieldMap) interface{}
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
	Default       func(Environment, FieldMap) interface{}
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
	Default          func(Environment, FieldMap) interface{}
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

// addSimpleField adds or overrides a new simple field with the given data and returns the Field
func (m *Model) addSimpleField(name string, params SimpleFieldParams, fieldType fieldtype.Type, typ reflect.Type) *Field {
	if params.GoType != nil {
		typ = reflect.TypeOf(params.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, fieldType, params.JSON, params.String)
	fInfo := &Field{
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
		groupOperator: strutils.GetDefaultString(params.GroupOperator, "sum"),
		noCopy:        params.NoCopy,
		structField:   structField,
		fieldType:     fieldType,
		defaultFunc:   params.Default,
		translate:     params.Translate,
	}
	m.fields.add(fInfo)
	return fInfo
}

// addStringField adds or overrides a new string field with the given data and returns the Field
func (m *Model) addStringField(name string, params StringFieldParams, fieldType fieldtype.Type, typ reflect.Type) *Field {
	if params.GoType != nil {
		typ = reflect.TypeOf(params.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, fieldType, params.JSON, params.String)
	fInfo := &Field{
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
		groupOperator: strutils.GetDefaultString(params.GroupOperator, "sum"),
		noCopy:        params.NoCopy,
		structField:   structField,
		size:          params.Size,
		fieldType:     fieldType,
		defaultFunc:   params.Default,
		translate:     params.Translate,
	}
	m.fields.add(fInfo)
	return fInfo
}

// addForeignKeyField adds or overrides a new FK field with the given data and returns the Field
func (m *Model) addForeignKeyField(name string, params ForeignKeyFieldParams, fieldType fieldtype.Type, typ reflect.Type) *Field {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new(int64)),
	}
	json, str := getJSONAndString(name, fieldType, params.JSON, params.String)
	onDelete := SetNull
	if params.OnDelete != "" {
		onDelete = params.OnDelete
	}
	noCopy := params.NoCopy
	required := params.Required
	if params.Embed {
		onDelete = Cascade
		required = true
		noCopy = true
	}
	fInfo := &Field{
		model:            m,
		acl:              security.NewAccessControlList(),
		name:             name,
		json:             json,
		description:      str,
		help:             params.Help,
		stored:           params.Stored,
		required:         required,
		index:            params.Index,
		compute:          params.Compute,
		depends:          params.Depends,
		relatedPath:      params.Related,
		noCopy:           noCopy,
		structField:      structField,
		embed:            params.Embed,
		relatedModelName: params.RelationModel,
		fieldType:        fieldType,
		onDelete:         onDelete,
		defaultFunc:      params.Default,
		translate:        params.Translate,
	}
	m.fields.add(fInfo)
	return fInfo
}

// addReverseField adds or overrides a new reverse field with the given data and returns the Field
func (m *Model) addReverseField(name string, params ReverseFieldParams, fieldType fieldtype.Type, typ reflect.Type) *Field {
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, fieldType, params.JSON, params.String)
	fInfo := &Field{
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
		defaultFunc:      params.Default,
		translate:        params.Translate,
	}
	m.fields.add(fInfo)
	return fInfo
}

// AddBinaryField adds a database stored binary field with the given name to this Model.
// Binary fields are mapped to string type in go.
func (m *Model) AddBinaryField(name string, params SimpleFieldParams) *Field {
	return m.addSimpleField(name, params, fieldtype.Binary, reflect.TypeOf(*new(string)))
}

// AddBooleanField adds a boolean field with the given name to this Model.
func (m *Model) AddBooleanField(name string, params SimpleFieldParams) *Field {
	return m.addSimpleField(name, params, fieldtype.Boolean, reflect.TypeOf(true))
}

// AddCharField adds a single line text field with the given name to this Model.
// Char fields are mapped to strings in go. There is no limitation in the size
// of the string, unless specified in the parameters.
func (m *Model) AddCharField(name string, params StringFieldParams) *Field {
	return m.addStringField(name, params, fieldtype.Char, reflect.TypeOf(*new(string)))
}

// AddDateField adds a date field with the given name to this Model.
// Date fields are mapped to Date type.
func (m *Model) AddDateField(name string, params SimpleFieldParams) *Field {
	return m.addSimpleField(name, params, fieldtype.Date, reflect.TypeOf(*new(types.Date)))
}

// AddDateTimeField adds a datetime field with the given name to this Model.
// DateTime fields are mapped to DateTime type.
func (m *Model) AddDateTimeField(name string, params SimpleFieldParams) *Field {
	return m.addSimpleField(name, params, fieldtype.DateTime, reflect.TypeOf(*new(types.DateTime)))
}

// AddFloatField adds a float field with the given name to this Model.
// Float fields are mapped to go float64 type and stored as numeric in database.
func (m *Model) AddFloatField(name string, params FloatFieldParams) *Field {
	typ := reflect.TypeOf(*new(float64))
	if params.GoType != nil {
		typ = reflect.TypeOf(params.GoType).Elem()
	}
	structField := reflect.StructField{
		Name: name,
		Type: typ,
	}
	json, str := getJSONAndString(name, fieldtype.Float, params.JSON, params.String)
	fInfo := &Field{
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
		groupOperator: strutils.GetDefaultString(params.GroupOperator, "sum"),
		noCopy:        params.NoCopy,
		structField:   structField,
		digits:        params.Digits,
		fieldType:     fieldtype.Float,
		defaultFunc:   params.Default,
		translate:     params.Translate,
	}
	m.fields.add(fInfo)
	return fInfo
}

// AddHTMLField adds an html field with the given name to this Model.
// HTML fields are mapped to string type in go.
func (m *Model) AddHTMLField(name string, params StringFieldParams) *Field {
	return m.addStringField(name, params, fieldtype.HTML, reflect.TypeOf(*new(string)))
}

// AddIntegerField adds an integer field with the given name to this Model.
// Integer fields are mapped to int64 type in go.
func (m *Model) AddIntegerField(name string, params SimpleFieldParams) *Field {
	return m.addSimpleField(name, params, fieldtype.Integer, reflect.TypeOf(*new(int64)))
}

// AddMany2ManyField adds a many2many field with the given name to this Model.
func (m *Model) AddMany2ManyField(name string, params Many2ManyFieldParams) *Field {
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
		log.Panic("Many2many relation must have different 'm2m_ours' and 'm2m_theirs'",
			"model", m.name, "field", name, "ours", our, "theirs", their)
	}

	modelNames := []string{our, their}
	sort.Strings(modelNames)
	m2mRelModName := params.M2MLinkModelName
	if m2mRelModName == "" {
		m2mRelModName = fmt.Sprintf("%s%sRel", modelNames[0], modelNames[1])
	}
	m2mRelModel, m2mOurField, m2mTheirField := createM2MRelModelInfo(m2mRelModName, our, their)

	json, str := getJSONAndString(name, fieldtype.Float, params.JSON, params.String)
	fInfo := &Field{
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
		fieldType:        fieldtype.Many2Many,
		defaultFunc:      params.Default,
		translate:        params.Translate,
	}
	m.fields.add(fInfo)
	return fInfo
}

// AddMany2OneField adds a many2one field with the given name to this Model.
func (m *Model) AddMany2OneField(name string, params ForeignKeyFieldParams) *Field {
	return m.addForeignKeyField(name, params, fieldtype.Many2One, reflect.TypeOf(*new(int64)))
}

// AddOne2ManyField adds a one2many field with the given name to this Model.
func (m *Model) AddOne2ManyField(name string, params ReverseFieldParams) *Field {
	return m.addReverseField(name, params, fieldtype.One2Many, reflect.TypeOf(*new([]int64)))
}

// AddOne2OneField adds a one2one field with the given name to this Model.
func (m *Model) AddOne2OneField(name string, params ForeignKeyFieldParams) *Field {
	fInfo := m.addForeignKeyField(name, params, fieldtype.One2One, reflect.TypeOf(*new(int64)))
	fInfo.unique = true
	return fInfo
}

// AddRev2OneField adds a rev2one field with the given name to this Model.
func (m *Model) AddRev2OneField(name string, params ReverseFieldParams) *Field {
	return m.addReverseField(name, params, fieldtype.Rev2One, reflect.TypeOf(*new(int64)))
}

// AddSelectionField adds a selection field with the given name to this Model.
func (m *Model) AddSelectionField(name string, params SelectionFieldParams) *Field {
	structField := reflect.StructField{
		Name: name,
		Type: reflect.TypeOf(*new(types.Selection)),
	}
	json, str := getJSONAndString(name, fieldtype.Float, params.JSON, params.String)
	fInfo := &Field{
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
		fieldType:   fieldtype.Selection,
		defaultFunc: params.Default,
		translate:   params.Translate,
	}
	m.fields.add(fInfo)
	return fInfo
}

// AddTextField adds a multi line text field with the given name to this Model.
// Text fields are mapped to strings in go. There is no limitation in the size
// of the string, unless specified in the parameters.
func (m *Model) AddTextField(name string, params StringFieldParams) *Field {
	return m.addStringField(name, params, fieldtype.Text, reflect.TypeOf(*new(string)))
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
func (f *Field) SetCompute(value string) *Field {
	f.compute = value
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
