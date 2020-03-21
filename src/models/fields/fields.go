// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package fields

import (
	"fmt"
	"log"
	"sort"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	"github.com/hexya-erp/hexya/src/tools/strutils"
)

// A FieldDefinition is a struct that declares a new field in a fields collection;
type FieldDefinition interface {
	// DeclareField creates a field for the given FieldsCollection with the given name and returns the created field.
	DeclareField(*models.FieldsCollection, string) *models.Field
}

// A Binary is a field for storing binary data, such as images.
//
// Clients are expected to handle binary fields as file uploads.
//
// TypeBinary fields are stored in the database. Consider other disk based
// alternatives if you have a large amount of data to store.
type Binary struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	GoType          interface{}
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a binary field for the given models.FieldsCollection with the given name.
func (bf Binary) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	return models.CreateFieldFromStruct(fc, &bf, name, fieldtype.Binary, new(string))
}

// A Boolean is a field for storing true/false values.
//
// Clients are expected to handle boolean fields as checkboxes.
type Boolean struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	GoType          interface{}
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a boolean field for the given models.FieldsCollection with the given name.
func (bf Boolean) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	if bf.Default == nil {
		bf.Default = models.DefaultValue(false)
	}
	return models.CreateFieldFromStruct(fc, &bf, name, fieldtype.Boolean, new(bool))
}

// A Char is a field for storing short text. There is no
// default max size, but it can be forced by setting the Size value.
//
// Clients are expected to handle TypeChar fields as single line inputs.
type Char struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	Size            int
	GoType          interface{}
	Translate       bool
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a char field for the given models.FieldsCollection with the given name.
func (cf Char) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &cf, name, fieldtype.Char, new(string))
	fInfo.SetProperty("size", cf.Size)
	return fInfo
}

// A Date is a field for storing dates without time.
//
// Clients are expected to handle Date fields with a date picker.
type Date struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	GroupOperator   string
	NoCopy          bool
	GoType          interface{}
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a date field for the given models.FieldsCollection with the given name.
func (df Date) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &df, name, fieldtype.Date, new(dates.Date))
	fInfo.SetProperty("groupOperator", strutils.GetDefaultString(df.GroupOperator, "sum"))
	return fInfo
}

// A DateTime is a field for storing dates with time.
//
// Clients are expected to handle DateTime fields with a date and time picker.
type DateTime struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	GroupOperator   string
	NoCopy          bool
	GoType          interface{}
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a datetime field for the given models.FieldsCollection with the given name.
func (df DateTime) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &df, name, fieldtype.DateTime, new(dates.DateTime))
	fInfo.SetProperty("groupOperator", strutils.GetDefaultString(df.GroupOperator, "sum"))
	return fInfo
}

// A Float is a field for storing decimal numbers.
type Float struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	GroupOperator   string
	NoCopy          bool
	Digits          nbutils.Digits
	GoType          interface{}
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField adds this datetime field for the given models.FieldsCollection with the given name.
func (ff Float) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	if ff.Default == nil {
		ff.Default = models.DefaultValue(0)
	}
	fInfo := models.CreateFieldFromStruct(fc, &ff, name, fieldtype.Float, new(float64))
	fInfo.SetProperty("groupOperator", strutils.GetDefaultString(ff.GroupOperator, "sum"))
	fInfo.SetProperty("digits", ff.Digits)
	return fInfo
}

// An HTML is a field for storing HTML formatted strings.
//
// Clients are expected to handle HTML fields with multi-line HTML editors.
type HTML struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	Size            int
	GoType          interface{}
	Translate       bool
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a html field for the given models.FieldsCollection with the given name.
func (tf HTML) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &tf, name, fieldtype.HTML, new(string))
	fInfo.SetProperty("size", tf.Size)
	return fInfo
}

// An Integer is a field for storing non decimal numbers.
type Integer struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	GroupOperator   string
	NoCopy          bool
	GoType          interface{}
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a datetime field for the given models.FieldsCollection with the given name.
func (i Integer) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	if i.Default == nil {
		i.Default = models.DefaultValue(0)
	}
	fInfo := models.CreateFieldFromStruct(fc, &i, name, fieldtype.Integer, new(int64))
	fInfo.SetProperty("groupOperator", strutils.GetDefaultString(i.GroupOperator, "sum"))
	return fInfo
}

