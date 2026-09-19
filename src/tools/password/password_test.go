// Copyright 2020 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswords(t *testing.T) {
	var hashed string
	t.Run("Testing password hashing and verifying", func(t *testing.T) {
		t.Run("Hashing a password should not fail", func(t *testing.T) {
			var err error
			hashed, err = Hash("secret")
			assert.Nil(t, err)
		})
		t.Run("Verifying the password should work", func(t *testing.T) {
			assert.True(t, Verify("secret", hashed))
		})
		t.Run("Verifiying with wrong password should fail", func(t *testing.T) {
			assert.False(t, Verify("wrong-password", hashed))
		})
	})
}
