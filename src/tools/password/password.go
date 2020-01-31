// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package password provides functions to Hash and Verify PBKDF2/SHA256 passwords.
// The hashes are in '$pbkdf2-sha256$<N iter>$<salt>$<key>' format.
package password

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// Iterations is the number of iterations used when hashing a password
var Iterations = 25000

// Hash returns a hashed string password in PBKDF2/SHA256 format
func Hash(password string) (string, error) {
	randByte := make([]byte, 8)

	_, err := rand.Read(randByte)
	if err != nil {
		return "", err
	}

	base64RandByte := base64.StdEncoding.EncodeToString(randByte)
	salt := []byte(base64RandByte)

	dk := pbkdf2.Key([]byte(password), salt, Iterations, 32, sha256.New)

	hashedPW := fmt.Sprintf("$pbkdf2-sha256$%d$%s$%s", Iterations, string(salt), base64.StdEncoding.EncodeToString(dk))
	return hashedPW, nil
}

// Verify returns true if the given password matches with the given hash.
func Verify(password, hash string) bool {
	split := strings.Split(strings.TrimPrefix(hash, "$"), "$")
	salt := []byte(split[2])
	iter, _ := strconv.Atoi(split[1])

	dk := pbkdf2.Key([]byte(password), salt, iter, 32, sha256.New)

	hashedPW := fmt.Sprintf("$pbkdf2-sha256$%d$%s$%s", iter, string(salt), base64.StdEncoding.EncodeToString(dk))

	if subtle.ConstantTimeCompare([]byte(hash), []byte(hashedPW)) == 0 {
		return false
	}

	return true
}
