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

	user.AddMethod("ComputeAge",
		`ComputeAge is a sample method layer for testing`,
		func(rc *models.RecordCollection) (models.FieldMap, []models.FieldNamer) {
			res := models.FieldMap{
				"Age": rc.Get("Profile").(*models.RecordCollection).Get("Age"),
			}
			return res, []models.FieldNamer{models.FieldName("Age")}
		})

	user.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{String: "Name", Help: "The user's username", Unique: true,
			NoCopy: true, OnChange: user.Methods().MustGet("ComputeDecoratedName")},
		"DecoratedName": models.CharField{Compute: user.Methods().MustGet("ComputeDecoratedName")},
		"Email":         models.CharField{Help: "The user's email address", Size: 100, Index: true},
		"Password":      models.CharField{NoCopy: true},
		"Status": models.IntegerField{JSON: "status_json", GoType: new(int16),
			Default: models.DefaultValue(int16(12))},
		"IsStaff":  models.BooleanField{},
		"IsActive": models.BooleanField{},
		"Profile":  models.Many2OneField{RelationModel: models.Registry.MustGet("Profile")},
		"Age": models.IntegerField{Compute: user.Methods().MustGet("ComputeAge"),
			Inverse: user.Methods().MustGet("InverseSetAge"),
			Depends: []string{"Profile", "Profile.Age"}, Stored: true},
		"Posts":     models.One2ManyField{RelationModel: models.Registry.MustGet("Post"), ReverseFK: "User"},
		"PMoney":    models.FloatField{Related: "Profile.Money"},
		"LastPost":  models.Many2OneField{RelationModel: models.Registry.MustGet("Post"), Embed: true},
		"Email2":    models.CharField{},
		"IsPremium": models.BooleanField{},
		"Nums":      models.IntegerField{GoType: new(int)},
		"Size":      models.FloatField{},
	})

	profile.AddFields(map[string]models.FieldDefinition{
		"Age":      models.IntegerField{GoType: new(int16)},
		"Gender":   models.SelectionField{Selection: types.Selection{"male": "Male", "female": "Female"}},
		"Money":    models.FloatField{},
		"User":     models.Many2OneField{RelationModel: models.Registry.MustGet("User")},
		"BestPost": models.One2OneField{RelationModel: models.Registry.MustGet("Post")},
		"City":     models.CharField{},
		"Country":  models.CharField{},
	})

	post.AddFields(map[string]models.FieldDefinition{
		"User":            models.Many2OneField{RelationModel: models.Registry.MustGet("User")},
		"Title":           models.CharField{},
		"Content":         models.HTMLField{},
		"Tags":            models.Many2ManyField{RelationModel: models.Registry.MustGet("Tag")},
		"BestPostProfile": models.Rev2OneField{RelationModel: models.Registry.MustGet("Profile"), ReverseFK: "BestPost"},
		"Abstract":        models.TextField{},
		"Attachment":      models.BinaryField{},
		"LastRead":        models.DateField{},
	})

	post.Methods().MustGet("Create").Extend("",
		func(rc *models.RecordCollection, data models.FieldMapper) *models.RecordCollection {
			res := rc.Super().Call("Create", data).(*models.RecordCollection)
			return res
		})

	tag.AddFields(map[string]models.FieldDefinition{
		"Name":        models.CharField{Constraint: tag.Methods().MustGet("CheckNameDescription")},
		"BestPost":    models.Many2OneField{RelationModel: models.Registry.MustGet("Post")},
		"Posts":       models.Many2ManyField{RelationModel: models.Registry.MustGet("Post")},
		"Parent":      models.Many2OneField{RelationModel: models.Registry.MustGet("Tag")},
		"Description": models.CharField{Constraint: tag.Methods().MustGet("CheckNameDescription")},
		"Rate":        models.FloatField{Constraint: tag.Methods().MustGet("CheckRate"), GoType: new(float32)},
	})

	addressMI.AddFields(map[string]models.FieldDefinition{
		"Street": models.CharField{GoType: new(string)},
		"Zip":    models.CharField{},
		"City":   models.CharField{},
	})
	profile.InheritModel(addressMI)

	activeMI.AddFields(map[string]models.FieldDefinition{
		"Active": models.BooleanField{},
	})

	models.Registry.MustGet("CommonMixin").InheritModel(activeMI)

	viewModel.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{},
		"City": models.CharField{},
	})
}
