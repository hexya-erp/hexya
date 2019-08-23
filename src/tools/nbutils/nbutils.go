// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package nbutils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/cockroachdb/apd/v2"
)

// CastToInteger casts the given val to int64 if it is
// a number type. Returns an error otherwise
func CastToInteger(val interface{}) (int64, error) {
	switch value := val.(type) {
	case int64:
		return value, nil
	case int, int8, int16, int32, uint, uint8, uint16, uint32, uint64, float32, float64:
		res, _ := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
		return res, nil
	case bool:
		if value {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("value %v cannot be casted to int64", val)
	}
}

// CastToFloat casts the given val to float64 if it is
// a number type. Panics otherwise
func CastToFloat(val interface{}) (float64, error) {
	switch value := val.(type) {
	case float64:
		return value, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
		res, _ := strconv.ParseFloat(fmt.Sprintf("%d", value), 64)
		return res, nil
	case bool:
		if value {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("value %v cannot be casted to float64", val)
	}
}

// Digits holds precision and scale information for a float (numeric) type:
//   - The precision: the total number of digits
//   - The scale: the number of digits to the right of the decimal point
//     (PostgresSQL definitions)
type Digits struct {
	Precision int8
	Scale     int8
}

// ToPrecision returns the given digits as a precision float:
//
// Digits{Precision: 6, Scale: 2} => 0.01
func (d Digits) ToPrecision() float64 {
	return math.Pow10(int(-d.Scale))
}

var ctx = apd.Context{
	MaxExponent: apd.MaxExponent,
	MinExponent: apd.MinExponent,
	Traps:       apd.DefaultTraps,
	Rounding:    apd.RoundHalfUp,
	Precision:   128,
}

// applyDecimalOperation applies fnct to value with the given precision and returns the result as a float
func applyDecimalOperation(value, precision float64, fnct func(d, x *apd.Decimal) (apd.Condition, error)) float64 {
	val, err := apd.New(0, 0).SetFloat64(value)
	if err != nil {
		panic(fmt.Errorf("error while rounding %f: %s", value, err))
	}
	prec, err := apd.New(0, 0).SetFloat64(precision)
	if err != nil {
		panic(fmt.Errorf("error while rounding precision %f: %s", precision, err))
	}
	normalized := apd.New(0, 0)
	_, err = ctx.Quo(normalized, val, prec)
	if err != nil {
		panic(fmt.Errorf("error while rounding %f: %s", value, err))
	}
	_, err = fnct(normalized, normalized)
	if err != nil {
		panic(fmt.Errorf("error while rounding %f: %s", value, err))
	}
	ctx.Mul(normalized, normalized, prec)
	res, err := normalized.Float64()
	if err != nil {
		panic(fmt.Errorf("error while rounding %f: %s", value, err))
	}
	return res

}

// Round rounds the given val to the given precision, which is a float such as :
//
// - 0.01 to round at the nearest 100th
// - 10 to round at the nearest ten
func Round(value float64, precision float64) float64 {
	return applyDecimalOperation(value, precision, ctx.RoundToIntegralExact)
}

// Ceil rounds up the given val to the given precision, which is a float such as :
//
// - 0.01 to round at the nearest 100th
// - 10 to round at the nearest ten
func Ceil(value float64, precision float64) float64 {
	return applyDecimalOperation(value, precision, ctx.Ceil)
}

// Floor rounds down the given val to the given precision, which is a float such as :
//
// - 0.01 to round at the nearest 100th
// - 10 to round at the nearest ten
func Floor(value, precision float64) float64 {
	return applyDecimalOperation(value, precision, ctx.Floor)
}

// Compare 'value1' and 'value2' after rounding them according to the
// given precision, which is a float such as :
//
// - 0.01 to round at the nearest 100th
// - 10 to round at the nearest ten
//
// The returned values are per the following table:
//
//    value1 > value2 : 1
//    value1 == value2: 0
//    value1 < value2 : -1
//
// A value is considered lower/greater than another value
// if their rounded value is different. This is not the same as having a
// non-zero difference!
//
// Example: 1.432 and 1.431 are equal at 2 digits precision,
// so this method would return 0
// However 0.006 and 0.002 are considered different (this method returns 1)
// because they respectively round to 0.01 and 0.0, even though
// 0.006-0.002 = 0.004 which would be considered zero at 2 digits precision.
//
// Warning: IsZero(value1-value2) is not equivalent to
// Compare(value1,value2) == _, true, as the former will round after
// computing the difference, while the latter will round before, giving
// different results for e.g. 0.006 and 0.002 at 2 digits precision.
func Compare(value1, value2 float64, precision float64) int8 {
	if Round(value1, precision) == Round(value2, precision) {
		return 0
	}
	if Round(value1, precision) > Round(value2, precision) {
		return 1
	}
	return -1
}

// IsZero returns true if 'value' is small enough to be treated as
// zero at the given precision , which is a float such as :
//
// - 0.01 to round at the nearest 100th
// - 10 to round at the nearest ten
//
// Warning: IsZero(value1-value2) is not equivalent to
// Compare(value1,value2) == _, true, as the former will round after
// computing the difference, while the latter will round before, giving
// different results for e.g. 0.006 and 0.002 at 2 digits precision.
func IsZero(value float64, precision float64) bool {
	if math.Abs(Round(value, precision)) < precision {
		return true
	}
	return false
}
