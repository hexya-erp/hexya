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
	"sync"
)

var methodsCache = &_methodsCache{
	cache: make(map[method]*methodInfo),
	cacheByFunc: make(map[reflect.Value]*methodLayer),
}

// method is the key to find a method in the methodsCache.
type method struct {
	modelName string
	name      string
}

// methodsCache is the methodInfo collection
type _methodsCache struct {
	sync.RWMutex
	cache       map[method]*methodInfo
	cacheByFunc map[reflect.Value]*methodLayer
	done        bool
}

/*
get returns the methodInfo of the given method.
*/
func (mc *_methodsCache) get(ref method) (mi *methodInfo, ok bool) {
	mi, ok = mc.cache[ref]
	return
}

/*
getByFunc returns the methodInfo that includes the given function as a layer.
*/
func (mc *_methodsCache) getByFunc(fnctPtr interface{}) (ml *methodLayer, ok bool) {
	ml, ok = mc.cacheByFunc[reflect.ValueOf(fnctPtr).Elem()]
	return
}

/*
set adds the given methodInfo to the methodsCache.
*/
func (mc *_methodsCache) set(ref method, methInfo *methodInfo) {
	mc.cache[ref] = methInfo
	mc.cacheByFunc[methInfo.topLayer.funcValue] = methInfo.topLayer
}

func (mc *_methodsCache) addLayer(fnVal reflect.Value, methLayer *methodLayer) {
	mc.cacheByFunc[fnVal] = methLayer
}

// methodInfo is a RecordSet method info
type methodInfo struct {
	ref        method
	methodType reflect.Type
	topLayer   *methodLayer
	nextLayer  map[*methodLayer]*methodLayer
}

/*
addMethodLayer adds the given layer to this methodInfo.
*/
func (methInfo *methodInfo) addMethodLayer(val reflect.Value) {
	ml := methodLayer{
		funcValue: val,
		methInfo:  methInfo,
	}
	methInfo.nextLayer[&ml] = methInfo.topLayer
	methInfo.topLayer = &ml
	methodsCache.addLayer(ml.funcValue, &ml)
}

func (methInfo *methodInfo) getNextLayer(methodLayer *methodLayer) *methodLayer {
	return methInfo.nextLayer[methodLayer]
}

// methodLayer is one layer of a method, that is one function defined in a module
type methodLayer struct {
	methInfo  *methodInfo
	funcValue reflect.Value
}

/*
newMethodInfo creates a new method ref with the given func value as first layer.
First argument of given function must implement RecordSet.
*/
func newMethodInfo(ref method, val reflect.Value) *methodInfo {
	funcType := val.Type()
	if funcType.NumIn() == 0 || funcType.In(0) != reflect.TypeOf((*RecordSet)(nil)).Elem() {
		panic(fmt.Errorf("Function must have `RecordSet` as first argument to be used as method."))
	}

	methInfo := methodInfo{
		ref:        ref,
		methodType: val.Type(),
		nextLayer:  make(map[*methodLayer]*methodLayer),
	}
	methInfo.topLayer = &methodLayer{
		funcValue: val,
		methInfo:  &methInfo,
	}
	return &methInfo
}

/*
CreateMethod creates a new method on given model name and adds the
given fnctPtr a first layer for this method. This function must have a RecordSet as
first argument.
*/
func CreateMethod(modelName, name string, fnctPtr interface{}) {
	if methodsCache.done {
		panic(fmt.Errorf("CreateMethod must be run before BootStrap"))
	}

	val := reflect.ValueOf(fnctPtr)
	ind := reflect.Indirect(val)
	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Func {
		panic(fmt.Errorf("CreateMethod: `fnctPtr` must be a pointer to a function"))
	}
	ref := method{
		modelName: modelName,
		name:      name,
	}
	methodsCache.set(ref, newMethodInfo(ref, ind))
}

/*
ExtendMethod adds a methodLayer on the top of the given method.
modelName and name identify the method to modify and fnctPtr is a pointer
to the function to add. This function must have the same signature
as the one used to create the method.
*/
func ExtendMethod(modelName, name string, fnctPtr interface{}) {
	ref := method{
		modelName: modelName,
		name:      name,
	}
	methInfo, ok := methodsCache.get(ref)
	if !ok {
		panic(fmt.Errorf("Unknwon method %+v", ref))
	}
	val := reflect.ValueOf(fnctPtr)
	ind := reflect.Indirect(val)
	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Func {
		panic(fmt.Errorf("ExtendMethod: `fnctPtr` must be a pointer to a function"))
	}
	if methInfo.methodType != ind.Type() {
		panic(fmt.Errorf("Function signature does not match. Received: %s, Expected: %s", methInfo.methodType, ind.Type()))
	}
	methInfo.addMethodLayer(ind)
}
