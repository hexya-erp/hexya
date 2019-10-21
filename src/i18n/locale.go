// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
)

// A Currency with symbol, position and decimals
type Currency interface {
	// Symbol returns the currency symbol when printing amounts
	Symbol() string
	// Position returns 'before' or 'after' depending on where the symbol must be printed
	Position() string
	// DecimalPlaces for this currency
	DecimalPlaces() int
	// Round returns the given value rounded according to this currency
	Round(float64) float64
}

// A LangDirection defines the direction of a language
// either left-to-right or right-to-left
type LangDirection string

const (
	// LangDirectionLTR defines a language written from left to right
	LangDirectionLTR LangDirection = "ltr"
	// LangDirectionRTL defines a language written from right to left
	LangDirectionRTL LangDirection = "rtl"
)

// NumberGrouping represents grouping values of a number as follows:
//  - it splits a number into groups of N, N being a value in the slice
//  - the values define groups from right to left
//  - all values should be positive
//  - 0 at the end means repetition of previous int
//  - if the last value is not a 0, the grouping will end
//    e.g. :
//       3       -> 123456,789
//       3,0     -> 123,456,789
//       3,2     -> 1234,56,789
//       3,2,0   -> 12,34,56,789
type NumberGrouping []int

// MarshalJSON function for the NumberGrouping type that should marshal as string.
func (nb NumberGrouping) MarshalJSON() ([]byte, error) {
	res := bytes.NewBufferString(`"[`)
	for i, n := range nb {
		res.WriteString(fmt.Sprintf("%d", n))
		if i != len(nb)-1 {
			res.WriteByte(',')
		}
	}
	res.WriteString(`]"`)
	return res.Bytes(), nil
}

// Locale defines the parameters of a language locale
type Locale struct {
	Name         string         `json:"name"`
	Code         string         `json:"code"`
	ISOCode      string         `json:"iso_code"`
	WeekStart    time.Weekday   `json:"week_start"`
	DateFormat   string         `json:"date_format"`
	Direction    LangDirection  `json:"lang_direction"`
	ThousandsSep string         `json:"thousands_sep"`
	TimeFormat   string         `json:"time_format"`
	DecimalPoint string         `json:"decimal_point"`
	Grouping     NumberGrouping `json:"grouping"`
}

// Check returns an error if this locale is not valid
func (l *Locale) Check() error {
	if l.ISOCode == "" {
		return errors.New("locale should have an iso code")
	}
	if l.Name == "" {
		return errors.New("locale should have a name")
	}
	if l.Direction == "" {
		return errors.New("locale should have a direction")
	}
	return nil
}

// FormatFloat formats the given number according to this Locale, with the given digits
func (l *Locale) FormatFloat(number float64, digits nbutils.Digits) string {
	number = nbutils.Round(number, digits.ToPrecision())
	format := fmt.Sprintf("%%.%df", digits.Scale)
	numStr := fmt.Sprintf(format, number)
	parts := strings.Split(numStr, ".")
	intPart := parts[0]
	var decPart string
	if len(parts) > 1 {
		decPart = parts[1]
	}
	// Add "thousands" separators
	var (
		lastGrouping int
		keepGrouping bool
	)
	groups := []string{intPart}
	// Iterate on each group
	for _, n := range l.Grouping {
		if n == 0 {
			keepGrouping = true
			break
		}
		var ok bool
		groups, ok = groupDigits(groups, n)
		if !ok {
			break
		}
		lastGrouping = n
	}
	// Continue grouping if applicable
	if keepGrouping {
		ok := true
		for ok {
			groups, ok = groupDigits(groups, lastGrouping)
		}
	}
	res := strings.Join(groups, l.ThousandsSep)
	// Add decimal part if any
	if decPart != "" {
		res += l.DecimalPoint + decPart
	}
	return res
}

// FormatMonetary formats the given value according to this Locale and given currency
func (l *Locale) FormatMonetary(value float64, curr Currency) string {
	digs := nbutils.Digits{Precision: 16, Scale: int8(curr.DecimalPlaces())}
	amount := l.FormatFloat(value, digs)
	if curr.Position() == "before" {
		return fmt.Sprintf("%s %s", curr.Symbol(), amount)
	}
	return fmt.Sprintf("%s %s", amount, curr.Symbol())
}

// FormatDate returns the given date formatted according to this Locale
func (l *Locale) FormatDate(date dates.Date) string {
	return date.Format(l.DateFormat)
}

// FormatTime returns the time part of the given datetime formatted
// according to this Locale
func (l *Locale) FormatTime(datetime dates.DateTime) string {
	return datetime.Format(l.TimeFormat)
}

// FormatDateTime returns the given datetime formatted
// according to this Locale
func (l *Locale) FormatDateTime(datetime dates.DateTime) string {
	return fmt.Sprintf("%s %s", datetime.Format(l.DateFormat), datetime.Format(l.TimeFormat))
}

// groupDigits splits groups[0] at its last N digits and returns a new slice with, in order:
// - the remainder of the split
// - the n grouped digits
// - the rest of groups
//
// The second returned value is true if the split occurred, false if there are no more
// splits to do.
func groupDigits(groups []string, n int) ([]string, bool) {
	str := groups[0]
	if len(str) <= n {
		return groups, false
	}
	group := str[len(str)-n:]
	str = str[:len(str)-n]
	res := []string{str, group}
	if len(groups) > 1 {
		res = append(res, groups[1:]...)
	}
	return res, true
}

// GetLocale returns a Locale struct describing a language's rules
// at first call, the data file containing all languages parameters is read
// if the language is not loaded, it returns a Locale similar to English (en_US)
func GetLocale(lang string) *Locale {
	out, ok := locales[lang]
	if !ok {
		base := strings.Split(lang, "_")[0]
		out, ok = locales[base]
		if !ok {
			return &Locale{
				Name:         fmt.Sprintf("UNKNOWN_LOCALE (%s)", lang),
				Code:         "C",
				ISOCode:      "C",
				Direction:    LangDirectionLTR,
				DateFormat:   `%m/%d/%Y`,
				TimeFormat:   `%H:%M:%S`,
				ThousandsSep: `,`,
				DecimalPoint: `.`,
				Grouping:     NumberGrouping{3, 0},
			}
		}
	}
	return out
}

// RegisterLocale registers a new locale
func RegisterLocale(loc *Locale) error {
	if err := loc.Check(); err != nil {
		return err
	}
	if _, exists := locales[loc.ISOCode]; exists {
		return fmt.Errorf("locale with ISO code %s already exists", loc.ISOCode)
	}
	locales[loc.ISOCode] = loc
	updateAllLanguageList()
	return nil
}

// OverrideLocale overrides the locale with the same ISO code as loc with loc.
// If such a locale does not exist, an error is returned and the locale is not
// registered.
func OverrideLocale(loc *Locale) error {
	if err := loc.Check(); err != nil {
		return err
	}
	if _, exists := locales[loc.ISOCode]; !exists {
		return fmt.Errorf("locale with ISO code %s does not exist", loc.ISOCode)
	}
	locales[loc.ISOCode] = loc
	updateAllLanguageList()
	return nil
}
