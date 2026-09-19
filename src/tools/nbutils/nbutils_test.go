// Copyright 2019 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package nbutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCastToFloat(t *testing.T) {
	t.Run("Testing cast to float", func(t *testing.T) {
		val, err := CastToFloat(12)
		assert.EqualValues(t, val, 12)
		assert.IsType(t, float64(1), val)
		assert.Nil(t, err)
		val, err = CastToFloat(12.85)
		assert.EqualValues(t, val, 12.85)
		assert.IsType(t, float64(1), val)
		assert.Nil(t, err)
		val, err = CastToFloat(int32(12))
		assert.EqualValues(t, val, 12)
		assert.IsType(t, float64(1), val)
		assert.Nil(t, err)
		val, err = CastToFloat(true)
		assert.EqualValues(t, val, 1)
		assert.IsType(t, float64(1), val)
		assert.Nil(t, err)
		val, err = CastToFloat(false)
		assert.EqualValues(t, val, 0)
		assert.IsType(t, float64(1), val)
		assert.Nil(t, err)
		val, err = CastToFloat("12")
		assert.EqualValues(t, val, 0)
		assert.IsType(t, float64(1), val)
		assert.NotNil(t, err)
	})
}

func TestCastToInteger(t *testing.T) {
	t.Run("Testing cast to integer", func(t *testing.T) {
		val, err := CastToInteger(12)
		assert.EqualValues(t, val, 12)
		assert.IsType(t, int64(1), val)
		assert.Nil(t, err)
		val, err = CastToInteger(12.52)
		assert.EqualValues(t, val, 12)
		assert.IsType(t, int64(1), val)
		assert.Nil(t, err)
		val, err = CastToInteger(int64(12))
		assert.EqualValues(t, val, 12)
		assert.IsType(t, int64(1), val)
		assert.Nil(t, err)
		val, err = CastToInteger(true)
		assert.EqualValues(t, val, 1)
		assert.IsType(t, int64(1), val)
		assert.Nil(t, err)
		val, err = CastToInteger(false)
		assert.EqualValues(t, val, 0)
		assert.IsType(t, int64(1), val)
		assert.Nil(t, err)
		val, err = CastToInteger("12")
		assert.EqualValues(t, val, 0)
		assert.IsType(t, int64(1), val)
		assert.NotNil(t, err)
	})
}

func TestRound(t *testing.T) {
	t.Run("Testing round", func(t *testing.T) {
		assert.EqualValues(t, Round(12.23, 0.1), 12.2)
		assert.EqualValues(t, Round(12.25, 0.1), 12.3)
		assert.EqualValues(t, Round(12.2499, 0.1), 12.2)
		assert.EqualValues(t, Round(-61.160000000000004, 0.01), -61.16)
	})
}

func TestIsZero(t *testing.T) {
	t.Run("Testing is zero", func(t *testing.T) {
		assert.True(t, IsZero(0, 1))
		assert.True(t, IsZero(0.1, 1))
		assert.True(t, IsZero(0.01, 0.1))
		assert.False(t, IsZero(0.1, 0.1))
		assert.False(t, IsZero(0.01, 0.01))
	})
}

func TestDigits(t *testing.T) {
	t.Run("Testing digits to precision", func(t *testing.T) {
		assert.EqualValues(t, Digits{Precision: 12, Scale: 4}.ToPrecision(), 0.0001)
		assert.EqualValues(t, Digits{Precision: 12, Scale: 1}.ToPrecision(), 0.1)
		assert.EqualValues(t, Digits{Precision: 12, Scale: 0}.ToPrecision(), 1)
	})
}

func TestFloor(t *testing.T) {
	t.Run("Testing floor", func(t *testing.T) {
		assert.EqualValues(t, Floor(12.23, 0.1), 12.2)
		assert.EqualValues(t, Floor(12.25, 0.1), 12.2)
		assert.EqualValues(t, Floor(12.2499, 0.1), 12.2)
		assert.EqualValues(t, Floor(-61.160000000000004, 0.01), -61.17)
	})
}

func TestCeil(t *testing.T) {
	t.Run("Testing ceil", func(t *testing.T) {
		assert.EqualValues(t, Ceil(12.23, 0.1), 12.3)
		assert.EqualValues(t, Ceil(12.25, 0.1), 12.3)
		assert.EqualValues(t, Ceil(12.2499, 0.1), 12.3)
		assert.EqualValues(t, Ceil(-61.160000000000004, 0.01), -61.16)
	})
}

func TestCompare(t *testing.T) {
	t.Run("Testing compare", func(t *testing.T) {
		assert.EqualValues(t, Compare(13, 13, 1), 0)
		assert.EqualValues(t, Compare(13, 13.1, 1), 0)
		assert.EqualValues(t, Compare(13, 13.01, 0.1), 0)
		assert.EqualValues(t, Compare(13, 13.1, 0.1), -1)
		assert.EqualValues(t, Compare(13, 13.01, 0.01), -1)
		assert.EqualValues(t, Compare(13.01, 13, 0.01), 1)
	})
}
