// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"sort"

	"github.com/hexya-erp/hexya/src/tools/typesutils"
)

// Union returns a new RecordCollection that is the union of this RecordCollection
// and the given `other` RecordCollection. The result is guaranteed to be a
// set of unique records. The order of the records is kept.
func (rc *RecordCollection) Union(other RecordSet) *RecordCollection {
	if !rc.IsValid() {
		return rc
	}
	if rc.ModelName() != other.ModelName() {
		log.Panic("Unable to union RecordCollections of different models", "this", rc.ModelName(),
			"other", other.ModelName())
	}
	rc.Fetch()
	idMap := make(map[int64]bool)
	origIds := append(rc.ids, other.Ids()...)
	for _, id := range origIds {
		idMap[id] = true
	}
	ids := make([]int64, len(idMap))
	i := 0
	for _, id := range origIds {
		if _, ok := idMap[id]; !ok {
			continue
		}
		ids[i] = id
		delete(idMap, id)
		i++
	}
	return newRecordCollection(rc.Env(), rc.ModelName()).withIds(ids)
}

// Subtract returns a RecordSet with the Records that are in this
// RecordCollection but not in the given 'other' one.
// The result is guaranteed to be a set of unique records.
func (rc *RecordCollection) Subtract(other RecordSet) *RecordCollection {
	if !rc.IsValid() {
		return rc
	}
	if rc.ModelName() != other.ModelName() {
		log.Panic("Unable to subtract RecordCollections of different models", "this", rc.ModelName(),
			"other", other.ModelName())
	}
	rc.Fetch()
	idMap := make(map[int64]bool)
	for _, id := range rc.ids {
		idMap[id] = true
	}
	for _, id := range other.Ids() {
		delete(idMap, id)
	}
	ids := make([]int64, len(idMap))
	i := 0
	for _, id := range rc.ids {
		if _, ok := idMap[id]; !ok {
			continue
		}
		ids[i] = id
		i++
	}
	return newRecordCollection(rc.Env(), rc.ModelName()).withIds(ids)
}

// Intersect returns a new RecordCollection with only the records that are both
// in this RecordCollection and in the other RecordSet.
func (rc *RecordCollection) Intersect(other RecordSet) *RecordCollection {
	if !rc.IsValid() {
		return rc
	}
	if rc.ModelName() != other.ModelName() {
		log.Panic("Unable to intersect RecordCollections of different models", "this", rc.ModelName(),
			"other", other.ModelName())
	}
	rc.Fetch()
	idMap := make(map[int64]bool)
	for _, id := range rc.ids {
		for _, ido := range other.Ids() {
			if ido == id {
				idMap[id] = true
				break
			}
		}
	}
	ids := make([]int64, len(idMap))
	i := 0
	for _, id := range rc.ids {
		if _, ok := idMap[id]; !ok {
			continue
		}
		ids[i] = id
		i++
	}
	return newRecordCollection(rc.Env(), rc.ModelName()).withIds(ids)
}

// CartesianProduct returns the cartesian product of this RecordCollection with others.
//
// This function panics if all records are not pf the same model
func (rc *RecordCollection) CartesianProduct(records ...RecordSet) []*RecordCollection {
	recSlices := make([][]*RecordCollection, len(records)+1)
	recSlices[0] = rc.Records()
	for i, rec := range records {
		recSlices[i+1] = rec.Collection().Records()
	}
	return cartesianProductSlices(recSlices...)
}

// Equals returns true if this RecordCollection is the same as other
// i.e. they are of the same model and have the same ids
func (rc *RecordCollection) Equals(other RecordSet) bool {
	if rc.ModelName() != other.ModelName() {
		return false
	}
	if rc.Len() != other.Len() {
		return false
	}
	theseIds := make(map[int64]bool)
	for _, id := range rc.Ids() {
		theseIds[id] = true
	}
	for _, id := range other.Ids() {
		if !theseIds[id] {
			return false
		}
		delete(theseIds, id)
	}
	if len(theseIds) != 0 {
		return false
	}
	return true
}

// Sorted returns a new RecordCollection sorted according to the given less function.
//
// The less function should return true if rs1 < rs2
func (rc *RecordCollection) Sorted(less func(rs1 RecordSet, rs2 RecordSet) bool) *RecordCollection {
	if !rc.IsValid() {
		return rc
	}
	records := rc.Records()
	sort.Slice(records, func(i, j int) bool {
		return less(records[i], records[j])
	})
	var ids []int64
	for _, rec := range records {
		ids = append(ids, rec.ids[0])
	}
	return newRecordCollection(rc.Env(), rc.ModelName()).withIds(ids)
}

// SortedDefault returns a new record set with the same records as rc but sorted according
// to the default order of this model
func (rc *RecordCollection) SortedDefault() *RecordCollection {
	return rc.Sorted(func(rs1 RecordSet, rs2 RecordSet) bool {
		for _, order := range Registry.MustGet(rs1.ModelName()).defaultOrder {
			if eq, _ := typesutils.AreEqual(rs1.Collection().Get(order.field), rs2.Collection().Get(order.field)); eq {
				continue
			}
			lt, _ := typesutils.IsLessThan(rs1.Collection().Get(order.field), rs2.Collection().Get(order.field))
			return (lt && !order.desc) || (!lt && order.desc)
		}
		return false
	})
}

// SortedByField returns a new record set with the same records as rc but sorted by the given field.
// If reverse is true, the sort is done in reversed order
func (rc *RecordCollection) SortedByField(name FieldName, reverse bool) *RecordCollection {
	return rc.Sorted(func(rs1 RecordSet, rs2 RecordSet) bool {
		lt, err := typesutils.IsLessThan(rs1.Collection().Get(name), rs2.Collection().Get(name))
		if err != nil {
			log.Panic("Unable to sort recordset", "recordset", rc, "error", err)
		}
		if reverse {
			return !lt
		}
		return lt
	})
}

// Filtered returns a new record set with only the elements of this record set
// for which test is true.
//
// Note that if this record set is not fully loaded, this function will call the database
// to load the fields before doing the filtering. In this case, it might be more efficient
// to search the database directly with the filter condition.
func (rc *RecordCollection) Filtered(test func(rs RecordSet) bool) *RecordCollection {
	if !rc.IsValid() {
		return rc
	}
	res := rc.Env().Pool(rc.ModelName())
	for _, rec := range rc.Records() {
		if !test(rec) {
			continue
		}
		res = res.Union(rec)
	}
	return res
}
