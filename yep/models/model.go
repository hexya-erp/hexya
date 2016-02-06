/*   Copyright (C) 2016 by Nicolas Piganeau
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package models

import (
	"reflect"
	"fmt"
)

/*
Model is the interface defining a YEP model. A Model defines a business object
and is in charge of data persistence and all calculations associated with this
model.

In YEP, models are defined incrementally by each loaded module which can add
fields and/or methods to the Model.
*/
type Model interface {
	// Name of the Model
	Name() string
	// Returns the Field given by its name
	Field(string) Field
	// Adds or update a Field in the Model
	AddField(string, Field) error
	// Adds a MethodFunc to a method, creating it if it does not exist
	AddMethod(string, MethodFunc) error
	// Searches the database for records matching the domain
	Search(Query) RecordSet
}

/*
BaseModel is the main implementation of Model as a database persistent model.
*/
type BaseModel struct {
	name      string
	fields    map[string]Field
	methods   map[string]MethodStack
}

func (m *BaseModel) Name() string {
	return m.name
}

func (m *BaseModel) Field(fieldName string) Field {
	field, ok := m.fields[fieldName]
	if !ok {
		return nil
	}
	return field
}

func (m *BaseModel) AddField(fieldName string, field Field) error {
	oldField, exists := m.fields[fieldName]
	if !exists{
		m.fields[fieldName] = field
		return nil
	} else {
		if reflect.TypeOf(oldField) != reflect.TypeOf(field){
			return fmt.Errorf("Unable to update fields of type %s with field of type %s",
				reflect.TypeOf(oldField).Name(), reflect.TypeOf(field).Name())
		}
		err := oldField.Update(field)
		return err
	}
}
