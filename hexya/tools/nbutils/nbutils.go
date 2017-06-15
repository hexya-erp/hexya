// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package nbutils

import (
	"fmt"
	"strconv"
)

// CastToInteger casts the given val to int64 if it is
// a number type. Returns an error otherwise
func CastToInteger(val interface{}) (int64, error) {
	var res int64
	switch value := val.(type) {
	case int64:
		res = value
	case int, int8, int16, int32, uint, uint8, uint16, uint32, uint64, float32, float64:
		res, _ = strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
	default:
		return 0, fmt.Errorf("Value %v cannot be casted to int64", val)
	}
	return res, nil
}

// CastToFloat casts the given val to float64 if it is
// a number type. Panics otherwise
func CastToFloat(val interface{}) (float64, error) {
	var res float64
	switch value := val.(type) {
	case float64:
		res = value
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
		res, _ = strconv.ParseFloat(fmt.Sprintf("%d", value), 64)
	default:
		return 0, fmt.Errorf("Value %v cannot be casted to float64", val)
	}
	return res, nil
}
