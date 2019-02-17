package dates

import (
	"encoding/json"
	"testing"
	"time"
)
import . "github.com/smartystreets/goconvey/convey"

func checkDate(date Date) {
	So(date.Year(), ShouldEqual, 2017)
	So(date.Month(), ShouldEqual, 8)
	So(date.Day(), ShouldEqual, 1)
}

func TestDate(t *testing.T) {
	Convey("Testing Date objects", t, func() {
		date, err := ParseDateWithLayout(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		dateTime, _ := ParseDateTimeWithLayout(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		Convey("Parsing should be correct", func() {
			So(err, ShouldBeNil)
			checkDate(date)
		})
		Convey("Direct parsing functions should work", func() {
			So(func() { ParseDate("2017-08-01") }, ShouldNotPanic)
			So(func() { ParseDate("2017-08-01 11:23:32") }, ShouldPanic)
		})
		Convey("Marshaling and String should work", func() {
			So(date.String(), ShouldEqual, "2017-08-01")
			data, _ := json.Marshal(date)
			So(string(data), ShouldEqual, "\"2017-08-01\"")
		})
		Convey("Marshaling zero", func() {
			data, _ := json.Marshal(Date{})
			So(string(data), ShouldEqual, "false")
		})
		Convey("Scanning date strings", func() {
			dateScan := &Date{}
			err := dateScan.Scan("2017-08-01 10:02:57")
			So(err, ShouldBeNil)
			checkDate(*dateScan)
			So(dateScan.Equal(date), ShouldBeTrue)
			dateScan.Scan("")
			So(dateScan.IsZero(), ShouldBeTrue)
			err = dateScan.Scan("2017-08-01")
			So(err, ShouldBeNil)
			checkDate(*dateScan)
		})
		Convey("Scanning date time.Time", func() {
			dateScan := &Date{}
			dateScan.Scan(date.Time)
			checkDate(*dateScan)
			dateScan.Scan(time.Time{})
			So(dateScan.IsZero(), ShouldBeTrue)
		})
		Convey("Scanning date wrong type", func() {
			dateScan := &Date{}
			err := dateScan.Scan([]string{"foo", "bar"})
			So(err, ShouldNotBeNil)
		})
		Convey("Checking ToDate", func() {
			So(date.ToDateTime().Equal(dateTime), ShouldBeTrue)
		})
		Convey("Valuing Date", func() {
			val, err := date.Value()
			So(err, ShouldBeNil)
			ti, ok := val.(time.Time)
			So(ok, ShouldBeTrue)
			So(ti.Equal(date.Time), ShouldBeTrue)

		})
		Convey("Valuing empty Date", func() {
			val, err := Date{}.Value()
			So(err, ShouldBeNil)
			ti, ok := val.(time.Time)
			So(ok, ShouldBeTrue)
			So(ti.IsZero(), ShouldBeTrue)

		})
		Convey("Today() should not panic", func() {
			So(func() { Today() }, ShouldNotPanic)
		})
	})
	Convey("Checking operations and comparisons on Date and DateTime", t, func() {
		date1 := ParseDate("2017-08-01")
		date2 := ParseDate("2017-08-03")
		Convey("Comparing dates", func() {
			So(date2.Greater(date1), ShouldBeTrue)
			So(date2.GreaterEqual(date1), ShouldBeTrue)
			So(date2.GreaterEqual(date2), ShouldBeTrue)
			So(date2.Lower(date1), ShouldBeFalse)
			So(date2.LowerEqual(date1), ShouldBeFalse)
			So(date2.LowerEqual(date2), ShouldBeTrue)
		})
		Convey("Adding durations to dates", func() {
			So(date1.AddDate(0, 2, 3).Equal(ParseDate("2017-10-04")), ShouldBeTrue)
			So(date1.AddWeeks(2).Equal(ParseDate("2017-08-15")), ShouldBeTrue)
		})
		Convey("Changing dates", func() {
			dateCpy := date1.Copy()
			So(dateCpy.SetMonth(10).SetDay(4).Equal(ParseDate("2017-10-04")), ShouldBeTrue)
			So(dateCpy.SetYear(1996).SetMonth(time.February).SetDay(30).Equal(ParseDate("1996-03-01")), ShouldBeTrue)
			So(dateCpy.AddWeeks(2).StartOfMonth().Equal(ParseDate("2017-08-01")), ShouldBeTrue)
			So(dateCpy.StartOfYear().Equal(ParseDate("2017-01-01")), ShouldBeTrue)
			So(dateCpy.SetUnix(123456789).Equal(ParseDate("1973-11-29")), ShouldBeTrue)
		})
	})
}
