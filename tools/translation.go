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

type LangDirection string

const (
	LANG_DIRECTION_LTR LangDirection = "ltr"
	LANG_DIRECTION_RTL LangDirection = "rtl"
)

type LangParameters struct {
	DateFormat   string        `json:"date_format"`
	Direction    LangDirection `json:"lang_direction"`
	ThousandsSep string        `json:"thousands_sep"`
	TimeFormat   string        `json:"time_format"`
	DecimalPoint string        `json:"decimal_point"`
	ID           int64         `json:"id"`
	Grouping     string        `json:"grouping"`
}
