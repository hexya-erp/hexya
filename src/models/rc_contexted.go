// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

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
	rc.checkContextedUniqueValues(fi, values)
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

// checkContextedUnique panics if another record than those of this
// RecordCollection already has, in the current context, one of the values of
// the given FieldMap for a unique contexted field.
//
// Uniqueness of contexted fields is checked here instead of by a SQL
// constraint, since all the values of a contexted field are stored in a single
// JSON document. As a consequence, it is not race safe: two concurrent
// transactions may both pass this check and commit duplicate values.
func (rc *RecordCollection) checkContextedUnique(fMap FieldMap) {
	for _, fi := range rc.model.fields.uniqueCtxFields {
		value, ok := fMap.Get(rc.model.FieldName(fi.name))
		if !ok {
			continue
		}
		rc.checkContextedUniqueValue(fi, value)
	}
}

// checkContextedUniqueValue panics if another record than those of this
// RecordCollection already has the given value for the given unique contexted
// field in the current context.
func (rc *RecordCollection) checkContextedUniqueValue(fi *Field, value any) {
	if rc.hasNegIds || value == nil || reflect.ValueOf(value).IsZero() {
		return
	}
	fName := rc.model.FieldName(fi.name)
	if len(rc.Ids()) > 1 {
		// All the records of this RecordCollection are given the same value.
		log.Panic(fmt.Sprintf("%s must be unique", fi.name), "model", rc.model.name,
			"field", fi.name, "value", value, "ids", rc.Ids())
	}
	cond := rc.Model().Field(fName).Equals(value)
	if len(rc.Ids()) > 0 {
		cond = cond.AndNot().Field(ID).In(rc.Ids())
	}
	if !rc.model.Search(*rc.env, cond).Sudo().IsEmpty() {
		log.Panic(fmt.Sprintf("%s must be unique", fi.name), "model", rc.model.name,
			"field", fi.name, "value", value)
	}
}

// checkContextedUniqueValues panics if another record than those of this
// RecordCollection already has, in one of the contexts of the given contexted
// value document, the value of this document for the given unique contexted
// field.
func (rc *RecordCollection) checkContextedUniqueValues(fi *Field, values ContextedValue) {
	if !fi.unique {
		return
	}
	for _, leaf := range contextedLeaves(values, nil) {
		rSet := rc
		for ctxName, ctxValue := range leaf.contexts {
			rSet = rSet.WithContext(ctxName, ctxValue)
		}
		rSet.checkContextedUniqueValue(fi, leaf.value)
	}
}
