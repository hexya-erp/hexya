// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package typesutils

import (
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

func TestIsZero(t *testing.T) {
	Convey("Testing IsZero function", t, func() {
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
			So(IsZero(dummyRecordSet{}), ShouldBeTrue)
		})
	})
}
