// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package typesutils

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyRecordSet struct{}

func (d *dummyRecordSet) ModelName() string { return "" }
func (d *dummyRecordSet) Ids() []int64      { return []int64{} }
func (d *dummyRecordSet) Len() int          { return 0 }
func (d *dummyRecordSet) IsEmpty() bool {
	return true
}
func (d *dummyRecordSet) IsNotEmpty() bool {
	return false
}

var _ RecordSet = new(dummyRecordSet)

func TestIsZero(t *testing.T) {
	t.Run("Testing IsZero function", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			assert.True(t, IsZero(nil))
		})
		t.Run("Strings", func(t *testing.T) {
			assert.True(t, IsZero(""))
			assert.False(t, IsZero("Hi"))
		})
		t.Run("Floats", func(t *testing.T) {
			assert.True(t, IsZero(float64(0.0)))
			assert.False(t, IsZero(float64(12.4)))
		})
		t.Run("Structs", func(t *testing.T) {
			type demoStruct struct {
				field1 string
				field2 int8
				field3 float32
			}
			assert.True(t, IsZero(demoStruct{}))
			assert.False(t, IsZero(demoStruct{field1: "Hello"}))
		})
		t.Run("Pointers", func(t *testing.T) {
			var nilPointer *string
			assert.True(t, IsZero(nilPointer))
			notNilString := "Hey !"
			assert.False(t, IsZero(&notNilString))
		})
		t.Run("RecordSets", func(t *testing.T) {
			assert.True(t, IsZero(new(dummyRecordSet)))
		})
	})
}

func TestAreEqual(t *testing.T) {
	t.Run("Testing ArEqual function", func(t *testing.T) {
		t.Run("Different types should return an error", func(t *testing.T) {
			res, err := AreEqual(true, 1)
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparison.Error())
		})
		t.Run("Unsupported type", func(t *testing.T) {
			res, err := AreEqual([]int{1, 2}, []int{1, 2})
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparisonType.Error())
			res, err = AreEqual(12, []int{1, 2})
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparisonType.Error())
		})
		t.Run("Bool", func(t *testing.T) {
			res, err := AreEqual(true, false)
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = AreEqual(true, true)
			assert.True(t, res)
			assert.Nil(t, err)
		})
		t.Run("Complex", func(t *testing.T) {
			res, err := AreEqual(complex(2, 3), complex(3, 4))
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = AreEqual(complex(2, 3), complex(2, 3))
			assert.True(t, res)
			assert.Nil(t, err)
		})
		t.Run("Int and UInt", func(t *testing.T) {
			res, err := AreEqual(int(1), int(3))
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = AreEqual(int(1), int(1))
			assert.True(t, res)
			assert.Nil(t, err)
			res, err = AreEqual(uint(1), uint(1))
			assert.True(t, res)
			assert.Nil(t, err)
			res, err = AreEqual(int8(1), uint16(1))
			assert.True(t, res)
			assert.Nil(t, err)
			res, err = AreEqual(uint8(1), int32(1))
			assert.True(t, res)
			assert.Nil(t, err)
		})
		t.Run("Float", func(t *testing.T) {
			res, err := AreEqual(float64(1), float64(3))
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = AreEqual(float64(1), float64(1))
			assert.True(t, res)
			assert.Nil(t, err)
			res, err = AreEqual(float32(1), float64(1))
			assert.True(t, res)
			assert.Nil(t, err)
		})
		t.Run("String", func(t *testing.T) {
			res, err := AreEqual("Hello", "World")
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = AreEqual("Hello", "Hello")
			assert.True(t, res)
			assert.Nil(t, err)
		})
	})
}

