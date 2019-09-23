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

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/jmoiron/sqlx"
)

// RecordCollection is a generic struct representing several
// records of a model.
type RecordCollection struct {
	model      *Model
	query      *Query
	env        *Environment
	prefetchRC *RecordCollection
	ids        []int64
	fetched    bool
	filtered   bool
}

// String returns the string representation of a RecordSet
func (rc *RecordCollection) String() string {
	idsStr := make([]string, len(rc.Ids()))
	for i, id := range rc.Ids() {
		idsStr[i] = strconv.Itoa(int(id))
	}
	rsIds := strings.Join(idsStr, ",")
	return fmt.Sprintf("%s(%s)", rc.model.name, rsIds)
}

// Env returns the RecordSet's Environment
func (rc *RecordCollection) Env() Environment {
	res := *rc.env
	return res
}

// ModelName returns the model name of the RecordSet
func (rc *RecordCollection) ModelName() string {
	return rc.model.name
}

// Ids returns the ids of the RecordSet, fetching from db if necessary.
func (rc *RecordCollection) Ids() []int64 {
	rc.Fetch()
	return rc.ids
}

// clone returns a pointer to a new RecordCollection identical to this one.
func (rc *RecordCollection) clone() *RecordCollection {
	rSet := *rc
	res := &rSet
	res.query = rc.query.clone(res)
	return res
}

// create inserts a new record in the database with the given data.
// data can be either a FieldMap or a struct pointer of the same model as rs.
// This function is private and low level. It should not be called directly.
// Instead use rs.Call("Create")
func (rc *RecordCollection) create(data RecordData) *RecordCollection {
	defer func() {
		if r := recover(); r != nil {
			panic(rc.substituteSQLErrorMessage(r))
		}
	}()
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Create"))
	// process create data for FK relations if any
	data = rc.createFKRelationRecords(data)

	newData := data.Underlying().Copy()
	rc.applyDefaults(newData, true)
	fMap := newData.Underlying().FieldMap
	rc.applyContexts()
	rc.addAccessFieldsCreateData(&fMap)
	fMap = rc.addEmbeddedfields(fMap)
	rc.model.convertValuesToFieldType(&fMap, true)
	fMap = rc.addContextsFieldsValues(fMap)
	// clean our fMap from ID and non stored fields
	fMap.RemovePKIfZero()
	storedFieldMap := filterMapOnStoredFields(rc.model, fMap)
	// insert in DB
	var createdId int64
	sql, args := rc.query.insertQuery(storedFieldMap)
	rc.env.cr.Get(&createdId, sql, args...)

	rc.env.cache.addRecord(rc.model, createdId, storedFieldMap, rc.query.ctxArgsSlug())
	rSet := rc.withIds([]int64{createdId})
	// update reverse relation fields
	rSet.updateRelationFields(fMap)
	// update related fields
	rSet.updateRelatedFields(fMap)
	// process create data for reverse relations if any
	rSet.createReverseRelationRecords(data)
	// compute stored fields
	rSet.processInverseMethods(data)
	rSet.processTriggers(fMap.Keys())
	rSet.CheckConstraints()
	return rSet
}

// createReverseRelationRecords creates the reverse records of relation fields when
// the given data contains such directive.
func (rc *RecordCollection) createReverseRelationRecords(data RecordData) {
	for f, dd := range data.Underlying().ToCreate {
		fi := rc.model.getRelatedFieldInfo(f)
		if !fi.fieldType.IsReverseRelationType() {
			continue
		}
		for _, d := range dd {
			rc.createRelatedRecord(f, d)
		}
	}
}

// createFKRelationRecords creates the FK records of relation fields when
// the given data contains such directive.
func (rc *RecordCollection) createFKRelationRecords(data RecordData) *ModelData {
	res := data.Underlying().Copy()
	for f, dd := range data.Underlying().ToCreate {
		fi := rc.model.getRelatedFieldInfo(f)
		if !fi.fieldType.IsFKRelationType() && fi.fieldType != fieldtype.Many2Many {
			continue
		}
		relRS := rc.env.Pool(fi.relatedModelName)
		for _, d := range dd {
			created := rc.createRelatedFKRecord(fi, d)
			if fi.fieldType == fieldtype.Many2Many {
				relRS = relRS.Union(created)
				continue
			}
			relRS = created
		}
		res.Set(f, relRS)
		delete(res.ToCreate, f)
	}
	return res
}

// addEmbeddedFields adds FK fields of embedded records into the fMap so that
// they will be automatically created during related field update.
func (rc *RecordCollection) addEmbeddedfields(fMap FieldMap) FieldMap {
	for fName, fi := range rc.model.fields.registryByName {
		if !fi.embed {
			continue
		}
		if _, ok := fMap[fi.json]; ok {
			continue
		}
		fMap[fmt.Sprintf("%s%sID", fName, ExprSep)] = nil
	}
	return fMap
}

// applyDefaults adds the default values to the given ModelData values which
// are not already set.
//
// If create is true, default values are not set for computed fields.
func (rc *RecordCollection) applyDefaults(md *ModelData, create bool) {
	defaults := rc.WithContext("hexya_ignore_computed_defaults", create).Call("DefaultGet").(RecordData).Underlying()
	defaults.MergeWith(md)
	*md = *defaults
}

