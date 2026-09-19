// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package dates

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func checkDate(t *testing.T, date Date) {
	assert.EqualValues(t, date.Year(), 2017)
	assert.EqualValues(t, date.Month(), 8)
	assert.EqualValues(t, date.Day(), 1)
}

func TestDate(t *testing.T) {
	t.Run("Testing Date objects", func(t *testing.T) {
		date, err := ParseDateWithLayout(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		dateTime, _ := ParseDateTimeWithLayout(DefaultServerDateTimeFormat, "2017-08-01 10:02:57")
		t.Run("Parsing should be correct", func(t *testing.T) {
			assert.Nil(t, err)
			checkDate(t, date)
		})
		t.Run("Direct parsing functions should work", func(t *testing.T) {
			assert.NotPanics(t, func() { ParseDate("2017-08-01") })
			assert.Panics(t, func() { ParseDate("2017-08-01 11:23:32") })
		})
		t.Run("Marshaling and String should work", func(t *testing.T) {
			assert.EqualValues(t, date.String(), "2017-08-01")
			data, _ := json.Marshal(date)
			assert.EqualValues(t, string(data), "\"2017-08-01\"")
		})
		t.Run("Marshaling zero", func(t *testing.T) {
			data, _ := json.Marshal(Date{})
			assert.EqualValues(t, string(data), "false")
		})
		t.Run("Scanning date strings", func(t *testing.T) {
			dateScan := &Date{}
			err := dateScan.Scan("2017-08-01 10:02:57")
			assert.Nil(t, err)
			checkDate(t, *dateScan)
			assert.True(t, dateScan.Equal(date))
			dateScan.Scan("")
			assert.True(t, dateScan.IsZero())
			err = dateScan.Scan("2017-08-01")
			assert.Nil(t, err)
			checkDate(t, *dateScan)
		})
		t.Run("Scanning date time.Time", func(t *testing.T) {
			dateScan := &Date{}
			dateScan.Scan(date.Time)
			checkDate(t, *dateScan)
			dateScan.Scan(time.Time{})
			assert.True(t, dateScan.IsZero())
		})
		t.Run("Scanning date wrong type", func(t *testing.T) {
			dateScan := &Date{}
			err := dateScan.Scan([]string{"foo", "bar"})
			assert.NotNil(t, err)
		})
		t.Run("Checking ToDate", func(t *testing.T) {
			assert.True(t, date.ToDateTime().Equal(dateTime))
		})
		t.Run("Valuing Date", func(t *testing.T) {
			val, err := date.Value()
			assert.Nil(t, err)
			ti, ok := val.(time.Time)
			assert.True(t, ok)
			assert.True(t, ti.Equal(date.Time))

		})
		t.Run("Valuing empty Date", func(t *testing.T) {
			val, err := Date{}.Value()
			assert.Nil(t, err)
			ti, ok := val.(time.Time)
			assert.True(t, ok)
			assert.True(t, ti.IsZero())

		})
		t.Run("Today() should not panic", func(t *testing.T) {
			assert.NotPanics(t, func() { Today() })
		})
	})
	t.Run("Checking operations and comparisons on Date and DateTime", func(t *testing.T) {
		date1 := ParseDate("2017-08-01")
		date2 := ParseDate("2017-08-03")
		t.Run("Comparing dates", func(t *testing.T) {
			assert.True(t, date2.Greater(date1))
			assert.True(t, date2.GreaterEqual(date1))
			assert.True(t, date2.GreaterEqual(date2))
			assert.False(t, date2.Lower(date1))
			assert.False(t, date2.LowerEqual(date1))
			assert.True(t, date2.LowerEqual(date2))
		})
		t.Run("Adding durations to dates", func(t *testing.T) {
			assert.True(t, date1.AddDate(0, 2, 3).Equal(ParseDate("2017-10-04")))
			assert.True(t, date1.AddWeeks(2).Equal(ParseDate("2017-08-15")))
		})
		t.Run("Changing dates", func(t *testing.T) {
			dateCpy := date1.Copy()
			assert.True(t, dateCpy.SetMonth(10).SetDay(4).Equal(ParseDate("2017-10-04")))
			assert.True(t, dateCpy.SetYear(1996).SetMonth(time.February).SetDay(30).Equal(ParseDate("1996-03-01")))
			assert.True(t, dateCpy.AddWeeks(2).StartOfMonth().Equal(ParseDate("2017-08-01")))
			assert.True(t, dateCpy.StartOfYear().Equal(ParseDate("2017-01-01")))
		})
	})
}
