package dates

import (
	"encoding/json"
	"testing"
	"time"
)
import . "github.com/smartystreets/goconvey/convey"

func checkDateTime(dateTime DateTime) {
	So(dateTime.Year(), ShouldEqual, 2017)
	So(dateTime.Month(), ShouldEqual, 8)
	So(dateTime.Day(), ShouldEqual, 1)
	So(dateTime.Hour(), ShouldEqual, 10)
	So(dateTime.Minute(), ShouldEqual, 2)
	So(dateTime.Second(), ShouldEqual, 57)
}

func TestDateTime(t *testing.T) {
	Convey("Testing DateTime objects", t, func() {
		dateTime, err := ParseDateTimeWithLayout(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		date, _ := ParseDateWithLayout(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		Convey("Parsing should be correct", func() {
			So(err, ShouldBeNil)
			checkDateTime(dateTime)
		})
		Convey("Direct parsing functions should work", func() {
			So(func() { ParseDateTime("2017-08-01 10:02:57") }, ShouldNotPanic)
			So(func() { ParseDateTime("2017-08-01") }, ShouldPanic)
		})
		Convey("Marshaling and String should work", func() {
			So(dateTime.String(), ShouldEqual, "2017-08-01 10:02:57")
			data, _ := json.Marshal(dateTime)
			So(string(data), ShouldEqual, "\"2017-08-01 10:02:57\"")
		})
		Convey("Marshaling zero", func() {
			data, _ := json.Marshal(DateTime{})
			So(string(data), ShouldEqual, "false")
		})
		Convey("Scanning datetime strings", func() {
			dtScan := &DateTime{}
			err := dtScan.Scan("2017-08-01 10:02:57")
			So(err, ShouldBeNil)
			checkDateTime(*dtScan)
			So(dtScan.Equal(dateTime), ShouldBeTrue)
			dtScan.Scan("")
			So(dtScan.IsZero(), ShouldBeTrue)
		})
		Convey("Scanning datetime time.Time", func() {
			dtScan := &DateTime{}
			dtScan.Scan(dateTime.Time)
			checkDateTime(*dtScan)
			dtScan.Scan(time.Time{})
			So(dtScan.IsZero(), ShouldBeTrue)
		})
		Convey("Scanning datetime wrong type", func() {
			dtScan := &DateTime{}
			err := dtScan.Scan([]string{"foo", "bar"})
			So(err, ShouldNotBeNil)
		})
		Convey("Checking ToDate", func() {
			So(dateTime.ToDate().Equal(date), ShouldBeTrue)
		})
		Convey("Valuing Datetime", func() {
			val, err := dateTime.Value()
			So(err, ShouldBeNil)
			ti, ok := val.(time.Time)
			So(ok, ShouldBeTrue)
			So(ti.Equal(date.Time), ShouldBeTrue)
		})
		Convey("Valuing empty Datetime", func() {
			val, err := DateTime{}.Value()
			So(err, ShouldBeNil)
			ti, ok := val.(time.Time)
			So(ok, ShouldBeTrue)
			So(ti.IsZero(), ShouldBeTrue)
		})
		Convey("Now() should not panic", func() {
			So(func() { Now() }, ShouldNotPanic)
		})
	})
	Convey("Checking operations and comparisons on Date and DateTime", t, func() {
		dateTime1 := ParseDateTime("2017-08-01 10:34:23")
		dateTime2 := ParseDateTime("2017-08-01 10:43:11")
		Convey("Comparing datetimes", func() {
			So(dateTime2.Greater(dateTime1), ShouldBeTrue)
			So(dateTime2.GreaterEqual(dateTime1), ShouldBeTrue)
			So(dateTime2.GreaterEqual(dateTime2), ShouldBeTrue)
			So(dateTime2.Lower(dateTime1), ShouldBeFalse)
			So(dateTime2.LowerEqual(dateTime1), ShouldBeFalse)
			So(dateTime2.LowerEqual(dateTime2), ShouldBeTrue)
		})
		Convey("Adding durations to datetimes", func() {
			So(dateTime1.AddDate(0, 2, 3).Equal(ParseDateTime("2017-10-04 10:34:23")), ShouldBeTrue)
			So(dateTime1.Add(2*time.Hour+11*time.Minute).Equal(ParseDateTime("2017-08-01 12:45:23")), ShouldBeTrue)
			So(dateTime1.AddWeeks(2).Equal(ParseDateTime("2017-08-15 10:34:23")), ShouldBeTrue)
		})
		Convey("Timezone tests", func() {
			dt1, _ := dateTime1.WithTimezone("Etc/GMT")
			So(dt1.Equal(dateTime1.UTC()), ShouldBeTrue)
			dt2, _ := dateTime1.WithTimezone("Africa/Tripoli")
			So(dt2.String() == ParseDateTime("2017-08-01 12:34:23").String(), ShouldBeTrue)
			dt3, _ := dateTime1.WithTimezone("America/Argentina/Buenos_Aires")
			So(dt3.String() == ParseDateTime("2017-08-01 7:34:23").String(), ShouldBeTrue)
			date, err := dateTime1.WithTimezone("invalid/tzCode")
			So(date == dateTime1, ShouldBeTrue)
			So(err, ShouldNotBeNil)
			values := TimeZones()
			So(values, ShouldContain, "America/Scoresbysund")
		})
		Convey("Changing dates", func() {
			dateCpy := dateTime1.Copy()
			So(dateCpy.SetMonth(10).SetDay(4).Equal(ParseDateTime("2017-10-04 10:34:23")), ShouldBeTrue)
			So(dateCpy.SetYear(1996).SetMonth(time.February).SetDay(30).SetHour(-2).SetMinute(50).SetSecond(-7).
				Equal(DateTime{Time: time.Date(1996, 02, 29, 22, 49, 53, 0, time.UTC)}), ShouldBeTrue)
			So(dateCpy.StartOfHour().Equal(ParseDateTime("2017-08-01 10:00:00")), ShouldBeTrue)
			So(dateCpy.StartOfDay().Equal(ParseDateTime("2017-08-01 00:00:00")), ShouldBeTrue)
			So(dateCpy.AddWeeks(2).StartOfMonth().Equal(ParseDateTime("2017-08-01 00:00:00")), ShouldBeTrue)
			So(dateCpy.StartOfYear().Equal(ParseDateTime("2017-01-01 00:00:00")), ShouldBeTrue)
			So(dateCpy.SetUnix(123456789).Equal(ParseDateTime("1973-11-29 21:33:09")), ShouldBeTrue)
		})
	})
}
