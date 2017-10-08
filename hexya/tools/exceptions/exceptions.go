// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package exceptions provides error types used throughout Hexya
package exceptions

// UserError is an error that must rollback the current transaction and
// be displayed as a warning to the user.
type UserError struct {
	Message string
	Debug   string
}

// Error method for the UserError type.
// Returns the message.
func (u UserError) Error() string {
	return u.Message
}
