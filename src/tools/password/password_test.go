// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPasswords(t *testing.T) {
	var hashed string
	Convey("Testing password hashing and verifying", t, func() {
		Convey("Hashing a password should not fail", func() {
			var err error
			hashed, err = Hash("secret")
			So(err, ShouldBeNil)
		})
		Convey("Verifying the password should work", func() {
			So(Verify("secret", hashed), ShouldBeTrue)
		})
		Convey("Verifiying with wrong password should fail", func() {
			So(Verify("wrong-password", hashed), ShouldBeFalse)
		})
	})
}
