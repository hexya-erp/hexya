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
	"github.com/npiganeau/yep/yep/models/types"
)

// An Environment stores various contextual data used by the models:
// - the database cursor (current open transaction),
// - the current user ID (for access rights checking)
// - the current context (for storing arbitrary metadata).
// The Environment also stores caches.
type Environment struct {
	cr      *Cursor
	uid     int64
	context *types.Context
	cache   *cache
}

// Cr returns a pointer to the Cursor of the Environment
func (env Environment) Cr() *Cursor {
	return env.cr
}

// Uid returns the user id of the Environment
func (env Environment) Uid() int64 {
	return env.uid
}

// Context returns the Context of the Environment
func (env Environment) Context() *types.Context {
	return env.context
}

// Commit the transaction of this environment.
//
// WARNING: Do NOT call Commit on Environment instances that you
// did not create yourself with NewEnvironment. The framework will
// automatically commit the Environment.
func (env Environment) Commit() {
	env.cr.tx.Commit()
}

// Rollback the transaction of this environment.
//
// WARNING: Do NOT call Rollback on Environment instances that you
// did not create yourself with NewEnvironment. Just panic instead
// for the framework to roll back automatically for you.
func (env Environment) Rollback() {
	env.cr.tx.Rollback()
}

// NewEnvironment returns a new Environment with the given parameters
// in a new DB transaction.
//
// WARNING: Callers to NewEnvironment should ensure to either call Commit()
// or Rollback() on the returned Environment after operation to release
// the database connection.
func NewEnvironment(uid int64, context ...types.Context) Environment {
	var ctx types.Context
	if len(context) > 0 {
		ctx = context[0]
	}
	env := Environment{
		cr:      newCursor(db),
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
