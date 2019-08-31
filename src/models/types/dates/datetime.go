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
	// DefaultServerDateTimeFormat is the Go layout for DateTime objects
	DefaultServerDateTimeFormat = "2006-01-02 15:04:05"
)

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
		return time.Time{}, nil
	}
	return d.Time, nil
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

// Now returns the current date/time with UTC timezone
func Now() DateTime {
	return DateTime{time.Now().UTC()}
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

// LoadLocation returns the Location with the given name.
//
// If the name is "" or "UTC", LoadLocation returns UTC.
// If the name is "Local", LoadLocation returns Local.
//
// Otherwise, the name is taken to be a location name corresponding to a file
// in the IANA Time Zone database, such as "America/New_York".
func LoadLocation(name string) (*time.Location, error) {
	return time.LoadLocation(name)
}

var _ driver.Valuer = DateTime{}
var _ sql.Scanner = new(DateTime)

// UTC returns d with the location set to UTC.
func (d DateTime) UTC() DateTime {
	return DateTime{
		Time: d.Time.UTC(),
	}
}

// WithTimezone returns d with a location corresponding to the given timezone identifier (IANA Time Zone database)
// returns an error if the name is not found
func (d DateTime) WithTimezone(tz string) (DateTime, error) {
	loc, err := LoadLocation(tz)
	if err != nil {
		return d, err
	}
	return d.In(loc), nil
}

// In returns d with the location information set to loc.
// In panics if loc is nil.
func (d DateTime) In(loc *time.Location) DateTime {
	return DateTime{Time: d.Time.In(loc)}
}

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
	return d.Time.Sub(t.Time)
}

// AddDate adds the given years, months or days to the current DateTime
func (d DateTime) AddDate(years, months, days int) DateTime {
	return DateTime{
		Time: d.Time.AddDate(years, months, days),
	}
}

// Copy returns a copy of d
func (d DateTime) Copy() DateTime {
	return DateTime{
		Time: d.Time,
	}
}

// SetYear changes the year value of d
// returns d for chained calls
func (d DateTime) SetYear(year int) DateTime {
	d.Time = time.Date(year, d.Month(), d.Day(), d.Hour(), d.Minute(), d.Second(), 0, d.Location())
	return d
}

// SetMonth changes the month value of d
// returns d for chained calls
func (d DateTime) SetMonth(month time.Month) DateTime {
	d.Time = time.Date(d.Year(), month, d.Day(), d.Hour(), d.Minute(), d.Second(), 0, d.Location())
	return d
}

// SetDay changes the day value of d
// returns d for chained calls
func (d DateTime) SetDay(day int) DateTime {
	d.Time = time.Date(d.Year(), d.Month(), day, d.Hour(), d.Minute(), d.Second(), 0, d.Location())
	return d
}

// SetHour changes the hour value of d
// returns d for chained calls
func (d DateTime) SetHour(hour int) DateTime {
	d.Time = time.Date(d.Year(), d.Month(), d.Day(), hour, d.Minute(), d.Second(), 0, d.Location())
	return d
}

// SetMinute changes the minute value of d
// returns d for chained calls
func (d DateTime) SetMinute(min int) DateTime {
	d.Time = time.Date(d.Year(), d.Month(), d.Day(), d.Hour(), min, d.Second(), 0, d.Location())
	return d
}

// SetSecond changes the second value of d
// returns d for chained calls
func (d DateTime) SetSecond(sec int) DateTime {
	d.Time = time.Date(d.Year(), d.Month(), d.Day(), d.Hour(), d.Minute(), sec, 0, d.Location())
	return d
}

// AddWeeks adds the given amount of weeks to d
func (d DateTime) AddWeeks(amount int) DateTime {
	return DateTime{
		Time: d.Time.AddDate(0, 0, 7*amount),
	}
}

// StartOfYear returns the DateTime corresponding to the first day of d's year at 00:00
func (d DateTime) StartOfYear() DateTime {
	return DateTime{
		Time: time.Date(d.Year(), 1, 1, 0, 0, 0, 0, d.Location()),
	}
}

// StartOfMonth returns the DateTime corresponding to the first day of d's current month at 00:00
func (d DateTime) StartOfMonth() DateTime {
	return DateTime{
		Time: time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location()),
	}
}

// StartOfDay returns the DateTime corresponding to the beginning of the day, at 00:00
func (d DateTime) StartOfDay() DateTime {
	return DateTime{
		Time: time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location()),
	}
}

// StartOfHour returns the DateTime corresponding to the beginning of the current hour
func (d DateTime) StartOfHour() DateTime {
	return DateTime{
		Time: time.Date(d.Year(), d.Month(), d.Day(), d.Hour(), 0, 0, 0, d.Location()),
	}
}

// SetUnix returns the DateTime corresponding to the given unix timestamp
func (d DateTime) SetUnix(sec int64) DateTime {
	return DateTime{
		Time: time.Unix(sec, 0).In(d.Location()),
	}
}
