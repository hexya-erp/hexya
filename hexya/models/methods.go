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

	"github.com/hexya-erp/hexya/hexya/models/security"
)

// A MethodsCollection is a collection of methods for use in a model
type MethodsCollection struct {
	model        *Model
	registry     map[string]*Method
	powerGroups  map[*security.Group]bool
	bootstrapped bool
}

// get returns the Method with the given method name.
func (mc *MethodsCollection) get(methodName string) (*Method, bool) {
	mi, ok := mc.registry[methodName]
	if !ok {
		// We didn't find the method, but maybe it exists in mixins
		miMethod, found := mc.model.findMethodInMixin(methodName)
		if !found || mc.bootstrapped {
			return nil, false
		}
		// The method exists in a mixin so we create it here with our layer.
		// Bootstrap will take care of putting them the right way round afterwards.
		mi = copyMethod(mc.model, miMethod)
		mc.set(methodName, mi)
	}
	return mi, true
}

// MustGet returns the Method of the given method. It panics if the
// method is not found.
func (mc *MethodsCollection) MustGet(methodName string) *Method {
	methInfo, exists := mc.get(methodName)
	if !exists {
		log.Panic("Unknown method in model", "model", mc.model.name, "method", methodName)
	}
	return methInfo
}

// set adds the given Method to the MethodsCollection.
func (mc *MethodsCollection) set(methodName string, methInfo *Method) {
	mc.registry[methodName] = methInfo
}

// AllowAllToGroup grants the given group access to all the methods of this collection
// This method must be called before bootstrap, or will have no effect.
func (mc *MethodsCollection) AllowAllToGroup(group *security.Group) {
	mc.powerGroups[group] = true
}

// RevokeAllFromGroup revokes permissions on all methods given by AllowAllToGroup
// It simply removes the group from the groups allowed on every method, but does
// not change any specific permission granted on a per method basis.
// This method must be called before bootstrap, or will have no effect.
func (mc *MethodsCollection) RevokeAllFromGroup(group *security.Group) {
	delete(mc.powerGroups, group)
}

