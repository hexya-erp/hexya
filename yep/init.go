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

package yep

import (
	_ "github.com/lib/pq"
	"github.com/npiganeau/yep/config"
	"github.com/npiganeau/yep/yep/ir"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/orm"
	"github.com/npiganeau/yep/yep/server"
)

func init() {
	orm.Debug = config.ORM_DEBUG
	orm.RegisterDataBase("default", config.DB_DRIVER, config.DB_SOURCE, 20)
	models.BootStrap(config.DB_FORCE)
	server.LoadInternalResources()
	ir.BootStrap()
	server.PostInit()
}