// A Many2Many is a field for storing many-to-many relations.
//
// Clients are expected to handle many2many fields with a table or with tags.
type Many2Many struct {
	JSON             string
	String           string
	Help             string
	Stored           bool
	Required         bool
	ReadOnly         bool
	RequiredFunc     func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc     func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc    func(models.Environment) (bool, models.Conditioner)
	Index            bool
	Compute          models.Methoder
	Depends          []string
	Related          string
	NoCopy           bool
	RelationModel    models.Modeler
	M2MLinkModelName string
	M2MOurField      string
	M2MTheirField    string
	OnChange         models.Methoder
	OnChangeWarning  models.Methoder
	OnChangeFilters  models.Methoder
	Constraint       models.Methoder
	Filter           models.Conditioner
	Inverse          models.Methoder
	Default          func(models.Environment) interface{}
}

// DeclareField creates a many2many field for the given models.FieldsCollection with the given name.
func (mf Many2Many) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &mf, name, fieldtype.Many2Many, new([]int64))
	our := mf.M2MOurField
	if our == "" {
		our = fc.Model().Name()
	}
	their := mf.M2MTheirField
	if their == "" {
		their = mf.RelationModel.Underlying().Name()
	}
	if our == their {
		log.Panic("Many2many relation must have different 'M2MOurField' and 'M2MTheirField'",
			"model", fc.Model().Name(), "field", name, "ours", our, "theirs", their)
	}

	modelNames := []string{fc.Model().Name(), mf.RelationModel.Underlying().Name()}
	sort.Strings(modelNames)
	m2mRelModName := mf.M2MLinkModelName
	if m2mRelModName == "" {
		m2mRelModName = fmt.Sprintf("%s%sRel", modelNames[0], modelNames[1])
	}
	m2mRelModel, m2mOurField, m2mTheirField := models.CreateM2MRelModelInfo(m2mRelModName, fc.Model().Name(), mf.RelationModel.Underlying().Name(), our, their, fc.Model().IsMixin())

	if mf.Filter != nil {
		fInfo.SetProperty("filter", mf.Filter.Underlying())
	}
	fInfo.SetProperty("relationModel", mf.RelationModel.Underlying())
	fInfo.SetProperty("m2mRelModel", m2mRelModel)
	fInfo.SetProperty("m2mOurField", m2mOurField)
	fInfo.SetProperty("m2mTheirField", m2mTheirField)
	return fInfo
}

// A Many2One is a field for storing many-to-one relations,
// i.e. the FK to another model.
//
// Clients are expected to handle many2one fields with a combo-box.
type Many2One struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	RelationModel   models.Modeler
	Embed           bool
	OnDelete        models.OnDeleteAction
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Filter          models.Conditioner
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a many2one field for the given models.FieldsCollection with the given name.
func (mf Many2One) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &mf, name, fieldtype.Many2One, new(int64))
	onDelete := models.SetNull
	if mf.OnDelete != "" {
		onDelete = mf.OnDelete
	}
	noCopy := mf.NoCopy
	required := mf.Required
	if mf.Embed {
		onDelete = models.Cascade
		noCopy = true
		required = false
	}
	if mf.Filter != nil {
		fInfo.SetProperty("filter", mf.Filter.Underlying())
	}
	fInfo.SetProperty("relationModel", mf.RelationModel.Underlying())
	fInfo.SetProperty("onDelete", onDelete)
	fInfo.SetProperty("noCopy", noCopy)
	fInfo.SetProperty("required", required)
	fInfo.SetProperty("embed", mf.Embed)
	return fInfo
}

// A One2Many is a field for storing one-to-many relations.
//
// Clients are expected to handle one2many fields with a table.
type One2Many struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	Copy            bool
	RelationModel   models.Modeler
	ReverseFK       string
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Filter          models.Conditioner
	Inverse         models.Methoder
	Default         func(models.Environment) interface{}
}

