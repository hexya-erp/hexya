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

import "fmt"

/*
Field is the interface that an attribute of a Model must satisfy.

In YEP, fields can be created incrementally by calling Update on an existing
Field.
*/
type Field interface {
	fmt.Stringer
	// Returns the help string of this Field
	Help() string
	// Returns the name of the method to compute this Field or an empty string if it is not computed
	ComputeMethod() string
	// Returns true if this field is readonly
	ReadOnly() bool
	// Update this Field with the non-zero parameters of the given field
	Update(Field) error
}

/*
NewField creates a new Field from the given params struct
 */
func NewField(params interface{}) Field {
	switch p := params.(type){
	case Char:
		return &fieldChar{
			baseField: baseField{
				description: p.String,
				help: p.Help,
				computeMethod: p.ComputeMethod,
				readOnly: p.ReadOnly,
			},
		}
	default:
		return nil
	}
}

/*
BaseField is the base implementation of all fields
*/
type baseField struct {
	description   string
	help          string
	computeMethod string
	readOnly      bool
}

func (f *baseField) String() string {
	return f.description
}

func (f *baseField) Help() string {
	return f.help
}

func (f *baseField) ComputeMethod() string {
	return f.computeMethod
}

func (f *baseField) ReadOnly() bool {
	return f.readOnly
}

func (c *baseField) Update(field Field) error {
	if field.Help() != "" {
		c.help = field.Help()
	}
	if field.ComputeMethod() != "" {
		c.computeMethod = field.ComputeMethod()
	}
	c.readOnly = field.ReadOnly()
	return nil
}

/*
FieldChar is a varchar string field
*/
type fieldChar struct {
	baseField
}

/*
Char is a parameters struct for NewField function when creating a Char Field
*/
type Char struct {
	String        string
	Help          string
	ComputeMethod string
	ReadOnly      bool
}
