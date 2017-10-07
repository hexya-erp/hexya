// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package emailutils

import "regexp"

const SingleEmailRE string = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,63}$`

// IsValidAddress returns true if the given address is valid
// and contains only one address
func IsValidAddress(address string) bool {
	ok, err := regexp.MatchString(SingleEmailRE, address)
	if !ok || err != nil {
		return false
	}
	return true
}
