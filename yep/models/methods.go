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
	"sync"

	"github.com/npiganeau/yep/yep/models/security"
)

// MethodsCollection is the Method collection
type MethodsCollection struct {
	registry     map[string]*Method
	bootstrapped bool
}

// get returns the Method of the given method.
func (mc *MethodsCollection) get(methodName string) (mi *Method, ok bool) {
	mi, ok = mc.registry[methodName]
	return
}

// MustGet returns the Method of the given method. It panics if the
// method is not found.
func (mc *MethodsCollection) MustGet(methodName string) *Method {
	methInfo, ok := mc.get(methodName)
	if !ok {
		var model string
		for _, f := range mc.registry {
			model = f.model.name
			break
		}
		log.Panic("Unknown method in model", "model", model, "method", methodName)
	}
	return methInfo
}

// set adds the given Method to the MethodsCollection.
func (mc *MethodsCollection) set(methodName string, methInfo *Method) {
	mc.registry[methodName] = methInfo
}

// newMethodsCollection returns a pointer to a new MethodsCollection
func newMethodsCollection() *MethodsCollection {
	mc := MethodsCollection{
		registry: make(map[string]*Method),
	}
	return &mc
}

// A callerGroup is the concatenation of a caller method and a security group
// It is used to lookup execution permissions.
type callerGroup struct {
	caller *Method
	group  *security.Group
}

// A Method is a definition of a model's method
type Method struct {
	sync.RWMutex
	name          string
	model         *Model
	doc           string
	methodType    reflect.Type
	topLayer      *methodLayer
	nextLayer     map[*methodLayer]*methodLayer
	groups        map[*security.Group]bool
	groupsCallers map[callerGroup]bool
}

// addMethodLayer adds the given layer to this Method.
func (m *Method) addMethodLayer(val reflect.Value, doc string) {
	m.Lock()
	defer m.Unlock()
	ml := methodLayer{
		funcValue: wrapFunctionForMethodLayer(val),
		method:    m,
		doc:       doc,
	}
	m.nextLayer[&ml] = m.topLayer
	m.topLayer = &ml
}

func (m *Method) getNextLayer(methodLayer *methodLayer) *methodLayer {
	return m.nextLayer[methodLayer]
}

// invertedLayers returns the list of method layers starting
// from the base methods and going up all inherited layers
func (m *Method) invertedLayers() []*methodLayer {
	var layersInv []*methodLayer
	for cl := m.topLayer; cl != nil; cl = m.getNextLayer(cl) {
		layersInv = append([]*methodLayer{cl}, layersInv...)
	}
	return layersInv
}

// AllowGroup grants the execution permission on this method to the given group
// If callers are defined, then the permission is granted only when this method
// is called from one of the callers, otherwise it is granted from any caller.
func (m *Method) AllowGroup(group *security.Group, callers ...*Method) *Method {
	m.Lock()
	defer m.Unlock()
	if len(callers) == 0 {
		m.groups[group] = true
		return m
	}
	for _, caller := range callers {
		m.groupsCallers[callerGroup{caller: caller, group: group}] = true
	}
	return m
}

// RevokeGroup revokes the execution permission on the method to the given group
// if it has been given previously, otherwise does nothing.
// Note that this methods revokes all permissions, whatever the caller.
func (m *Method) RevokeGroup(group *security.Group) *Method {
	m.Lock()
	defer m.Unlock()
	delete(m.groups, group)
	for cg := range m.groupsCallers {
		if cg.group == group {
			delete(m.groupsCallers, cg)
		}
	}
	return m
}

// methodLayer is one layer of a method, that is one function defined in a module
type methodLayer struct {
	method    *Method
	mixedIn   bool
	funcValue reflect.Value
	doc       string
}

