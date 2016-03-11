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

import "github.com/npiganeau/yep/yep/orm"

/*
Context is a map of objects that is passed along from function to function
during a transaction.
*/
type Context map[string]interface{}

/*
Environment holds the context data for a transaction.
*/
type Environment interface {
	Cr() orm.Ormer
	Uid() int64
	Context() Context
	WithContext(ctx Context, replace ...bool) Environment
	Sudo(...int64) Environment
	Pool(interface{}) RecordSet
}

/*
envStruct implements Environment. It is immutable.
*/
type envStruct struct {
	cr      orm.Ormer
	uid     int64
	context Context
}

/*
Cr return the Ormer of the Environment
*/
func (env envStruct) Cr() orm.Ormer {
	return env.cr
}

/*
Uid returns the user id of the Environment
*/
func (env envStruct) Uid() int64 {
	return env.uid
}

/*
Context returns the Context of the Environment
*/
func (env envStruct) Context() Context {
	return env.context
}

/*
WithContext returns a new Environment with its context updated by ctx.
If replace is true, then the context is replaced by the given ctx instead of
being updated.
*/
func (env envStruct) WithContext(ctx Context, replace ...bool) Environment {
	if len(replace) > 0 && replace[0] {
		return NewEnvironment(env.cr, env.uid, ctx)
	}
	newCtx := env.context
	for key, value := range ctx {
		newCtx[key] = value
	}
	return NewEnvironment(env.cr, env.uid, newCtx)
}

/*
Sudo returns a new Environment with the given userId or the superuser id if not specified
*/
func (env envStruct) Sudo(userId ...int64) Environment {
	var uid int64
	if len(userId) > 0 {
		uid = userId[0]
	} else {
		uid = 1
	}
	return NewEnvironment(env.cr, uid, env.context)
}

/*
NewEnvironment returns a new environment with the given parameters.
*/
func NewEnvironment(cr orm.Ormer, uid int64, context ...Context) Environment {
	var ctx Context
	if len(context) > 0 {
		ctx = context[0]
	}
	env := envStruct{
		cr:      cr,
		uid:     uid,
		context: ctx,
	}
	return env
}

/*
Pool returns an empty RecordSet from the given table name string or struct pointer
*/
func (env envStruct) Pool(tableNameOrStructPtr interface{}) RecordSet {
	return NewRecordSet(env, tableNameOrStructPtr)
}

var _ Environment = envStruct{}
