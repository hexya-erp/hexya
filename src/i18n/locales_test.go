// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"encoding/json"
	"testing"

	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	. "github.com/smartystreets/goconvey/convey"
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
	Convey("Testing locale's methods", t, func() {
		Convey("GetLocale by complete or simplified code", func() {
			frFR := GetLocale("fr_FR")
			So(frFR.Code, ShouldEqual, "fr_FR")
			fr := GetLocale("fr")
			So(fr.Code, ShouldEqual, "fr_FR")
		})
		Convey("FormatFloat", func() {
			fr := GetLocale("fr")
			val := 1234567890.123456
			So(fr.FormatFloat(val, nbutils.Digits{12, 3}), ShouldEqual, "1 234 567 890,123")
			en := GetLocale("en")
			So(en.FormatFloat(val, nbutils.Digits{12, 3}), ShouldEqual, "1,234,567,890.123")
			en.Grouping = NumberGrouping{3, 2}
			So(en.FormatFloat(val, nbutils.Digits{12, 3}), ShouldEqual, "12345,67,890.123")
			en.Grouping = NumberGrouping{3, 2, 0}
			So(en.FormatFloat(val, nbutils.Digits{12, 3}), ShouldEqual, "1,23,45,67,890.123")
			en.Grouping = NumberGrouping{2, 3, 2}
			So(en.FormatFloat(val, nbutils.Digits{12, 5}), ShouldEqual, "123,45,678,90.12346")
			en.Grouping = NumberGrouping{2, 3, 2, 0}
			So(en.FormatFloat(val, nbutils.Digits{12, 3}), ShouldEqual, "1,23,45,678,90.123")
		})
		Convey("FormatDate, FormatTime, FormatDateTime", func() {
			en := GetLocale("en")
			fr := GetLocale("fr")
			date := dates.ParseDate("2003-07-12")
			dateTime := dates.ParseDateTime("2003-07-12 15:02:00")
			So(fr.FormatDate(date), ShouldEqual, "12/07/2003")
			So(en.FormatDate(date), ShouldEqual, "07/12/2003")
			en.TimeFormatGo = "03:04:05 PM"
			So(fr.FormatDateTime(dateTime), ShouldEqual, "12/07/2003 15:02:00")
			So(en.FormatDateTime(dateTime), ShouldEqual, "07/12/2003 03:02:00 PM")
			So(fr.FormatTime(dateTime), ShouldEqual, "15:02:00")
			So(en.FormatTime(dateTime), ShouldEqual, "03:02:00 PM")
		})
		Convey("FormatMonetary", func() {
			fr := GetLocale("fr")
			ja := GetLocale("ja")
			eur := currency{decimals: 2, symbol: "€", position: "after"}
			yen := currency{decimals: 0, symbol: "¥", position: "before"}
			So(fr.FormatMonetary(1234567.789, eur), ShouldEqual, "1 234 567,79 €")
			So(ja.FormatMonetary(1234567.789, yen), ShouldEqual, "¥ 1,234,568")
		})
	})
	Convey("Testing number grouping JSON Marshalling", t, func() {
		b, err := json.Marshal(NumberGrouping{1, 2, 3})
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, `"[1,2,3]"`)
	})
}
