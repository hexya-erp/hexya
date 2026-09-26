// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"maps"
	"sort"
)

// ctxValueKey is the key under which the value of a branch of a contexted
// value document is stored.
const ctxValueKey = "_"

// A ContextedValue holds all the values of a contexted field for a record.
//
// It is a tree of nested objects where each level is alternatively a context
// name and a context value. The value of each branch is stored in the branch
// node under the ctxValueKey key.
//
// For instance, a field with the "lang" and "company" contexts could hold:
//
//	{
//	    "_": "Product",
//	    "lang": {
//	        "fr_FR": { "_": "Produit" }
//	    },
//	    "company": {
//	        "3": {
//	            "_": "Item",
//	            "lang": { "fr_FR": { "_": "Article" } }
//	        }
//	    }
//	}
type ContextedValue map[string]any

// get returns the value of this ContextedValue at the given path.
//
// Second returned value is false if there is no value at the given path.
func (cv ContextedValue) get(path []string) (any, bool) {
	node := map[string]any(cv)
	for i, key := range path {
		val, ok := node[key]
		if !ok {
			return nil, false
		}
		if i == len(path)-1 {
			return val, true
		}
		node, ok = val.(map[string]any)
		if !ok {
			return nil, false
		}
	}
	return nil, false
}

// set sets the given value at the given path of this ContextedValue,
// creating the intermediate nodes if necessary.
func (cv ContextedValue) set(path []string, value any) {
	node := map[string]any(cv)
	for _, key := range path[:len(path)-1] {
		next, ok := node[key].(map[string]any)
		if !ok {
			next = make(map[string]any)
			node[key] = next
		}
		node = next
	}
	node[path[len(path)-1]] = value
}

// resolve returns the value of this ContextedValue for the given context
// values, falling back to less specific values and finally to the default
// value if it is not set for the given contexts.
//
// Second returned value is false if no value at all could be found.
func (cv ContextedValue) resolve(ctxNames []string, ctxValues map[string]string) (any, bool) {
	for _, path := range contextedPaths(ctxNames, ctxValues) {
		if val, ok := cv.get(path); ok && val != nil {
			return val, true
		}
	}
	return nil, false
}

// contextNames returns the sorted list of the context names of this field.
func (f *Field) contextNames() []string {
	res := make([]string, 0, len(f.contexts))
	for name := range f.contexts {
		res = append(res, name)
	}
	sort.Strings(res)
	return res
}

// contextedPaths returns all the paths at which a value may be found in a
// ContextedValue document, from the most specific to the most generic.
//
// ctxNames is the list of the context names defined on the field and ctxValues
// holds the value of each context in the current environment. Contexts that
// have no value in ctxValues are not taken into account.
//
// Context names are sorted alphabetically and the last one takes precedence
// over the others. Each returned path ends with ctxValueKey so that it points
// directly at a value. The last returned path is always the default one, i.e.
// the value that does not depend on any context.
func contextedPaths(ctxNames []string, ctxValues map[string]string) [][]string {
	var names []string
	for _, name := range ctxNames {
		if ctxValues[name] == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	res := make([][]string, 0, 1<<len(names))
	for weight := 1<<len(names) - 1; weight >= 0; weight-- {
		var path []string
		for i, name := range names {
			if weight&(1<<i) == 0 {
				continue
			}
			path = append(path, name, ctxValues[name])
		}
		res = append(res, append(path, ctxValueKey))
	}
	return res
}

// cacheFieldKey returns the key under which the value of the given field must
// be stored in the cache for the given context slug.
//
// Only contexted fields are keyed by context, since the value of the other
// fields does not depend on the environment's context.
func cacheFieldKey(mi *Model, jsonName, ctxSlug string) string {
	if ctxSlug == "" {
		return jsonName
	}
	fi, ok := mi.fields.Get(jsonName)
	if !ok || !fi.isContextedField() {
		return jsonName
	}
	return jsonName + ContextSep + ctxSlug
}

// contextedLeaf is a value of a contexted value document together with the
// context in which it is set.
type contextedLeaf struct {
	contexts map[string]string
	value    any
}

// contextedLeaves returns all the values of the given contexted value document
// node, together with the context in which each of them is set. contexts is
// the context of the given node.
func contextedLeaves(node map[string]any, contexts map[string]string) []contextedLeaf {
	var res []contextedLeaf
	for key, val := range node {
		if key == ctxValueKey {
			res = append(res, contextedLeaf{contexts: contexts, value: val})
			continue
		}
		ctxValues, ok := val.(map[string]any)
		if !ok {
			continue
		}
		for ctxValue, child := range ctxValues {
			childNode, ok := child.(map[string]any)
			if !ok {
				continue
			}
			newContexts := make(map[string]string, len(contexts)+1)
			maps.Copy(newContexts, contexts)
			newContexts[key] = ctxValue
			res = append(res, contextedLeaves(childNode, newContexts)...)
		}
	}
	return res
}
