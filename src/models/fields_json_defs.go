
package models

import (

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/types"
)

// A JSONField is a field for storing jsonb.
type JSONField struct {
	JSON          string
	String        string
	Help          string
	Stored        bool
	Required      bool
	ReadOnly      bool
	RequiredFunc  func(Environment) (bool, Conditioner)
	ReadOnlyFunc  func(Environment) (bool, Conditioner)
	InvisibleFunc func(Environment) (bool, Conditioner)
	Unique        bool
	Index         bool
	Compute       Methoder
	Depends       []string
	Related       string
	NoCopy        bool
	GoType        interface{}
	Translate     bool
	OnChange      Methoder
	Constraint    Methoder
	Inverse       Methoder
	Contexts      FieldContexts
	Default       func(Environment) interface{}
}

// DeclareField creates a html field for the given FieldsCollection with the given name.
func (jsonf JSONField) DeclareField(fc *FieldsCollection, name string) *Field {
	fInfo := genericDeclareField(fc, &jsonf, name, fieldtype.JSON, new(types.JSONText))
	return fInfo
}


