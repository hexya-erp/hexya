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

	"github.com/npiganeau/yep/yep/tools/logging"
)

// methodsCache is the methodInfo collection
type methodsCollection struct {
	registry     map[string]*methodInfo
	bootstrapped bool
}

// get returns the methodInfo of the given method.
func (mc *methodsCollection) get(methodName string) (mi *methodInfo, ok bool) {
	mi, ok = mc.registry[methodName]
	return
}

// mustGet returns the methodInfo of the given method. It panics if the
// method is not found.
func (mc *methodsCollection) mustGet(methodName string) *methodInfo {
	methInfo, ok := mc.get(methodName)
	if !ok {
		var model string
		for _, f := range mc.registry {
			model = f.mi.name
			break
		}
		logging.LogAndPanic(log, "Unknown field in model", "model", model, "method", methodName)
	}
	return methInfo
}

// set adds the given methodInfo to the methodsCollection.
func (mc *methodsCollection) set(methodName string, methInfo *methodInfo) {
	mc.registry[methodName] = methInfo
}

// newMethodsCollection returns a pointer to a new methodsCollection
func newMethodsCollection() *methodsCollection {
	mc := methodsCollection{
		registry: make(map[string]*methodInfo),
	}
	return &mc
}

// A methodInfo is a definition of a model's method
type methodInfo struct {
	name       string
	mi         *modelInfo
	doc        string
	methodType reflect.Type
	topLayer   *methodLayer
	nextLayer  map[*methodLayer]*methodLayer
}

// addMethodLayer adds the given layer to this methodInfo.
func (methInfo *methodInfo) addMethodLayer(val reflect.Value, doc string) {
	ml := methodLayer{
		funcValue: wrapFunctionForMethodLayer(val),
		methInfo:  methInfo,
		doc:       doc,
	}
	methInfo.nextLayer[&ml] = methInfo.topLayer
	methInfo.topLayer = &ml
}

func (methInfo *methodInfo) getNextLayer(methodLayer *methodLayer) *methodLayer {
	return methInfo.nextLayer[methodLayer]
}

// invertedLayers returns the list of method layers starting
// from the base methods and going up all inherited layers
func (methInfo *methodInfo) invertedLayers() []*methodLayer {
	var layersInv []*methodLayer
	for cl := methInfo.topLayer; cl != nil; cl = methInfo.getNextLayer(cl) {
		layersInv = append([]*methodLayer{cl}, layersInv...)
	}
	return layersInv
}

// methodLayer is one layer of a method, that is one function defined in a module
type methodLayer struct {
	methInfo  *methodInfo
	mixedIn   bool
	funcValue reflect.Value
	doc       string
}

// newMethodInfo creates a new method ref with the given func value as first layer.
// First argument of given function must implement RecordSet.
func newMethodInfo(mi *modelInfo, methodName, doc string, val reflect.Value) *methodInfo {
	funcType := val.Type()
	if funcType.NumIn() == 0 || !funcType.In(0).Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
		logging.LogAndPanic(log, "Function must have a `RecordSet` as first argument to be used as method.", "model", mi.name, "method", methodName, "type", funcType.In(0))
	}

	methInfo := methodInfo{
		mi:         mi,
		name:       methodName,
		methodType: val.Type(),
		nextLayer:  make(map[*methodLayer]*methodLayer),
	}
	methInfo.topLayer = &methodLayer{
		funcValue: wrapFunctionForMethodLayer(val),
		methInfo:  &methInfo,
		doc:       doc,
	}
	return &methInfo
}