// newMethod creates a new method ref with the given func value as first layer.
// First argument of given function must implement RecordSet.
func newMethod(m *Model, methodName, doc string, val reflect.Value) *Method {
	funcType := val.Type()
	if funcType.NumIn() == 0 || !funcType.In(0).Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
		log.Panic("Function must have a `RecordSet` as first argument to be used as method.", "model", m.name, "method", methodName, "type", funcType.In(0))
	}

	method := Method{
		model:         m,
		name:          methodName,
		methodType:    val.Type(),
		nextLayer:     make(map[*methodLayer]*methodLayer),
		groups:        make(map[*security.Group]bool),
		groupsCallers: make(map[callerGroup]bool),
	}
	method.topLayer = &methodLayer{
		funcValue: wrapFunctionForMethodLayer(val),
		method:    &method,
		doc:       doc,
	}
	return &method
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

// AddMethod creates a new method on given model name and adds the given fnct
// as first layer for this method. Given fnct function must have a RecordSet as
// first argument.
// It returns a pointer to the newly created Method instance.
func (m *Model) AddMethod(methodName, doc string, fnct interface{}) *Method {
	if m.methods.bootstrapped {
		log.Panic("Create/ExtendMethod must be run before BootStrap", "model", m.name, "method", methodName)
	}
	val := reflect.ValueOf(fnct)
	if val.Kind() != reflect.Func {
		log.Panic("fnct parameter must be a function", "model", m.name, "method", methodName, "fnct", fnct)
	}
	_, exists := m.methods.get(methodName)
	if exists {
		log.Panic("Call to AddMethod with an existing method name", "model", m.name, "method", methodName)
	}
	newMethod := newMethod(m, methodName, doc, reflect.ValueOf(fnct))
	m.methods.set(methodName, newMethod)
	return newMethod
}

// ExtendMethod adds the given fnct function as a new layer on the given
// method of the given model.
// fnct must be of the same signature as the first layer of this method.
func (m *Model) ExtendMethod(methodName, doc string, fnct interface{}) {
	m.checkMethodAndFnctType(methodName, fnct)
	methInfo, exists := m.methods.get(methodName)
	if !exists {
		// We didn't find the method, but maybe it exists in mixins
		var found bool
		allMixIns := append(Registry.commonMixins, m.mixins...)
		for _, mixin := range allMixIns {
			_, ok := mixin.methods.get(methodName)
			if ok {
				found = true
				break
			}
		}
		if !found {
			log.Panic("Call to ExtendMethod on non existent method", "model", m.name, "method", methodName)
		}
		// The method exists in a mixin so we create it here with our layer.
		// Bootstrap will take care of putting them the right way round afterwards.
		methInfo = newMethod(m, methodName, doc, reflect.ValueOf(fnct))
	}
	val := reflect.ValueOf(fnct)
	for i := 1; i < methInfo.methodType.NumIn(); i++ {
		if methInfo.methodType.In(i) != val.Type().In(i) {
			log.Panic("Function signature does not match", "model", m.name, "method", methodName,
				"argument", i, "expected", methInfo.methodType.In(i), "received", val.Type().In(i))
		}
	}
	if methInfo.methodType.NumOut() > 0 && methInfo.methodType.Out(0) != val.Type().Out(0) {
		log.Panic("Function return type does not match", "model", m.name, "method", methodName,
			"expected", methInfo.methodType.Out(0), "received", val.Type().Out(0))
	}
	if methInfo.methodType.IsVariadic() != val.Type().IsVariadic() {
		log.Panic("Variadic mismatch", "model", m.name, "method", methodName,
			"base_is_variadic", methInfo.methodType.IsVariadic(), "ext_is_variadic", val.Type().IsVariadic())
	}
	if !exists {
		m.methods.set(methodName, methInfo)
		return
	}
	methInfo.addMethodLayer(val, doc)
}

// checkMethodAndFnctType checks whether the given arguments are valid for
// AddMethod or ExtendMethod
func (m *Model) checkMethodAndFnctType(methodName string, fnct interface{}) {
	if m.methods.bootstrapped {
		log.Panic("Create/ExtendMethod must be run before BootStrap", "model", m.name, "method", methodName)
	}

	val := reflect.ValueOf(fnct)
	if val.Kind() != reflect.Func {
		log.Panic("fnct parameter must be a function", "model", m.name, "method", methodName, "fnct", fnct)
	}
}
