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

// A BaseCaller is a struct that is meant to be embedded inside
// another struct so as to ease implementation of Caller interface.
// It provides helper functions that can be called by the embedding
// struct with all the implementation of the model's callstack.
type BaseCaller struct {
	mi        *modelInfo
	callStack []*methodLayer
}

// Call calls the given method name methName on the given Caller
// with the given arguments and return the result as interface{}.
func (bc *BaseCaller) Call(methName string, rs RecordSet, args ...interface{}) interface{} {
	methInfo, ok := bc.mi.methods.get(methName)
	if !ok {
		tools.LogAndPanic(log, "Unknown method in model", "method", methName, "model", bc.mi.name)
	}
	methLayer := methInfo.topLayer
	return bc.call(methLayer, rs, args...)
}

// Super calls the next method Layer.
// This method is meant to be used inside a method layer function to call its parent.
func (bc *BaseCaller) Super(rs RecordSet, args ...interface{}) interface{} {
	if len(bc.callStack) == 0 {
		tools.LogAndPanic(log, "Empty call stack", "model", bc.mi.name)
	}
	currentLayer := bc.callStack[0]
	methInfo := currentLayer.methInfo
	methLayer := methInfo.getNextLayer(currentLayer)
	if methLayer == nil {
		// No parent
		return nil
	}

	return bc.call(methLayer, rs, args...)
}

// MethodType returns the type of the method given by methName
func (bc *BaseCaller) MethodType(methName string) reflect.Type {
	methInfo, ok := bc.mi.methods.get(methName)
	if !ok {
		tools.LogAndPanic(log, "Unknown method in model", "model", bc.mi.name, "method", methName)
	}
	return methInfo.methodType
}

// call is a wrapper around reflect.Value.Call() to use with interface{} type.
// This is a method
func (bc *BaseCaller) call(methLayer *methodLayer, rs RecordSet, args ...interface{}) interface{} {
	fnVal := methLayer.funcValue
	fnTyp := fnVal.Type()

	rsVal := reflect.ValueOf(rs)
	inVals := []reflect.Value{rsVal}
	for i := 1; i < fnTyp.NumIn(); i++ {
		if i > len(args) {
			tools.LogAndPanic(log, "Not enough argument while calling method", "model", methLayer.methInfo.mi.name, "method", methLayer.methInfo.name, "args", args, "expected", fnTyp.NumIn())
		}
		inVals = append(inVals, reflect.ValueOf(args[i-1]))
	}

	bc.callStack = append([]*methodLayer{methLayer}, bc.callStack...)
	retVal := fnVal.Call(inVals)
	bc.callStack = bc.callStack[1:]

	if len(retVal) == 0 {
		return nil
	}
	return retVal[0].Interface()
}
