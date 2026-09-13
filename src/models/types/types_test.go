// Copyright 2025 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package types

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/hexya-erp/hexya/src/models/types/dates"
	. "github.com/smartystreets/goconvey/convey"
)

// testRecordSet is a dummy RecordSet implementation used to check
// that passing a RecordSet to a Context panics.
type testRecordSet struct{}

func (r testRecordSet) ModelName() string                           { return "TestModel" }
func (r testRecordSet) Ids() []int64                                { return []int64{1} }
func (r testRecordSet) Len() int                                    { return 1 }
func (r testRecordSet) IsEmpty() bool                               { return false }
func (r testRecordSet) IsNotEmpty() bool                            { return true }
func (r testRecordSet) Call(_ string, _ ...interface{}) interface{} { return nil }

func TestContext(t *testing.T) {
	Convey("Testing Context creation and access", t, func() {
		Convey("A new Context should be empty", func() {
			ctx := NewContext()
			So(ctx.IsEmpty(), ShouldBeTrue)
			So(ctx.HasKey("foo"), ShouldBeFalse)
			So(ctx.Get("foo"), ShouldBeNil)
		})
		Convey("WithKey should set values", func() {
			ctx := NewContext().WithKey("foo", "bar").WithKey("baz", 3)
			So(ctx.IsEmpty(), ShouldBeFalse)
			So(ctx.HasKey("foo"), ShouldBeTrue)
			So(ctx.Get("foo"), ShouldEqual, "bar")
			So(ctx.Get("baz"), ShouldEqual, 3)
		})
		Convey("WithKey should overwrite existing values", func() {
			ctx := NewContext().WithKey("foo", "bar").WithKey("foo", "qux")
			So(ctx.Get("foo"), ShouldEqual, "qux")
		})
		Convey("WithKey should panic with a RecordSet value", func() {
			So(func() { NewContext().WithKey("rs", testRecordSet{}) }, ShouldPanic)
		})
	})
	Convey("Testing Context typed getters with existing values", t, func() {
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
		Convey("GetString should return the string value", func() {
			So(ctx.GetString("string"), ShouldEqual, "bar")
		})
		Convey("GetDate and GetDateTime should return the date values", func() {
			So(ctx.GetDate("date").Equal(date), ShouldBeTrue)
			So(ctx.GetDateTime("datetime").Equal(dateTime), ShouldBeTrue)
		})
		Convey("GetInteger should cast numbers to int64", func() {
			So(ctx.GetInteger("int"), ShouldEqual, int64(3))
			So(ctx.GetInteger("int64"), ShouldEqual, int64(4))
			// Non integral float values cannot be parsed as integers and yield 0
			So(ctx.GetInteger("float"), ShouldEqual, int64(0))
			So(ctx.GetInteger("boolTrue"), ShouldEqual, int64(1))
		})
		Convey("GetFloat should cast numbers to float64", func() {
			So(ctx.GetFloat("float"), ShouldEqual, 5.4)
			So(ctx.GetFloat("int"), ShouldEqual, 3.0)
			So(ctx.GetFloat("int64"), ShouldEqual, 4.0)
			So(ctx.GetFloat("boolTrue"), ShouldEqual, 1.0)
		})
		Convey("GetBool should return true only for values equal to 1", func() {
			So(ctx.GetBool("boolTrue"), ShouldBeTrue)
			So(ctx.GetBool("boolFalse"), ShouldBeFalse)
			So(ctx.GetBool("one"), ShouldBeTrue)
			So(ctx.GetBool("int"), ShouldBeFalse)
			So(ctx.GetBool("string"), ShouldBeFalse)
		})
	})
	Convey("Testing Context typed getters with missing keys", t, func() {
		ctx := NewContext()
		So(ctx.GetString("nokey"), ShouldEqual, "")
		So(ctx.GetDate("nokey").IsZero(), ShouldBeTrue)
		So(ctx.GetDateTime("nokey").IsZero(), ShouldBeTrue)
		So(ctx.GetInteger("nokey"), ShouldEqual, int64(0))
		So(ctx.GetFloat("nokey"), ShouldEqual, 0.0)
		So(ctx.GetBool("nokey"), ShouldBeFalse)
		So(ctx.GetStringSlice("nokey"), ShouldResemble, []string{})
		So(ctx.GetIntegerSlice("nokey"), ShouldResemble, []int64{})
		So(ctx.GetFloatSlice("nokey"), ShouldResemble, []float64{})
	})
	Convey("Testing Context typed getters error paths", t, func() {
		ctx := NewContext().
			WithKey("int", 3).
			WithKey("string", "bar").
			WithKey("badSlice", []string{"a"})
		So(func() { ctx.GetString("int") }, ShouldPanic)
		So(func() { ctx.GetDate("int") }, ShouldPanic)
		So(func() { ctx.GetDateTime("int") }, ShouldPanic)
		So(func() { ctx.GetInteger("string") }, ShouldPanic)
		So(func() { ctx.GetFloat("string") }, ShouldPanic)
		So(func() { ctx.GetIntegerSlice("int") }, ShouldPanic)
		So(func() { ctx.GetFloatSlice("int") }, ShouldPanic)
		So(func() { ctx.GetIntegerSlice("badSlice") }, ShouldPanic)
		So(func() { ctx.GetFloatSlice("badSlice") }, ShouldPanic)
	})
	Convey("Testing Context slice getters", t, func() {
		Convey("GetStringSlice", func() {
			ctx := NewContext().
				WithKey("strings", []string{"a", "b"}).
				WithKey("interfaces", []interface{}{"c", "d"}).
				WithKey("other", []int{1, 2})
			So(ctx.GetStringSlice("strings"), ShouldResemble, []string{"a", "b"})
			So(ctx.GetStringSlice("interfaces"), ShouldResemble, []string{"c", "d"})
			So(ctx.GetStringSlice("other"), ShouldBeNil)
			So(func() { ctx.GetStringSlice("interfaces") }, ShouldNotPanic)
		})
		Convey("GetStringSlice should panic with non string elements", func() {
			ctx := NewContext().WithKey("interfaces", []interface{}{"c", 3})
			So(func() { ctx.GetStringSlice("interfaces") }, ShouldPanic)
		})
		Convey("GetIntegerSlice", func() {
			ctx := NewContext().
				WithKey("ints", []int{1, 2}).
				WithKey("int64s", []int64{3, 4}).
				WithKey("floats", []float64{5.2, 6.8}).
				WithKey("interfaces", []interface{}{7, int64(8)})
			So(ctx.GetIntegerSlice("ints"), ShouldResemble, []int64{1, 2})
			So(ctx.GetIntegerSlice("int64s"), ShouldResemble, []int64{3, 4})
			So(ctx.GetIntegerSlice("floats"), ShouldResemble, []int64{0, 0})
			So(ctx.GetIntegerSlice("interfaces"), ShouldResemble, []int64{7, 8})
		})
		Convey("GetFloatSlice", func() {
			ctx := NewContext().
				WithKey("floats", []float64{5.2, 6.8}).
				WithKey("ints", []int{1, 2}).
				WithKey("interfaces", []interface{}{7, 8.5})
			So(ctx.GetFloatSlice("floats"), ShouldResemble, []float64{5.2, 6.8})
			So(ctx.GetFloatSlice("ints"), ShouldResemble, []float64{1, 2})
			So(ctx.GetFloatSlice("interfaces"), ShouldResemble, []float64{7, 8.5})
		})
	})
	Convey("Testing Context mutation methods", t, func() {
		Convey("Delete should remove the given key", func() {
			ctx := NewContext().WithKey("foo", "bar").WithKey("baz", 3)
			ctx = ctx.Delete("foo")
			So(ctx.HasKey("foo"), ShouldBeFalse)
			So(ctx.HasKey("baz"), ShouldBeTrue)
		})
		Convey("Pop should return and remove the given key", func() {
			ctx := NewContext().WithKey("foo", "bar")
			So(ctx.Pop("foo"), ShouldEqual, "bar")
			So(ctx.HasKey("foo"), ShouldBeFalse)
			So(ctx.Pop("nokey"), ShouldBeNil)
		})
		Convey("Cleaned should remove keys with the given prefix", func() {
			ctx := NewContext().
				WithKey("default_foo", "bar").
				WithKey("default_baz", 3).
				WithKey("lang", "fr_FR")
			res := ctx.Cleaned("default_")
			So(res.HasKey("default_foo"), ShouldBeFalse)
			So(res.HasKey("default_baz"), ShouldBeFalse)
			So(res.GetString("lang"), ShouldEqual, "fr_FR")
			So(ctx.HasKey("default_foo"), ShouldBeTrue)
		})
		Convey("Update should merge the other context", func() {
			ctx := NewContext().WithKey("foo", "bar")
			other := NewContext().WithKey("baz", 3).WithKey("foo", "qux")
			ctx.Update(other)
			So(ctx.GetString("foo"), ShouldEqual, "qux")
			So(ctx.GetInteger("baz"), ShouldEqual, int64(3))
		})
		Convey("Copy should return an independent context", func() {
			ctx := NewContext().WithKey("foo", "bar")
			newCtx := ctx.Copy()
			So(newCtx.GetString("foo"), ShouldEqual, "bar")
			newCtx.Delete("foo")
			So(newCtx.HasKey("foo"), ShouldBeFalse)
			So(ctx.HasKey("foo"), ShouldBeTrue)
		})
		Convey("ToMap should return a copy of the values", func() {
			ctx := NewContext().WithKey("foo", "bar")
			m := ctx.ToMap()
			So(m, ShouldResemble, map[string]interface{}{"foo": "bar"})
			delete(m, "foo")
			So(ctx.HasKey("foo"), ShouldBeTrue)
		})
		Convey("String should render the values map", func() {
			ctx := NewContext().WithKey("foo", "bar")
			So(ctx.String(), ShouldEqual, "map[foo:bar]")
		})
	})
}

