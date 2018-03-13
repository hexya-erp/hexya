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
	"strconv"
	"strings"

	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
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
		i++
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

// create inserts a new record in the database with the given data.
// data can be either a FieldMap or a struct pointer of the same model as rs.
// This function is private and low level. It should not be called directly.
// Instead use rs.Call("Create")
func (rc *RecordCollection) create(data FieldMapper) *RecordCollection {
	defer func() {
		if r := recover(); r != nil {
			panic(rc.substituteSQLErrorMessage(r))
		}
	}()
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Create"))
	fMap := data.FieldMap()
	fMap = filterMapOnAuthorizedFields(rc.model, fMap, rc.env.uid, security.Write)
	rc.applyDefaults(&fMap, true)
	rc.addAccessFieldsCreateData(&fMap)
	rc.model.convertValuesToFieldType(&fMap)
	fMap = rc.createEmbeddedRecords(fMap)
	// clean our fMap from ID and non stored fields
	fMap.RemovePKIfZero()
	storedFieldMap := filterMapOnStoredFields(rc.model, fMap)
	// insert in DB
	var createdId int64
	sql, args := rc.query.insertQuery(storedFieldMap)
	rc.env.cr.Get(&createdId, sql, args...)

	rc.env.cache.addRecord(rc.model, createdId, storedFieldMap)
	rSet := rc.withIds([]int64{createdId})
	// update reverse relation fields
	rSet.updateRelationFields(fMap)
	// compute stored fields
	rSet.processInverseMethods(fMap)
	rSet.processTriggers(fMap)
	rSet.checkConstraints()
	return rSet
}

// createEmbeddedRecords creates the records that are embedded in this
// one if they don't already exist. It returns the given fMap with the
// ids inserted for the embedded records.
func (rc *RecordCollection) createEmbeddedRecords(fMap FieldMap) FieldMap {
	type modelAndValues struct {
		model  string
		values FieldMap
	}
	embeddedData := make(map[string]modelAndValues)
	// 1. We create entries in our map for each embedded field if they don't already have an id
	for fName, fi := range rc.model.fields.registryByName {
		if !fi.embed {
			continue
		}
		if id, ok := fMap[fi.json].(int64); ok && id != int64(0) {
			continue
		}
		embeddedData[fName] = modelAndValues{
			model:  fi.relatedModelName,
			values: make(FieldMap),
		}
	}
	// 2. We populate our map with the values for each embedded record
	for fName, value := range fMap {
		fi := rc.Model().fields.MustGet(fName)
		if fi.relatedPath == "" {
			continue
		}
		exprs := strings.Split(fi.relatedPath, ExprSep)
		if len(exprs) != 2 {
			continue
		}
		fm, ok := embeddedData[exprs[0]]
		if !ok {
			continue
		}
		fm.values[exprs[1]] = value
	}
	// 3. We create the embedded records
	for fieldName, vals := range embeddedData {
		// We do not call "create" directly to have the caller set in the callstack for permissions
		res := rc.env.Pool(vals.model).Call("Create", vals.values)
		if resRS, ok := res.(RecordSet); ok {
			fMap[fieldName] = resRS.Ids()[0]
		}
	}
	return fMap
}

// applyDefaults adds the default value to the given fMap values which
// are equal to their Go type zero value. If requiredOnly is true, default
// value is set only if the field is required (and equal to zero value).
func (rc *RecordCollection) applyDefaults(fMap *FieldMap, requiredOnly bool) {
	for fName, fi := range Registry.MustGet(rc.ModelName()).fields.registryByJSON {
		if fi.defaultFunc == nil {
			continue
		}
		val := reflect.ValueOf((*fMap)[fName])
		if !fi.isReadOnly() && (!val.IsValid() || val == reflect.Zero(val.Type())) {
			if fi.required || !requiredOnly {
				(*fMap)[fName] = fi.defaultFunc(rc.Env())
			}
		}
	}
}

