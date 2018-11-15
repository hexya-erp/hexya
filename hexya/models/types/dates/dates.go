// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package dates

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const (
	// DefaultServerDateFormat is the Go layout for Date objects
	DefaultServerDateFormat = "2006-01-02"
	// DefaultServerDateTimeFormat is the Go layout for DateTime objects
	DefaultServerDateTimeFormat = "2006-01-02 15:04:05"
)

// Date type that JSON marshal and unmarshals as "YYYY-MM-DD"
type Date struct {
	time.Time
}

// String method for Date.
func (d Date) String() string {
	bs, _ := d.MarshalJSON()
	return strings.Trim(string(bs), "\"")
}

// MarshalJSON for Date type
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("false"), nil
	}
	dateStr := d.Time.Format(DefaultServerDateFormat)
	dateStr = fmt.Sprintf(`"%s"`, dateStr)
	return []byte(dateStr), nil
}

// Value formats our Date for storing in database
// Especially handles empty Date.
func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return driver.Value(time.Time{}), nil
	}
	return driver.Value(d.Time), nil
}

// Scan casts the database output to a Date
func (d *Date) Scan(src interface{}) error {
	switch t := src.(type) {
	case time.Time:
		d.Time = t
		return nil
	case string:
		if t == "" {
			*d = Date{}
			return nil
		}
		val, err := ParseDateWithLayout(DefaultServerDateFormat, t)
		if err != nil {
			val, err = ParseDateWithLayout(DefaultServerDateTimeFormat, t)
		}
		*d = val
		return err
	}
	return fmt.Errorf("date data is not time.Time but %T", src)
}

var _ driver.Valuer = Date{}
var _ sql.Scanner = new(Date)

// Equal reports whether d and other represent the same day
func (d Date) Equal(other Date) bool {
	return d.String() == other.String()
}

// Greater returns true if d is strictly greater than other
func (d Date) Greater(other Date) bool {
	return d.Sub(other) > 0
}

// GreaterEqual returns true if d is greater than or equal to other
func (d Date) GreaterEqual(other Date) bool {
	return d.Sub(other) >= 0
}

// Lower returns true if d is strictly lower than other
func (d Date) Lower(other Date) bool {
	return d.Sub(other) < 0
}

// LowerEqual returns true if d is lower than or equal to other
func (d Date) LowerEqual(other Date) bool {
	return d.Sub(other) <= 0
}

// AddDate adds the given years, months or days to the current date
func (d Date) AddDate(years, months, days int) Date {
	return Date{
		Time: d.Time.AddDate(years, months, days),
	}
}

// Sub returns the duration t-u. If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration, the maximum (or minimum) duration
// will be returned.
// To compute t-d for a duration d, use t.Add(-d).
func (d Date) Sub(t Date) time.Duration {
	return  d.Time.Sub(t.Time)
}

// Today returns the current date
func Today() Date {
	return Date{time.Now()}
}

// ParseDate returns a date from the given string value
// that is formatted with the default YYYY-MM-DD format.
//
// It panics in case the parsing cannot be done.
func ParseDate(value string) Date {
	d, err := ParseDateWithLayout(DefaultServerDateFormat, value)
	if err != nil {
		panic(err)
	}
	return d
}

// ParseDateWithLayout returns a date from the given string value
// that is formatted with layout.
func ParseDateWithLayout(layout, value string) (Date, error) {
	t, err := time.Parse(layout, value)
	return Date{Time: t}, err
}

// DateTime type that JSON marshals and unmarshals as "YYYY-MM-DD HH:MM:SS"
type DateTime struct {
	time.Time
}

// String method for DateTime.
func (d DateTime) String() string {
	bs, _ := d.MarshalJSON()
	return strings.Trim(string(bs), "\"")
}

// ToDate returns the Date of this DateTime
func (d DateTime) ToDate() Date {
	return Date{d.Time}
}

// MarshalJSON for DateTime type
func (d DateTime) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("false"), nil
	}
	dateStr := d.Time.Format(DefaultServerDateTimeFormat)
	dateStr = fmt.Sprintf(`"%s"`, dateStr)
	return []byte(dateStr), nil
}

// Value formats our DateTime for storing in database
// Especially handles empty DateTime.
func (d DateTime) Value() (driver.Value, error) {
	if d.IsZero() {
		return driver.Value(time.Time{}), nil
	}
	return driver.Value(d.Time), nil
}

// Scan casts the database output to a DateTime
func (d *DateTime) Scan(src interface{}) error {
	switch t := src.(type) {
	case time.Time:
		d.Time = t
		return nil
	case string:
		if t == "" {
			*d = DateTime{}
			return nil
		}
		val, err := ParseDateTimeWithLayout(DefaultServerDateTimeFormat, t)
		*d = val
		return err
	}
	return fmt.Errorf("DateTime data is not time.Time but %T", src)
}

// Now returns the current date/time
func Now() DateTime {
	return DateTime{time.Now()}
}

// ParseDateTime returns a datetime from the given string value
// that is formatted with the default YYYY-MM-DD HH:MM:SSformat.
//
// It panics in case the parsing cannot be done.
func ParseDateTime(value string) DateTime {
	dt, err := ParseDateTimeWithLayout(DefaultServerDateTimeFormat, value)
	if err != nil {
		panic(err)
	}
	return dt
}

// ParseDateTimeWithLayout returns a datetime from the given string value
// that is formatted with layout.
func ParseDateTimeWithLayout(layout, value string) (DateTime, error) {
	t, err := time.Parse(layout, value)
	return DateTime{Time: t}, err
}

var _ driver.Valuer = DateTime{}
var _ sql.Scanner = new(DateTime)

// Equal reports whether d and other represent the same time instant
func (d DateTime) Equal(other DateTime) bool {
	return d.Time.Equal(other.Time)
}

// Greater returns true if d is strictly greater than other
func (d DateTime) Greater(other DateTime) bool {
	return d.Sub(other) > 0
}

// GreaterEqual returns true if d is greater than or equal to other
func (d DateTime) GreaterEqual(other DateTime) bool {
	return d.Sub(other) >= 0
}

// Lower returns true if d is strictly lower than other
func (d DateTime) Lower(other DateTime) bool {
	return d.Sub(other) < 0
}

// LowerEqual returns true if d is lower than or equal to other
func (d DateTime) LowerEqual(other DateTime) bool {
	return d.Sub(other) <= 0
}

// Add adds the given duration to this DateTime
func (d DateTime) Add(duration time.Duration) DateTime {
	return DateTime{
		Time: d.Time.Add(duration),
	}
}

// Sub returns the duration t-u. If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration, the maximum (or minimum) duration
// will be returned.
// To compute t-d for a duration d, use t.Add(-d).
func (d DateTime) Sub(t DateTime) time.Duration {
	return  d.Time.Sub(t.Time)
}

// AddDate adds the given years, months or days to the current DateTime
func (d DateTime) AddDate(years, months, days int) DateTime {
	return DateTime{
		Time: d.Time.AddDate(years, months, days),
	}
}
