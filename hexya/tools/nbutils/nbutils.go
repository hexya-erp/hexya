// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package nbutils

import (
	"fmt"
	"math"
	"strconv"
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
// Digits{Scale: 6, Precision: 2} => 0.01
func (d Digits) ToPrecision() float64 {
	return float64(10 ^ (-d.Precision))
}

// Round rounds the given val to the given precision, which is a float such as :
//
// - 0.01 to round at the nearest 100th
// - 10 to round at the nearest ten
//
// This function uses the future Go 1.10 implementation
func Round(value float64, precision float64) float64 {
	val := value / precision
	const (
		mask  = 0x7FF
		shift = 64 - 11 - 1
		bias  = 1023

		signMask = 1 << 63
		fracMask = (1 << shift) - 1
		halfMask = 1 << (shift - 1)
		one      = bias << shift
	)

	bits := math.Float64bits(val)
	e := uint(bits>>shift) & mask
	switch {
	case e < bias:
		// Round abs(x)<1 including denormals.
		bits &= signMask // +-0
		if e == bias-1 {
			bits |= one // +-1
		}
	case e < bias+shift:
		// Round any abs(x)>=1 containing a fractional component [0,1).
		e -= bias
		bits += halfMask >> e
		bits &^= fracMask >> e
	}
	return math.Float64frombits(bits) * precision
}

// Round32 rounds the given val to the given precision.
// This function is just a wrapper for Round() casted to float32
func Round32(val float32, precision float64) float32 {
	return float32(Round(float64(val), precision))
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

// Compare32 'value1' and 'value2' after rounding them according to the
// given precision. This function is just a wrapper for Compare() with float32 values
func Compare32(value1, value2 float32, precision float64) int8 {
	return Compare(float64(value1), float64(value2), precision)
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