// getDefaults returns the default values for a new record, taking into account
// the context and fields default functions.
//
// If create is true, default values are not given for computed fields.
func (rc *RecordCollection) getDefaults(create bool) *ModelData {
	md := NewModelData(rc.model)

	// 1. Create a map with default values from context
	ctxDefaults := make(FieldMap)
	for ctxKey, ctxVal := range rc.env.context.ToMap() {
		if !strings.HasPrefix(ctxKey, "default_") {
			continue
		}
		fJSON := strings.TrimPrefix(ctxKey, "default_")
		if _, exists := rc.model.fields.Get(fJSON); !exists {
			continue
		}
		ctxDefaults[fJSON] = ctxVal
	}

	// 2. Apply defaults from context (if exists) or default function
	for fName, fi := range Registry.MustGet(rc.ModelName()).fields.registryByJSON {
		if !fi.isSettable() {
			continue
		}
		if create && (fi.isComputedField() || (fi.isRelatedField() && !fi.isContextedField())) {
			continue
		}
		if val, exists := ctxDefaults[fi.json]; exists {
			md.Set(fName, val)
			continue
		}
		if fi.defaultFunc == nil {
			continue
		}
		md.Set(fName, fi.defaultFunc(rc.Env()))
	}
	return md
}

// applyContexts adds filtering on contexts when applicable to this RecordSet query.
func (rc *RecordCollection) applyContexts() *RecordCollection {
	ctxCond := newCondition()
	var ctxOrders []string
	for _, fi := range rc.model.fields.registryByName {
		if fi.contexts == nil {
			continue
		}
		for ctxName, ctxFunc := range fi.contexts {
			path := fmt.Sprintf("%sHexyaContexts%s%s", fi.name, ExprSep, ctxName)
			ctxOrders = append(ctxOrders, fmt.Sprintf("%s DESC", path))
			cond := rc.model.Field(path).IsNull()
			if !rc.env.context.GetBool("hexya_default_contexts") {
				cond = cond.Or().Field(path).Equals(ctxFunc)
			}
			ctxCond = ctxCond.AndCond(cond)
		}
	}
	rc.query.ctxCond = ctxCond
	rc.query.ctxOrders = ctxOrders
	return rc
}

// addContextsFieldsValues adds the contexts to the given fMap so that the resulting set can be filtered
func (rc *RecordCollection) addContextsFieldsValues(fMap FieldMap) FieldMap {
	res := make(FieldMap)
	for k, v := range fMap {
		res[k] = v
		fi := rc.model.getRelatedFieldInfo(k)
		if fi.contexts != nil {
			for ctxName, ctxFunc := range fi.contexts {
				path := fmt.Sprintf("%sHexyaContexts.%s", fi.name, ctxName)
				res[path] = ctxFunc(rc)
			}
		}
	}
	return res
}

// CheckConstraints executes the constraint method for each field defined
// in the given fMap with the corresponding value.
// Each method is only executed once, even if it is called by several fields.
// It panics as soon as one constraint fails.
func (rc *RecordCollection) CheckConstraints() {
	if rc.env.context.GetBool("hexya_skip_check_constraints") {
		return
	}
	methods := make(map[string]bool)
	for _, fi := range rc.model.fields.registryByJSON {
		if fi.constraint != "" {
			methods[fi.constraint] = true
		}
	}
	if len(methods) == 0 {
		return
	}
	for method := range methods {
		for _, rec := range rc.Records() {
			rec.Call(method)
		}
	}
}

// addAccessFieldsCreateData adds appropriate CreateDate and CreateUID fields to
// the given FieldMap.
func (rc *RecordCollection) addAccessFieldsCreateData(fMap *FieldMap) {
	if !rc.model.isSystem() {
		(*fMap)["CreateDate"] = dates.Now()
		(*fMap)["CreateUID"] = rc.env.uid
	}
}

// update updates the database with the given data and returns the number of updated rows.
// It panics in case of error.
// It returns without changes if rc is empty
// This function is private and low level. It should not be called directly.
// Instead use rs.Call("Write")
func (rc *RecordCollection) update(data RecordData) bool {
	if rc.ForceLoad("ID").IsEmpty() {
		return true
	}
	rSet := rc.addRecordRuleConditions(rc.env.uid, security.Write)
	// process create data for FK relations if any
	data = rc.createFKRelationRecords(data)
	fMap := data.Underlying().Copy().FieldMap
	rSet.addAccessFieldsUpdateData(&fMap)
	rSet.applyContexts()
	fMap = rSet.addContextsFieldsValues(fMap)
	// We process inverse method before we convert RecordSets to ids
	rSet.processInverseMethods(data)
	rSet.model.convertValuesToFieldType(&fMap, true)
	// clean our fMap from ID and non stored fields
	fMap.RemovePK()
	storedFieldMap := filterMapOnStoredFields(rSet.model, fMap)
	rSet.doUpdate(storedFieldMap)
	// Let's fetch once for all
	rSet.Fetch()
	// write reverse relation fields
	rSet.updateRelationFields(fMap)
	// write related fields
	rSet.updateRelatedFields(fMap)
	// process create data for reverse relations if any
	rSet.createReverseRelationRecords(data)
	// compute stored fields
	rSet.processTriggers(fMap.Keys())
	rSet.CheckConstraints()
	return true
}

// addAccessFieldsUpdateData adds appropriate WriteDate and WriteUID fields to
// the given FieldMap.
func (rc *RecordCollection) addAccessFieldsUpdateData(fMap *FieldMap) {
	if !rc.model.isSystem() {
		(*fMap)["WriteDate"] = dates.Now()
		(*fMap)["WriteUID"] = rc.env.uid
	}
}

