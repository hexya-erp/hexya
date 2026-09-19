// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package dates

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func checkDateTime(t *testing.T, dateTime DateTime) {
	assert.EqualValues(t, dateTime.Year(), 2017)
	assert.EqualValues(t, dateTime.Month(), 8)
	assert.EqualValues(t, dateTime.Day(), 1)
	assert.EqualValues(t, dateTime.Hour(), 10)
	assert.EqualValues(t, dateTime.Minute(), 2)
	assert.EqualValues(t, dateTime.Second(), 57)
}

func TestDateTime(t *testing.T) {
	t.Run("Testing DateTime objects", func(t *testing.T) {
		dateTime, err := ParseDateTimeWithLayout(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		date, _ := ParseDateWithLayout(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		t.Run("Parsing should be correct", func(t *testing.T) {
			assert.Nil(t, err)
			checkDateTime(t, dateTime)
		})
		t.Run("Direct parsing functions should work", func(t *testing.T) {
			assert.NotPanics(t, func() { ParseDateTime("2017-08-01 10:02:57") })
			assert.Panics(t, func() { ParseDateTime("2017-08-01") })
		})
		t.Run("Marshaling and String should work", func(t *testing.T) {
			assert.EqualValues(t, dateTime.String(), "2017-08-01 10:02:57")
			data, _ := json.Marshal(dateTime)
			assert.EqualValues(t, string(data), "\"2017-08-01 10:02:57\"")
		})
		t.Run("Marshaling zero", func(t *testing.T) {
			data, _ := json.Marshal(DateTime{})
			assert.EqualValues(t, string(data), "false")
		})
		t.Run("Scanning datetime strings", func(t *testing.T) {
			dtScan := &DateTime{}
			err := dtScan.Scan("2017-08-01 10:02:57")
			assert.Nil(t, err)
			checkDateTime(t, *dtScan)
			assert.True(t, dtScan.Equal(dateTime))
			dtScan.Scan("")
			assert.True(t, dtScan.IsZero())
		})
		t.Run("Scanning datetime time.Time", func(t *testing.T) {
			dtScan := &DateTime{}
			dtScan.Scan(dateTime.Time)
			checkDateTime(t, *dtScan)
			dtScan.Scan(time.Time{})
			assert.True(t, dtScan.IsZero())
		})
		t.Run("Scanning datetime wrong type", func(t *testing.T) {
			dtScan := &DateTime{}
			err := dtScan.Scan([]string{"foo", "bar"})
			assert.NotNil(t, err)
		})
		t.Run("Checking ToDate", func(t *testing.T) {
			assert.True(t, dateTime.ToDate().Equal(date))
		})
		t.Run("Valuing Datetime", func(t *testing.T) {
			val, err := dateTime.Value()
			assert.Nil(t, err)
			ti, ok := val.(time.Time)
			assert.True(t, ok)
			assert.True(t, ti.Equal(date.Time))
		})
		t.Run("Valuing empty Datetime", func(t *testing.T) {
			val, err := DateTime{}.Value()
			assert.Nil(t, err)
			ti, ok := val.(time.Time)
			assert.True(t, ok)
			assert.True(t, ti.IsZero())
		})
		t.Run("Now() should not panic", func(t *testing.T) {
			assert.NotPanics(t, func() { Now() })
		})
	})
	t.Run("Checking operations and comparisons on Date and DateTime", func(t *testing.T) {
		dateTime1 := ParseDateTime("2017-08-01 10:34:23")
		dateTime2 := ParseDateTime("2017-08-01 10:43:11")
		t.Run("Comparing datetimes", func(t *testing.T) {
			assert.True(t, dateTime2.Greater(dateTime1))
			assert.True(t, dateTime2.GreaterEqual(dateTime1))
			assert.True(t, dateTime2.GreaterEqual(dateTime2))
			assert.False(t, dateTime2.Lower(dateTime1))
			assert.False(t, dateTime2.LowerEqual(dateTime1))
			assert.True(t, dateTime2.LowerEqual(dateTime2))
		})
		t.Run("Adding durations to datetimes", func(t *testing.T) {
			assert.True(t, dateTime1.AddDate(0, 2, 3).Equal(ParseDateTime("2017-10-04 10:34:23")))
			assert.True(t, dateTime1.Add(2*time.Hour+11*time.Minute).Equal(ParseDateTime("2017-08-01 12:45:23")))
			assert.True(t, dateTime1.AddWeeks(2).Equal(ParseDateTime("2017-08-15 10:34:23")))
		})
		t.Run("Timezone tests", func(t *testing.T) {
			dt1, _ := dateTime1.WithTimezone("Etc/GMT")
			assert.True(t, dt1.Equal(dateTime1.UTC()))
			dt2, _ := dateTime1.WithTimezone("Africa/Tripoli")
			assert.True(t, dt2.String() == ParseDateTime("2017-08-01 12:34:23").String())
			dt3, _ := dateTime1.WithTimezone("America/Argentina/Buenos_Aires")
			assert.True(t, dt3.String() == ParseDateTime("2017-08-01 7:34:23").String())
			date, err := dateTime1.WithTimezone("invalid/tzCode")
			assert.True(t, date == dateTime1)
			assert.NotNil(t, err)
			values := TimeZones()
			assert.Contains(t, values, "America/Scoresbysund")
		})
		t.Run("Changing dates", func(t *testing.T) {
			dateCpy := dateTime1.Copy()
			assert.True(t, dateCpy.SetMonth(10).SetDay(4).Equal(ParseDateTime("2017-10-04 10:34:23")))
			assert.True(t, dateCpy.SetYear(1996).SetMonth(time.February).SetDay(30).SetHour(-2).SetMinute(50).SetSecond(-7).
				Equal(DateTime{Time: time.Date(1996, 02, 29, 22, 49, 53, 0, time.UTC)}))
			assert.True(t, dateCpy.StartOfHour().Equal(ParseDateTime("2017-08-01 10:00:00")))
			assert.True(t, dateCpy.StartOfDay().Equal(ParseDateTime("2017-08-01 00:00:00")))
			assert.True(t, dateCpy.AddWeeks(2).StartOfMonth().Equal(ParseDateTime("2017-08-01 00:00:00")))
			assert.True(t, dateCpy.StartOfYear().Equal(ParseDateTime("2017-01-01 00:00:00")))
		})
	})
}
