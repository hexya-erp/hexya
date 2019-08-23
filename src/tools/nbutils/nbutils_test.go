// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package nbutils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCastToFloat(t *testing.T) {
	Convey("Testing cast to float", t, func() {
		val, err := CastToFloat(12)
		So(val, ShouldEqual, 12)
		So(val, ShouldHaveSameTypeAs, float64(1))
		So(err, ShouldBeNil)
		val, err = CastToFloat(float64(12))
		So(val, ShouldEqual, 12)
		So(val, ShouldHaveSameTypeAs, float64(1))
		So(err, ShouldBeNil)
		val, err = CastToFloat(int64(12))
		So(val, ShouldEqual, 12)
		So(val, ShouldHaveSameTypeAs, float64(1))
		So(err, ShouldBeNil)
		val, err = CastToFloat(true)
		So(val, ShouldEqual, 1)
		So(val, ShouldHaveSameTypeAs, float64(1))
		So(err, ShouldBeNil)
		val, err = CastToFloat(false)
		So(val, ShouldEqual, 0)
		So(val, ShouldHaveSameTypeAs, float64(1))
		So(err, ShouldBeNil)
		val, err = CastToFloat("12")
		So(val, ShouldEqual, 0)
		So(val, ShouldHaveSameTypeAs, float64(1))
		So(err, ShouldNotBeNil)
	})
}

func TestCastToInteger(t *testing.T) {
	Convey("Testing cast to integer", t, func() {
		val, err := CastToInteger(12)
		So(val, ShouldEqual, 12)
		So(val, ShouldHaveSameTypeAs, int64(1))
		So(err, ShouldBeNil)
		val, err = CastToInteger(float64(12))
		So(val, ShouldEqual, 12)
		So(val, ShouldHaveSameTypeAs, int64(1))
		So(err, ShouldBeNil)
		val, err = CastToInteger(int64(12))
		So(val, ShouldEqual, 12)
		So(val, ShouldHaveSameTypeAs, int64(1))
		So(err, ShouldBeNil)
		val, err = CastToInteger(true)
		So(val, ShouldEqual, 1)
		So(val, ShouldHaveSameTypeAs, int64(1))
		So(err, ShouldBeNil)
		val, err = CastToInteger(false)
		So(val, ShouldEqual, 0)
		So(val, ShouldHaveSameTypeAs, int64(1))
		So(err, ShouldBeNil)
		val, err = CastToInteger("12")
		So(val, ShouldEqual, 0)
		So(val, ShouldHaveSameTypeAs, int64(1))
		So(err, ShouldNotBeNil)
	})
}

func TestRound(t *testing.T) {
	Convey("Testing round", t, func() {
		So(Round(12.23, 0.1), ShouldEqual, 12.2)
		So(Round(12.25, 0.1), ShouldEqual, 12.3)
		So(Round(12.2499, 0.1), ShouldEqual, 12.2)
		So(Round(-61.160000000000004, 0.01), ShouldEqual, -61.16)
	})
}

func TestIsZero(t *testing.T) {
	Convey("Testing is zero", t, func() {
		So(IsZero(0, 1), ShouldBeTrue)
		So(IsZero(0.1, 1), ShouldBeTrue)
		So(IsZero(0.01, 0.1), ShouldBeTrue)
		So(IsZero(0.1, 0.1), ShouldBeFalse)
		So(IsZero(0.01, 0.01), ShouldBeFalse)
	})
}

func TestDigits(t *testing.T) {
	Convey("Testing digits to precision", t, func() {
		So(Digits{Precision: 12, Scale: 4}.ToPrecision(), ShouldEqual, 0.0001)
		So(Digits{Precision: 12, Scale: 1}.ToPrecision(), ShouldEqual, 0.1)
		So(Digits{Precision: 12, Scale: 0}.ToPrecision(), ShouldEqual, 1)
	})
}

func TestFloor(t *testing.T) {
	Convey("Testing floor", t, func() {
		So(Floor(12.23, 0.1), ShouldEqual, 12.2)
		So(Floor(12.25, 0.1), ShouldEqual, 12.2)
		So(Floor(12.2499, 0.1), ShouldEqual, 12.2)
		So(Floor(-61.160000000000004, 0.01), ShouldEqual, -61.17)
	})
}

func TestCeil(t *testing.T) {
	Convey("Testing ceil", t, func() {
		So(Ceil(12.23, 0.1), ShouldEqual, 12.3)
		So(Ceil(12.25, 0.1), ShouldEqual, 12.3)
		So(Ceil(12.2499, 0.1), ShouldEqual, 12.3)
		So(Ceil(-61.160000000000004, 0.01), ShouldEqual, -61.16)
	})
}

func TestCompare(t *testing.T) {
	Convey("Testing compare", t, func() {
		So(Compare(13, 13, 1), ShouldEqual, 0)
		So(Compare(13, 13.1, 1), ShouldEqual, 0)
		So(Compare(13, 13.01, 0.1), ShouldEqual, 0)
		So(Compare(13, 13.1, 0.1), ShouldEqual, -1)
		So(Compare(13, 13.01, 0.01), ShouldEqual, -1)
		So(Compare(13.01, 13, 0.01), ShouldEqual, 1)
	})
}