// doUpdate just updates the database records pointed at by
// this RecordCollection with the given fieldMap. It also
// updates the cache for the record
func (rc *RecordCollection) doUpdate(fMap FieldMap) {
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Write"))
	if rc.IsEmpty() {
		log.Panic("Trying to update an empty RecordSet", "model", rc.ModelName(), "values", fMap)
	}
	defer func() {
		if r := recover(); r != nil {
			panic(rc.substituteSQLErrorMessage(r))
		}
	}()
	// update DB
	if len(fMap) == 2 {
		_, okWD := fMap["write_date"]
		_, okWU := fMap["write_uid"]
		if okWD && okWU {
			// We only have write_date and write_uid to update, so we ignore
			return
		}
	}
	sql, args := rc.query.updateQuery(fMap)
	res := rc.env.cr.Execute(sql, args...)
	if num, _ := res.RowsAffected(); num == 0 {
		log.Panic("Unexpected noop on update (num = 0)", "model", rc.ModelName(), "values", fMap, "query", sql, "args", args)
	}
	for _, rec := range rc.Records() {
		for k, v := range fMap {
			rc.env.cache.updateEntry(rc.model, rec.Ids()[0], k, v, rc.query.ctxArgsSlug())
		}
	}
}

// updateRelationFields updates reverse relations fields of the
// given fMap.
func (rc *RecordCollection) updateRelationFields(fMap FieldMap) {
	rc.Fetch()
	for field, value := range fMap {
		fi := rc.model.getRelatedFieldInfo(field)
		switch fi.fieldType {
		case fieldtype.One2Many:
			// We take only the first record since updating all records
			// will override each other
			if rc.Len() > 1 {
				log.Warn("Updating one2many relation on multiple record at once", "model", rc.ModelName(), "field", field)
			}
			curRS := rc.env.Pool(fi.relatedModelName).Search(fi.relatedModel.Field("ID").In(rc.Get(fi.name).(RecordSet).Collection()))
			newRS := rc.env.Pool(fi.relatedModelName).Search(fi.relatedModel.Field("ID").In(value.([]int64)))
			// Remove ReverseFK for Records that are no longer our children
			toRemove := curRS.Subtract(newRS)
			if toRemove.Len() > 0 {
				toRemove.Set(fi.reverseFK, nil)
			}
			// Add new children records
			toAdd := newRS.Subtract(curRS)
			if toAdd.Len() > 0 {
				toAdd.Set(fi.reverseFK, rc.Records()[0])
			}

		case fieldtype.Rev2One:
		case fieldtype.Many2Many:
			delQuery := fmt.Sprintf(`DELETE FROM %s WHERE %s IN (?)`, fi.m2mRelModel.tableName, fi.m2mOurField.json)
			rc.env.cr.Execute(delQuery, rc.ids)
			for _, id := range rc.ids {
				rc.env.cache.removeM2MLinks(fi, id)
				query := fmt.Sprintf(`INSERT INTO %s (%s, %s) VALUES (?, ?)`, fi.m2mRelModel.tableName,
					fi.m2mOurField.json, fi.m2mTheirField.json)
				for _, relId := range value.([]int64) {
					rc.env.cr.Execute(query, id, relId)
				}
				rc.env.cache.addM2MLink(fi, id, value.([]int64))
			}
		}
	}
}

// updateRelatedFields updates related fields of the given fMap.
func (rc *RecordCollection) updateRelatedFields(fMap FieldMap) {
	type rsRef struct {
		model *Model
		id    int64
	}

	rc.Fetch()
	fMap = rc.substituteRelatedFieldsInMap(fMap)
	fields := rc.addIntermediatePaths(fMap.Keys())
	rc.loadRelatedRecords(fields)

	// Ordered fields will show shorter paths before longer paths
	sort.Strings(fields)
	// Create an update map for each record to update
	updateMap := make(map[rsRef]FieldMap)
	for _, rec := range rc.Records() {
		createdPaths := make(map[string]bool)
		// Create related records
		for _, path := range fields {
			vals, prefix := rec.relatedRecordMap(fMap, path)
			if createdPaths[prefix] {
				continue
			}
			model, id, _, err := rec.env.cache.getStrictRelatedRef(rec.model, rec.ids[0], path, rc.query.ctxArgsSlug())
			if err != nil {
				// Record does not exist, we create it on the fly instead of updating
				fp := rec.model.getRelatedFieldInfo(prefix)
				nr := rec.createRelatedRecord(prefix, NewModelDataFromRS(rec.env.Pool(fp.relatedModelName), vals))
				rec.env.cache.setX2MValue(rec.model.name, rec.ids[0], prefix, nr.Ids()[0], rc.query.ctxArgsSlug())
				createdPaths[prefix] = true
				continue
			}
			if len(vals) == 0 {
				continue
			}
			updateMap[rsRef{model, id}] = vals
		}
	}

	// Create default value for contexted field if we do not have one yet
	rc.loadRelatedRecords(fields)
	for _, rec := range rc.Records() {
		for _, path := range fields {
			if _, _, _, err := rec.env.cache.getStrictRelatedRef(rec.model, rec.ids[0], path, ""); err == nil {
				continue
			}
			// We have no default value
			fi := rec.model.getRelatedFieldInfo(path)
			if fi.ctxType != ctxValue {
				continue
			}
			// This is a contexted field and we have no default value so we create it
			vals, prefix := rc.relatedRecordMap(fMap, path)
			//
			field := strings.TrimPrefix(path, prefix+ExprSep)
			defVals := FieldMap{
				"record_id": vals["record_id"],
				field:       vals[field],
			}
			fp := rc.model.getRelatedFieldInfo(prefix)
			nr := rc.createRelatedRecord(prefix, NewModelDataFromRS(rc.env.Pool(fp.relatedModelName), defVals))
			rc.env.cache.setX2MValue(rc.model.name, rc.ids[0], prefix, nr.Ids()[0], "")
		}
	}

	// Make the update for each record
	for ref, upMap := range updateMap {
		rs := rc.env.Pool(ref.model.name).withIds([]int64{ref.id})
		rs.Call("Write", NewModelDataFromRS(rs, upMap))
	}
}

