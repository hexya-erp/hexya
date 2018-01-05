// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package typesutils

import "reflect"

// RecordSet is an approximation of models.RecordSet so as not to import models.
type RecordSet interface {
	// ModelName returns the name of the model of this RecordSet
	ModelName() string
	// Ids returns the ids in this set of Records
	Ids() []int64
	// Len returns the number of records in this RecordSet
	Len() int
	// IsEmpty returns true if this RecordSet has no records
	IsEmpty() bool
}

// IsZero returns true if the given value is the zero value of its type or nil
func IsZero(value interface{}) bool {
	if value == nil {
		return true
	}
	if rc, ok := value.(RecordSet); ok {
		return rc.IsEmpty()
	}
	val := reflect.ValueOf(value)
	return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
}
