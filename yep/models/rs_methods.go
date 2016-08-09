/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package models

import (
	"fmt"
	"reflect"

	"github.com/npiganeau/yep/yep/tools"
)

// Call calls the given method name methName with the given arguments and return the
// result as interface{}.
func (rs RecordSet) Call(methName string, args ...interface{}) interface{} {
	methInfo, ok := rs.mi.methods.get(methName)
	if !ok {
		tools.LogAndPanic(log, "Unknown method in model", "method", methName, "model", rs.ModelName())
	}
	methLayer := methInfo.topLayer

	rs.callStack = append([]*methodLayer{methLayer}, rs.callStack...)
	return rs.call(methLayer, args...)
}

// call is a wrapper around reflect.Value.Call() to use with interface{} type.
func (rs RecordSet) call(methLayer *methodLayer, args ...interface{}) interface{} {
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

// Super calls the next method Layer after the given funcPtr.
// This method is meant to be used inside a method layer function to call its parent.
func (rs RecordSet) Super(args ...interface{}) interface{} {
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
func (rs RecordSet) MethodType(methName string) reflect.Type {
	methInfo, ok := rs.mi.methods.get(methName)
	if !ok {
		tools.LogAndPanic(log, "Unknown method in model", "model", rs.ModelName(), "method", methName)
	}
	return methInfo.methodType
}
