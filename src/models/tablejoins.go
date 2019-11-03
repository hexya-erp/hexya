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

package models

import "fmt"

// tableJoin represents a join in a SQL query
// tableName should be escaped already in the struct
type tableJoin struct {
	tableName  string
	joined     bool
	field      FieldName
	otherTable *tableJoin
	otherField FieldName
	alias      string
	expr       FieldName
}

// sqlString returns the sql string for the tableJoin Clause
func (t tableJoin) sqlString() string {
	if !t.joined {
		return fmt.Sprintf("%s %s ", t.tableName, t.alias)
	}
	return fmt.Sprintf("LEFT JOIN %s %s ON %s.%s=%s.%s ", t.tableName, t.alias, t.otherTable.alias,
		t.otherField.JSON(), t.alias, t.field.JSON())
}