// loadRelatedRecords loads all records pointed at by the given fMap keys
// relatively to this record collection if they are not already in cache.
//
// This method also loads default values for contexted fields
func (rc *RecordCollection) loadRelatedRecords(fields []string) {
	var toLoad []string

	// load contexted fields default value
	if rc.query.ctxArgsSlug() != "" {
		for _, field := range fields {
			exprs := strings.Split(field, ExprSep)
			if len(exprs) <= 1 {
				continue
			}
			if !rc.env.cache.checkIfInCache(rc.model, rc.ids, []string{field}, "", true) {
				toLoad = append(toLoad, field)
			}
		}
		if len(toLoad) > 0 {
			rc.WithContext("hexya_default_contexts", true).Load(toLoad...)
		}
	}
	// Load contexted fields with rc's context
	toLoad = []string{}
	for _, field := range fields {
		exprs := strings.Split(field, ExprSep)
		if len(exprs) <= 1 {
			continue
		}
		if !rc.env.cache.checkIfInCache(rc.model, rc.ids, []string{field}, rc.query.ctxArgsSlug(), true) {
			toLoad = append(toLoad, field)
		}
	}
	// Load related paths if not loaded already
	if len(toLoad) > 0 {
		rc.Call("Load", toLoad)
	}
}

// relatedRecordMap return a ModelData to create or update the related record
// defined by path from this RecordCollection, using fMap values. The second record
// value is the path to the related record.
//
// field must be a field of the related record (and not an M2O field pointing to it)
func (rc *RecordCollection) relatedRecordMap(fMap FieldMap, field string) (FieldMap, string) {
	exprs := strings.Split(field, ExprSep)
	prefix := strings.Join(exprs[:len(exprs)-1], ExprSep)
	res := make(FieldMap)
	for f, v := range fMap {
		fExprs := strings.Split(f, ExprSep)
		if strings.HasPrefix(f, prefix+ExprSep) {
			key := strings.Join(fExprs[len(exprs)-1:], ExprSep)
			res[key] = v
		}
	}
	return res, prefix
}

// substituteSQLErrorMessage changes the message from the given recover data
// if it comes from the database with the message defined in this model
func (rc *RecordCollection) substituteSQLErrorMessage(r interface{}) interface{} {
	err, ok := r.(error)
	if !ok {
		return r
	}
	for constraintName, constraint := range rc.model.sqlConstraints {
		if strings.Contains(err.Error(), constraintName) {
			res := adapters[db.DriverName()].substituteErrorMessage(err, constraint.errorString)
			return res
		}
	}
	return r
}

// unlink deletes the database record of this RecordSet and returns the number of deleted rows.
// This function is private and low level. It should not be called directly.
// Instead use rs.Unlink() or rs.Call("Unlink")
func (rc *RecordCollection) unlink() int64 {
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Unlink"))
	rSet := rc.addRecordRuleConditions(rc.env.uid, security.Unlink)
	ids := rSet.Ids()
	if rSet.IsEmpty() {
		return 0
	}
	// get recomputate data to update after unlinking
	compData := rc.retrieveComputeData(rc.model.fields.allJSONNames())
	sql, args := rSet.query.deleteQuery()
	res := rSet.env.cr.Execute(sql, args...)
	num, _ := res.RowsAffected()
	for _, id := range ids {
		rc.env.cache.invalidateRecord(rc.model, id)
	}
	// Update stored fields that referenced this recordset
	rc.updateStoredFields(compData)
	return num
}

// Search returns a new RecordSet filtering on the current one with the
// additional given Condition
func (rc *RecordCollection) Search(cond *Condition) *RecordCollection {
	rSetVal := *rc
	rSetVal.query = rc.query.clone(&rSetVal)
	rSetVal.query.cond = rSetVal.query.cond.AndCond(cond)
	return &rSetVal
}

// Limit returns a new RecordSet with only the first 'limit' records.
func (rc *RecordCollection) Limit(limit int) *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone(&rSet)
	rSet.query.limit = limit
	return &rSet
}

// Offset returns a new RecordSet with only the records starting at offset
func (rc *RecordCollection) Offset(offset int) *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone(&rSet)
	rSet.query.offset = offset
	return &rSet
}

// OrderBy returns a new RecordSet ordered by the given ORDER BY expressions
func (rc *RecordCollection) OrderBy(exprs ...string) *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone(&rSet)
	rSet.query.orders = append(rSet.query.orders, exprs...)
	return &rSet
}

