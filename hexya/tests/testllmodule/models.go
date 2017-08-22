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

package testllmodule

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
)

func declareModels() {
	user := models.NewModel("User")
	profile := models.NewModel("Profile")
	post := models.NewModel("Post")
	tag := models.NewModel("Tag")
	addressMI := models.NewMixinModel("AddressMixIn")
	activeMI := models.NewMixinModel("ActiveMixIn")
	viewModel := models.NewManualModel("UserView")

	user.AddCharField("Name", models.StringFieldParams{String: "Name", Help: "The user's username", Unique: true})
	user.AddCharField("Email", models.StringFieldParams{Help: "The user's email address", Size: 100, Index: true})
	user.AddCharField("Password", models.StringFieldParams{NoCopy: true})
	user.AddIntegerField("Status", models.SimpleFieldParams{JSON: "status_json", GoType: new(int16)})
	user.AddBooleanField("IsStaff", models.SimpleFieldParams{})
	user.AddBooleanField("IsActive", models.SimpleFieldParams{})
	user.AddMany2OneField("Profile", models.ForeignKeyFieldParams{RelationModel: models.Registry.MustGet("Profile")})
	user.AddIntegerField("Age", models.SimpleFieldParams{Compute: user.Methods().MustGet("ComputeAge"),
		Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)})
	user.AddOne2ManyField("Posts", models.ReverseFieldParams{RelationModel: models.Registry.MustGet("Post"),
		ReverseFK: "User"})
	user.AddFloatField("PMoney", models.FloatFieldParams{Related: "Profile.Money"})
	user.AddMany2OneField("LastPost", models.ForeignKeyFieldParams{
		RelationModel: models.Registry.MustGet("Post"), Embed: true})
	user.AddCharField("Email2", models.StringFieldParams{})
	user.AddBooleanField("IsPremium", models.SimpleFieldParams{})
	user.AddIntegerField("Nums", models.SimpleFieldParams{GoType: new(int)})

	user.AddMethod("computeAge",
		`ComputeAge is a sample method layer for testing`,
		func(rc models.RecordCollection) (models.FieldMap, []models.FieldNamer) {
			res := models.FieldMap{
				"Age": rc.Get("Profile").(models.RecordCollection).Get("Age"),
			}
			return res, []models.FieldNamer{models.FieldName("Age")}
		})

	profile.AddIntegerField("Age", models.SimpleFieldParams{GoType: new(int16)})
	profile.AddFloatField("Money", models.FloatFieldParams{})
	profile.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: models.Registry.MustGet("User")})
	profile.AddSelectionField("Gender", models.SelectionFieldParams{Selection: types.Selection{"male": "Male", "female": "Female"}})
	profile.AddOne2OneField("BestPost", models.ForeignKeyFieldParams{RelationModel: models.Registry.MustGet("Post")})
	profile.AddCharField("City", models.StringFieldParams{})
	profile.AddCharField("Country", models.StringFieldParams{})

	post.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: models.Registry.MustGet("User")})
	post.AddCharField("Title", models.StringFieldParams{})
	post.AddTextField("Content", models.StringFieldParams{})
	post.AddMany2ManyField("Tags", models.Many2ManyFieldParams{RelationModel: models.Registry.MustGet("Tag")})

	post.Methods().MustGet("Create").Extend("",
		func(rc models.RecordCollection, data models.FieldMapper) models.RecordCollection {
			res := rc.Super().Call("Create", data).(models.RecordCollection)
			return res
		})

	tag.AddCharField("Name", models.StringFieldParams{})
	tag.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: models.Registry.MustGet("Tag")})
	tag.AddMany2OneField("BestPost", models.ForeignKeyFieldParams{RelationModel: models.Registry.MustGet("Post")})
	tag.AddMany2ManyField("Posts", models.Many2ManyFieldParams{RelationModel: models.Registry.MustGet("Post")})
	tag.AddCharField("Description", models.StringFieldParams{})

	addressMI.AddCharField("Street", models.StringFieldParams{})
	addressMI.AddCharField("Zip", models.StringFieldParams{})
	addressMI.AddCharField("City", models.StringFieldParams{})
	profile.InheritModel(addressMI)

	activeMI.AddBooleanField("Active", models.SimpleFieldParams{})
	models.Registry.MustGet("CommonMixin").InheritModel(activeMI)

	viewModel.AddCharField("Name", models.StringFieldParams{})
	viewModel.AddCharField("City", models.StringFieldParams{})
}
