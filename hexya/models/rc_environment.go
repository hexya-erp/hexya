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
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
)

// WithEnv returns a copy of the current RecordCollection with the given Environment.
func (rc *RecordCollection) WithEnv(env Environment) *RecordCollection {
	rSet := rc.clone()
	rSet.env = &env
	rSet.applyContexts()
	return rSet
}

// WithContext returns a copy of the current RecordCollection with
// its context extended by the given key and value.
func (rc *RecordCollection) WithContext(key string, value interface{}) *RecordCollection {
	newCtx := rc.env.context.Copy().WithKey(key, value)
	newEnv := *rc.env
	newEnv.context = newCtx
	return rc.WithEnv(newEnv)
}

// WithNewContext returns a copy of the current RecordCollection with its context
// replaced by the given one.
func (rc *RecordCollection) WithNewContext(context *types.Context) *RecordCollection {
	newEnv := *rc.env
	newEnv.context = context
	return rc.WithEnv(newEnv)
}

// Sudo returns a new RecordCollection with the given userId
// or the superuser id if not specified
func (rc *RecordCollection) Sudo(userId ...int64) *RecordCollection {
	uid := security.SuperUserID
	if len(userId) > 0 {
		uid = userId[0]
	}
	newEnv := *rc.env
	newEnv.uid = uid
	return rc.WithEnv(newEnv)
}