func TestIsLessThan(t *testing.T) {
	t.Run("Testing IsLessThan function", func(t *testing.T) {
		t.Run("Different types should return an error", func(t *testing.T) {
			res, err := IsLessThan(true, 1)
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparison.Error())
		})
		t.Run("Unsupported type", func(t *testing.T) {
			res, err := IsLessThan([]int{1, 2}, []int{1, 2})
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparisonType.Error())
			res, err = IsLessThan(12, []int{1, 2})
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparisonType.Error())
		})
		t.Run("Bool", func(t *testing.T) {
			res, err := IsLessThan(true, false)
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparisonType.Error())
			res, err = IsLessThan(true, true)
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparisonType.Error())
		})
		t.Run("Complex", func(t *testing.T) {
			res, err := IsLessThan(complex(2, 3), complex(3, 4))
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparisonType.Error())
			res, err = IsLessThan(complex(2, 3), complex(2, 3))
			assert.False(t, res)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), errBadComparisonType.Error())
		})
		t.Run("Int and UInt", func(t *testing.T) {
			res, err := IsLessThan(int(1), int(3))
			assert.True(t, res)
			assert.Nil(t, err)
			res, err = IsLessThan(int(1), int(1))
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = IsLessThan(uint(3), uint(1))
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = IsLessThan(int8(1), uint16(2))
			assert.True(t, res)
			assert.Nil(t, err)
			res, err = IsLessThan(uint8(1), int32(4))
			assert.True(t, res)
			assert.Nil(t, err)
		})
		t.Run("Float", func(t *testing.T) {
			res, err := IsLessThan(float64(1), float64(3))
			assert.True(t, res)
			assert.Nil(t, err)
			res, err = IsLessThan(float64(1), float64(1))
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = IsLessThan(float32(1), float64(2))
			assert.True(t, res)
			assert.Nil(t, err)
		})
		t.Run("String", func(t *testing.T) {
			res, err := IsLessThan("Hello", "World")
			assert.True(t, res)
			assert.Nil(t, err)
			res, err = IsLessThan("Hello", "Hello")
			assert.False(t, res)
			assert.Nil(t, err)
			res, err = IsLessThan("World", "Hello")
			assert.False(t, res)
			assert.Nil(t, err)
		})
	})
}

type convertTestCase struct {
	value  any
	target any
	isRS   bool
	result any
	err    string
}

var convertTestCases = []convertTestCase{
	{value: 1, target: new(int), result: 1},
	{value: nil, target: new(int), result: 0},
	{value: true, target: new(bool), result: true},
	{value: false, target: new(bool), result: false},
	{value: 0, target: new(bool), result: false},
	{value: 1, target: new(bool), result: true},
	{value: 1, target: new(float32), result: float32(1)},
	{value: []byte("1"), target: new(float32), result: float32(1)},
	{value: []byte("1"), target: new(float64), result: float64(1)},
	{value: "1", target: new(sql.NullFloat64), result: sql.NullFloat64{Float64: 1, Valid: true}},
	{value: 1, target: new(int64), result: int64(1), isRS: true},
	{value: []any{}, target: new(int64), result: int64(0), isRS: true},
	{value: []any{}, target: new([]int64), result: []int64{}, isRS: true},
	{value: []int64{1, 2}, target: new([]int64), result: []int64{1, 2}, isRS: true},
	{value: []int64{1}, target: new(int64), result: int64(1), isRS: true},
	{value: (*any)(nil), target: new(int64), result: int64(0), isRS: true},
	{value: (*any)(nil), target: new([]int64), result: []int64{}, isRS: true},
}

var convertErrorCases = []convertTestCase{
	{value: "SOMESTRING", target: new(sql.NullFloat64), err: "unable to scan into target Type: converting driver.Value type string (\"SOMESTRING\") to a float64: invalid syntax"},
	{value: []byte("STRING"), target: new(float32), err: "strconv.ParseFloat: parsing \"STRING\": invalid syntax"},
	{value: []byte("STRING"), target: new(float64), err: "strconv.ParseFloat: parsing \"STRING\": invalid syntax"},
	{value: []any{1}, target: new([]int64), isRS: true, err: "non empty []interface{} given"},
	{value: "ST", target: new(float32), isRS: true, err: "expected number value, got ST: value ST cannot be casted to int64"},
	{value: 1, target: new(float64), isRS: true, err: "non consistent type"},
	{value: false, target: new(int), err: "impossible conversion of false (bool) to int"},
}

func TestConvert(t *testing.T) {
	t.Run("Testing Convert", func(t *testing.T) {
		for _, tc := range convertTestCases {
			var target = tc.target
			err := Convert(tc.value, target, tc.isRS)
			assert.Nil(t, err)
			targetVal := reflect.ValueOf(target).Elem()
			assert.EqualValues(t, targetVal.Type(), reflect.TypeOf(tc.result))
			assert.True(t, reflect.DeepEqual(targetVal.Interface(), tc.result))
		}
	})
	t.Run("Testing conversion errors", func(t *testing.T) {
		for _, tc := range convertErrorCases {
			var target = tc.target
			err := Convert(tc.value, target, tc.isRS)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), tc.err)
		}
	})
}
