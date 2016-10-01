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
	"github.com/jmoiron/sqlx"
	"github.com/npiganeau/yep/yep/tools"
)

// An Environment stores various contextual data used by the models:
// - the database cursor (current open transaction),
// - the current user ID (for access rights checking)
// - the current context (for storing arbitrary metadata).
// The Environment also stores caches.
type Environment struct {
	cr      *sqlx.Tx
	uid     int64
	context *tools.Context
	cache   *cache
}

// Cr returns a pointer to the transaction of the Environment
func (env Environment) Cr() *sqlx.Tx {
	return env.cr
}

// Uid returns the user id of the Environment
func (env Environment) Uid() int64 {
	return env.uid
}

// Context returns the Context of the Environment
func (env Environment) Context() tools.Context {
	return *env.context
}

// WithContext returns a new Environment with its context updated by ctx.
// If replace is true, then the context is replaced by the given ctx instead of
// being updated.
func (env Environment) WithContext(ctx tools.Context, replace ...bool) Environment {
	if len(replace) > 0 && replace[0] {
		env.context = &ctx
		return env
	}
	newCtx := *env.context
	for key, value := range ctx {
		newCtx[key] = value
	}
	env.context = &newCtx
	return env
}

// Sudo returns a new Environment with the given userId
// or the superuser id if not specified
func (env Environment) Sudo(userId ...int64) Environment {
	if len(userId) > 0 {
		env.uid = userId[0]
	} else {
		env.uid = 1
	}
	return env
}

// NewEnvironment returns a new Environment with the given parameters
// in a new DB transaction.
//
// WARNING: Callers to NewEnvironment should ensure to either commit
// or rollback the returned Environment.Cr() after operation to release
// the database connection.
func NewEnvironment(uid int64, context ...tools.Context) Environment {
	var ctx tools.Context
	if len(context) > 0 {
		ctx = context[0]
	}
	env := Environment{
		cr:      db.MustBegin(),
		uid:     uid,
		context: &ctx,
		cache:   newCache(),
	}
	return env
}

// Pool returns an empty RecordCollection for the given modelName
func (env Environment) Pool(modelName string) RecordCollection {
	return newRecordCollection(env, modelName)
}
