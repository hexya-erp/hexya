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

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"os"
)

func init() {
	if os.Getenv("YEP_DB_DRIVER") == "" {
		return
	}
	models.CreateModel("Test__User")
	models.ExtendModel("Test__User", new(struct {
		ID            int64
		UserName      string `yep:"unique;string(Name);help(The user's username)"`
		DecoratedName string `yep:"compute(computeDecoratedName)"`
		Email         string `yep:"size(100);help(The user's email address);index"`
		Password      string
		Status        int16 `yep:"json(status_json)"`
		IsStaff       bool
		IsActive      bool
		Profile       pool.Test__ProfileSet `yep:"type(many2one)"` //;on_delete(set_null)"`
		Age           int16                 `yep:"compute(computeAge);store;depends(Profile.Age,Profile)"`
		Posts         pool.Test__PostSet    `yep:"type(one2many);fk(User)"`
		Nums          int
		PMoney        float64            `yep:"related(Profile.Money)"`
		LastPost      pool.Test__PostSet `yep:"type(many2one);embed"`
	}))

	models.ExtendModel("Test__User", new(struct {
		Email2    string
		IsPremium bool
	}))

	models.CreateModel("Test__Profile")
	models.ExtendModel("Test__Profile", new(struct {
		Age      int16
		Money    float64
		User     pool.Test__UserSet `yep:"type(many2one)"`
		BestPost pool.Test__PostSet `yep:"type(one2one)"`
	}))
	models.ExtendModel("Test__Profile", new(struct {
		City    string
		Country string
	}))

	models.CreateModel("Test__Post")
	models.ExtendModel("Test__Post", new(struct {
		User    pool.Test__UserSet `yep:"type(many2one)"`
		Title   string
		Content string            `yep:"type(text)"`
		Tags    pool.Test__TagSet `yep:"type(many2many)"`
	}))

	models.CreateModel("Test__Tag")
	models.ExtendModel("Test__Tag", new(struct {
		Name     string
		BestPost pool.Test__PostSet `yep:"type(many2one)"`
		Posts    pool.Test__PostSet `yep:"type(many2many)"`
	}))

	models.ExtendModel("Test__Tag", new(struct {
		Description string
	}))
}
