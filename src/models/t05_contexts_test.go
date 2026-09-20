// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextedPaths(t *testing.T) {
	t.Run("No context value should only give the default path", func(t *testing.T) {
		assert.EqualValues(t, contextedPaths([]string{"lang"}, nil), [][]string{{ctxValueKey}})
		assert.EqualValues(t, contextedPaths([]string{"lang"}, map[string]string{"lang": ""}),
			[][]string{{ctxValueKey}})
		assert.EqualValues(t, contextedPaths(nil, map[string]string{"lang": "fr_FR"}),
			[][]string{{ctxValueKey}})
	})
	t.Run("One context", func(t *testing.T) {
		assert.EqualValues(t, contextedPaths([]string{"lang"}, map[string]string{"lang": "fr_FR"}),
			[][]string{
				{"lang", "fr_FR", ctxValueKey},
				{ctxValueKey},
			})
	})
	t.Run("Two contexts should be sorted alphabetically, the last one having priority", func(t *testing.T) {
		ctxValues := map[string]string{"company": "3", "lang": "fr_FR"}
		assert.EqualValues(t, contextedPaths([]string{"lang", "company"}, ctxValues),
			[][]string{
				{"company", "3", "lang", "fr_FR", ctxValueKey},
				{"lang", "fr_FR", ctxValueKey},
				{"company", "3", ctxValueKey},
				{ctxValueKey},
			})
	})
	t.Run("Contexts without a value should be ignored", func(t *testing.T) {
		ctxValues := map[string]string{"lang": "fr_FR"}
		assert.EqualValues(t, contextedPaths([]string{"lang", "company"}, ctxValues),
			[][]string{
				{"lang", "fr_FR", ctxValueKey},
				{ctxValueKey},
			})
	})
	t.Run("Three contexts", func(t *testing.T) {
		ctxValues := map[string]string{"company": "3", "lang": "fr_FR", "user": "8"}
		assert.EqualValues(t, contextedPaths([]string{"user", "lang", "company"}, ctxValues),
			[][]string{
				{"company", "3", "lang", "fr_FR", "user", "8", ctxValueKey},
				{"lang", "fr_FR", "user", "8", ctxValueKey},
				{"company", "3", "user", "8", ctxValueKey},
				{"user", "8", ctxValueKey},
				{"company", "3", "lang", "fr_FR", ctxValueKey},
				{"lang", "fr_FR", ctxValueKey},
				{"company", "3", ctxValueKey},
				{ctxValueKey},
			})
	})
}

func TestContextedValue(t *testing.T) {
	ctxNames := []string{"company", "lang"}
	newValue := func() ContextedValue {
		return ContextedValue{
			ctxValueKey: "Chair",
			"lang": map[string]any{
				"fr_FR": map[string]any{ctxValueKey: "Siège"},
			},
			"company": map[string]any{
				"3": map[string]any{
					ctxValueKey: "Seat",
					"lang": map[string]any{
						"fr_FR": map[string]any{ctxValueKey: "Fauteuil"},
					},
				},
			},
		}
	}
	t.Run("Resolving values", func(t *testing.T) {
		cv := newValue()
		for _, tc := range []struct {
			ctxValues map[string]string
			expected  string
		}{
			{nil, "Chair"},
			{map[string]string{"lang": "fr_FR"}, "Siège"},
			{map[string]string{"lang": "de_DE"}, "Chair"},
			{map[string]string{"company": "3"}, "Seat"},
			{map[string]string{"company": "3", "lang": "fr_FR"}, "Fauteuil"},
			{map[string]string{"company": "3", "lang": "de_DE"}, "Seat"},
			{map[string]string{"company": "7", "lang": "fr_FR"}, "Siège"},
			{map[string]string{"company": "7", "lang": "de_DE"}, "Chair"},
		} {
			res, ok := cv.resolve(ctxNames, tc.ctxValues)
			assert.True(t, ok)
			assert.EqualValues(t, res, tc.expected)
		}
	})
	t.Run("Resolving an empty value", func(t *testing.T) {
		cv := make(ContextedValue)
		res, ok := cv.resolve(ctxNames, map[string]string{"lang": "fr_FR"})
		assert.False(t, ok)
		assert.Nil(t, res)
	})
	t.Run("Setting values should create the missing levels", func(t *testing.T) {
		cv := newValue()
		paths := contextedPaths(ctxNames, map[string]string{"company": "7", "lang": "de_DE"})
		cv.set(paths[0], "Stuhl der Firma 7")
		res, ok := cv.resolve(ctxNames, map[string]string{"company": "7", "lang": "de_DE"})
		assert.True(t, ok)
		assert.EqualValues(t, res, "Stuhl der Firma 7")
		// Other branches are left untouched
		res, _ = cv.resolve(ctxNames, map[string]string{"company": "3", "lang": "fr_FR"})
		assert.EqualValues(t, res, "Fauteuil")
		res, _ = cv.resolve(ctxNames, nil)
		assert.EqualValues(t, res, "Chair")
	})
	t.Run("Setting the default value", func(t *testing.T) {
		cv := newValue()
		cv.set([]string{ctxValueKey}, "Stool")
		res, _ := cv.resolve(ctxNames, map[string]string{"lang": "de_DE"})
		assert.EqualValues(t, res, "Stool")
		res, _ = cv.resolve(ctxNames, map[string]string{"lang": "fr_FR"})
		assert.EqualValues(t, res, "Siège")
	})
}
