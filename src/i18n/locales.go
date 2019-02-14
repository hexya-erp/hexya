// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"errors"
	"fmt"
	"time"
)

// A LangDirection defines the direction of a language
// either left-to-right or right-to-left
type LangDirection string

const (
	// LangDirectionLTR defines a language written from left to right
	LangDirectionLTR LangDirection = "ltr"
	// LangDirectionRTL defines a language written from right to left
	LangDirectionRTL LangDirection = "rtl"
)

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

// GetLocale returns a Locale struct describing a language's rules
// at first call, the data file containing all languages parameters is read
// if the language is not loaded, it returns a Locale similar to English (en_US)
func GetLocale(lang string) *Locale {
	out, ok := locales[lang]
	if !ok {
		return &Locale{
			Name:         fmt.Sprintf("UNKNOWN_LOCALE (%s)", lang),
			Direction:    LangDirectionLTR,
			DateFormat:   `%m/%d/%Y`,
			TimeFormat:   `%H:%M:%S`,
			ThousandsSep: `,`,
			DecimalPoint: `.`,
			Grouping:     []int{3, 0},
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

// locales lists all available locales by ISO code.
var locales = map[string]*Locale{
	"it": {
		Name:         "Italian / Italiano",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"lo": {
		Name:         "Lao / ພາສາລາວ",
		DateFormat:   "%d/%m/y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"lv": {
		Name:         "Latvian / latviešu valoda",
		DateFormat:   "%Y.%m.%d.",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"pt": {
		Name:         "Portuguese / Português",
		DateFormat:   "%d-%m-%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"sl": {
		Name:         "Slovenian / slovenščina",
		DateFormat:   "%d. %m. %Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"fi": {
		Name:         "Finnish / Suomi",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H.%M.%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"fr_CH": {
		Name:         "French (CH) / Français (CH)",
		DateFormat:   "%d. %m. %Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "'",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"gu": {
		Name:         "Gujarati / ગુજરાતી",
		DateFormat:   "%A %d %b %Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%I:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{}},
	"et": {
		Name:         "Estonian / Eesti keel",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"kab": {
		Name:         "Kabyle / Taqbaylit",
		DateFormat:   "%m/%d/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%I:%M:%S %p",
		DecimalPoint: ".",
		Grouping:     []int{}},
	"my": {
		Name:         "Burmese / ဗမာစာ",
		DateFormat:   "%Y %b %d %A",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%I:%M:%S %p",
		DecimalPoint: ".",
		Grouping:     []int{3, 3}},
	"es_AR": {
		Name:         "Spanish (AR) / Español (AR)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"es_BO": {
		Name:         "Spanish (BO) / Español (BO)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"am_ET": {
		Name:         "Amharic / አምሃርኛ",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%I:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"ca_ES": {
		Name:         "Catalan / Català",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"en_GB": {
		Name:         "English (UK)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"es": {
		Name:         "Spanish / Español",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"sq": {
		Name:         "Albanian / Shqip",
		DateFormat:   "%Y-%b-%d",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%I.%M.%S.",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"mk": {
		Name:         "Macedonian / македонски јазик",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"es_PY": {
		Name:         "Spanish (PY) / Español (PY)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"hu": {
		Name:         "Hungarian / Magyar",
		DateFormat:   "%Y-%m-%d",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H.%M.%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"sr_RS": {
		Name:         "Serbian (Cyrillic) / српски",
		DateFormat:   "%d.%m.%Y.",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{}},
	"es_MX": {
		Name:         "Spanish (MX) / Español (MX)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"te": {
		Name:         "Telugu / తెలుగు",
		DateFormat:   "%B %d %A %Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%p%I.%M.%S",
		DecimalPoint: ".",
		Grouping:     []int{}},
	"en": {
		Name:         "English",
		DateFormat:   "%m/%d/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"fr_CA": {
		Name:         "French (CA) / Français (CA)",
		DateFormat:   "%Y-%m-%d",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"de": {
		Name:         "German / Deutsch",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"pt_BR": {
		Name:         "Portuguese (BR) / Português (BR)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"ru": {
		Name:         "Russian / русский язык",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"es_CO": {
		Name:         "Spanish (CO) / Español (CO)",
		DateFormat:   "%d-%m-%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"bg": {
		Name:         "Bulgarian / български език",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H,%M,%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"fr_BE": {
		Name:         "French (BE) / Français (BE)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"lt": {
		Name:         "Lithuanian / Lietuvių kalba",
		DateFormat:   "%Y.%m.%d",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"es_VE": {
		Name:         "Spanish (VE) / Español (VE)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"en_CA": {
		Name:         "English (CA)",
		DateFormat:   "%Y-%m-%d",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"gl": {
		Name:         "Galician / Galego",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{}},
	"de_CH": {
		Name:         "German (CH) / Deutsch (CH)",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "'",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 3}},
	"id": {
		Name:         "Indonesian / Bahasa Indonesia",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"es_DO": {
		Name:         "Spanish (DO) / Español (DO)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"hr": {
		Name:         "Croatian / hrvatski jezik",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{}},
	"nl": {
		Name:         "Dutch / Nederlands",
		DateFormat:   "%d-%m-%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{}},
	"en_AU": {
		Name:         "English (AU)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"es_PE": {
		Name:         "Spanish (PE) / Español (PE)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"ka": {
		Name:         "Georgian / ქართული ენა",
		DateFormat:   "%m/%d/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"pl": {
		Name:         "Polish / Język polski",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{}},
	"es_EC": {
		Name:         "Spanish (EC) / Español (EC)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"ja": {
		Name:         "Japanese / 日本語",
		DateFormat:   "%Y年%m月%d日",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H時%M分%S秒",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"km": {
		Name:         "Khmer / ភាសាខ្មែរ",
		DateFormat:   "%d %B %Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"fa": {
		Name:         "Persian / فارس",
		DateFormat:   "%Y/%m/%d",
		Direction:    LangDirectionRTL,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"sr@latin": {
		Name:         "Serbian (Latin) / srpski",
		DateFormat:   "%m/%d/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%I:%M:%S %p",
		DecimalPoint: ".",
		Grouping:     []int{}},
	"es_CL": {
		Name:         "Spanish (CL) / Español (CL)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"ar": {
		Name:         "Arabic / الْعَرَبيّة",
		DateFormat:   "%d %b, %Y",
		Direction:    LangDirectionRTL,
		ThousandsSep: ",",
		TimeFormat:   "%I:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"el_GR": {
		Name:         "Greek / Ελληνικά",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%I:%M:%S %p",
		DecimalPoint: ",",
		Grouping:     []int{}},
	"he": {
		Name:         "Hebrew / עִבְרִי",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionRTL,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"es_GT": {
		Name:         "Spanish (GT) / Español (GT)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"fr": {
		Name:         "French / Français",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"mn": {
		Name:         "Mongolian / монгол",
		DateFormat:   "%Y.%m.%d",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"ko_KP": {
		Name:         "Korean (KP) / 한국어 (KP)",
		DateFormat:   "%m/%d/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%I:%M:%S %p",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"nb_NO": {
		Name:         "Norwegian Bokmål / Norsk bokmål",
		DateFormat:   "%d. %b %Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"sk": {
		Name:         "Slovak / Slovenský jazyk",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"es_PA": {
		Name:         "Spanish (PA) / Español (PA)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"bs": {
		Name:         "Bosnian / bosanski jezik",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{}},
	"zh_HK": {
		Name:         "Chinese (HK)",
		DateFormat:   "%Y年%m月%d日 %A",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%I時%M分%S秒",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"hi": {
		Name:         "Hindi / हिंदी",
		DateFormat:   "%A %d",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%I:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{}},
	"zh_CN": {
		Name:         "Chinese (Simplified) / 简体中文",
		DateFormat:   "%Y年%m月%d日",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H时%M分%S秒",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"es_CR": {
		Name:         "Spanish (CR) / Español (CR)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"uk": {
		Name:         "Ukrainian / українська",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"cs_CZ": {
		Name:         "Czech / Čeština",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"ko": {
		Name:         "Korean (KR) / 한국어 (KR)",
		DateFormat:   "%Y년 %m월 %d일",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H시 %M분 %S초",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"tr": {
		Name:         "Turkish / Türkçe",
		DateFormat:   "%d-%m-%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"da_DK": {
		Name:         "Danish / Dansk",
		DateFormat:   "%d-%m-%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"nl_BE": {
		Name:         "Dutch (BE) / Nederlands (BE)",
		DateFormat:   "%d-%m-%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"ro": {
		Name:         "Romanian / română",
		DateFormat:   "%d.%m.%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"eu_ES": {
		Name:         "Basque / Euskara",
		DateFormat:   "%a, %Y.eko %bren %da",
		Direction:    LangDirectionLTR,
		ThousandsSep: "",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{}},
	"sv": {
		Name:         "Swedish / svenska",
		DateFormat:   "%Y-%m-%d",
		Direction:    LangDirectionLTR,
		ThousandsSep: " ",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"vi": {
		Name:         "Vietnamese / Tiếng Việt",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"zh_TW": {
		Name:         "Chinese (Traditional) / 正體字",
		DateFormat:   "%Y年%m月%d日",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H時%M分%S秒",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}},
	"es_UY": {
		Name:         "Spanish (UY) / Español (UY)",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ".",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ",",
		Grouping:     []int{3, 0}},
	"th": {
		Name:         "Thai / ภาษาไทย",
		DateFormat:   "%d/%m/%Y",
		Direction:    LangDirectionLTR,
		ThousandsSep: ",",
		TimeFormat:   "%H:%M:%S",
		DecimalPoint: ".",
		Grouping:     []int{3, 0}}}