// newMethodsCollection returns a pointer to a new MethodsCollection
func newMethodsCollection() *MethodsCollection {
	mc := MethodsCollection{
		registry:    make(map[string]*Method),
		powerGroups: make(map[*security.Group]bool),
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
	if m.topLayer != nil {
		m.nextLayer[&ml] = m.topLayer
	}
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
func (m *Method) AllowGroup(group *security.Group, callers ...Methoder) *Method {
	m.Lock()
	defer m.Unlock()
	if len(callers) == 0 {
		m.groups[group] = true
		return m
	}
	for _, caller := range callers {
		m.groupsCallers[callerGroup{caller: caller.Underlying(), group: group}] = true
	}
	return m
}

// Underlying returns the underlysing method data object
func (m *Method) Underlying() *Method {
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

var _ Methoder = new(Method)

// methodLayer is one layer of a method, that is one function defined in a module
type methodLayer struct {
	method    *Method
	mixedIn   bool
	funcValue reflect.Value
	doc       string
}

// copyMethod creates a new method without any method layer for
// the given model by taking data from the given method.
func copyMethod(m *Model, method *Method) *Method {
	return &Method{
		model:         m,
		name:          method.name,
		methodType:    method.methodType,
		nextLayer:     make(map[*methodLayer]*methodLayer),
		groups:        make(map[*security.Group]bool),
		groupsCallers: make(map[callerGroup]bool),
	}
}

// wrapFunctionForMethodLayer take the given fnct Value and wrap it in a
// func(RecordCollection, args...) function Value suitable for use in a
// methodLayer.
func wrapFunctionForMethodLayer(fnctVal reflect.Value) reflect.Value {
	wrapperType := reflect.TypeOf(func(*RecordCollection, ...interface{}) []interface{} { return nil })
	if fnctVal.Type() == wrapperType {
		// fnctVal is already wrapped, we just return it
		return fnctVal
	}
	methodLayerFunction := func(rc *RecordCollection, args ...interface{}) []interface{} {
		argZeroType := fnctVal.Type().In(0)
		argsVals := make([]reflect.Value, len(args)+1)
		argsVals[0] = reflect.New(argZeroType).Elem()
		switch argZeroType {
		case reflect.TypeOf(new(RecordCollection)):
			argsVals[0].Set(reflect.ValueOf(rc))
		default:
			argsVals[0].Field(0).Set(reflect.ValueOf(rc))
		}
		var fns []FieldNamer
		for i := 0; i < fnctVal.Type().NumIn()-1; i++ {
			var arg interface{}
			if len(args) < i+1 && fnctVal.Type().IsVariadic() && i == fnctVal.Type().NumIn()-2 {
				// We handle here the case of a variadic function whose last argument is []FieldNamer
				// and for which we did not have any values but we received some from previous arg conversion.
				argType := fnctVal.Type().In(i + 1)
				if argType.Elem().Name() != "FieldNamer" || fns == nil {
					break
				}
				arg = []FieldNamer{}
				argsVals = append(argsVals, reflect.Value{})
			} else {
				arg = args[i]
			}
			if argFn, ok := arg.([]FieldNamer); ok && fns != nil {
				// Result of previous arg conversion gave fieldNames that we add to the list
				// of this arg if it is actually a []FieldNamer.
				arg = append(argFn, fns...)
			}
			argsVals[i+1], fns = convertFunctionArg(rc, fnctVal.Type().In(i+1), arg)
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

// convertFunctionArg converts the given argument to match that of fnctArgType.
// Second argument is a list of field names to reset if the argument is a FieldMapper
func convertFunctionArg(rc *RecordCollection, fnctArgType reflect.Type, arg interface{}) (reflect.Value, []FieldNamer) {
	var val reflect.Value
	switch at := arg.(type) {
	case Conditioner:
		if fnctArgType.Kind() == reflect.Interface {
			// Target is a Conditioner nothing to change
			return reflect.ValueOf(at), nil
		}
		val = reflect.New(fnctArgType).Elem()
		val.Field(0).Set(reflect.ValueOf(at.Underlying()))
		return val, nil
	case FieldMapper:
		if fnctArgType.Kind() == reflect.Interface {
			// Target is a FieldMapper nothing to change
			return reflect.ValueOf(at), nil
		}
		if fnctArgType.Kind() == reflect.Map {
			// Target is a FieldMap, so we give the FieldMap of this FieldMapper
			return reflect.ValueOf(at.FieldMap()), nil
		}
		// => Target is a struct pointer *h.MyModelData
		if fm, ok := at.(FieldMap); ok {
			// Given arg is a FieldMap, so we map to our struct
			val = reflect.New(fnctArgType.Elem())
			MapToStruct(rc, val.Interface(), fm)
			return val, fm.FieldNames()
		}
		// Given arg is already a struct pointer
		return reflect.ValueOf(arg), nil
	default:
		return reflect.ValueOf(arg), nil
	}
}

// AddMethod creates a new method on given model name and adds the given fnct
// as first layer for this method. Given fnct function must have a RecordSet as
// first argument.
// It returns a pointer to the newly created Method instance.
func (m *Model) AddMethod(methodName, doc string, fnct interface{}) *Method {
	meth := m.AddEmptyMethod(methodName)
	meth.declareMethod(doc, fnct)
	return meth
}

// AddEmptyMethod creates a new method withoud function layer
// The resulting method cannot be called until DeclareMethod is called
func (m *Model) AddEmptyMethod(methodName string) *Method {
	if m.methods.bootstrapped {
		log.Panic("Create/ExtendMethod must be run before BootStrap", "model", m.name, "method", methodName)
	}
	_, exists := m.methods.get(methodName)
	if exists {
		log.Panic("Call to AddMethod with an existing method name", "model", m.name, "method", methodName)
	}
	meth := &Method{
		model:         m,
		name:          methodName,
		nextLayer:     make(map[*methodLayer]*methodLayer),
		groups:        make(map[*security.Group]bool),
		groupsCallers: make(map[callerGroup]bool),
	}
	m.methods.set(methodName, meth)
	return meth
}

// DeclareMethod overrides the given Method by :
// - setting documentation string to doc
// - setting fnct as the first layer
func (m *Method) DeclareMethod(doc string, fnct interface{}) *Method {
	return m.declareMethod(doc, fnct)
}

// declareMethod is the actual implementation of DeclareMethod
// so that it can be called without triggering code generation
func (m *Method) declareMethod(doc string, fnct interface{}) *Method {
	if m.topLayer != nil {
		log.Panic("Call to AddMethod with an existing method name", "model", m.model.name, "method", m.name)
	}
	m.checkMethodAndFnctType(fnct)
	m.doc = doc
	val := reflect.ValueOf(fnct)
	m.addMethodLayer(val, doc)
	m.methodType = val.Type()
	return m
}

// Extend adds the given fnct function as a new layer on this method.
// fnct must be of the same signature as the first layer of this method.
func (m *Method) Extend(doc string, fnct interface{}) *Method {
	m.checkMethodAndFnctType(fnct)
	methInfo := m
	val := reflect.ValueOf(fnct)
	if methInfo.methodType.NumIn() != val.Type().NumIn() {
		log.Panic("Number of args do not match", "model", m.model.name, "method", m.name,
			"no_arguments", val.Type().NumIn(), "expected", methInfo.methodType.NumIn())
	}
	for i := 1; i < methInfo.methodType.NumIn(); i++ {
		if !checkTypesMatch(methInfo.methodType.In(i), val.Type().In(i)) {
			log.Panic("Function signature does not match", "model", m.model.name, "method", m.name,
				"argument", i, "expected", methInfo.methodType.In(i), "received", val.Type().In(i))
		}
	}
	if methInfo.methodType.NumOut() != val.Type().NumOut() {
		log.Panic("Number of returns do not match", "model", m.model.name, "method", m.name,
			"no_arguments", val.Type().NumOut(), "expected", methInfo.methodType.NumOut())
	}
	for i := 1; i < methInfo.methodType.NumOut(); i++ {
		if !checkTypesMatch(methInfo.methodType.Out(i), val.Type().Out(i)) {
			log.Panic("Function return type does not match", "model", m.model.name, "method", m.name,
				"expected", methInfo.methodType.Out(i), "received", val.Type().Out(i))
		}
	}
	if methInfo.methodType.IsVariadic() != val.Type().IsVariadic() {
		log.Panic("Variadic mismatch", "model", m.name, "method", m.name,
			"base_is_variadic", methInfo.methodType.IsVariadic(), "ext_is_variadic", val.Type().IsVariadic())
	}
	methInfo.addMethodLayer(val, doc)
	return methInfo
}

// checkTypesMatch returns true if both given types match
// Two types match if :
// - both types are the same
// - type2 implements type1 or vice-versa
// - if one type is a pointer to a RecordCollection and the second
// one implements the RecordSet interface.
// - if one type is a FieldMap and the other implements FieldMapper
// - if one type is a Condition and the other implements Conditioner
func checkTypesMatch(type1, type2 reflect.Type) bool {
	if type1 == type2 {
		return true
	}
	if type1 == reflect.TypeOf(new(RecordCollection)) && type2.Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
		return true
	}
	if type2 == reflect.TypeOf(new(RecordCollection)) && type1.Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
		return true
	}
	if type1 == reflect.TypeOf(FieldMap{}) && type2.Implements(reflect.TypeOf((*FieldMapper)(nil)).Elem()) {
		return true
	}
	if type2 == reflect.TypeOf(FieldMap{}) && type1.Implements(reflect.TypeOf((*FieldMapper)(nil)).Elem()) {
		return true
	}
	if type2.Kind() == reflect.Interface && type1.Implements(type2) {
		return true
	}
	if type1.Kind() == reflect.Interface && type2.Implements(type1) {
		return true
	}
	return false
}

// findMethodInMixin recursively goes through all mixins
// to find the method with the given name. Returns true if
// it found one, false otherwise.
func (m *Model) findMethodInMixin(methodName string) (*Method, bool) {
	for _, mixin := range m.mixins {
		if method, ok := mixin.methods.get(methodName); ok {
			return method, true
		}
		if method, ok := mixin.findMethodInMixin(methodName); ok {
			return method, true
		}
	}
	return nil, false
}

// checkMethodAndFnctType checks whether the given arguments are valid for
// AddMethod or ExtendMethod. It panics if this is not the case
func (m *Method) checkMethodAndFnctType(fnct interface{}) {
	if m.model.methods.bootstrapped {
		log.Panic("Create/ExtendMethod must be run before BootStrap", "model", m.name, "method", m.name)
	}
	val := reflect.ValueOf(fnct)
	if val.Kind() != reflect.Func {
		log.Panic("fnct parameter must be a function", "model", m.name, "method", m.name, "fnct", fnct)
	}
	funcType := val.Type()
	if funcType.NumIn() == 0 || !funcType.In(0).Implements(reflect.TypeOf((*RecordSet)(nil)).Elem()) {
		log.Panic("Function must have a `RecordSet` as first argument to be used as method.",
			"model", m.model.name, "method", m.name, "type", funcType.In(0))
	}
}
