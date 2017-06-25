// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBootStrap(t *testing.T) {
	// Creating a dummy table to check that it is correctly removed by Bootstrap
	dbExecuteNoTx("CREATE TABLE IF NOT EXISTS shouldbedeleted (id serial NOT NULL PRIMARY KEY)")

	Convey("Database creation should run fine", t, func() {
		Convey("Dummy table should exist", func() {
			So(testAdapter.tables(), ShouldContainKey, "shouldbedeleted")
		})
		Convey("Bootstrap should not panic", func() {
			So(BootStrap, ShouldNotPanic)
			So(SyncDatabase, ShouldNotPanic)
		})
		Convey("Boostrapping twice should panic", func() {
			So(BootStrap, ShouldPanic)
		})
		Convey("Creating SQL view should run fine", func() {
			So(func() {
				dbExecuteNoTx(`DROP VIEW IF EXISTS user_view;
					CREATE VIEW user_view AS (
						SELECT u.id, u.name, p.city, u.active
						FROM "user" u
							LEFT JOIN "profile" p ON p.id = u.profile_id
					)`)
			}, ShouldNotPanic)
		})
		Convey("All models should have a DB table", func() {
			dbTables := testAdapter.tables()
			for tableName, mi := range Registry.registryByTableName {
				if mi.isMixin() || mi.isManual() {
					continue
				}
				So(dbTables[tableName], ShouldBeTrue)
			}
		})
		Convey("All DB tables should have a model", func() {
			for dbTable := range testAdapter.tables() {
				So(Registry.registryByTableName, ShouldContainKey, dbTable)
			}
		})
		Convey("Table constraints should have been created", func() {
			So(testAdapter.constraints("%_mancon"), ShouldHaveLength, 1)
			So(testAdapter.constraints("%_mancon")[0], ShouldEqual, "nums_premium_user_mancon")
		})
	})
	Convey("Making small changes to test DB sync", t, func() {
		Convey("Modifying Required and Default values", func() {
			numsField := Registry.MustGet("User").Fields().MustGet("Nums")
			numsField.SetDefault(nil).SetIndex(false)
			So(numsField.defaultFunc, ShouldBeNil)
			So(numsField.index, ShouldBeFalse)
			profileField := Registry.MustGet("User").Fields().MustGet("Profile")
			profileField.SetRequired(false)
			So(profileField.required, ShouldBeFalse)
			So(SyncDatabase, ShouldNotPanic)
		})
	})

	Convey("Truncating all tables...", t, func() {
		for tn, mi := range Registry.registryByTableName {
			if mi.isMixin() || mi.isManual() {
				continue
			}
			dbExecuteNoTx(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE`, tn))
		}
	})
}
