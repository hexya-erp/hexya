// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package dates

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func checkDate(date Date) {
	So(date.Year(), ShouldEqual, 2017)
	So(date.Month(), ShouldEqual, 8)
	So(date.Day(), ShouldEqual, 1)
}

func checkDateTime(dateTime DateTime) {
	So(dateTime.Year(), ShouldEqual, 2017)
	So(dateTime.Month(), ShouldEqual, 8)
	So(dateTime.Day(), ShouldEqual, 1)
	So(dateTime.Hour(), ShouldEqual, 10)
	So(dateTime.Minute(), ShouldEqual, 2)
	So(dateTime.Second(), ShouldEqual, 57)
}

func TestDates(t *testing.T) {
	Convey("Testing Date objects", t, func() {
		date, err1 := ParseDate(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		dateTime, err2 := ParseDateTime(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		Convey("Parsing should be correct", func() {
			So(err1, ShouldBeNil)
			checkDate(date)
			So(err2, ShouldBeNil)
			checkDateTime(dateTime)
		})
		Convey("Marshaling and String should work", func() {
			So(date.String(), ShouldEqual, "2017-08-01")
			data, _ := json.Marshal(date)
			So(string(data), ShouldEqual, "\"2017-08-01\"")
			So(dateTime.String(), ShouldEqual, "2017-08-01 10:02:57")
			data, _ = json.Marshal(dateTime)
			So(string(data), ShouldEqual, "\"2017-08-01 10:02:57\"")
		})
		Convey("Marshaling zero", func() {
			data, _ := json.Marshal(Date{})
			So(string(data), ShouldEqual, "false")
			data, _ = json.Marshal(DateTime{})
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
		Convey("Valuing Date", func() {
			val, err := date.Value()
			So(err, ShouldBeNil)
			t, ok := val.(time.Time)
			So(ok, ShouldBeTrue)
			So(t.Equal(date.Time), ShouldBeTrue)

		})
		Convey("Valuing empty Date", func() {
			val, err := Date{}.Value()
			So(err, ShouldBeNil)
			t, ok := val.(time.Time)
			So(ok, ShouldBeTrue)
			So(t.IsZero(), ShouldBeTrue)

		})
		Convey("Valuing Datetime", func() {
			val, err := dateTime.Value()
			So(err, ShouldBeNil)
			t, ok := val.(time.Time)
			So(ok, ShouldBeTrue)
			So(t.Equal(date.Time), ShouldBeTrue)
		})
		Convey("Valuing empty Datetime", func() {
			val, err := DateTime{}.Value()
			So(err, ShouldBeNil)
			t, ok := val.(time.Time)
			So(ok, ShouldBeTrue)
			So(t.IsZero(), ShouldBeTrue)
		})
	})
}
