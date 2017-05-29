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

	"github.com/hexya-erp/hexya/hexya/models/security"
)

// Call calls the given method name methName on the given RecordCollection
// with the given arguments and returns (only) the first result as interface{}.
func (rc RecordCollection) Call(methName string, args ...interface{}) interface{} {
	res := rc.CallMulti(methName, args...)
	if len(res) == 0 {
		return nil
	}
	return res[0]
}

// CallMulti calls the given method name methName on the given RecordCollection
// with the given arguments and return the result as []interface{}.
func (rc RecordCollection) CallMulti(methName string, args ...interface{}) []interface{} {
	methInfo, ok := rc.model.methods.get(methName)
	if !ok {
		log.Panic("Unknown method in model", "method", methName, "model", rc.model.name)
	}
	methLayer := rc.getExistingLayer(methInfo)
	rSet := rc
	if methLayer == nil {
		methLayer = methInfo.topLayer
		newEnv := rc.Env()
		newEnv.callStack = append([]*methodLayer{methLayer}, newEnv.callStack...)
		rSet = rSet.WithEnv(newEnv)
	}
	return rSet.callMulti(methLayer, args...)
}

// getExistingLayer returns the first methodLayer in this RecordCollection call stack
// that matches with the given method. Returns nil, if none was found.
func (rc RecordCollection) getExistingLayer(methInfo *Method) *methodLayer {
	for _, ml := range rc.env.callStack {
		if ml.method == methInfo {
			return ml
		}
	}
	return nil
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
func (rc RecordCollection) Super() RecordCollection {
	if len(rc.env.callStack) == 0 {
		log.Panic("Empty call stack", "model", rc.model.name)
	}
	currentLayer := rc.env.callStack[0]
	methInfo := currentLayer.method
	methLayer := methInfo.getNextLayer(currentLayer)
	if methLayer == nil {
		// No parent
		log.Panic("Called Super() on a base method", "model", rc.model.name, "method", methInfo.name)
	}
	newEnv := rc.Env()
	newEnv.callStack = append([]*methodLayer{methLayer}, newEnv.callStack...)
	return rc.WithEnv(newEnv)
}

// MethodType returns the type of the method given by methName
func (rc RecordCollection) MethodType(methName string) reflect.Type {
	methInfo, ok := rc.model.methods.get(methName)
	if !ok {
		log.Panic("Unknown method in model", "model", rc.model.name, "method", methName)
	}
	return methInfo.methodType
}

// callMulti is a wrapper around reflect.Value.Call() to use with interface{} type.
func (rc RecordCollection) callMulti(methLayer *methodLayer, args ...interface{}) []interface{} {
	rc.checkExecutionPermission(methLayer.method)
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

// checkExecutionPermission panics if the current user is not allowed to
// execute the given method
func (rc RecordCollection) checkExecutionPermission(method *Method) {
	var caller *Method
	if len(rc.env.callStack) > 1 {
		caller = rc.env.callStack[1].method
	}
	if caller == method {
		// We are calling Super on the same method, so it's ok
		return
	}
	userGroups := security.Registry.UserGroups(rc.env.uid)
	for group := range userGroups {
		if method.groups[group] {
			return
		}
		if caller == nil {
			continue
		}
		if method.groupsCallers[callerGroup{caller: caller, group: group}] {
			return
		}
	}
	log.Panic("You are not allowed to execute this method", "model", rc.ModelName(), "method", method.name, "uid", rc.env.uid)
}
