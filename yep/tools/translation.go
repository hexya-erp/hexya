// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tools

// A LangDirection defines the direction of a language
// either left-to-right or right-to-left
type LangDirection string

const (
	// LangDirectionLTR defines a language written from left to right
	LangDirectionLTR LangDirection = "ltr"
	// LangDirectionRTL defines a language written from right to left
	LangDirectionRTL LangDirection = "rtl"
)

// LangParameters defines the parameters of a language locale
type LangParameters struct {
	DateFormat   string        `json:"date_format"`
	Direction    LangDirection `json:"lang_direction"`
	ThousandsSep string        `json:"thousands_sep"`
	TimeFormat   string        `json:"time_format"`
	DecimalPoint string        `json:"decimal_point"`
	ID           int64         `json:"id"`
	Grouping     string        `json:"grouping"`
}
