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
	"reflect"

	"github.com/npiganeau/yep/yep/tools"
)

// methodsCache is the methodInfo collection
type methodsCollection struct {
	cache        map[string]*methodInfo
	cacheByFunc  map[reflect.Value]*methodLayer
	bootstrapped bool
}

// get returns the methodInfo of the given method.
func (mc *methodsCollection) get(methodName string) (mi *methodInfo, ok bool) {
	mi, ok = mc.cache[methodName]
	return
}

// getByFunc returns the methodInfo that includes the given function as a layer.
func (mc *methodsCollection) getByFunc(fnctPtr interface{}) (ml *methodLayer, ok bool) {
	ml, ok = mc.cacheByFunc[reflect.ValueOf(fnctPtr).Elem()]
	return
}

//set adds the given methodInfo to the methodsCollection.
func (mc *methodsCollection) set(methodName string, methInfo *methodInfo) {
	mc.cache[methodName] = methInfo
	mc.cacheByFunc[methInfo.topLayer.funcValue] = methInfo.topLayer
}

// addLayer adds the given function Value in the given methodLayer
func (mc *methodsCollection) addLayer(fnVal reflect.Value, methLayer *methodLayer) {
	mc.cacheByFunc[fnVal] = methLayer
}

// newMethodsCollection returns a pointer to a new methodsCollection
func newMethodsCollection() *methodsCollection {
	mc := methodsCollection{
		cache:       make(map[string]*methodInfo),
		cacheByFunc: make(map[reflect.Value]*methodLayer),
	}
	return &mc
}

// A methodInfo is a definition of a model's method
type methodInfo struct {
	name       string
	mi         *modelInfo
	methodType reflect.Type
	topLayer   *methodLayer
	nextLayer  map[*methodLayer]*methodLayer
}

// addMethodLayer adds the given layer to this methodInfo.
func (methInfo *methodInfo) addMethodLayer(val reflect.Value) {
	ml := methodLayer{
		funcValue: val,
		methInfo:  methInfo,
	}
	methInfo.nextLayer[&ml] = methInfo.topLayer
	methInfo.topLayer = &ml
	methInfo.mi.methods.addLayer(ml.funcValue, &ml)
}

func (methInfo *methodInfo) getNextLayer(methodLayer *methodLayer) *methodLayer {
	return methInfo.nextLayer[methodLayer]
}

// methodLayer is one layer of a method, that is one function defined in a module
type methodLayer struct {
	methInfo  *methodInfo
	funcValue reflect.Value
}

// newMethodInfo creates a new method ref with the given func value as first layer.
// First argument of given function must implement RecordSet.
func newMethodInfo(mi *modelInfo, methodName string, val reflect.Value) *methodInfo {
	funcType := val.Type()
	if funcType.NumIn() == 0 || funcType.In(0) != reflect.TypeOf((*RecordCollection)(nil)).Elem() {
		tools.LogAndPanic(log, "Function must have `RecordSet` as first argument to be used as method.", "model", mi.name, "method", methodName)
	}

	methInfo := methodInfo{
		mi:         mi,
		name:       methodName,
		methodType: val.Type(),
		nextLayer:  make(map[*methodLayer]*methodLayer),
	}
	methInfo.topLayer = &methodLayer{
		funcValue: val,
		methInfo:  &methInfo,
	}
	return &methInfo
}

// CreateMethod creates a new method on given model name and adds the given fnct
// as first layer for this method. Given fnct function must have a RecordSet as
// first argument.
func CreateMethod(modelName, methodName string, fnct interface{}) {
	mi := checkMethodAndFnctType(modelName, methodName, fnct)
	_, exists := mi.methods.get(methodName)
	if exists {
		tools.LogAndPanic(log, "Call to CreateMethod with an existing method name", "model", modelName, "method", methodName)
	}
	mi.methods.set(methodName, newMethodInfo(mi, methodName, reflect.ValueOf(fnct)))
}

// ExtendMethod adds the given fnct function as a new layer on the given
// method of the given model.
// fnct must be of the same signature as the first layer of this method.
func ExtendMethod(modelName, methodName string, fnct interface{}) {
	mi := checkMethodAndFnctType(modelName, methodName, fnct)
	methInfo, exists := mi.methods.get(methodName)
	if !exists {
		tools.LogAndPanic(log, "Call to ExtendMethod on non existant method", "model", modelName, "method", methodName)
	}
	val := reflect.ValueOf(fnct)
	if methInfo.methodType != val.Type() {
		tools.LogAndPanic(log, "Function signature does not match", "model", modelName, "method", methodName,
			"received", methInfo.methodType, "expected", val.Type())
	}
	methInfo.addMethodLayer(val)
}

// checkMethodAndFnctType checks whether the given arguments are valid for
// CreateMethod or ExtendMethod
func checkMethodAndFnctType(modelName, methodName string, fnct interface{}) *modelInfo {
	mi, ok := modelRegistry.get(modelName)
	if !ok {
		tools.LogAndPanic(log, "Unknown model", "model", modelName)
	}
	if mi.methods.bootstrapped {
		tools.LogAndPanic(log, "Create/ExtendMethod must be run before BootStrap", "model", modelName, "method", methodName)
	}

	val := reflect.ValueOf(fnct)
	if val.Kind() != reflect.Func {
		tools.LogAndPanic(log, "fnct parameter must be a function", "model", modelName, "method", methodName, "fnct", fnct)
	}
	return mi
}
