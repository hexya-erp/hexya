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

package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"sort"

	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/logging"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
)

// RecordSet identifies a type that holds a set of records of
// a given model. Dummy interface in types module.
type RecordSet interface {
	// ModelName returns the name of the model of this RecordSet
	ModelName() string
	// Ids returns the ids in this set of Records
	Ids() []int64
	// Len returns the number of records in this RecordSet
	Len() int
	// IsEmpty returns true if this RecordSet has no records
	IsEmpty() bool
	// Call executes the given method (as string) with the given arguments
	Call(string, ...interface{}) interface{}
}

var log logging.Logger

// A Context is a map of objects that is passed along from function to function
// during a transaction. A Context is read only.
type Context struct {
	values map[string]interface{}
}

// Copy returns a shallow copy of the Context
func (c Context) Copy() *Context {
	newCtx := NewContext()
	for k, v := range c.values {
		newCtx.values[k] = v
	}
	return newCtx
}

// Get returns the value of the given key in this Context
// It returns nil if the key is not in this context
func (c *Context) Get(key string) interface{} {
	value := c.values[key]
	return value
}

// GetString returns the value of the given key in
// this Context as a string.
// It returns an empty string if there is no such key in the context.
// It panics if the value is not of type string
func (c *Context) GetString(key string) string {
	if !c.HasKey(key) {
		return ""
	}
	return c.Get(key).(string)
}

// GetDate returns the value of the given key in
// this Context as a Date.
// It returns an null Date if there is no such key in the context.
// It panics if the value is not of type Date
func (c *Context) GetDate(key string) dates.Date {
	if !c.HasKey(key) {
		return dates.Date{}
	}
	return c.Get(key).(dates.Date)
}

// GetDateTime returns the value of the given key in
// this Context as a DateTime.
// It returns an null DateTime if there is no such key in the context.
// It panics if the value is not of type DateTime
func (c *Context) GetDateTime(key string) dates.DateTime {
	if !c.HasKey(key) {
		return dates.DateTime{}
	}
	return c.Get(key).(dates.DateTime)
}

// GetInteger returns the value of the given key in
// this Context as an int64.
// It returns 0 if there is no such key in the context.
// It panics if the value cannot be casted to int64
func (c *Context) GetInteger(key string) int64 {
	if !c.HasKey(key) {
		return 0
	}
	val := c.Get(key)
	res, err := nbutils.CastToInteger(val)
	if err != nil {
		log.Panic(err.Error(), "ContextKey", key)
	}
	return res
}

// GetFloat returns the value of the given key in
// this Context as a float64.
// It returns 0 if there is no such key in the context.
// It panics if the value cannot be casted to float64
func (c *Context) GetFloat(key string) float64 {
	if !c.HasKey(key) {
		return 0
	}
	val := c.Get(key)
	res, err := nbutils.CastToFloat(val)
	if err != nil {
		log.Panic(err.Error(), "ContextKey", key)
	}
	return res
}

// GetStringSlice returns the value of the given key in
// this Context as a []string.
// It returns an empty slice if there is no such key in the context.
// It panics if the value is not a slice or if any value
// is not a string
func (c *Context) GetStringSlice(key string) []string {
	if !c.HasKey(key) {
		return []string{}
	}
	val := c.Get(key)
	var res []string
	switch value := val.(type) {
	case []string:
		res = value
	case []interface{}:
		res = make([]string, len(value))
		for i, v := range value {
			res[i] = v.(string)
		}
	}
	return res
}

// GetIntegerSlice returns the value of the given key in
// this Context as a []int64.
// It returns an empty slice if there is no such key in the context.
// It panics if the value is not a slice or if any value
// cannot be casted to int64
func (c *Context) GetIntegerSlice(key string) []int64 {
	if !c.HasKey(key) {
		return []int64{}
	}
	val := c.Get(key)
	rVal := reflect.ValueOf(val)
	if rVal.Kind() != reflect.Slice {
		log.Panic("Value in Context is not a slice", "key", key, "value", val)
	}
	res := make([]int64, rVal.Len())
	var err error
	for i := 0; i < rVal.Len(); i++ {
		res[i], err = nbutils.CastToInteger(rVal.Index(i).Interface())
		if err != nil {
			log.Panic(err.Error(), "ContextKey", key)
		}
	}
	return res
}

