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
	"time"

	"github.com/hexya-erp/hexya/hexya/models/security"
)

// Call calls the given method name methName on the given RecordCollection
// with the given arguments and returns (only) the first result as interface{}.
func (rc *RecordCollection) Call(methName string, args ...interface{}) interface{} {
	res := rc.CallMulti(methName, args...)
	if len(res) == 0 {
		return nil
	}
	return res[0]
}

// CallMulti calls the given method name methName on the given RecordCollection
// with the given arguments and return the result as []interface{}.
func (rc *RecordCollection) CallMulti(methName string, args ...interface{}) []interface{} {
	log.Debug("Calling Recordset method", "model", rc.model.name, "method", methName, "ids", rc.ids, "args", args)
	startTime := time.Now()
	methInfo, ok := rc.model.methods.Get(methName)
	if !ok {
		log.Panic("Unknown method in model", "method", methName, "model", rc.model.name)
	}

	methLayer := methInfo.topLayer
	if rc.env.super {
		if !ok {
			log.Panic("Missing layer", "method", methName, "model", rc.model.name)
		}
		methLayer = methInfo.getNextLayer(rc.env.currentLayer)
	}

	newEnv := rc.Env()
	newEnv.super = false
	rSet := rc.WithEnv(newEnv)
	rSet.env.currentLayer = methLayer
	if rc.env.currentLayer != nil && rc.env.currentLayer.method != methInfo {
		rSet.env.previousMethod = rc.env.currentLayer.method
	}
	res := rSet.callMulti(methLayer, args...)
	for i, r := range res {
		switch r.(type) {
		case RecordSet:
			res[i].(RecordSet).Collection().env.currentLayer = rc.env.currentLayer
			res[i].(RecordSet).Collection().env.previousMethod = rc.env.previousMethod
		}
	}
	log.Debug("Called Recordset method", "model", rc.ModelName(), "method", methName, "ids", rc.ids, "duration", time.Now().Sub(startTime), "args", args)
	return res
}

// Super returns a RecordSet with a modified callstack so that call to the current
// method will execute the next method layer.
//
// This method is meant to be used inside a method layer function to call its parent,
// such as:
//
//    func (rs models.RecordCollection) MyMethod() string {
//        res := rs.Super().MyMethod()
//        res += " ok!"
//        return res
//    }
//
// Calls to a different method than the current method will call its next layer only
// if the current method has been called from a layer of the other method. Otherwise,
// it will be the same as calling the other method directly.
func (rc *RecordCollection) Super() *RecordCollection {
	newEnv := rc.Env()
	newEnv.super = true
	return rc.WithEnv(newEnv)
}

// MethodType returns the type of the method given by methName
func (rc *RecordCollection) MethodType(methName string) reflect.Type {
	methInfo, ok := rc.model.methods.Get(methName)
	if !ok {
		log.Panic("Unknown method in model", "model", rc.model.name, "method", methName)
	}
	return methInfo.methodType
}

// callMulti is a wrapper around reflect.Value.Call() to use with interface{} type.
func (rc *RecordCollection) callMulti(methLayer *methodLayer, args ...interface{}) []interface{} {
	rc.CheckExecutionPermission(methLayer.method)
	inVals := make([]reflect.Value, len(args)+1)
	inVals[0] = reflect.ValueOf(rc)
	for i, arg := range args {
		inVals[i+1] = reflect.ValueOf(arg)
	}

	retVal := methLayer.funcValue.Call(inVals)[0]

	res := make([]interface{}, retVal.Len())
	for i := 0; i < retVal.Len(); i++ {
		res[i] = retVal.Index(i).Interface()
	}
	return res
}

// CheckExecutionPermission panics if the current user is not allowed to
// execute the given method.
//
// If dontPanic is false, this function will panic, otherwise it returns true
// if the user has the execution permission and false otherwise.
func (rc *RecordCollection) CheckExecutionPermission(method *Method, dontPanic ...bool) bool {
	caller := rc.env.previousMethod
	if rc.env.currentLayer != nil && rc.env.currentLayer.method != method {
		caller = rc.env.currentLayer.method
	}

	if caller == method {
		// We are calling Super on the same method, so it's ok
		return true
	}
	userGroups := security.Registry.UserGroups(rc.env.uid)
	for group := range userGroups {
		if method.groups[group] {
			return true
		}
		if caller == nil {
			continue
		}
		if method.groupsCallers[callerGroup{caller: caller, group: group}] {
			return true
		}
	}
	if len(dontPanic) > 0 && dontPanic[0] {
		return false
	}
	log.Panic("You are not allowed to execute this method", "model", rc.ModelName(),
		"method", fmt.Sprintf("%s.%s()", method.model.name, method.name), "uid", rc.env.uid,
		"methodCaller", fmt.Sprintf("%s.%s()", caller.model.name, caller.name))
	// Unreachable
	return false
}
