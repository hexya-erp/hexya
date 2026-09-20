// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

// contextValues returns the value of each context of this field evaluated
// against the given RecordSet.
//
// Contexts which evaluate to an empty string are not returned. An empty map
// is returned if the "hexya_default_contexts" key is set in the environment's
// context, so that the default values of contexted fields are used.
func (f *Field) contextValues(rs RecordSet) map[string]string {
	res := make(map[string]string)
	if rs.Env().Context().GetBool("hexya_default_contexts") {
		return res
	}
	for name, ctxFunc := range f.contexts {
		val := ctxFunc(rs)
		if val == "" {
			continue
		}
		res[name] = val
	}
	return res
}

// contextValues returns the value of all the contexts of all the contexted
// fields of this RecordCollection's model in the current environment.
//
// Related fields are taken into account through the field they point at, so
// that the value of a related contexted field is cached by context too.
func (rc *RecordCollection) contextValues() map[string]string {
	res := make(map[string]string)
	for _, fi := range rc.model.fields.registryByJSON {
		target := fi
		if fi.isRelatedField() {
			target = rc.model.getRelatedFieldInfo(fi.relatedPath)
		}
		if !target.isContextedField() {
			continue
		}
		for name, val := range target.contextValues(rc) {
			res[name] = val
		}
	}
	return res
}

// ctxSlug returns a unique string identifying the context values that apply
// to the contexted fields of this RecordCollection's model.
//
// It returns an empty string if this model has no contexted field or if none
// of the contexts has a value in this environment.
func (rc *RecordCollection) ctxSlug() string {
	ctxValues := rc.contextValues()
	if len(ctxValues) == 0 {
		return ""
	}
	names := make([]string, 0, len(ctxValues))
	for name := range ctxValues {
		names = append(names, name)
	}
	sort.Strings(names)
	res := make([]string, len(names))
	for i, name := range names {
		res[i] = fmt.Sprintf("%s=%s", name, ctxValues[name])
	}
	return strings.Join(res, ContextSep)
}

// contextedFieldInfo returns the Field of the given contexted field of this
// model. It panics if the field is not a contexted field of this model.
func (rc *RecordCollection) contextedFieldInfo(field FieldName) *Field {
	fi := rc.model.getRelatedFieldInfo(field)
	if !fi.isContextedField() {
		log.Panic("Field is not a contexted field", "model", rc.model.name, "field", field.Name())
	}
	if fi.model != rc.model {
		log.Panic("Contexted values can only be accessed on the field's own model",
			"model", rc.model.name, "field", field.Name())
	}
	return fi
}

// GetContextedValues returns the values of the given contexted field of the
// first record of this RecordCollection for all the contexts it is set for.
//
// It panics if the given field is not a contexted field of this model.
func (rc *RecordCollection) GetContextedValues(field FieldName) ContextedValue {
	rc.EnsureOne()
	fi := rc.contextedFieldInfo(field)
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Load"))
	adapter := adapters[db.DriverName()]
	var doc []byte
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE id = ?`,
		fi.json, adapter.quoteTableName(rc.model.tableName))
	rc.env.cr.Get(&doc, query, rc.ids[0])
	res := make(ContextedValue)
	if len(doc) == 0 {
		return res
	}
	if err := json.Unmarshal(doc, &res); err != nil {
		log.Panic("Unable to unmarshal contexted value", "model", rc.model.name,
			"field", field.Name(), "error", err)
	}
	return res
}

// SetContextedValues sets all the contexted values of the given field of the
// records of this RecordCollection at once, overriding any existing value.
//
// It panics if the given field is not a contexted field of this model.
func (rc *RecordCollection) SetContextedValues(field FieldName, values ContextedValue) {
	fi := rc.contextedFieldInfo(field)
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Write"))
	if rc.IsEmpty() {
		return
	}
	doc, err := json.Marshal(values)
	if err != nil {
		log.Panic("Unable to marshal contexted value", "model", rc.model.name,
			"field", field.Name(), "error", err)
	}
	adapter := adapters[db.DriverName()]
	query := fmt.Sprintf(`UPDATE %s SET %s = ?::jsonb WHERE id IN (?)`,
		adapter.quoteTableName(rc.model.tableName), fi.json)
	rc.env.cr.Execute(query, string(doc), rc.Ids())
	for _, id := range rc.Ids() {
		rc.env.cache.removeContextedEntries(rc.model, id, fi.json, "")
	}
}

// GetTranslations returns the translations of the given contexted field of the
// first record of this RecordCollection, as a map of language to value.
//
// The value that does not depend on any language is returned under the empty
// string key.
func (rc *RecordCollection) GetTranslations(field FieldName) map[string]string {
	res := make(map[string]string)
	cv := rc.GetContextedValues(field)
	if val, ok := cv.get([]string{ctxValueKey}); ok && val != nil {
		res[""] = fmt.Sprintf("%v", val)
	}
	langs, _ := cv["lang"].(map[string]any)
	for lang := range langs {
		val, ok := cv.get([]string{"lang", lang, ctxValueKey})
		if !ok || val == nil {
			continue
		}
		res[lang] = fmt.Sprintf("%v", val)
	}
	return res
}

// SetTranslations sets the translations of the given contexted field of the
// records of this RecordCollection from the given map of language to value.
//
// The value under the empty string key is set as the value that does not
// depend on any language. The translations of the languages that are not in
// the given map are left unchanged, as well as the other contexts.
func (rc *RecordCollection) SetTranslations(field FieldName, values map[string]string) {
	for _, rec := range rc.Records() {
		cv := rec.GetContextedValues(field)
		for lang, val := range values {
			path := []string{"lang", lang, ctxValueKey}
			if lang == "" {
				path = []string{ctxValueKey}
			}
			cv.set(path, val)
		}
		rec.SetContextedValues(field, cv)
	}
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