type testXMLStruct struct {
	XMLName xml.Name `xml:"test"`
	Context *Context `xml:"context,attr"`
}

func TestContextSerialization(t *testing.T) {
	Convey("Testing Context JSON marshalling", t, func() {
		ctx := NewContext().WithKey("foo", "bar").WithKey("baz", 3)
		data, err := json.Marshal(ctx)
		So(err, ShouldBeNil)
		Convey("Unmarshalling back should give the same values", func() {
			var newCtx Context
			So(json.Unmarshal(data, &newCtx), ShouldBeNil)
			So(newCtx.ToMap(), ShouldResemble, map[string]interface{}{
				"foo": "bar",
				"baz": float64(3),
			})
		})
		Convey("Unmarshalling invalid JSON should fail", func() {
			var newCtx Context
			So(json.Unmarshal([]byte(`{"foo":}`), &newCtx), ShouldNotBeNil)
		})
	})
	Convey("Testing Context XML attribute unmarshalling", t, func() {
		Convey("UnmarshalXMLAttr should populate the values", func() {
			var ctx Context
			So(ctx.UnmarshalXMLAttr(xml.Attr{Value: `{"foo":"bar","baz":3}`}), ShouldBeNil)
			So(ctx.GetString("foo"), ShouldEqual, "bar")
			So(ctx.GetInteger("baz"), ShouldEqual, int64(3))
		})
		Convey("UnmarshalXMLAttr should fail with invalid JSON", func() {
			var ctx Context
			So(ctx.UnmarshalXMLAttr(xml.Attr{Value: `{"foo":}`}), ShouldNotBeNil)
		})
		Convey("Unmarshalling a struct with a Context attribute should work", func() {
			var ts testXMLStruct
			So(xml.Unmarshal([]byte(`<test context="{&#34;foo&#34;:&#34;bar&#34;}"/>`), &ts), ShouldBeNil)
			So(ts.Context.GetString("foo"), ShouldEqual, "bar")
		})
	})
	Convey("Testing Context database serialization", t, func() {
		ctx := NewContext().WithKey("foo", "bar")
		Convey("Value should return a JSON encoded context", func() {
			val, err := ctx.Value()
			So(err, ShouldBeNil)
			// MarshalJSON has a pointer receiver, so it is not used here
			So(string(val.([]byte)), ShouldEqual, `{}`)
		})
		Convey("Scan should accept a string", func() {
			var newCtx Context
			So(newCtx.Scan(`{"foo":"bar"}`), ShouldBeNil)
			So(newCtx.GetString("foo"), ShouldEqual, "bar")
		})
		Convey("Scan should accept a byte slice", func() {
			var newCtx Context
			So(newCtx.Scan([]byte(`{"foo":"bar"}`)), ShouldBeNil)
			So(newCtx.GetString("foo"), ShouldEqual, "bar")
		})
		Convey("Scan should accept a map", func() {
			var newCtx Context
			So(newCtx.Scan(map[string]interface{}{"foo": "bar"}), ShouldBeNil)
			So(newCtx.GetString("foo"), ShouldEqual, "bar")
		})
		Convey("Scan should fail with an unsupported type", func() {
			var newCtx Context
			So(newCtx.Scan(3), ShouldNotBeNil)
		})
		Convey("Scan should fail with malformed JSON", func() {
			var newCtx Context
			So(newCtx.Scan([]byte(`{"foo":}`)), ShouldNotBeNil)
		})
		Convey("Marshalling and Scan should round trip", func() {
			data, err := json.Marshal(ctx)
			So(err, ShouldBeNil)
			var newCtx Context
			So(newCtx.Scan(data), ShouldBeNil)
			So(newCtx.ToMap(), ShouldResemble, ctx.ToMap())
		})
	})
}

func TestSelection(t *testing.T) {
	Convey("Testing Selection marshalling", t, func() {
		Convey("Keys should be marshalled in sorted order", func() {
			sel := Selection{"b": "B", "a": "A", "c": "C"}
			data, err := json.Marshal(sel)
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `[["a","A"],["b","B"],["c","C"]]`)
		})
		Convey("An empty Selection should be marshalled as null", func() {
			data, err := json.Marshal(Selection{})
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `null`)
		})
	})
}
