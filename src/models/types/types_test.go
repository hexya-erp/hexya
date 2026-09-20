// Copyright 2025 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package types

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/types/dates"
)

// testRecordSet is a dummy RecordSet implementation used to check
// that passing a RecordSet to a Context panics.
type testRecordSet struct{}

func (r testRecordSet) ModelName() string           { return "TestModel" }
func (r testRecordSet) Ids() []int64                { return []int64{1} }
func (r testRecordSet) Len() int                    { return 1 }
func (r testRecordSet) IsEmpty() bool               { return false }
func (r testRecordSet) IsNotEmpty() bool            { return true }
func (r testRecordSet) Call(_ string, _ ...any) any { return nil }

func TestContext(t *testing.T) {
	t.Run("Testing Context creation and access", func(t *testing.T) {
		t.Run("A new Context should be empty", func(t *testing.T) {
			ctx := NewContext()
			assert.True(t, ctx.IsEmpty())
			assert.False(t, ctx.HasKey("foo"))
			assert.Nil(t, ctx.Get("foo"))
		})
		t.Run("WithKey should set values", func(t *testing.T) {
			ctx := NewContext().WithKey("foo", "bar").WithKey("baz", 3)
			assert.False(t, ctx.IsEmpty())
			assert.True(t, ctx.HasKey("foo"))
			assert.EqualValues(t, ctx.Get("foo"), "bar")
			assert.EqualValues(t, ctx.Get("baz"), 3)
		})
		t.Run("WithKey should overwrite existing values", func(t *testing.T) {
			ctx := NewContext().WithKey("foo", "bar").WithKey("foo", "qux")
			assert.EqualValues(t, ctx.Get("foo"), "qux")
		})
		t.Run("WithKey should panic with a RecordSet value", func(t *testing.T) {
			assert.Panics(t, func() { NewContext().WithKey("rs", testRecordSet{}) })
		})
	})
	t.Run("Testing Context typed getters with existing values", func(t *testing.T) {
		date := dates.ParseDate("2025-01-15")
		dateTime := dates.ParseDateTime("2025-01-15 10:20:30")
		ctx := NewContext().
			WithKey("string", "bar").
			WithKey("date", date).
			WithKey("datetime", dateTime).
			WithKey("int", 3).
			WithKey("int64", int64(4)).
			WithKey("float", 5.4).
			WithKey("boolTrue", true).
			WithKey("boolFalse", false).
			WithKey("one", 1)
		t.Run("GetString should return the string value", func(t *testing.T) {
			assert.EqualValues(t, ctx.GetString("string"), "bar")
		})
		t.Run("GetDate and GetDateTime should return the date values", func(t *testing.T) {
			assert.True(t, ctx.GetDate("date").Equal(date))
			assert.True(t, ctx.GetDateTime("datetime").Equal(dateTime))
		})
		t.Run("GetInteger should cast numbers to int64", func(t *testing.T) {
			assert.EqualValues(t, ctx.GetInteger("int"), int64(3))
			assert.EqualValues(t, ctx.GetInteger("int64"), int64(4))
			assert.EqualValues(t, ctx.GetInteger("float"), int64(5))
			assert.EqualValues(t, ctx.GetInteger("boolTrue"), int64(1))
		})
		t.Run("GetFloat should cast numbers to float64", func(t *testing.T) {
			assert.EqualValues(t, ctx.GetFloat("float"), 5.4)
			assert.EqualValues(t, ctx.GetFloat("int"), 3.0)
			assert.EqualValues(t, ctx.GetFloat("int64"), 4.0)
			assert.EqualValues(t, ctx.GetFloat("boolTrue"), 1.0)
		})
		t.Run("GetBool should return true only for non zero number values", func(t *testing.T) {
			assert.True(t, ctx.GetBool("boolTrue"))
			assert.False(t, ctx.GetBool("boolFalse"))
			assert.True(t, ctx.GetBool("one"))
			assert.True(t, ctx.GetBool("int"))
			assert.False(t, ctx.GetBool("string"))
		})
	})
	t.Run("Testing Context typed getters with missing keys", func(t *testing.T) {
		ctx := NewContext()
		assert.EqualValues(t, ctx.GetString("nokey"), "")
		assert.True(t, ctx.GetDate("nokey").IsZero())
		assert.True(t, ctx.GetDateTime("nokey").IsZero())
		assert.EqualValues(t, ctx.GetInteger("nokey"), int64(0))
		assert.EqualValues(t, ctx.GetFloat("nokey"), 0.0)
		assert.False(t, ctx.GetBool("nokey"))
		assert.Equal(t, ctx.GetStringSlice("nokey"), []string{})
		assert.Equal(t, ctx.GetIntegerSlice("nokey"), []int64{})
		assert.Equal(t, ctx.GetFloatSlice("nokey"), []float64{})
	})
	t.Run("Testing Context typed getters error paths", func(t *testing.T) {
		ctx := NewContext().
			WithKey("int", 3).
			WithKey("string", "bar").
			WithKey("badSlice", []string{"a"})
		assert.Panics(t, func() { ctx.GetString("int") })
		assert.Panics(t, func() { ctx.GetDate("int") })
		assert.Panics(t, func() { ctx.GetDateTime("int") })
		assert.Panics(t, func() { ctx.GetInteger("string") })
		assert.Panics(t, func() { ctx.GetFloat("string") })
		assert.Panics(t, func() { ctx.GetIntegerSlice("int") })
		assert.Panics(t, func() { ctx.GetFloatSlice("int") })
		assert.Panics(t, func() { ctx.GetIntegerSlice("badSlice") })
		assert.Panics(t, func() { ctx.GetFloatSlice("badSlice") })
	})
	t.Run("Testing Context slice getters", func(t *testing.T) {
		t.Run("GetStringSlice", func(t *testing.T) {
			ctx := NewContext().
				WithKey("strings", []string{"a", "b"}).
				WithKey("interfaces", []any{"c", "d"}).
				WithKey("other", []int{1, 2})
			assert.Equal(t, ctx.GetStringSlice("strings"), []string{"a", "b"})
			assert.Equal(t, ctx.GetStringSlice("interfaces"), []string{"c", "d"})
			assert.Nil(t, ctx.GetStringSlice("other"))
			assert.NotPanics(t, func() { ctx.GetStringSlice("interfaces") })
		})
		t.Run("GetStringSlice should panic with non string elements", func(t *testing.T) {
			ctx := NewContext().WithKey("interfaces", []any{"c", 3})
			assert.Panics(t, func() { ctx.GetStringSlice("interfaces") })
		})
		t.Run("GetIntegerSlice", func(t *testing.T) {
			ctx := NewContext().
				WithKey("ints", []int{1, 2}).
				WithKey("int64s", []int64{3, 4}).
				WithKey("floats", []float64{5.2, 6.8}).
				WithKey("interfaces", []any{7, int64(8)})
			assert.Equal(t, ctx.GetIntegerSlice("ints"), []int64{1, 2})
			assert.Equal(t, ctx.GetIntegerSlice("int64s"), []int64{3, 4})
			assert.Equal(t, ctx.GetIntegerSlice("floats"), []int64{5, 6})
			assert.Equal(t, ctx.GetIntegerSlice("interfaces"), []int64{7, 8})
		})
		t.Run("GetFloatSlice", func(t *testing.T) {
			ctx := NewContext().
				WithKey("floats", []float64{5.2, 6.8}).
				WithKey("ints", []int{1, 2}).
				WithKey("interfaces", []any{7, 8.5})
			assert.Equal(t, ctx.GetFloatSlice("floats"), []float64{5.2, 6.8})
			assert.Equal(t, ctx.GetFloatSlice("ints"), []float64{1, 2})
			assert.Equal(t, ctx.GetFloatSlice("interfaces"), []float64{7, 8.5})
		})
	})
	t.Run("Testing Context mutation methods", func(t *testing.T) {
		t.Run("Delete should remove the given key", func(t *testing.T) {
			ctx := NewContext().WithKey("foo", "bar").WithKey("baz", 3)
			ctx = ctx.Delete("foo")
			assert.False(t, ctx.HasKey("foo"))
			assert.True(t, ctx.HasKey("baz"))
		})
		t.Run("Pop should return and remove the given key", func(t *testing.T) {
			ctx := NewContext().WithKey("foo", "bar")
			assert.EqualValues(t, ctx.Pop("foo"), "bar")
			assert.False(t, ctx.HasKey("foo"))
			assert.Nil(t, ctx.Pop("nokey"))
		})
		t.Run("Cleaned should remove keys with the given prefix", func(t *testing.T) {
			ctx := NewContext().
				WithKey("default_foo", "bar").
				WithKey("default_baz", 3).
				WithKey("lang", "fr_FR")
			res := ctx.Cleaned("default_")
			assert.False(t, res.HasKey("default_foo"))
			assert.False(t, res.HasKey("default_baz"))
			assert.EqualValues(t, res.GetString("lang"), "fr_FR")
			assert.True(t, ctx.HasKey("default_foo"))
		})
		t.Run("Update should merge the other context", func(t *testing.T) {
			ctx := NewContext().WithKey("foo", "bar")
			other := NewContext().WithKey("baz", 3).WithKey("foo", "qux")
			ctx.Update(other)
			assert.EqualValues(t, ctx.GetString("foo"), "qux")
			assert.EqualValues(t, ctx.GetInteger("baz"), int64(3))
		})
		t.Run("Copy should return an independent context", func(t *testing.T) {
			ctx := NewContext().WithKey("foo", "bar")
			newCtx := ctx.Copy()
			assert.EqualValues(t, newCtx.GetString("foo"), "bar")
			newCtx.Delete("foo")
			assert.False(t, newCtx.HasKey("foo"))
			assert.True(t, ctx.HasKey("foo"))
		})
		t.Run("ToMap should return a copy of the values", func(t *testing.T) {
			ctx := NewContext().WithKey("foo", "bar")
			m := ctx.ToMap()
			assert.Equal(t, m, map[string]any{"foo": "bar"})
			delete(m, "foo")
			assert.True(t, ctx.HasKey("foo"))
		})
		t.Run("String should render the values map", func(t *testing.T) {
			ctx := NewContext().WithKey("foo", "bar")
			assert.EqualValues(t, ctx.String(), "map[foo:bar]")
		})
	})
}

