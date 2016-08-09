/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
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

// Create is a shortcut function for rs.Call("Create") on the current RecordSet.
// Data can be either a struct pointer or a FieldMap.
func (rs RecordSet) Create(data interface{}) *RecordSet {
	return rs.Call("Create", data).(*RecordSet)
}

// Write is a shortcut for rs.Call("Write") on the current RecordSet.
// Data can be either a struct pointer or a FieldMap.
func (rs RecordSet) Write(data interface{}) bool {
	return rs.Call("Write", data).(bool)
}

// Unlink is a shortcut for rs.Call("Unlink") on the current RecordSet.
func (rs RecordSet) Unlink() int64 {
	return rs.Call("Unlink").(int64)
}