// GetFloatSlice returns the value of the given key in
// this Context as a []float64.
// It returns an empty slice if there is no such key in the context.
// It panics if the value is not a slice or if any value
// cannot be casted to float64
func (c *Context) GetFloatSlice(key string) []float64 {
	if !c.HasKey(key) {
		return []float64{}
	}
	val := c.Get(key)
	rVal := reflect.ValueOf(val)
	if rVal.Kind() != reflect.Slice {
		log.Panic("Value in Context is not a slice", "key", key, "value", val)
	}
	res := make([]float64, rVal.Len())
	var err error
	for i := 0; i < rVal.Len(); i++ {
		res[i], err = nbutils.CastToFloat(rVal.Index(i).Interface())
		if err != nil {
			log.Panic(err.Error(), "ContextKey", key)
		}
	}
	return res
}

// GetBool returns the value of the given key in
// this Context as a bool.
// It returns false if there is no such key in the context.
// It panics if the value cannot be casted to bool
func (c *Context) GetBool(key string) bool {
	if !c.HasKey(key) {
		return false
	}
	val := c.Get(key)
	res, _ := nbutils.CastToFloat(val)
	return res == 1
}

// HasKey returns true if this Context has the given key
func (c *Context) HasKey(key string) bool {
	_, exists := c.values[key]
	return exists
}

// WithKey returns a copy of this context with the given key/value.
// If key already exists, it is overwritten.
func (c Context) WithKey(key string, value interface{}) *Context {
	if _, ok := value.(RecordSet); ok {
		log.Panic("Recordset passed in Context. Pass ID instead", "key", key, "value", value)
	}
	c.values[key] = value
	return &c
}

// IsEmpty returns true if this Context has no entries.
func (c Context) IsEmpty() bool {
	if len(c.values) == 0 {
		return true
	}
	return false
}

// ToMap returns a copy of the map of values of this context
func (c Context) ToMap() map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range c.values {
		res[k] = v
	}
	return res
}

// UnmarshalXMLAttr is the XML unmarshalling method of Context.
func (c *Context) UnmarshalXMLAttr(attr xml.Attr) error {
	var cm map[string]interface{}
	err := json.Unmarshal([]byte(attr.Value), &cm)
	(*c).values = cm
	return err
}

// MarshalJSON method for Context
func (c *Context) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.values)
}

// UnmarshalJSON method for Context
func (c *Context) UnmarshalJSON(data []byte) error {
	var cm map[string]interface{}
	err := json.Unmarshal(data, &cm)
	(*c).values = cm
	return err
}

// String function for Context type
func (c Context) String() string {
	return fmt.Sprintf("%v", c.values)
}

// Value JSON encode our Context for storing in the database.
func (c Context) Value() (driver.Value, error) {
	bytes, err := json.Marshal(c)
	return driver.Value(bytes), err
}

// Scan JSON decodes the value of the database into a Context
func (c *Context) Scan(src interface{}) error {
	var data []byte
	switch s := src.(type) {
	case string:
		data = []byte(s)
	case []byte:
		data = s
	case map[string]interface{}:
		c.values = s
		return nil
	default:
		return fmt.Errorf("invalid type for Context: %T", src)
	}
	var ctx Context
	err := json.Unmarshal(data, &ctx)
	if err != nil {
		return err
	}
	*c = ctx
	return nil

}

var _ driver.Valuer = Context{}
var _ sql.Scanner = &Context{}
var _ xml.UnmarshalerAttr = &Context{}
var _ json.Marshaler = &Context{}
var _ json.Unmarshaler = &Context{}

// NewContext returns a new Context instance
func NewContext() *Context {
	values := make(map[string]interface{})
	return &Context{
		values: values,
	}
}

// A Selection is a set of possible (key, label) values for a model
// "selection" field.
type Selection map[string]string

// MarshalJSON function for the Selection type
func (s Selection) MarshalJSON() ([]byte, error) {
	keys := make([]string, len(s))
	var i int
	for k := range s {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var selSlice [][2]string
	for _, key := range keys {
		selSlice = append(selSlice, [2]string{0: key, 1: s[key]})
	}
	return json.Marshal(selSlice)
}

var _ json.Marshaler = Selection{}

func init() {
	log = logging.GetLogger("types")
}