type testXMLStruct struct {
	XMLName xml.Name `xml:"test"`
	Context *Context `xml:"context,attr"`
}

func TestContextSerialization(t *testing.T) {
	t.Run("Testing Context JSON marshalling", func(t *testing.T) {
		ctx := NewContext().WithKey("foo", "bar").WithKey("baz", 3)
		data, err := json.Marshal(ctx)
		assert.Nil(t, err)
		t.Run("Unmarshalling back should give the same values", func(t *testing.T) {
			var newCtx Context
			assert.Nil(t, json.Unmarshal(data, &newCtx))
			assert.Equal(t, newCtx.ToMap(), map[string]any{
				"foo": "bar",
				"baz": float64(3),
			})
		})
		t.Run("Unmarshalling invalid JSON should fail", func(t *testing.T) {
			var newCtx Context
			assert.NotNil(t, json.Unmarshal([]byte(`{"foo":}`), &newCtx))
		})
	})
	t.Run("Testing Context XML attribute unmarshalling", func(t *testing.T) {
		t.Run("UnmarshalXMLAttr should populate the values", func(t *testing.T) {
			var ctx Context
			assert.Nil(t, ctx.UnmarshalXMLAttr(xml.Attr{Value: `{"foo":"bar","baz":3}`}))
			assert.EqualValues(t, ctx.GetString("foo"), "bar")
			assert.EqualValues(t, ctx.GetInteger("baz"), int64(3))
		})
		t.Run("UnmarshalXMLAttr should fail with invalid JSON", func(t *testing.T) {
			var ctx Context
			assert.NotNil(t, ctx.UnmarshalXMLAttr(xml.Attr{Value: `{"foo":}`}))
		})
		t.Run("Unmarshalling a struct with a Context attribute should work", func(t *testing.T) {
			var ts testXMLStruct
			assert.Nil(t, xml.Unmarshal([]byte(`<test context="{&#34;foo&#34;:&#34;bar&#34;}"/>`), &ts))
			assert.EqualValues(t, ts.Context.GetString("foo"), "bar")
		})
	})
	t.Run("Testing Context database serialization", func(t *testing.T) {
		ctx := NewContext().WithKey("foo", "bar")
		t.Run("Value should return a JSON encoded context", func(t *testing.T) {
			val, err := ctx.Value()
			assert.Nil(t, err)
			assert.EqualValues(t, string(val.([]byte)), `{"foo":"bar"}`)
		})
		t.Run("Scan should accept a string", func(t *testing.T) {
			var newCtx Context
			assert.Nil(t, newCtx.Scan(`{"foo":"bar"}`))
			assert.EqualValues(t, newCtx.GetString("foo"), "bar")
		})
		t.Run("Scan should accept a byte slice", func(t *testing.T) {
			var newCtx Context
			assert.Nil(t, newCtx.Scan([]byte(`{"foo":"bar"}`)))
			assert.EqualValues(t, newCtx.GetString("foo"), "bar")
		})
		t.Run("Scan should accept a map", func(t *testing.T) {
			var newCtx Context
			assert.Nil(t, newCtx.Scan(map[string]any{"foo": "bar"}))
			assert.EqualValues(t, newCtx.GetString("foo"), "bar")
		})
		t.Run("Scan should fail with an unsupported type", func(t *testing.T) {
			var newCtx Context
			assert.NotNil(t, newCtx.Scan(3))
		})
		t.Run("Scan should fail with malformed JSON", func(t *testing.T) {
			var newCtx Context
			assert.NotNil(t, newCtx.Scan([]byte(`{"foo":}`)))
		})
		t.Run("Marshalling and Scan should round trip", func(t *testing.T) {
			data, err := json.Marshal(ctx)
			assert.Nil(t, err)
			var newCtx Context
			assert.Nil(t, newCtx.Scan(data))
			assert.Equal(t, newCtx.ToMap(), ctx.ToMap())
		})
	})
}

func TestSelection(t *testing.T) {
	t.Run("Testing Selection marshalling", func(t *testing.T) {
		t.Run("Keys should be marshalled in sorted order", func(t *testing.T) {
			sel := Selection{"b": "B", "a": "A", "c": "C"}
			data, err := json.Marshal(sel)
			assert.Nil(t, err)
			assert.EqualValues(t, string(data), `[["a","A"],["b","B"],["c","C"]]`)
		})
		t.Run("An empty Selection should be marshalled as null", func(t *testing.T) {
			data, err := json.Marshal(Selection{})
			assert.Nil(t, err)
			assert.EqualValues(t, string(data), `null`)
		})
	})
}
