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

package testmodule

import (
	// Import models definition of this module
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/server"
	_ "github.com/npiganeau/yep/yep/tests/test_module/defs"
)

// Module data declaration
const (
	MODULE_NAME string = "test_module"
)

func init() {
	declareMethods()
	server.RegisterModule(&server.Module{
		Name: MODULE_NAME,
		PostInit: func() {
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				env.Cr().Execute(`DROP VIEW IF EXISTS user_view;
					CREATE VIEW user_view AS (
						SELECT u.id, u.name, p.city, u.active
						FROM "user" u
							LEFT JOIN "profile" p ON p.id = u.profile_id
					)`)
			})
		},
	})
}
