// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package typesutils

import (
	"database/sql"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
	Convey("Testing IsZero function", t, func() {
		Convey("nil", func() {
			So(IsZero(nil), ShouldBeTrue)
		})
		Convey("Strings", func() {
			So(IsZero(""), ShouldBeTrue)
			So(IsZero("Hi"), ShouldBeFalse)
		})
		Convey("Floats", func() {
			So(IsZero(float64(0.0)), ShouldBeTrue)
			So(IsZero(float64(12.4)), ShouldBeFalse)
		})
		Convey("Structs", func() {
			type demoStruct struct {
				field1 string
				field2 int8
				field3 float32
			}
			So(IsZero(demoStruct{}), ShouldBeTrue)
			So(IsZero(demoStruct{field1: "Hello"}), ShouldBeFalse)
		})
		Convey("Pointers", func() {
			var nilPointer *string
			So(IsZero(nilPointer), ShouldBeTrue)
			notNilString := "Hey !"
			So(IsZero(&notNilString), ShouldBeFalse)
		})
		Convey("RecordSets", func() {
			So(IsZero(new(dummyRecordSet)), ShouldBeTrue)
		})
	})
}

func TestAreEqual(t *testing.T) {
	Convey("Testing ArEqual function", t, func() {
		Convey("Different types should return an error", func() {
			res, err := AreEqual(true, 1)
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparison.Error())
		})
		Convey("Unsupported type", func() {
			res, err := AreEqual([]int{1, 2}, []int{1, 2})
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparisonType.Error())
			res, err = AreEqual(12, []int{1, 2})
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparisonType.Error())
		})
		Convey("Bool", func() {
			res, err := AreEqual(true, false)
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = AreEqual(true, true)
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
		Convey("Complex", func() {
			res, err := AreEqual(complex(2, 3), complex(3, 4))
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = AreEqual(complex(2, 3), complex(2, 3))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
		Convey("Int and UInt", func() {
			res, err := AreEqual(int(1), int(3))
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = AreEqual(int(1), int(1))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
			res, err = AreEqual(uint(1), uint(1))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
			res, err = AreEqual(int8(1), uint16(1))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
			res, err = AreEqual(uint8(1), int32(1))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
		Convey("Float", func() {
			res, err := AreEqual(float64(1), float64(3))
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = AreEqual(float64(1), float64(1))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
			res, err = AreEqual(float32(1), float64(1))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
		Convey("String", func() {
			res, err := AreEqual("Hello", "World")
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = AreEqual("Hello", "Hello")
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestIsLessThan(t *testing.T) {
	Convey("Testing IsLessThan function", t, func() {
		Convey("Different types should return an error", func() {
			res, err := IsLessThan(true, 1)
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparison.Error())
		})
		Convey("Unsupported type", func() {
			res, err := IsLessThan([]int{1, 2}, []int{1, 2})
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparisonType.Error())
			res, err = IsLessThan(12, []int{1, 2})
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparisonType.Error())
		})
		Convey("Bool", func() {
			res, err := IsLessThan(true, false)
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparisonType.Error())
			res, err = IsLessThan(true, true)
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparisonType.Error())
		})
		Convey("Complex", func() {
			res, err := IsLessThan(complex(2, 3), complex(3, 4))
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparisonType.Error())
			res, err = IsLessThan(complex(2, 3), complex(2, 3))
			So(res, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errBadComparisonType.Error())
		})
		Convey("Int and UInt", func() {
			res, err := IsLessThan(int(1), int(3))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
			res, err = IsLessThan(int(1), int(1))
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = IsLessThan(uint(3), uint(1))
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = IsLessThan(int8(1), uint16(2))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
			res, err = IsLessThan(uint8(1), int32(4))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
		Convey("Float", func() {
			res, err := IsLessThan(float64(1), float64(3))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
			res, err = IsLessThan(float64(1), float64(1))
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = IsLessThan(float32(1), float64(2))
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
		})
		Convey("String", func() {
			res, err := IsLessThan("Hello", "World")
			So(res, ShouldBeTrue)
			So(err, ShouldBeNil)
			res, err = IsLessThan("Hello", "Hello")
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
			res, err = IsLessThan("World", "Hello")
			So(res, ShouldBeFalse)
			So(err, ShouldBeNil)
		})
	})
}

type convertTestCase struct {
	value  interface{}
	target interface{}
	isRS   bool
	result interface{}
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
	{value: []interface{}{}, target: new(int64), result: int64(0), isRS: true},
	{value: []interface{}{}, target: new([]int64), result: []int64{}, isRS: true},
	{value: []int64{1, 2}, target: new([]int64), result: []int64{1, 2}, isRS: true},
	{value: []int64{1}, target: new(int64), result: int64(1), isRS: true},
	{value: (*interface{})(nil), target: new(int64), result: int64(0), isRS: true},
	{value: (*interface{})(nil), target: new([]int64), result: []int64{}, isRS: true},
}

var convertErrorCases = []convertTestCase{
	{value: "SOMESTRING", target: new(sql.NullFloat64), err: "unable to scan into target Type: converting driver.Value type string (\"SOMESTRING\") to a float64: invalid syntax"},
	{value: []byte("STRING"), target: new(float32), err: "strconv.ParseFloat: parsing \"STRING\": invalid syntax"},
	{value: []byte("STRING"), target: new(float64), err: "strconv.ParseFloat: parsing \"STRING\": invalid syntax"},
	{value: []interface{}{1}, target: new([]int64), isRS: true, err: "non empty []interface{} given"},
	{value: "ST", target: new(float32), isRS: true, err: "expected number value, got ST: value ST cannot be casted to int64"},
	{value: 1, target: new(float64), isRS: true, err: "non consistent type"},
	{value: false, target: new(int), err: "impossible conversion of false (bool) to int"},
}

func TestConvert(t *testing.T) {
	Convey("Testing Convert", t, func() {
		for _, tc := range convertTestCases {
			var target = tc.target
			err := Convert(tc.value, target, tc.isRS)
			So(err, ShouldBeNil)
			targetVal := reflect.ValueOf(target).Elem()
			So(targetVal.Type(), ShouldEqual, reflect.TypeOf(tc.result))
			So(reflect.DeepEqual(targetVal.Interface(), tc.result), ShouldBeTrue)
		}
	})
	Convey("Testing conversion errors", t, func() {
		for _, tc := range convertErrorCases {
			var target = tc.target
			err := Convert(tc.value, target, tc.isRS)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, tc.err)
		}
	})
}