// wrapFunctionForMethodLayer take the given fnct Value and wrap it in a
// func(RecordCollection, args...) function Value suitable for use in a
// methodLayer.
func wrapFunctionForMethodLayer(fnctVal reflect.Value) reflect.Value {
	wrapperType := reflect.TypeOf(func(RecordCollection, ...interface{}) []interface{} { return nil })
	if fnctVal.Type() == wrapperType {
		// fnctVal is already wrapped, we just return it
		return fnctVal
	}
	methodLayerFunction := func(rc RecordCollection, args ...interface{}) []interface{} {
		argZeroType := fnctVal.Type().In(0)
		argsVals := make([]reflect.Value, len(args)+1)
		argsVals[0] = reflect.New(argZeroType).Elem()
		if argZeroType == reflect.TypeOf(RecordCollection{}) {
			argsVals[0].Set(reflect.ValueOf(rc))
		} else {
			argsVals[0].Field(0).Set(reflect.ValueOf(rc))
		}
		for i, arg := range args {
			argsVals[i+1] = reflect.ValueOf(arg)
		}

		var retVal []reflect.Value
		if fnctVal.Type().IsVariadic() && len(argsVals) == fnctVal.Type().NumIn() {
			retVal = fnctVal.CallSlice(argsVals)
		} else {
			retVal = fnctVal.Call(argsVals)
		}

		res := make([]interface{}, len(retVal))
		for i, val := range retVal {
			res[i] = val.Interface()
		}
		return res
	}
	return reflect.ValueOf(methodLayerFunction)
}

// CreateMethod creates a new method on given model name and adds the given fnct
// as first layer for this method. Given fnct function must have a RecordSet as
// first argument.
func CreateMethod(modelName, methodName, doc string, fnct interface{}) {
	mi := checkMethodAndFnctType(modelName, methodName, fnct)
	_, exists := mi.methods.get(methodName)
	if exists {
		logging.LogAndPanic(log, "Call to CreateMethod with an existing method name", "model", modelName, "method", methodName)
	}
	mi.methods.set(methodName, newMethodInfo(mi, methodName, doc, reflect.ValueOf(fnct)))
}

// ExtendMethod adds the given fnct function as a new layer on the given
// method of the given model.
// fnct must be of the same signature as the first layer of this method.
func ExtendMethod(modelName, methodName, doc string, fnct interface{}) {
	mi := checkMethodAndFnctType(modelName, methodName, fnct)
	methInfo, exists := mi.methods.get(methodName)
	if !exists {
		// We didn't find the method, but maybe it exists in mixins
		var found bool
		allMixIns := append(modelRegistry.commonMixins, mi.mixins...)
		for _, mixin := range allMixIns {
			_, ok := mixin.methods.get(methodName)
			if ok {
				found = true
				break
			}
		}
		if !found {
			logging.LogAndPanic(log, "Call to ExtendMethod on non existent method", "model", modelName, "method", methodName)
		}
		// The method exists in a mixin so we create it here with our layer.
		// Bootstrap will take care of putting them the right way round afterwards.
		methInfo = newMethodInfo(mi, methodName, doc, reflect.ValueOf(fnct))
	}
	val := reflect.ValueOf(fnct)
	for i := 1; i < methInfo.methodType.NumIn(); i++ {
		if methInfo.methodType.In(i) != val.Type().In(i) {
			logging.LogAndPanic(log, "Function signature does not match", "model", modelName, "method", methodName,
				"argument", i, "expected", methInfo.methodType.In(i), "received", val.Type().In(i))
		}
	}
	if methInfo.methodType.NumOut() > 0 && methInfo.methodType.Out(0) != val.Type().Out(0) {
		logging.LogAndPanic(log, "Function return type does not match", "model", modelName, "method", methodName,
			"expected", methInfo.methodType.Out(0), "received", val.Type().Out(0))
	}
	if methInfo.methodType.IsVariadic() != val.Type().IsVariadic() {
		logging.LogAndPanic(log, "Variadic mismatch", "model", modelName, "method", methodName,
			"base_is_variadic", methInfo.methodType.IsVariadic(), "ext_is_variadic", val.Type().IsVariadic())
	}
	if !exists {
		mi.methods.set(methodName, methInfo)
		return
	}
	methInfo.addMethodLayer(val, doc)
}

// checkMethodAndFnctType checks whether the given arguments are valid for
// CreateMethod or ExtendMethod
func checkMethodAndFnctType(modelName, methodName string, fnct interface{}) *modelInfo {
	mi := modelRegistry.mustGet(modelName)
	if mi.methods.bootstrapped {
		logging.LogAndPanic(log, "Create/ExtendMethod must be run before BootStrap", "model", modelName, "method", methodName)
	}

	val := reflect.ValueOf(fnct)
	if val.Kind() != reflect.Func {
		logging.LogAndPanic(log, "fnct parameter must be a function", "model", modelName, "method", methodName, "fnct", fnct)
	}
	return mi
}
