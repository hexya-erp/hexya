// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"reflect"
)

// recordSetWrappers is a map that stores the available types of RecordSet
var recordSetWrappers map[string]reflect.Type

// RegisterRecordSetWrapper registers the object passed as obj as the RecordSet type
// for the given model.
//
// - typ must be a struct that embeds *RecordCollection
// - modelName must be the name of a model that exists in the registry
func RegisterRecordSetWrapper(modelName string, obj interface{}) {
	Registry.MustGet(modelName)
	typ := reflect.TypeOf(obj)
	if typ.Kind() != reflect.Struct {
		log.Panic("trying to register a non struct type as Wrapper", "modelName", modelName, "type", typ)
	}
	if typ.Field(0).Type != reflect.TypeOf(new(RecordCollection)) {
		log.Panic("trying to register a struct that don't embed *RecordCollection", "modelName", modelName, "type", typ)
	}
	recordSetWrappers[modelName] = typ
}

// Wrap returns the given RecordCollection embedded into a RecordSet Wrapper type
//
// If modelName is defined, wrap in a modelName Wrapper type instead (use for mixins).
func (rc *RecordCollection) Wrap(modelName ...string) interface{} {
	modName := rc.ModelName()
	if len(modelName) > 0 {
		modName = modelName[0]
	}
	typ, ok := recordSetWrappers[modName]
	if !ok {
		log.Panic("unable to wrap RecordCollection", "model", modName)
	}
	val := reflect.New(typ).Elem()
	val.Field(0).Set(reflect.ValueOf(rc))
	return val.Interface()
}

// recordSetWrappers is a map that stores the available types of ModelData
var modelDataWrappers map[string]reflect.Type

// RegisterModelDataWrapper registers the object passed as obj as the ModelData type
// for the given model.
//
// - typ must be a struct that embeds ModelData
// - modelName must be the name of a model that exists in the registry
func RegisterModelDataWrapper(modelName string, obj interface{}) {
	Registry.MustGet(modelName)
	typ := reflect.TypeOf(obj)
	if typ.Kind() != reflect.Struct {
		log.Panic("trying to register a non struct type as Wrapper", "modelName", modelName, "type", typ)
	}
	if typ.Field(0).Type != reflect.TypeOf(new(ModelData)) {
		log.Panic("trying to register a struct that don't embed ModelData", "modelName", modelName, "type", typ)
	}
	modelDataWrappers[modelName] = typ
}

// Wrap returns the given ModelData embedded into a RecordSet Wrapper type.
// This method returns a pointer.
func (md ModelData) Wrap() interface{} {
	typ, ok := modelDataWrappers[md.model.name]
	if !ok {
		return &md
	}
	val := reflect.New(typ)
	val.Elem().Field(0).Set(reflect.ValueOf(&md))
	return val.Interface()
}