// DeclareField creates a one2many field for the given models.FieldsCollection with the given name.
func (of One2Many) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &of, name, fieldtype.One2Many, new([]int64))
	if of.Filter != nil {
		fInfo.SetProperty("filter", of.Filter.Underlying())
	}
	fInfo.SetProperty("relationModel", of.RelationModel.Underlying())
	fInfo.SetProperty("reverseFK", of.ReverseFK)
	if !of.Copy {
		fInfo.SetProperty("noCopy", true)
	}
	return fInfo
}

// A One2One is a field for storing one-to-one relations,
// i.e. the FK to another model with a unique constraint.
//
// Clients are expected to handle one2one fields with a combo-box.
type One2One struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	RelationModel   models.Modeler
	Embed           bool
	OnDelete        models.OnDeleteAction
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Filter          models.Conditioner
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a one2one field for the given models.FieldsCollection with the given name.
func (of One2One) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &of, name, fieldtype.One2One, new(int64))
	onDelete := models.SetNull
	if of.OnDelete != "" {
		onDelete = of.OnDelete
	}
	noCopy := of.NoCopy
	required := of.Required
	if of.Embed {
		onDelete = models.Cascade
		required = false
		noCopy = true
	}
	if of.Filter != nil {
		fInfo.SetProperty("filter", of.Filter.Underlying())
	}
	fInfo.SetProperty("relationModel", of.RelationModel.Underlying())
	fInfo.SetProperty("onDelete", onDelete)
	fInfo.SetProperty("noCopy", noCopy)
	fInfo.SetProperty("required", required)
	fInfo.SetProperty("embed", of.Embed)
	return fInfo
}

// A Rev2One is a field for storing reverse one-to-one relations,
// i.e. the relation on the model without FK.
//
// Clients are expected to handle rev2one fields with a combo-box.
type Rev2One struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	Copy            bool
	RelationModel   models.Modeler
	ReverseFK       string
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Filter          models.Conditioner
	Inverse         models.Methoder
	Default         func(models.Environment) interface{}
}

// DeclareField creates a rev2one field for the given models.FieldsCollection with the given name.
func (rf Rev2One) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &rf, name, fieldtype.Rev2One, new(int64))
	if rf.Filter != nil {
		fInfo.SetProperty("filter", rf.Filter.Underlying())
	}
	fInfo.SetProperty("relationModel", rf.RelationModel.Underlying())
	fInfo.SetProperty("reverseFK", rf.ReverseFK)
	if !rf.Copy {
		fInfo.SetProperty("noCopy", true)
	}
	return fInfo
}

// A Selection is a field for storing a value from a preset list.
//
// Clients are expected to handle selection fields with a combo-box or radio buttons.
type Selection struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	Selection       types.Selection
	SelectionFunc   func() types.Selection
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a selection field for the given models.FieldsCollection with the given name.
func (sf Selection) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &sf, name, fieldtype.Selection, new(string))
	fInfo.SetProperty("selection", sf.Selection)
	fInfo.SetProperty("selectionFunc", sf.SelectionFunc)
	return fInfo
}

// A Text is a field for storing long text. There is no
// default max size, but it can be forced by setting the Size value.
//
// Clients are expected to handle text fields as multi-line inputs.
type Text struct {
	JSON            string
	String          string
	Help            string
	Stored          bool
	Required        bool
	ReadOnly        bool
	RequiredFunc    func(models.Environment) (bool, models.Conditioner)
	ReadOnlyFunc    func(models.Environment) (bool, models.Conditioner)
	InvisibleFunc   func(models.Environment) (bool, models.Conditioner)
	Unique          bool
	Index           bool
	Compute         models.Methoder
	Depends         []string
	Related         string
	NoCopy          bool
	Size            int
	GoType          interface{}
	Translate       bool
	OnChange        models.Methoder
	OnChangeWarning models.Methoder
	OnChangeFilters models.Methoder
	Constraint      models.Methoder
	Inverse         models.Methoder
	Contexts        models.FieldContexts
	Default         func(models.Environment) interface{}
}

// DeclareField creates a text field for the given models.FieldsCollection with the given name.
func (tf Text) DeclareField(fc *models.FieldsCollection, name string) *models.Field {
	fInfo := models.CreateFieldFromStruct(fc, &tf, name, fieldtype.Text, new(string))
	fInfo.SetProperty("size", tf.Size)
	return fInfo
}
