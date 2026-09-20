// Copyright 2019 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
)

type currency struct {
	symbol   string
	position string
	decimals int
}

func (c currency) Symbol() string {
	return c.symbol
}

func (c currency) Position() string {
	return c.position
}

func (c currency) DecimalPlaces() int {
	return c.decimals
}

func (c currency) Round(val float64) float64 {
	return nbutils.Round(val, nbutils.Digits{Scale: int8(c.decimals)}.ToPrecision())
}

var _ Currency = currency{}

func TestLocale(t *testing.T) {
	t.Run("Testing locale's methods", func(t *testing.T) {
		t.Run("GetLocale by complete or simplified code", func(t *testing.T) {
			frFR := GetLocale("fr_FR")
			assert.EqualValues(t, frFR.Code, "fr_FR")
			fr := GetLocale("fr")
			assert.EqualValues(t, fr.Code, "fr_FR")
		})
		t.Run("FormatFloat", func(t *testing.T) {
			fr := GetLocale("fr")
			val := 1234567890.123456
			assert.EqualValues(t, fr.FormatFloat(val, nbutils.Digits{Precision: 12, Scale: 3}), "1 234 567 890,123")
			en := GetLocale("en")
			assert.EqualValues(t, en.FormatFloat(val, nbutils.Digits{Precision: 12, Scale: 3}), "1,234,567,890.123")
			en.Grouping = NumberGrouping{3, 2}
			assert.EqualValues(t, en.FormatFloat(val, nbutils.Digits{Precision: 12, Scale: 3}), "12345,67,890.123")
			en.Grouping = NumberGrouping{3, 2, 0}
			assert.EqualValues(t, en.FormatFloat(val, nbutils.Digits{Precision: 12, Scale: 3}), "1,23,45,67,890.123")
			en.Grouping = NumberGrouping{2, 3, 2}
			assert.EqualValues(t, en.FormatFloat(val, nbutils.Digits{Precision: 12, Scale: 5}), "123,45,678,90.12346")
			en.Grouping = NumberGrouping{2, 3, 2, 0}
			assert.EqualValues(t, en.FormatFloat(val, nbutils.Digits{Precision: 12, Scale: 3}), "1,23,45,678,90.123")
		})
		t.Run("FormatDate, FormatTime, FormatDateTime", func(t *testing.T) {
			en := GetLocale("en")
			fr := GetLocale("fr")
			date := dates.ParseDate("2003-07-12")
			dateTime := dates.ParseDateTime("2003-07-12 15:02:00")
			assert.EqualValues(t, fr.FormatDate(date), "12/07/2003")
			assert.EqualValues(t, en.FormatDate(date), "07/12/2003")
			en.TimeFormatGo = "03:04:05 PM"
			assert.EqualValues(t, fr.FormatDateTime(dateTime), "12/07/2003 15:02:00")
			assert.EqualValues(t, en.FormatDateTime(dateTime), "07/12/2003 03:02:00 PM")
			assert.EqualValues(t, fr.FormatTime(dateTime), "15:02:00")
			assert.EqualValues(t, en.FormatTime(dateTime), "03:02:00 PM")
		})
		t.Run("FormatMonetary", func(t *testing.T) {
			fr := GetLocale("fr")
			ja := GetLocale("ja")
			eur := currency{decimals: 2, symbol: "€", position: "after"}
			yen := currency{decimals: 0, symbol: "¥", position: "before"}
			assert.EqualValues(t, fr.FormatMonetary(1234567.789, eur), "1 234 567,79 €")
			assert.EqualValues(t, ja.FormatMonetary(1234567.789, yen), "¥ 1,234,568")
		})
	})
	t.Run("Testing number grouping JSON Marshalling", func(t *testing.T) {
		b, err := json.Marshal(NumberGrouping{1, 2, 3})
		assert.Nil(t, err)
		assert.EqualValues(t, string(b), `"[1,2,3]"`)
	})
}
