// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package testmodule is a dummy module used to test the extraction of
// translatable strings from go code. It is inside a testdata directory
// so that it is ignored by the go tool.
package testmodule

// T is a dummy translation function
func T(src string, args ...interface{}) string {
	return src
}

// Hello returns translatable strings
func Hello() []string {
	return []string{
		T("Hello World"),
		T("Goodbye"),
		notT("Not translated"),
	}
}

// notT is not a translation function
func notT(src string) string {
	return src
}
