// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package emailutils

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidAddress(t *testing.T) {
	t.Run("Valid addresses", func(t *testing.T) {
		assert.True(t, IsValidAddress("john.doe@example.com"))
		assert.True(t, IsValidAddress("john_doe@example.com"))
		assert.True(t, IsValidAddress("john+doe@example.co.uk"))
		assert.True(t, IsValidAddress("john%doe@sub-domain.example.museum"))
		assert.True(t, IsValidAddress("j@e.io"))
	})
	t.Run("Invalid addresses", func(t *testing.T) {
		assert.False(t, IsValidAddress(""))
		assert.False(t, IsValidAddress("john.doe"))
		assert.False(t, IsValidAddress("@example.com"))
		assert.False(t, IsValidAddress("john.doe@"))
		assert.False(t, IsValidAddress("john.doe@example"))
		assert.False(t, IsValidAddress("john.doe@example.c"))
		assert.False(t, IsValidAddress("john doe@example.com"))
		assert.False(t, IsValidAddress("John Doe <john.doe@example.com>"))
	})
	t.Run("Several addresses should not be valid", func(t *testing.T) {
		assert.False(t, IsValidAddress("john.doe@example.com, jane.doe@example.com"))
		assert.False(t, IsValidAddress("john.doe@example.com\njane.doe@example.com"))
	})
}

func TestMakeMsgID(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	t.Run("Without additional id", func(t *testing.T) {
		msgID := MakeMsgID("")
		rx := regexp.MustCompile(fmt.Sprintf(`^<\d+\.%d\.\d+@%s>$`, os.Getpid(), regexp.QuoteMeta(hostname)))
		assert.Regexp(t, rx, msgID)
	})
	t.Run("With an additional id", func(t *testing.T) {
		msgID := MakeMsgID("myID")
		rx := regexp.MustCompile(fmt.Sprintf(`^<\d+\.%d\.\d+\.myID@%s>$`, os.Getpid(), regexp.QuoteMeta(hostname)))
		assert.Regexp(t, rx, msgID)
	})
	t.Run("Two message IDs should be different", func(t *testing.T) {
		assert.NotEqual(t, MakeMsgID("myID"), MakeMsgID("myID"))
	})
}

func TestGenerateTrackingMessageID(t *testing.T) {
	hostname, _ := os.Hostname()
	t.Run("Message ID should follow the expected pattern", func(t *testing.T) {
		msgID := GenerateTrackingMessageID("myObject")
		rx := regexp.MustCompile(fmt.Sprintf(`^<\d+\.\d+-hexya-myObject@%s>$`, regexp.QuoteMeta(hostname)))
		assert.Regexp(t, rx, msgID)
	})
	t.Run("Empty ident should work too", func(t *testing.T) {
		msgID := GenerateTrackingMessageID("")
		rx := regexp.MustCompile(fmt.Sprintf(`^<\d+\.\d+-hexya-@%s>$`, regexp.QuoteMeta(hostname)))
		assert.Regexp(t, rx, msgID)
	})
	t.Run("Two message IDs should be different", func(t *testing.T) {
		assert.NotEqual(t, GenerateTrackingMessageID("myObject"), GenerateTrackingMessageID("myObject"))
	})
}

func TestSplitAddresses(t *testing.T) {
	t.Run("Single address", func(t *testing.T) {
		assert.Equal(t, []string{"<john.doe@example.com>"}, SplitAddresses("john.doe@example.com"))
	})
	t.Run("Several addresses", func(t *testing.T) {
		assert.Equal(t, []string{"<john.doe@example.com>", "<jane.doe@example.com>"},
			SplitAddresses("john.doe@example.com, jane.doe@example.com"))
	})
	t.Run("Addresses with names", func(t *testing.T) {
		assert.Equal(t, []string{`"John Doe" <john.doe@example.com>`, `"Jane Doe" <jane.doe@example.com>`},
			SplitAddresses(`John Doe <john.doe@example.com>, "Jane Doe" <jane.doe@example.com>`))
	})
	t.Run("Invalid address lists should return an empty slice", func(t *testing.T) {
		assert.Equal(t, []string{}, SplitAddresses(""))
		assert.Equal(t, []string{}, SplitAddresses("not an address"))
		assert.Equal(t, []string{}, SplitAddresses("john.doe@example.com, invalid"))
	})
}