// GroupBy returns a new RecordSet grouped with the given GROUP BY expressions
func (rc *RecordCollection) GroupBy(fields ...FieldNamer) *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone(&rSet)
	exprs := make([]string, len(fields))
	for i, f := range fields {
		exprs[i] = string(f.FieldName())
	}
	rSet.query.groups = append(rSet.query.groups, exprs...)
	return &rSet
}

// Fetch query the database with the current filter and returns a RecordSet
// with the queries ids.
//
// Fetch is lazy and only return ids. Use Load() instead
// if you want to fetch all fields.
func (rc *RecordCollection) Fetch() *RecordCollection {
	if rc.fetched {
		return rc
	}
	if rc.query.isEmpty() {
		// We do not load empty queries to keep empty record sets empty
		// Call SearchAll instead to load all the records of the table
		return rc
	}
	return rc.Load("id")
}

// SearchAll returns a new RecordSet with all items of the table, regardless of the
// current RecordSet query. It is mainly meant to be used on an empty RecordSet
func (rc *RecordCollection) SearchAll() *RecordCollection {
	rSet := rc.env.Pool(rc.ModelName())
	rSet.query.fetchAll = true
	return rSet
}

// SearchCount fetch from the database the number of records that match the RecordSet conditions
// It panics in case of error
func (rc *RecordCollection) SearchCount() int {
	rSet := rc.Limit(0)
	rSet.applyDefaultOrder()
	rSet.applyContexts()
	addNameSearchesToCondition(rSet.model, rSet.query.cond)
	rSet = rSet.substituteRelatedInQuery()
	sql, args := rSet.query.countQuery()
	var res int
	rSet.env.cr.Get(&res, sql, args...)
	return res
}

// Load look up fields of the RecordCollection in cache and query the database
// for missing values which are then stored in cache.
func (rc *RecordCollection) Load(fields ...string) *RecordCollection {
	if len(fields) == 0 {
		fields = rc.model.fields.storedFieldNames()
	}
	if rc.env.cache.checkIfInCache(rc.model, rc.ids, fields, rc.query.ctxArgsSlug(), true) {
		return rc
	}
	return rc.ForceLoad(fields...)
}

// ForceLoad query all data of the RecordCollection and store in cache.
// fields are the fields to retrieve in the path format,
// i.e. "User.Profile.Age" or "user_id.profile_id.age".
//
// If no fields are given, all DB columns of the RecordCollection's
// model are retrieved as well as related fields. Non-DB fields must
// be explicitly given in fields to be retrieved.
func (rc *RecordCollection) ForceLoad(fieldNames ...string) *RecordCollection {
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Load"))
	if rc.query.isEmpty() {
		// Never load RecordSets without query.
		return rc
	}
	if len(rc.query.groups) > 0 {
		log.Panic("Trying to load a grouped query", "model", rc.model, "groups", rc.query.groups)
	}
	rSet := rc
	var prefetch bool
	if !rc.prefetchRC.IsEmpty() && len(rc.ids) > 0 {
		// We have a prefetch recordSet and our ids are already fetched
		prefetch = true
		rSet = rc.Union(rc.prefetchRC).WithEnv(rc.Env())
	}
	rSet = rSet.addRecordRuleConditions(rc.env.uid, security.Read)
	rSet.applyDefaultOrder()

	fields := make([]string, len(fieldNames))
	copy(fields, fieldNames)
	if len(fields) == 0 {
		fields = rSet.model.fields.storedFieldNames()
	}
	addNameSearchesToCondition(rSet.model, rSet.query.cond)
	rSet.applyContexts()
	subFields, _ := rSet.substituteRelatedFields(fields)
	rSet = rSet.substituteRelatedInQuery()
	dbFields := filterOnDBFields(rSet.model, subFields)
	sql, args, substs := rSet.query.selectQuery(dbFields)
	rows := dbQuery(rSet.env.cr.tx, sql, args...)
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		line := make(FieldMap)
		err := rSet.model.scanToFieldMap(rows, &line, substs)
		if err != nil {
			log.Panic(err.Error(), "model", rSet.ModelName(), "fields", fields)
		}
		rSet.env.cache.addRecord(rSet.model, line["id"].(int64), line, rc.query.ctxArgsSlug())
		ids = append(ids, line["id"].(int64))
	}

	rSet = rSet.withIds(ids)
	rSet.loadRelationFields(subFields)
	if prefetch {
		*rc = *rSet.Intersect(rc).WithEnv(rc.Env())
		return rc
	}
	return rSet
}

// applyDefaultOrder adds the model's default order if this query has no specific order defined
func (rc *RecordCollection) applyDefaultOrder() {
	if len(rc.query.orders) == 0 {
		rc.query.orders = make([]string, len(rc.model.defaultOrder))
		copy(rc.query.orders, rc.model.defaultOrder)
	}
}