// checkConstraints executes the constraint method for each field defined
// in the given fMap with the corresponding value.
// Each method is only executed once, even if it is called by several fields.
// It panics as soon as one constraint fails.
func (rc *RecordCollection) checkConstraints() {
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
// This function is private and low level. It should not be called directly.
// Instead use rs.Call("Write")
func (rc *RecordCollection) update(data FieldMapper, fieldsToUnset ...FieldNamer) bool {
	rSet := rc.addRecordRuleConditions(rc.env.uid, security.Write)
	fMap := data.FieldMap(fieldsToUnset...)
	rSet.addAccessFieldsUpdateData(&fMap)
	// We process inverse method before we convert RecordSets to ids
	rSet.processInverseMethods(fMap)
	rSet.model.convertValuesToFieldType(&fMap)
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
	// compute stored fields
	rSet.processTriggers(fMap)
	rSet.checkConstraints()
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
// invalidates the cache for the record
func (rc *RecordCollection) doUpdate(fMap FieldMap) {
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Write"))
	defer func() {
		if r := recover(); r != nil {
			panic(rc.substituteSQLErrorMessage(r))
		}
	}()
	fMap = filterMapOnAuthorizedFields(rc.model, fMap, rc.env.uid, security.Write)
	// update DB
	if len(fMap) > 0 {
		sql, args := rc.query.updateQuery(fMap)
		res := rc.env.cr.Execute(sql, args...)
		if num, _ := res.RowsAffected(); num == 0 {
			log.Panic("Trying to update an empty RecordSet", "model", rc.ModelName(), "values", fMap)
		}
	}
	for _, rec := range rc.Records() {
		for k, v := range fMap {
			rc.env.cache.updateEntry(rc.model, rec.Ids()[0], k, v)
		}
	}
}

// updateRelationFields updates reverse relations fields of the
// given fMap.
func (rc *RecordCollection) updateRelationFields(fMap FieldMap) {
	rc.Fetch()
	for field, value := range fMap {
		fi := rc.model.getRelatedFieldInfo(field)
		if !checkFieldPermission(fi, rc.env.uid, security.Write) {
			continue
		}
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
				toAdd.Set(fi.reverseFK, rc.ids[0])
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

// updateRelatedFields updates related non stored fields of the
// given fMap.
func (rc *RecordCollection) updateRelatedFields(fMap FieldMap) {
	rc.Fetch()
	var toLoad []string
	toSubstitute := make(map[string]string)
	for field := range fMap {
		fi := rc.model.fields.MustGet(field)
		if !fi.isRelatedField() {
			continue
		}
		if !checkFieldPermission(fi, rc.env.uid, security.Write) {
			continue
		}

		toSubstitute[field] = fi.relatedPath
		if !rc.env.cache.checkIfInCache(rc.model, rc.ids, []string{fi.relatedPath}) {
			toLoad = append(toLoad, field)
		}
	}
	// Load related paths if not loaded already
	if len(toLoad) > 0 {
		rc.Load(toLoad...)
	}
	// Create an update map for each record to update
	updateMap := make(map[cacheRef]FieldMap)
	for _, rec := range rc.Records() {
		for field, subst := range toSubstitute {
			ref, relField, _ := rec.env.cache.getRelatedRef(rec.model, rec.ids[0], subst)
			if _, exists := updateMap[ref]; !exists {
				updateMap[ref] = make(FieldMap)
			}
			updateMap[ref][relField] = fMap[field]
		}
	}

	// Make the update for each record
	for ref, upMap := range updateMap {
		rs := rc.env.Pool(ref.model.name).withIds([]int64{ref.id})
		rs.Call("Write", upMap)
	}
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
	sql, args := rSet.query.deleteQuery()
	res := rSet.env.cr.Execute(sql, args...)
	num, _ := res.RowsAffected()
	for _, id := range ids {
		rc.env.cache.invalidateRecord(rc.model, id)
	}
	return num
}

// Search returns a new RecordSet filtering on the current one with the
// additional given Condition
func (rc *RecordCollection) Search(cond *Condition) *RecordCollection {
	rSetVal := *rc
	rSetVal.query = rc.query.clone()
	rSetVal.query.cond = rSetVal.query.cond.AndCond(cond)
	return &rSetVal
}

// NoDistinct removes the DISTINCT keyword from this RecordSet query.
// By default, all queries are distinct.
func (rc *RecordCollection) NoDistinct() *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone()
	rSet.query.noDistinct = true
	return &rSet
}

// Limit returns a new RecordSet with only the first 'limit' records.
func (rc *RecordCollection) Limit(limit int) *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone()
	rSet.query.limit = limit
	return &rSet
}

// Offset returns a new RecordSet with only the records starting at offset
func (rc *RecordCollection) Offset(offset int) *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone()
	rSet.query.offset = offset
	return &rSet
}

// OrderBy returns a new RecordSet ordered by the given ORDER BY expressions
func (rc *RecordCollection) OrderBy(exprs ...string) *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone()
	rSet.query.orders = append(rSet.query.orders, exprs...)
	return &rSet
}

