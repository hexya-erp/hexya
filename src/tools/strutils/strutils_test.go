// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package strutils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnakeCase(t *testing.T) {
	t.Run("Simple cases", func(t *testing.T) {
		assert.Equal(t, "", SnakeCase(""))
		assert.Equal(t, "name", SnakeCase("Name"))
		assert.Equal(t, "my_field_name", SnakeCase("MyFieldName"))
		assert.Equal(t, "my_field_name", SnakeCase("myFieldName"))
	})
	t.Run("Acronyms", func(t *testing.T) {
		assert.Equal(t, "my_html_data", SnakeCase("MyHTMLData"))
		assert.Equal(t, "html", SnakeCase("HTML"))
		assert.Equal(t, "create_uid", SnakeCase("CreateUID"))
	})
}

func TestTitle(t *testing.T) {
	assert.Equal(t, "", Title(""))
	assert.Equal(t, "Name", Title("Name"))
	assert.Equal(t, "My HTML Data", Title("MyHTMLData"))
	assert.Equal(t, "my Field Name", Title("myFieldName"))
}

func TestGetDefaultString(t *testing.T) {
	assert.Equal(t, "def", GetDefaultString("", "def"))
	assert.Equal(t, "str", GetDefaultString("str", "def"))
	assert.Equal(t, "", GetDefaultString("", ""))
}

func TestStartsAndEndsWith(t *testing.T) {
	assert.True(t, StartsAndEndsWith("[1,2]", "[", "]"))
	assert.False(t, StartsAndEndsWith("[1,2)", "[", "]"))
	assert.False(t, StartsAndEndsWith("(1,2]", "[", "]"))
	assert.True(t, StartsAndEndsWith("anything", "", ""))
}

func TestMarshalToJSONString(t *testing.T) {
	t.Run("Strings are returned as is", func(t *testing.T) {
		assert.Equal(t, "my string", MarshalToJSONString("my string"))
	})
	t.Run("Other data is marshalled", func(t *testing.T) {
		assert.Equal(t, `[1,2,3]`, MarshalToJSONString([]int{1, 2, 3}))
		assert.Equal(t, `{"foo":"bar"}`, MarshalToJSONString(map[string]string{"foo": "bar"}))
		assert.Equal(t, "null", MarshalToJSONString(nil))
	})
	t.Run("Unmarshallable data panics", func(t *testing.T) {
		assert.Panics(t, func() { MarshalToJSONString(func() {}) })
	})
}

func TestHumanSize(t *testing.T) {
	assert.Equal(t, "0.00 bytes", HumanSize(0))
	assert.Equal(t, "512.00 bytes", HumanSize(512))
	assert.Equal(t, "1.00 KB", HumanSize(1024))
	assert.Equal(t, "1.50 MB", HumanSize(1024*1024*3/2))
	assert.Equal(t, "2.00 GB", HumanSize(2*1024*1024*1024))
	// Larger sizes stay in GB as it is the largest known unit
	assert.Equal(t, "2048.00 GB", HumanSize(2*1024*1024*1024*1024))
}

func TestSubstitute(t *testing.T) {
	assert.Equal(t, "Hello World", Substitute("Hello World", nil))
	assert.Equal(t, "Hello World", Substitute("Hello $name", map[string]string{"$name": "World"}))
	assert.Equal(t, "b b", Substitute("a a", map[string]string{"a": "b"}))
}

func TestDictToJSON(t *testing.T) {
	assert.Equal(t, `{"foo": "bar"}`, DictToJSON(`{'foo': 'bar'}`))
	assert.Equal(t, `{"foo": true, "bar": false}`, DictToJSON(`{'foo': True, 'bar': False}`))
	assert.Equal(t, `["a", "b"]`, DictToJSON(`('a', 'b')`))
}

func TestMakeUnique(t *testing.T) {
	assert.Equal(t, "foo", MakeUnique("foo", []string{"bar", "baz"}))
	assert.Equal(t, "foo1", MakeUnique("foo", []string{"foo"}))
	assert.Equal(t, "foo3", MakeUnique("foo", []string{"foo", "foo1", "foo2"}))
	assert.Equal(t, "1", MakeUnique("", []string{"foo"}))
}

func TestIsIn(t *testing.T) {
	assert.True(t, IsIn("foo", "bar", "foo"))
	assert.False(t, IsIn("foo", "bar", "baz"))
	assert.False(t, IsIn("foo"))
}

func TestTrimArgs(t *testing.T) {
	t.Run("Empty args", func(t *testing.T) {
		assert.Equal(t, []string{}, TrimArgs(nil))
	})
	t.Run("Short args are kept as is", func(t *testing.T) {
		assert.Equal(t, []string{"foo", "3", "true"}, TrimArgs([]any{"foo", 3, true}))
	})
	t.Run("Long args are trimmed", func(t *testing.T) {
		long := strings.Repeat("a", 40)
		assert.Equal(t, []string{strings.Repeat("a", 30) + "..."}, TrimArgs([]any{long}))
		exact := strings.Repeat("b", 30)
		assert.Equal(t, []string{exact}, TrimArgs([]any{exact}))
	})
}

func TestRemoveAccent(t *testing.T) {
	assert.Equal(t, "", RemoveAccent(""))
	assert.Equal(t, "Elephant", RemoveAccent("Éléphant"))
	assert.Equal(t, "aeiou", RemoveAccent("àéîôù"))
	assert.Equal(t, "no accent", RemoveAccent("no accent"))
}
