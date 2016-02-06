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
	"testing"

	"github.com/npiganeau/yep/tools/tests"
)

func TestModelCreation(t *testing.T) {
	testModel := BaseModel{
		name: "test.model",
		fields: map[string]Field{
			"name": NewField(Char{Help: "This is the name of the test model"}),
			"upper-name": NewField(Char{
				Help:          "Upper case version of name",
				ComputeMethod: "upper",
			}),
		},
	}
	tests.AssertEqual(t, testModel.Name(), "test.model", "BaseModel/Name")
	tests.AssertEqual(t, testModel.Field("name").Help(), "This is the name of the test model", "BaseModel/Field/Help")
	tests.AssertEqual(t, testModel.Field("upper-name").ComputeMethod(), "upper", "BaseModel/Field/ComputeMethod")

	updField := NewField(Char{ComputeMethod: "upperUpd"})
	testModel.AddField("upper-name", updField)
	tests.AssertEqual(t, testModel.Field("upper-name").ComputeMethod(), "upperUpd",
		"BaseModel/Field/ComputeMethod field not updated")
	tests.AssertEqual(t, testModel.Field("upper-name").Help(), "Upper case version of name",
		"BaseModel/Field/Help field should not have been updated")

}