// GroupBy returns a new RecordSet grouped with the given GROUP BY expressions
func (rc *RecordCollection) GroupBy(fields ...FieldNamer) *RecordCollection {
	rSet := *rc
	rSet.query = rSet.query.clone()
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
	addNameSearchesToCondition(rSet.model, rSet.query.cond)
	sql, args := rSet.query.countQuery()
	var res int
	rSet.env.cr.Get(&res, sql, args...)
	return res
}

// Load query all data of the RecordCollection and store in cache.
// fields are the fields to retrieve in the path format,
// i.e. "User.Profile.Age" or "user_id.profile_id.age".
//
// If no fields are given, all DB columns of the RecordCollection's
// model are retrieved. Non-DB fields must be explicitly given in
// fields to be retrieved.
func (rc *RecordCollection) Load(fields ...string) *RecordCollection {
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
		rSet = rc.prefetchRC
	}
	rSet = rSet.addRecordRuleConditions(rc.env.uid, security.Read)
	if len(rSet.query.orders) == 0 {
		rSet.query.orders = make([]string, len(rSet.model.defaultOrder))
		copy(rSet.query.orders, rSet.model.defaultOrder)
	}
	var results []FieldMap
	if len(fields) == 0 {
		fields = rSet.model.fields.storedFieldNames()
	}
	fields = filterOnAuthorizedFields(rSet.model, rSet.env.uid, fields, security.Read)
	addNameSearchesToCondition(rSet.model, rSet.query.cond)
	subFields, rSet := rSet.substituteRelatedFields(fields)
	dbFields := filterOnDBFields(rSet.model, subFields)
	sql, args := rSet.query.selectQuery(dbFields)
	rows := dbQuery(rSet.env.cr.tx, sql, args...)
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		line := make(FieldMap)
		err := rSet.model.scanToFieldMap(rows, &line)
		if err != nil {
			log.Panic(err.Error(), "model", rSet.ModelName(), "fields", fields)
		}
		results = append(results, line)
		rSet.env.cache.addRecord(rSet.model, line["id"].(int64), line)
		ids = append(ids, line["id"].(int64))
	}

	rSet = rSet.withIds(ids)
	rSet.loadRelationFields(fields)
	if prefetch {
		return rc
	}
	return rSet
}

