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

import "github.com/npiganeau/yep/yep/models"

func init() {
	user := models.NewModel("User")
	user.AddCharField("Name", models.StringFieldParams{String: "Name", Help: "The user's username", Unique: true})
	user.AddCharField("DecoratedName", models.StringFieldParams{Compute: "computeDecoratedName"})
	user.AddCharField("Email", models.StringFieldParams{Help: "The user's email address", Size: 100, Index: true})
	user.AddCharField("Password", models.StringFieldParams{})
	user.AddIntegerField("Status", models.SimpleFieldParams{JSON: "status_json", GoType: new(int16)})
	user.AddBooleanField("IsStaff", models.SimpleFieldParams{})
	user.AddBooleanField("IsActive", models.SimpleFieldParams{})
	user.AddMany2OneField("Profile", models.ForeignKeyFieldParams{RelationModel: "Profile"})
	user.AddIntegerField("Age", models.SimpleFieldParams{Compute: "computeAge", Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)})
	user.AddOne2ManyField("Posts", models.ReverseFieldParams{RelationModel: "Post", ReverseFK: "User"})
	user.AddFloatField("PMoney", models.FloatFieldParams{Related: "Profile.Money"})
	user.AddMany2OneField("LastPost", models.ForeignKeyFieldParams{RelationModel: "Post", Embed: true})
	user.AddCharField("Email2", models.StringFieldParams{})
	user.AddBooleanField("IsPremium", models.SimpleFieldParams{})
	user.AddIntegerField("Nums", models.SimpleFieldParams{GoType: new(int)})

	profile := models.NewModel("Profile")
	profile.AddIntegerField("Age", models.SimpleFieldParams{GoType: new(int16)})
	profile.AddFloatField("Money", models.FloatFieldParams{})
	profile.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: "User"})
	profile.AddOne2OneField("BestPost", models.ForeignKeyFieldParams{RelationModel: "Post"})
	profile.AddCharField("City", models.StringFieldParams{})
	profile.AddCharField("Country", models.StringFieldParams{})

	post := models.NewModel("Post")
	post.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: "User"})
	post.AddCharField("Title", models.StringFieldParams{})
	post.AddTextField("Content", models.StringFieldParams{})
	post.AddMany2ManyField("Tags", models.Many2ManyFieldParams{RelationModel: "Tag"})

	tag := models.NewModel("Tag")
	tag.AddCharField("Name", models.StringFieldParams{})
	tag.AddMany2OneField("BestPost", models.ForeignKeyFieldParams{RelationModel: "Post"})
	tag.AddMany2ManyField("Posts", models.Many2ManyFieldParams{RelationModel: "Post"})
	tag.AddCharField("Description", models.StringFieldParams{})

	addressMI := models.NewMixinModel("AddressMixIn")
	addressMI.AddCharField("Street", models.StringFieldParams{})
	addressMI.AddCharField("Zip", models.StringFieldParams{})
	addressMI.AddCharField("City", models.StringFieldParams{})
	profile.MixInModel(addressMI)

	activeMI := models.NewMixinModel("ActiveMixIn")
	activeMI.AddBooleanField("Active", models.SimpleFieldParams{})
	models.MixInAllModels(activeMI)

	viewModel := models.NewManualModel("UserView")
	viewModel.AddCharField("Name", models.StringFieldParams{})
	viewModel.AddCharField("City", models.StringFieldParams{})
}
