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

import "strings"

/*
ConvertModelName converts an Odoo dotted style model name (e.g. res.partner) into
a YEP Pascal cased style (e.g. ResPartner).
*/
func ConvertModelName(val string) string {
	var res string
	tokens := strings.Split(val, ".")
	for _, token := range tokens {
		res += strings.Title(token)
	}
	return res
}

/*
ConvertMethodName converts an Odoo snake style method name (e.g. search_read) into
a YEP Pascal cased style (e.g. SearchRead).
*/
func ConvertMethodName(val string) string {
	var res string
	tokens := strings.Split(val, "_")
	for _, token := range tokens {
		res += strings.Title(token)
	}
	return res
}
