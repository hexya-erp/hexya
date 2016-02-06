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

/*
RecordSet is a set of instances of a Model that can be manipulated.
 */
type RecordSet interface {
	// Returns the Model of this RecordSet
	Model() Model
	// Returns the Environment of this RecordSet
	Env() Environment
	// Returns true if this RecordSet contains exactly one record
	IsSingleton() bool
	// Returns the ids of the records in this RecordSet. Records that are not yet in database are given negative ids.
	Ids() []int
	// Scans the data of the Recordset into the given struct slice
	Scan([]interface{}) error
	// Scans the singleton in the given struct. Returns an error if this is not a Singleton
	ScanSingleton(interface{}) error
	// Call the method given by string on this recordset with the given arguments
	Call(string, ...interface{}) interface{}
}

/*
baseRecordSet is the main implementation of RecordSet
 */
type baseRecordSet struct {
	model Model
}

func (r *baseRecordSet) Model() Model {
	return r.model
}

func (r* baseRecordSet) Call(methodName string, args ...interface{}) interface{} {
	return nil
}