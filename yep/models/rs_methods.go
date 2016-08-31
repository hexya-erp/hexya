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

	"github.com/npiganeau/yep/yep/tools"
)

// Call calls the given method name methName with the given arguments and return the
// result as interface{}.
func (rs RecordCollection) Call(methName string, args ...interface{}) interface{} {
	methInfo, ok := rs.mi.methods.get(methName)
	if !ok {
		tools.LogAndPanic(log, "Unknown method in model", "method", methName, "model", rs.ModelName())
	}
	methLayer := methInfo.topLayer

	rs.callStack = append([]*methodLayer{methLayer}, rs.callStack...)
	return rs.call(methLayer, args...)
}

// call is a wrapper around reflect.Value.Call() to use with interface{} type.
func (rs RecordCollection) call(methLayer *methodLayer, args ...interface{}) interface{} {
	fnVal := methLayer.funcValue
	fnTyp := fnVal.Type()

	rsVal := reflect.ValueOf(rs)
	inVals := []reflect.Value{rsVal}
	methName := fmt.Sprintf("%s.%s()", methLayer.methInfo.mi.name, methLayer.methInfo.name)
	for i := 1; i < fnTyp.NumIn(); i++ {
		if i > len(args) {
			tools.LogAndPanic(log, "Not enough argument while calling method", "model", rs.mi.name, "method", methName, "args", args, "expected", fnTyp.NumIn())
		}
		inVals = append(inVals, reflect.ValueOf(args[i-1]))
	}
	retVal := fnVal.Call(inVals)
	if len(retVal) == 0 {
		return nil
	}
	return retVal[0].Interface()
}

// Super calls the next method Layer.
// This method is meant to be used inside a method layer function to call its parent.
func (rs RecordCollection) Super(args ...interface{}) interface{} {
	if len(rs.callStack) == 0 {
		tools.LogAndPanic(log, "Empty call stack", "model", rs.mi.name)
	}
	methLayer := rs.callStack[0]
	methInfo := methLayer.methInfo
	methLayer = methInfo.getNextLayer(methLayer)
	if methLayer == nil {
		// No parent
		return nil
	}

	rs.callStack[0] = methLayer
	return rs.call(methLayer, args...)
}

// MethodType returns the type of the method given by methName
func (rs RecordCollection) MethodType(methName string) reflect.Type {
	methInfo, ok := rs.mi.methods.get(methName)
	if !ok {
		tools.LogAndPanic(log, "Unknown method in model", "model", rs.ModelName(), "method", methName)
	}
	return methInfo.methodType
}