// loadRelationFields loads one2many, many2many and rev2one fields from the given fields
// names in this RecordCollection into the cache. fields of other types given in fields
// are ignored.
func (rc *RecordCollection) loadRelationFields(fields []string) {
	for _, id := range rc.ids {
		for _, fieldName := range fields {
			fi := rc.model.getRelatedFieldInfo(fieldName)
			switch fi.fieldType {
			case fieldtype.One2Many:
				relRC := rc.env.Pool(fi.relatedModelName).Search(rc.Model().Field(fi.reverseFK).Equals(id)).Fetch()
				rc.env.cache.updateEntry(rc.model, id, fieldName, relRC.ids)
			case fieldtype.Many2Many:
				query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`, fi.m2mTheirField.json,
					fi.m2mRelModel.tableName, fi.m2mOurField.json)
				var ids []int64
				rc.env.cr.Select(&ids, query, id)
				rc.env.cache.updateEntry(rc.model, id, fieldName, ids)
			case fieldtype.Rev2One:
				relRC := rc.env.Pool(fi.relatedModelName).Search(rc.Model().Field(fi.reverseFK).Equals(id)).Fetch()
				var relID int64
				if len(relRC.ids) > 0 {
					relID = relRC.ids[0]
				}
				rc.env.cache.updateEntry(rc.model, id, fieldName, relID)
			default:
				continue
			}
		}
	}
}

// Get returns the value of the given fieldName for the first record of this RecordCollection.
// It returns the type's zero value if the RecordCollection is empty.
func (rc *RecordCollection) Get(fieldName string) interface{} {
	rc.CheckExecutionPermission(rc.model.methods.MustGet("Load"))
	rc.Fetch()
	fi := rc.model.fields.MustGet(fieldName)
	var res interface{}

	switch {
	case !checkFieldPermission(fi, rc.env.uid, security.Read):
		res = nil
	case rc.IsEmpty():
		res = reflect.Zero(fi.structField.Type).Interface()
	case fi.isComputedField() && !fi.isStored():
		fMap := make(FieldMap)
		rc.computeFieldValues(&fMap, fi.json)
		res = fMap[fi.json]
	case fi.isRelatedField() && !fi.isStored():
		res, _ = rc.get(fi.relatedPath, false)
	default:
		// If value is not in cache we fetch the whole model to speed up later calls to Get,
		// except for the case of non stored relation fields, where we only load the requested field.
		all := !fi.fieldType.IsNonStoredRelationType()
		res, _ = rc.get(fieldName, all)
	}

	if res == nil {
		// res is nil if we do not have access rights on the field.
		// then return the field's type zero value
		res = reflect.Zero(fi.structField.Type).Interface()
	}

	if fi.isRelationField() {
		switch r := res.(type) {
		case *interface{}:
			// *interface{} is returned when the field is null
			res = newRecordCollection(rc.Env(), fi.relatedModel.name)
		case int64:
			res = newRecordCollection(rc.Env(), fi.relatedModel.name)
			if r != 0 {
				res = res.(RecordSet).Collection().withIds([]int64{r})
			}
		case []int64:
			res = newRecordCollection(rc.Env(), fi.relatedModel.name).withIds(r).SortedDefault()
		}
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
	if !rc.env.cache.checkIfInCache(rc.model, []int64{rc.ids[0]}, []string{field}) {
		if !all {
			rc.Load(field)
		} else {
			rc.Load()
		}
		dbCalled = true
	}
	return rc.env.cache.get(rc.model, rc.ids[0], field), dbCalled
}

// Set sets field given by fieldName to the given value. If the RecordSet has several
// Records, all of them will be updated. Each call to Set makes an update query in the
// database. It panics if it is called on an empty RecordSet.
func (rc *RecordCollection) Set(fieldName string, value interface{}) {
	fMap := make(FieldMap)
	fMap[fieldName] = value
	rc.Call("Write", fMap)
}

// InvalidateCache clears the cache for this RecordSet data, and immediately reloads the data from the DB.
func (rc *RecordCollection) InvalidateCache() {
	for _, rec := range rc.Records() {
		rc.env.cache.invalidateRecord(rc.model, rec.ids[0])
	}
	rc.Load()
}

// First populates structPtr with a copy of the first Record of the RecordCollection.
// structPtr must a pointer to a struct.
func (rc *RecordCollection) First(structPtr interface{}) {
	rc.Fetch()
	if err := checkStructPtr(structPtr); err != nil {
		log.Panic("Invalid structPtr given", "error", err, "model", rc.ModelName(), "received", structPtr)
	}
	if rc.IsEmpty() {
		return
	}
	typ := reflect.TypeOf(structPtr).Elem()
	fields := make([]string, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		fields[i] = typ.Field(i).Name
	}
	rc.Load(fields...)
	fMap := rc.env.cache.getRecord(rc.Model(), rc.ids[0])
	fMap = filterMapOnAuthorizedFields(rc.model, fMap, rc.env.uid, security.Read)
	MapToStruct(rc, structPtr, fMap)
}

// All fetches a copy of all records of the RecordCollection and populates structSlicePtr.
func (rc *RecordCollection) All(structSlicePtr interface{}) {
	rc.Fetch()
	if err := checkStructSlicePtr(structSlicePtr); err != nil {
		log.Panic("Invalid structPtr given", "error", err, "model", rc.ModelName(), "received", structSlicePtr)
	}
	val := reflect.ValueOf(structSlicePtr)
	// sspType is []*struct
	sspType := val.Type().Elem()
	// structType is struct
	structType := sspType.Elem().Elem()
	val.Elem().Set(reflect.MakeSlice(sspType, rc.Len(), rc.Len()))
	recs := rc.Records()
	rc.Load()
	for i := 0; i < rc.Len(); i++ {
		fMap := rc.env.cache.getRecord(rc.Model(), recs[i].ids[0])
		fMap = filterMapOnAuthorizedFields(rc.model, fMap, rc.env.uid, security.Read)
		newStructPtr := reflect.New(structType).Interface()
		MapToStruct(rc, newStructPtr, fMap)
		val.Elem().Index(i).Set(reflect.ValueOf(newStructPtr))
	}
}

// Aggregates returns the result of this RecordCollection query, which must by a grouped query.
func (rc *RecordCollection) Aggregates(fieldNames ...FieldNamer) []GroupAggregateRow {
	if len(rc.query.groups) == 0 {
		log.Panic("Trying to get aggregates of a non-grouped query", "model", rc.model)
	}
	rSet := rc.addRecordRuleConditions(rc.env.uid, security.Read)
	fields := filterOnAuthorizedFields(rSet.model, rSet.env.uid, convertToStringSlice(fieldNames), security.Read)
	subFields, rSet := rSet.substituteRelatedFields(fields)
	dbFields := filterOnDBFields(rSet.model, subFields, true)

	if len(rSet.query.orders) == 0 {
		rSet = rSet.OrderBy(rSet.query.groups...)
	}
	fieldsOperatorMap := rSet.fieldsGroupOperators(dbFields)
	sql, args := rSet.query.selectGroupQuery(fieldsOperatorMap)
	var res []GroupAggregateRow
	rows := dbQuery(rSet.env.cr.tx, sql, args...)
	defer rows.Close()

	for rows.Next() {
		vals := make(map[string]interface{})
		err := sqlx.MapScan(rows, vals)
		if err != nil {
			log.Panic(err.Error(), "model", rSet.ModelName(), "fields", fields)
		}
		cnt := vals["__count"].(int64)
		delete(vals, "__count")
		line := GroupAggregateRow{
			Values:    vals,
			Count:     int(cnt),
			Condition: getGroupCondition(rc.query.groups, vals, rc.query.cond),
		}
		res = append(res, line)
	}
	return res
}

// fieldsGroupOperators returns a map of fields to retrieve in a group by query.
// The returned map has a field as key, and sql aggregate function as value.
// it also includes 'field_count' for grouped fields
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

// withIdMap adds the given ids to this RecordCollection and returns it too.
// It overrides the current query with ("ID", "in", ids).
func (rc *RecordCollection) withIds(ids []int64) *RecordCollection {
	var newIds []int64
	for _, id := range ids {
		if id == 0 {
			continue
		}
		newIds = append(newIds, id)
	}
	rc.ids = newIds
	rc.fetched = true
	rc.filtered = false
	if len(newIds) > 0 {
		for _, id := range rc.ids {
			rc.env.cache.updateEntry(rc.model, id, "id", id)
		}
		rc.query.cond = rc.Model().Field("ID").In(newIds)
		rc.query.fetchAll = false
		rc.query.limit = 0
		rc.query.offset = 0
		rc.query.noDistinct = true
	}
	return rc
}

// T translates the given string to the language specified by
// the 'lang' key of rc.Env().Context(). If for any reason the
// string cannot be translated, then src is returned.
//
// You MUST pass a string literal as src to have it extracted automatically
//
// The given src will be passed to fmt.Sprintf with the optional args
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
	mi := Registry.MustGet(modelName)
	rc := RecordCollection{
		model: mi,
		query: newQuery(),
		env:   &env,
		ids:   make([]int64, 0),
	}
	rc.query.recordSet = &rc
	return &rc
}
