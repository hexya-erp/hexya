// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package typesutils

import (
	"reflect"

	"github.com/hexya-erp/hexya/hexya/models"
)

// IsZero returns true if the given value is the zero value of its type or nil
func IsZero(value interface{}) bool {
	if value == nil {
		return true
	}
	if rc, ok := value.(models.RecordSet); ok {
		return rc.IsEmpty()
	}
	val := reflect.ValueOf(value)
	return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
}