// loadRelationFields loads one2many, many2many and rev2one fields from the given fields
// names in this RecordCollection into the cache. fields of other types given in fields
// are ignored.
func (rc *RecordCollection) loadRelationFields(fields []string) {
	if len(fields) == 0 {
		return
	}
	sort.Strings(fields)

	for _, rec := range rc.Records() {
		id := rec.ids[0]
		for _, fieldName := range fields {
			fi := rc.model.getRelatedFieldInfo(fieldName)
			if !fi.fieldType.IsNonStoredRelationType() {
				continue
			}
			thisRC := rec
			exprs := strings.Split(fieldName, ExprSep)
			if len(exprs) > 1 {
				prefix := strings.Join(exprs[:len(exprs)-1], ExprSep)
				// We do not call "Load" directly to have caller method properly set
				thisRC.Call("Load", []string{prefix})
				thisRC = thisRC.Get(prefix).(RecordSet).Collection()
			}
			switch fi.fieldType {
			case fieldtype.One2Many:
				relRC := rc.env.Pool(fi.relatedModelName)
				// We do not call "Fetch" directly to have caller method properly set
				relRC = relRC.Search(relRC.Model().Field(fi.reverseFK).Equals(thisRC)).Call("Fetch").(RecordSet).Collection()
				rc.env.cache.updateEntry(rc.model, id, fieldName, relRC.ids, rc.query.ctxArgsSlug())
			case fieldtype.Many2Many:
				query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`, fi.m2mTheirField.json,
					fi.m2mRelModel.tableName, fi.m2mOurField.json)
				var ids []int64
				if thisRC.IsEmpty() {
					continue
				}
				rc.env.cr.Select(&ids, query, thisRC.ids[0])
				rc.env.cache.updateEntry(rc.model, id, fieldName, ids, rc.query.ctxArgsSlug())
			case fieldtype.Rev2One:
				relRC := rc.env.Pool(fi.relatedModelName)
				// We do not call "Fetch" directly to have caller method properly set
				relRC = relRC.Search(relRC.Model().Field(fi.reverseFK).Equals(thisRC)).Call("Fetch").(RecordSet).Collection()
				var relID int64
				if len(relRC.ids) > 0 {
					relID = relRC.ids[0]
				}
				rc.env.cache.updateEntry(rc.model, id, fieldName, relID, rc.query.ctxArgsSlug())
			}
		}
	}
}

// Get returns the value of the given fieldName for the first record of this RecordCollection.
// It returns the type's zero value if the RecordCollection is empty.
func (rc *RecordCollection) Get(fieldName string) interface{} {
	fi := rc.model.getRelatedFieldInfo(fieldName)
	if !rc.IsValid() {
		res := reflect.Zero(fi.structField.Type).Interface()
		if fi.isRelationField() {
			res = rc.convertToRecordSet(res, fi.relatedModelName)
		}
		return res
	}
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Load"))
	rc.Fetch()
	var res interface{}

	switch {
	case rc.IsEmpty():
		res = reflect.Zero(fi.structField.Type).Interface()
	case fi.isComputedField() && !fi.isStored():
		exprs := strings.Split(fieldName, ExprSep)
		prefix := strings.Join(exprs[:len(exprs)-1], ExprSep)
		relRC := rc
		if prefix != "" {
			relRC = rc.Get(prefix).(RecordSet).Collection()
		}
		fMap := make(FieldMap)
		relRC.computeFieldValues(&fMap, fi.json)
		res = fMap[fi.json]
	case fi.isRelatedField():
		res = rc.Get(rc.substituteRelatedInPath(fieldName))
	default:
		// If value is not in cache we fetch the whole model to speed up later calls to Get,
		// except for the case of non stored relation fields, where we only load the requested field.
		all := !fi.fieldType.IsNonStoredRelationType()
		res, _ = rc.get(fieldName, all)
	}

	if res == nil || res == (*interface{})(nil) {
		// res is nil if we do not have access rights on the field.
		// then return the field's type zero value
		res = reflect.Zero(fi.structField.Type).Interface()
	}

	if fi.isRelationField() {
		res = rc.convertToRecordSet(res, fi.relatedModelName)
	}
	return res
}

// ConvertToRecordSet the given val which can be of type *interface{}(nil) int64, []int64
// for the given related model name
func (rc *RecordCollection) convertToRecordSet(val interface{}, relatedModelName string) *RecordCollection {
	if rc.env == nil {
		return InvalidRecordCollection(relatedModelName)
	}
	res := newRecordCollection(rc.Env(), relatedModelName)
	switch r := val.(type) {
	case *interface{}, nil, bool:
	case []interface{}:
		if len(r) > 0 {
			ids := make([]int64, len(r))
			for i, v := range r {
				ids[i] = int64(v.(float64))
			}
			res = res.withIds(ids).SortedDefault()
		}
	case int64:
		if r != 0 {
			res = res.withIds([]int64{r})
		}
	case int:
		if r != 0 {
			res = res.withIds([]int64{int64(r)})
		}
	case float64:
		if r != 0 {
			res = res.withIds([]int64{int64(r)})
		}
	case []int64:
		res = res.withIds(r).SortedDefault()
	case []int:
		ids := make([]int64, len(r))
		for i, v := range r {
			ids[i] = int64(v)
		}
		res = res.withIds(ids).SortedDefault()
	case []float64:
		ids := make([]int64, len(r))
		for i, v := range r {
			ids[i] = int64(v)
		}
		res = res.withIds(ids).SortedDefault()
	case RecordSet:
		res = r.Collection()
	default:
		log.Panic("unexpected type", "value", r, "type", fmt.Sprintf("%T", r))
	}
	return res
}

// get returns the value of field for this RecordSet.
// It loads the cache if necessary before reading.
// If all is true, all fields of the model are loaded, otherwise only field.
//
// Second returned value is true if a call to the DB was necessary (i.e. not in cache)
func (rc *RecordCollection) get(field string, all bool) (interface{}, bool) {
	rc.Fetch()
	var dbCalled bool
	if !rc.env.cache.checkIfInCache(rc.model, []int64{rc.ids[0]}, []string{field}, rc.query.ctxArgsSlug(), true) {
		fields := []string{field}
		if all {
			fields = append(fields, rc.model.fields.storedFieldNames()...)
		}
		rc.Load(fields...)
		if rc.IsEmpty() {
			// rc might now be empty if it has just been deleted
			return nil, true
		}
		dbCalled = true
	}
	return rc.env.cache.get(rc.model, rc.ids[0], field, rc.query.ctxArgsSlug()), dbCalled
}

// Set sets field given by fieldName to the given value. If the RecordSet has several
// Records, all of them will be updated. Each call to Set makes an update query in the
// database. It panics if it is called on an empty RecordSet.
func (rc *RecordCollection) Set(fieldName string, value interface{}) {
	md := NewModelData(rc.model).Set(fieldName, value)
	rc.Call("Write", md)
}

// InvalidateCache clears the cache for this RecordSet data, and immediately reloads the data from the DB.
func (rc *RecordCollection) InvalidateCache() {
	for _, rec := range rc.Records() {
		rc.env.cache.invalidateRecord(rc.model, rec.ids[0])
	}
	rc.Load()
}

// First returns the values of the first Record of the RecordCollection as a ModelData.
//
// If this RecordCollection is empty, it returns an empty ModelData.
func (rc *RecordCollection) First() *ModelData {
	rc.Fetch()
	if rc.IsEmpty() {
		NewModelData(rc.model)
	}
	fields := rc.model.fields.allJSONNames()
	rc.Load(fields...)
	res := NewModelDataFromRS(rc)
	for _, f := range fields {
		res.Set(f, rc.Get(f))
	}
	return res
}

// All returns the values of all records of the RecordCollection as a slice of ModelData.
func (rc *RecordCollection) All() []*ModelData {
	rc.Fetch()
	res := make([]*ModelData, rc.Len())
	recs := rc.Records()
	for i := 0; i < rc.Len(); i++ {
		res[i] = recs[i].First()
	}
	return res
}

// Aggregates returns the result of this RecordCollection query, which must by a grouped query.
func (rc *RecordCollection) Aggregates(fieldNames ...FieldNamer) []GroupAggregateRow {
	if len(rc.query.groups) == 0 {
		log.Panic("Trying to get aggregates of a non-grouped query", "model", rc.model)
	}
	groups := make([]string, len(rc.query.groups))
	copy(groups, rc.query.groups)

	rSet := rc.addRecordRuleConditions(rc.env.uid, security.Read)
	rSet.applyContexts()
	fields := convertToStringSlice(fieldNames)
	subFields, substMap := rSet.substituteRelatedFields(fields)
	rSet = rSet.substituteRelatedInQuery()
	dbFields := filterOnDBFields(rSet.model, subFields, true)

	rSet = rSet.fixGroupByOrders(subFields...)

	fieldsOperatorMap := rSet.fieldsGroupOperators(dbFields)
	sql, args := rSet.query.selectGroupQuery(fieldsOperatorMap)
	var res []GroupAggregateRow
	rows := dbQuery(rSet.env.cr.tx, sql, args...)
	defer rows.Close()

	for rows.Next() {
		vals := make(FieldMap)
		err := sqlx.MapScan(rows, vals)
		if err != nil {
			log.Panic(err.Error(), "model", rSet.ModelName(), "fields", fields)
		}
		cnt := vals["__count"].(int64)
		delete(vals, "__count")
		vals = substituteKeys(vals, substMap)
		line := GroupAggregateRow{
			Values:    NewModelDataFromRS(rc, vals),
			Count:     int(cnt),
			Condition: getGroupCondition(groups, vals, rc.query.cond),
		}
		res = append(res, line)
	}
	return res
}

// fixGroupByOrders adds order by expressions to group by clause to have a correct query.
// It also adds a default order to the grouped fields if it does not exist.
func (rc *RecordCollection) fixGroupByOrders(fieldNames ...string) *RecordCollection {
	rSet := rc
	orderExprs := rc.query.getOrderByExpressions(false)
	ctxOrderExprs := rc.query.getCtxOrderByExpressions()
	groupExprs := rc.query.getGroupByExpressions()
	groupFields := make(map[string]bool)
	ctxGroupFields := make(map[string]bool)
	for _, g := range groupExprs {
		groupFields[strings.Join(g, ExprSep)] = true
	}
	fieldsMap := make(map[string]bool)
	for _, f := range fieldNames {
		fieldsMap[f] = true
	}
	for _, o := range orderExprs {
		oName := strings.Join(jsonizeExpr(rc.model, o), ExprSep)
		if !groupFields[oName] && !fieldsMap[oName] {
			rSet = rSet.GroupBy(FieldName(oName))
		}
	}
	for _, o := range ctxOrderExprs {
		oName := strings.Join(jsonizeExpr(rc.model, o), ExprSep)
		if !ctxGroupFields[oName] && !fieldsMap[oName] {
			rSet = rSet.clone()
			rSet.query.ctxGroups = append(rSet.query.ctxGroups, oName)
		}
	}
	if len(rc.query.orders) == 0 {
		rSet = rSet.OrderBy(rSet.query.groups...)
	}
	return rSet
}

// fieldsGroupOperators returns a map of fields to retrieve in a group by query.
// The returned map has a field as key, and sql aggregate function as value.
func (rc *RecordCollection) fieldsGroupOperators(fields []string) map[string]string {
	groups := make(map[string]bool)
	for _, g := range rc.query.groups {
		groups[rc.model.JSONizeFieldName(g)] = true
	}
	res := make(map[string]string)
	for _, dbf := range fields {
		if groups[dbf] {
			res[dbf] = ""
			continue
		}
		fi := rc.model.getRelatedFieldInfo(dbf)
		if fi.fieldType != fieldtype.Float && fi.fieldType != fieldtype.Integer {
			continue
		}
		res[dbf] = fi.groupOperator
	}
	return res
}

// Condition returns the query condition associated with this RecordSet.
func (rc *RecordCollection) Condition() *Condition {
	return rc.query.cond
}

// SQLFromCondition returns the WHERE clause sql and arguments corresponding to
// the given condition.
func (rc *RecordCollection) SQLFromCondition(c *Condition) (string, SQLParams) {
	return rc.query.conditionSQLClause(c)
}

// Records returns the slice of RecordCollection singletons that constitute this
// RecordCollection.
func (rc *RecordCollection) Records() []*RecordCollection {
	res := make([]*RecordCollection, rc.Len())
	for i, id := range rc.Ids() {
		newRC := newRecordCollection(rc.Env(), rc.ModelName())
		res[i] = newRC.withIds([]int64{id})
		res[i].prefetchRC = rc
	}
	return res
}

// EnsureOne panics if rc is not a singleton
func (rc *RecordCollection) EnsureOne() {
	if rc.Len() != 1 {
		log.Panic("Expected singleton", "model", rc.ModelName(), "received", rc)
	}
}

// IsEmpty returns true if rc is an empty RecordCollection
func (rc *RecordCollection) IsEmpty() bool {
	return !rc.IsValid() || rc.Len() == 0
}

// IsNotEmpty returns true if rc is not an empty RecordCollection
func (rc *RecordCollection) IsNotEmpty() bool {
	return !rc.IsEmpty()
}

// IsValid returns true if this RecordSet has been initialized.
func (rc *RecordCollection) IsValid() bool {
	if rc == nil {
		return false
	}
	if rc.model == nil {
		return false
	}
	if rc.query == nil {
		return false
	}
	if rc.env == nil {
		return false
	}
	return true
}

// Len returns the number of records in this RecordCollection
func (rc *RecordCollection) Len() int {
	rc.Fetch()
	return len(rc.ids)
}

// Model returns the Model instance of this RecordCollection
func (rc *RecordCollection) Model() *Model {
	return rc.model
}

// GetRecord returns the Recordset with the given externalID. It panics if the externalID does not exist.
func (rc *RecordCollection) GetRecord(externalID string) *RecordCollection {
	res := rc.Search(rc.model.Field("HexyaExternalID").Equals(externalID))
	if res.IsEmpty() {
		log.Panic("Unknown external ID", "model", rc.model.name, "externalID", externalID)
	}
	return res
}

// withIdMap adds the given ids to this RecordCollection and returns it too.
//
// It removes duplicates and overrides the current query with ("ID", "in", ids).
func (rc *RecordCollection) withIds(ids []int64) *RecordCollection {
	// Remove 0 and duplicate ids
	idsMap := make(map[int64]bool)
	var newIds []int64
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if !idsMap[id] {
			newIds = append(newIds, id)
		}
		idsMap[id] = true
	}

	// Update RecordCollection
	rc.ids = newIds
	rc.fetched = true
	rc.filtered = false
	if len(newIds) > 0 {
		for _, id := range rc.ids {
			rc.env.cache.updateEntry(rc.model, id, "id", id, rc.query.ctxArgsSlug())
		}
		rc.query.cond = rc.Model().Field("ID").In(newIds)
		rc.query.fetchAll = false
		rc.query.limit = 0
		rc.query.offset = 0
	}
	return rc
}

// T translates the given string to the language specified by
// the 'lang' key of rc.Env().Context(). If for any reason the
// string cannot be translated, then src is returned.
//
// You MUST pass a string literal as src to have it extracted automatically
//
// The translated string will be passed to fmt.Sprintf with the optional args
// before being returned.
func (rc *RecordCollection) T(src string, args ...interface{}) string {
	lang := rc.Env().Context().GetString("lang")
	transCode := i18n.TranslateCode(lang, "", src)
	return fmt.Sprintf(transCode, args...)
}

// Collection returns the underlying RecordCollection instance
// i.e. itself
func (rc *RecordCollection) Collection() *RecordCollection {
	return rc
}

var _ RecordSet = new(RecordCollection)

// newRecordCollection returns a new empty RecordCollection in the
// given environment for the given modelName
func newRecordCollection(env Environment, modelName string) *RecordCollection {
	rc := InvalidRecordCollection(modelName)
	rc.env = &env
	return rc
}

// InvalidRecordCollection returns an invalid RecordCollection without an environment.
//
// You should really not use this function, but use env.Pool("ModelName") instead.
func InvalidRecordCollection(modelName string) *RecordCollection {
	mi := Registry.MustGet(modelName)
	rc := RecordCollection{
		model: mi,
		query: newQuery(),
		ids:   make([]int64, 0),
	}
	rc.query.recordSet = &rc
	return &rc
}
