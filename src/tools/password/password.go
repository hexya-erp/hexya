// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package password provides functions to Hash and Verify PBKDF2/SHA256 passwords.
// The hashes are in '$pbkdf2-sha256$<N iter>$<salt>$<key>' format.
package password

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen        = 16
	keyLen         = 64
	encodePassword = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789./"
)

// Iterations is the number of iterations used when hashing a password
var Iterations = 25000

// Hash returns a hashed string password in PBKDF2/SHA256 format
func Hash(password string) (string, error) {
	encoding := base64.NewEncoding(encodePassword).WithPadding(base64.NoPadding)
	randByte := make([]byte, saltLen)

	_, err := rand.Read(randByte)
	if err != nil {
		return "", err
	}
	salt := make([]byte, encoding.EncodedLen(saltLen))
	encoding.Encode(salt, randByte)
	dk := pbkdf2.Key([]byte(password), salt, Iterations, keyLen, sha512.New)

	hashedPW := fmt.Sprintf("$pbkdf2-sha512$%d$%s$%s", Iterations, string(salt), encoding.EncodeToString(dk))
	return hashedPW, nil
}

// Verify returns true if the given password matches with the given hash.
func Verify(password, hash string) bool {
	encoding := base64.NewEncoding(encodePassword).WithPadding(base64.NoPadding)
	split := strings.Split(strings.TrimPrefix(hash, "$"), "$")
	salt := []byte(split[2])
	iter, _ := strconv.Atoi(split[1])

	dk := pbkdf2.Key([]byte(password), salt, iter, keyLen, sha512.New)

	hashedPW := fmt.Sprintf("$pbkdf2-sha512$%d$%s$%s", iter, string(salt), encoding.EncodeToString(dk))

	if subtle.ConstantTimeCompare([]byte(hash), []byte(hashedPW)) == 0 {
		return false
	}

	return true
}
