// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
	"github.com/hexya-erp/pool/q"
)

func TestContextedFieldsAPI(t *testing.T) {
	t.Run("Testing translated fields with the type safe API", func(t *testing.T) {
		assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			var tag m.TagSet
			t.Run("Creating a record sets the value for the default context", func(t *testing.T) {
				tag = h.Tag().Create(env, h.Tag().NewData().
					SetName("Contexted Tag").
					SetNote("Translated note"))
				assert.EqualValues(t, tag.Note(), "Translated note")
				assert.EqualValues(t, tag.GetTranslations(h.Tag().Fields().Note()), map[string]string{
					"": "Translated note",
				})
			})
			t.Run("Writing in a language should not modify the other languages", func(t *testing.T) {
				tag.WithContext("lang", "fr_FR").SetNote("Note traduite")
				assert.EqualValues(t, tag.Note(), "Translated note")
				assert.EqualValues(t, tag.WithContext("lang", "fr_FR").Note(), "Note traduite")
				assert.EqualValues(t, tag.WithContext("lang", "de_DE").Note(), "Translated note")

				tag.WithContext("lang", "de_DE").SetNote("Übersetzte Notiz")
				assert.EqualValues(t, tag.WithContext("lang", "fr_FR").Note(), "Note traduite")
				assert.EqualValues(t, tag.WithContext("lang", "de_DE").Note(), "Übersetzte Notiz")
				assert.EqualValues(t, tag.WithContext("lang", "es_ES").Note(), "Translated note")
			})
			t.Run("Writing with Write should behave the same way", func(t *testing.T) {
				tag.WithContext("lang", "fr_FR").Write(h.Tag().NewData().SetNote("Nouvelle note"))
				assert.EqualValues(t, tag.WithContext("lang", "fr_FR").Note(), "Nouvelle note")
				assert.EqualValues(t, tag.WithContext("lang", "de_DE").Note(), "Übersetzte Notiz")
				assert.EqualValues(t, tag.Note(), "Translated note")
				assert.EqualValues(t, tag.GetTranslations(h.Tag().Fields().Note()), map[string]string{
					"":      "Translated note",
					"de_DE": "Übersetzte Notiz",
					"fr_FR": "Nouvelle note",
				})
			})
			t.Run("Searching should be done on the value of the current language", func(t *testing.T) {
				tagFR := h.Tag().NewSet(env).WithContext("lang", "fr_FR").
					Search(q.Tag().Note().Equals("Nouvelle note"))
				assert.EqualValues(t, tagFR.Len(), 1)
				assert.True(t, tagFR.Equals(tag))
				assert.True(t, h.Tag().NewSet(env).WithContext("lang", "de_DE").
					Search(q.Tag().Note().Equals("Nouvelle note")).IsEmpty())
				assert.EqualValues(t, h.Tag().Search(env, q.Tag().Note().Equals("Translated note")).Len(), 1)
			})
			t.Run("Getting and setting all translations at once", func(t *testing.T) {
				tag.SetTranslations(h.Tag().Fields().Note(), map[string]string{
					"fr_FR": "Note en français",
					"it_IT": "Nota tradotta",
				})
				// The languages that are not in the given map are left unchanged
				assert.EqualValues(t, tag.GetTranslations(h.Tag().Fields().Note()), map[string]string{
					"":      "Translated note",
					"de_DE": "Übersetzte Notiz",
					"fr_FR": "Note en français",
					"it_IT": "Nota tradotta",
				})
				assert.EqualValues(t, tag.WithContext("lang", "it_IT").Note(), "Nota tradotta")
				// Languages that are not translated fall back on the default value
				assert.EqualValues(t, tag.WithContext("lang", "pt_PT").Note(), "Translated note")
			})
			t.Run("Forcing the default contexts should return the context-less value", func(t *testing.T) {
				tagFR := tag.WithContext("lang", "fr_FR")
				assert.EqualValues(t, tagFR.Note(), "Note en français")
				assert.EqualValues(t, tagFR.WithContext("hexya_default_contexts", true).Note(), "Translated note")
			})
			t.Run("Copying a record should copy all its translations", func(t *testing.T) {
				tagCopy := tag.Copy(h.Tag().NewData().SetName("Contexted Tag Copy"))
				assert.EqualValues(t, tagCopy.Note(), "Translated note")
				assert.EqualValues(t, tagCopy.WithContext("lang", "fr_FR").Note(), "Note en français")
				assert.EqualValues(t, tagCopy.WithContext("lang", "it_IT").Note(), "Nota tradotta")
				tagCopy.Unlink()
			})
			t.Run("Creating a record in a language should set the default value too", func(t *testing.T) {
				newTag := h.Tag().NewSet(env).WithContext("lang", "fr_FR").
					Create(h.Tag().NewData().
						SetName("Tag en français").
						SetNote("Note en français"))
				assert.EqualValues(t, newTag.Note(), "Note en français")
				assert.EqualValues(t, newTag.WithContext("lang", "de_DE").Note(), "Note en français")
				newTag.WithContext("lang", "de_DE").SetNote("Deutsche Notiz")
				assert.EqualValues(t, newTag.WithContext("lang", "de_DE").Note(), "Deutsche Notiz")
				assert.EqualValues(t, newTag.WithContext("lang", "fr_FR").Note(), "Note en français")
				newTag.Unlink()
			})
			t.Run("Unlinking a record should remove all its translations", func(t *testing.T) {
				tag.Unlink()
				assert.True(t, h.Tag().Search(env, q.Tag().Name().Equals("Contexted Tag")).IsEmpty())
			})
		}))
	})
	t.Run("Testing fields with several contexts with the type safe API", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			tag := h.Tag().Create(env, h.Tag().NewData().
				SetName("Multi Contexted Tag").
				SetSlogan("Chair"))
			tag.WithContext("lang", "fr_FR").SetSlogan("Siège")
			tag.WithContext("company", "3").SetSlogan("Seat")
			tag.WithContext("company", "3").WithContext("lang", "fr_FR").SetSlogan("Fauteuil")
			t.Run("Resolution should try each subset of contexts", func(t *testing.T) {
				assert.EqualValues(t, tag.Slogan(), "Chair")
				assert.EqualValues(t, tag.WithContext("lang", "fr_FR").Slogan(), "Siège")
				assert.EqualValues(t, tag.WithContext("company", "3").Slogan(), "Seat")
				assert.EqualValues(t, tag.WithContext("company", "3").WithContext("lang", "fr_FR").Slogan(),
					"Fauteuil")
				// The per company value is used when the language is unknown
				assert.EqualValues(t, tag.WithContext("company", "3").WithContext("lang", "de_DE").Slogan(),
					"Seat")
				// The generic translation is used for the other companies
				assert.EqualValues(t, tag.WithContext("company", "7").WithContext("lang", "fr_FR").Slogan(),
					"Siège")
				assert.EqualValues(t, tag.WithContext("company", "7").WithContext("lang", "de_DE").Slogan(),
					"Chair")
			})
			t.Run("Searching should operate on the resolved value", func(t *testing.T) {
				tagFR := h.Tag().NewSet(env).WithContext("company", "3").WithContext("lang", "fr_FR").
					Search(q.Tag().Slogan().Equals("Fauteuil"))
				assert.EqualValues(t, tagFR.Len(), 1)
				assert.True(t, tagFR.Equals(tag))
				assert.True(t, h.Tag().NewSet(env).WithContext("lang", "fr_FR").
					Search(q.Tag().Slogan().Equals("Fauteuil")).IsEmpty())
				assert.EqualValues(t, h.Tag().NewSet(env).WithContext("lang", "fr_FR").
					Search(q.Tag().Slogan().Equals("Siège")).Len(), 1)
			})
			t.Run("Reading a contexted field with First and All", func(t *testing.T) {
				data := tag.WithContext("lang", "fr_FR").First()
				assert.True(t, data.HasSlogan())
				assert.EqualValues(t, data.Slogan(), "Siège")
				allData := tag.WithContext("company", "3").All()
				assert.Len(t, allData, 1)
				assert.EqualValues(t, allData[0].Slogan(), "Seat")
			})
		}))
	})
	t.Run("Testing non string contexted fields with the type safe API", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			tag := h.Tag().Create(env, h.Tag().NewData().
				SetName("Priced Tag").
				SetPrice(12.34))
			assert.EqualValues(t, tag.Price(), 12.34)
			tag.WithContext("company", "3").SetPrice(56.78)
			assert.EqualValues(t, tag.Price(), 12.34)
			assert.EqualValues(t, tag.WithContext("company", "3").Price(), 56.78)
			assert.EqualValues(t, tag.WithContext("company", "7").Price(), 12.34)
			assert.EqualValues(t, h.Tag().NewSet(env).WithContext("company", "3").
				Search(q.Tag().Price().Greater(50)).Len(), 1)
			assert.True(t, h.Tag().Search(env, q.Tag().Price().Greater(50)).IsEmpty())
		}))
	})
	t.Run("Testing contexted fields on a related model", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			resume := h.Resume().Create(env, h.Resume().NewData().
				SetExperience("Professional experience"))
			user := h.User().Create(env, h.User().NewData().
				SetName("Contexted Resume User").
				SetResume(resume))
			resume.WithContext("lang", "fr_FR").SetExperience("Expérience professionnelle")
			assert.EqualValues(t, resume.Experience(), "Professional experience")
			assert.EqualValues(t, resume.WithContext("lang", "fr_FR").Experience(),
				"Expérience professionnelle")
			// The contexted value is also resolved when read through the embedded field
			assert.EqualValues(t, user.Experience(), "Professional experience")
			assert.EqualValues(t, user.WithContext("lang", "fr_FR").Experience(),
				"Expérience professionnelle")
			assert.EqualValues(t, user.WithContext("lang", "de_DE").Experience(),
				"Professional experience")
			assert.EqualValues(t, user.Resume().GetTranslations(h.Resume().Fields().Experience()),
				map[string]string{
					"":      "Professional experience",
					"fr_FR": "Expérience professionnelle",
				})
			t.Run("Writing through the embedded field targets the current context", func(t *testing.T) {
				user.WithContext("lang", "de_DE").SetExperience("Berufserfahrung")
				assert.EqualValues(t, user.WithContext("lang", "de_DE").Experience(), "Berufserfahrung")
				assert.EqualValues(t, user.WithContext("lang", "fr_FR").Experience(),
					"Expérience professionnelle")
				assert.EqualValues(t, user.Experience(), "Professional experience")
				assert.EqualValues(t, resume.WithContext("lang", "de_DE").Experience(), "Berufserfahrung")
				assert.EqualValues(t, resume.Experience(), "Professional experience")
			})
			t.Run("Writing through the embedded field with Write", func(t *testing.T) {
				user.WithContext("lang", "it_IT").Write(h.User().NewData().
					SetExperience("Esperienza professionale"))
				assert.EqualValues(t, user.WithContext("lang", "it_IT").Experience(),
					"Esperienza professionale")
				assert.EqualValues(t, user.Experience(), "Professional experience")
				assert.EqualValues(t, resume.GetTranslations(h.Resume().Fields().Experience()),
					map[string]string{
						"":      "Professional experience",
						"de_DE": "Berufserfahrung",
						"fr_FR": "Expérience professionnelle",
						"it_IT": "Esperienza professionale",
					})
				// The values are actually written in the database
				resume.Collection().InvalidateCache()
				assert.EqualValues(t, user.WithContext("lang", "de_DE").Experience(), "Berufserfahrung")
				assert.EqualValues(t, user.WithContext("lang", "it_IT").Experience(),
					"Esperienza professionale")
				assert.EqualValues(t, user.Experience(), "Professional experience")
			})
			t.Run("Writing through an embed on a record with no value yet", func(t *testing.T) {
				otherUser := h.User().Create(env, h.User().NewData().
					SetName("No Experience User"))
				assert.EqualValues(t, otherUser.Experience(), "")
				otherUser.WithContext("lang", "fr_FR").SetExperience("Expérience créée")
				assert.EqualValues(t, otherUser.WithContext("lang", "fr_FR").Experience(), "Expérience créée")
				// Only the current context is written, the default value is left untouched
				assert.EqualValues(t, otherUser.Experience(), "")
				otherUser.WithContext("lang", "de_DE").SetExperience("Erfahrung")
				assert.EqualValues(t, otherUser.WithContext("lang", "de_DE").Experience(), "Erfahrung")
				assert.EqualValues(t, otherUser.WithContext("lang", "fr_FR").Experience(), "Expérience créée")
				otherUser.Unlink()
			})
			t.Run("Searching on an embedded contexted field uses the current context", func(t *testing.T) {
				usersFR := h.User().NewSet(env).WithContext("lang", "fr_FR").
					Search(q.User().Experience().Equals("Expérience professionnelle"))
				assert.EqualValues(t, usersFR.Len(), 1)
				assert.True(t, usersFR.Equals(user))
				assert.True(t, h.User().NewSet(env).WithContext("lang", "de_DE").
					Search(q.User().Experience().Equals("Expérience professionnelle")).IsEmpty())
				assert.EqualValues(t, h.User().NewSet(env).WithContext("lang", "de_DE").
					Search(q.User().Experience().Equals("Berufserfahrung")).Len(), 1)
				assert.EqualValues(t, h.User().Search(env,
					q.User().Experience().Equals("Professional experience")).Len(), 1)
			})
		}))
	})
	t.Run("Testing unique contexted fields with the type safe API", func(t *testing.T) {
		t.Run("Records without any value should not conflict", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1"))
				h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2"))
			}))
		})
		t.Run("Creating two records with the same value should fail", func(t *testing.T) {
			err := models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2").SetMotto("Carpe diem"))
			})
			assert.ErrorContains(t, err, "Motto must be unique")
		})
		t.Run("Creating two records with the same value in the same language should fail", func(t *testing.T) {
			err := models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				tag1 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				tag2 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2").SetMotto("Seize the day"))
				tag1.WithContext("lang", "fr_FR").SetMotto("Cueille le jour")
				tag2.WithContext("lang", "fr_FR").SetMotto("Cueille le jour")
			})
			assert.ErrorContains(t, err, "Motto must be unique")
		})
		t.Run("The same value in different languages should be accepted", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				tag1 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				tag2 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2").SetMotto("Seize the day"))
				tag1.WithContext("lang", "fr_FR").SetMotto("Cueille le jour")
				tag2.WithContext("lang", "de_DE").SetMotto("Cueille le jour")
			}))
		})
		t.Run("A record may hold the same value in two of its own languages", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				tag := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				tag.WithContext("lang", "fr_FR").SetMotto("Cueille le jour")
				tag.WithContext("lang", "de_DE").SetMotto("Cueille le jour")
			}))
		})
		t.Run("A translation conflicting with another default value should fail", func(t *testing.T) {
			err := models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				tag2 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2").SetMotto("Seize the day"))
				tag2.WithContext("lang", "fr_FR").SetMotto("Carpe diem")
			})
			assert.ErrorContains(t, err, "Motto must be unique")
		})
		t.Run("Writing on another field should not fail", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				tag := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				tag.Write(h.Tag().NewData().SetDescription("Another description"))
			}))
		})
		t.Run("Writing a duplicate value on several records at once should fail", func(t *testing.T) {
			err := models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2").SetMotto("Seize the day"))
				h.Tag().Search(env, q.Tag().Name().Like("Unique Tag %")).
					Write(h.Tag().NewData().SetMotto("Same motto"))
			})
			assert.ErrorContains(t, err, "Motto must be unique")
		})
		t.Run("SetTranslations introducing a duplicate should fail", func(t *testing.T) {
			err := models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				tag1 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				tag2 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2").SetMotto("Seize the day"))
				tag1.SetTranslations(h.Tag().Fields().Motto(), map[string]string{"fr_FR": "Cueille le jour"})
				tag2.SetTranslations(h.Tag().Fields().Motto(), map[string]string{"fr_FR": "Cueille le jour"})
			})
			assert.ErrorContains(t, err, "Motto must be unique")
		})
		t.Run("Copying a record with a unique contexted field should fail", func(t *testing.T) {
			err := models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				tag := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetMotto("Carpe diem"))
				tag.Copy(h.Tag().NewData().SetName("Unique Tag 1 Copy"))
			})
			assert.ErrorContains(t, err, "Motto must be unique")
		})
		t.Run("Uniqueness of a field with two contexts", func(t *testing.T) {
			t.Run("Duplicates in the same company and language should fail", func(t *testing.T) {
				err := models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
					tag1 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetEmblem("Chair"))
					tag2 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2").SetEmblem("Seat"))
					tag1.WithContext("company", "3").WithContext("lang", "fr_FR").SetEmblem("Fauteuil")
					tag2.WithContext("company", "3").WithContext("lang", "fr_FR").SetEmblem("Fauteuil")
				})
				assert.ErrorContains(t, err, "Emblem must be unique")
			})
			t.Run("The same value in different companies should be accepted", func(t *testing.T) {
				assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
					tag1 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 1").SetEmblem("Chair"))
					tag2 := h.Tag().Create(env, h.Tag().NewData().SetName("Unique Tag 2").SetEmblem("Seat"))
					tag1.WithContext("company", "3").WithContext("lang", "fr_FR").SetEmblem("Fauteuil")
					tag2.WithContext("company", "7").WithContext("lang", "fr_FR").SetEmblem("Fauteuil")
				}))
			})
		})
	})
	t.Run("Testing group by queries on contexted fields", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			h.Tag().NewSet(env).SearchAll().Unlink()
			for range 2 {
				tag := h.Tag().Create(env, h.Tag().NewData().
					SetName("Grouped Tag").
					SetSlogan("Common slogan"))
				tag.WithContext("lang", "fr_FR").SetSlogan("Slogan commun")
			}
			h.Tag().Create(env, h.Tag().NewData().
				SetName("Grouped Tag").
				SetSlogan("Other slogan"))

			aggregates := h.Tag().NewSet(env).WithContext("lang", "fr_FR").SearchAll().
				GroupBy(h.Tag().Fields().Slogan()).Aggregates(h.Tag().Fields().Slogan())
			assert.Len(t, aggregates, 2)
			counts := make(map[string]int)
			for _, agg := range aggregates {
				assert.True(t, agg.Values().HasSlogan())
				counts[agg.Values().Slogan()] = agg.Count()
			}
			assert.EqualValues(t, counts, map[string]int{
				"Slogan commun": 2,
				"Other slogan":  1,
			})

			aggregates = h.Tag().NewSet(env).SearchAll().
				GroupBy(h.Tag().Fields().Slogan()).Aggregates(h.Tag().Fields().Slogan())
			assert.Len(t, aggregates, 2)
			counts = make(map[string]int)
			for _, agg := range aggregates {
				counts[agg.Values().Slogan()] = agg.Count()
			}
			assert.EqualValues(t, counts, map[string]int{
				"Common slogan": 2,
				"Other slogan":  1,
			})
		}))
	})
}
