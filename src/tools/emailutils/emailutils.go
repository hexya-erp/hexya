// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package emailutils

import (
	"fmt"
	"math/rand"
	"net/mail"
	"os"
	"regexp"
	"time"
)

// SingleEmailRE is the regular expression for a single email address
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

// MakeMsgID returns a string suitable for RFC 2822 compliant Message-ID, e.g:
//
// <142480216486.20800.16526388040877946887@nightshade.la.mastaler.com>
//
// Optional id if given is a string used to strengthen the uniqueness of the message id.
func MakeMsgID(id string) string {
	timeVal := time.Now().UnixNano() / 10000
	pid := os.Getpid()
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Uint64()
	if id != "" {
		id = "." + id
	}
	idHost, err := os.Hostname()
	if err != nil {
		idHost = "localhost"
	}
	return fmt.Sprintf("<%d.%d.%d%s@%s>", timeVal, pid, randInt, id, idHost)
}

// GenerateTrackingMessageID returns a string that can be used in the Message-ID RFC822 header field
//
// Used to track the replies related to a given object thanks to the "In-Reply-To"
// or "References" fields that Mail User Agents will set.
func GenerateTrackingMessageID(ident string) string {
	rand.Seed(time.Now().UnixNano())
	rnd := rand.Float64()
	rndStr := fmt.Sprintf("%.15f", rnd)[2:]
	hostname, _ := os.Hostname()
	return fmt.Sprintf("<%s.%d-hexya-%s@%s>", rndStr, time.Now().Unix(), ident, hostname)
}

// SplitAddresses returns a list of email addresses from a string of multiple addresses
func SplitAddresses(in string) []string {
	addresses, err := mail.ParseAddressList(in)
	if err != nil {
		return []string{}
	}
	res := make([]string, len(addresses))
	for i, addr := range addresses {
		res[i] = addr.String()
	}
	return res
}
