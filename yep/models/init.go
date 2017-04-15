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
	"github.com/inconshreveable/log15"
	"github.com/jmoiron/sqlx"
	"github.com/npiganeau/yep/yep/tools"
	"github.com/npiganeau/yep/yep/tools/logging"
)

var log log15.Logger

func init() {
	log = logging.GetLogger("models")
	sqlx.NameMapper = tools.SnakeCaseString
	// DB drivers
	adapters = make(map[string]dbAdapter)
	registerDBAdapter("postgres", new(postgresAdapter))
	// model registry
	Registry = newModelCollection()
	// declare base and common mixins
	declareBaseMixin()
	declareCommonMixin()
}
