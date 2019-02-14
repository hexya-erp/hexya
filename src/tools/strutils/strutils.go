// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hexya-erp/hexya/src/tools/logging"
)

var log logging.Logger

func init() {
	log = logging.GetLogger("strutils")
}

// SnakeCase convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func SnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// Title convert the given camelCase string to a title string.
// eg. MyHTMLData => My HTML Data
func Title(in string) string {

	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, ' ')
		}
		out = append(out, runes[i])
	}

	return string(out)
}

// GetDefaultString returns str if it is not an empty string or def otherwise
func GetDefaultString(str, def string) string {
	if str == "" {
		return def
	}
	return str
}

// StartsAndEndsWith returns true if the given string starts with prefix
// and ends with suffix.
func StartsAndEndsWith(str, prefix, suffix string) bool {
	return strings.HasPrefix(str, prefix) && strings.HasSuffix(str, suffix)
}

// MarshalToJSONString marshals the given data to its JSON representation and
// returns it as a string. It panics in case of error.
func MarshalToJSONString(data interface{}) string {
	if _, ok := data.(string); !ok {
		domBytes, err := json.Marshal(data)
		if err != nil {
			log.Panic("Unable to marshal given data", "error", err, "data", data)
		}
		return string(domBytes)
	}
	return data.(string)
}

// HumanSize returns the given size (in bytes) in a human readable format
func HumanSize(size int64) string {
	units := []string{"bytes", "KB", "MB", "GB"}
	s, i := float64(size), 0
	for s >= 1024 && i < len(units)-1 {
		s /= 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", s, units[i])
}

// Substitute substitutes each occurrence of each key of mapping in str by the
// corresponding mapping value and returns the substituted string.
func Substitute(str string, mapping map[string]string) string {
	for key, val := range mapping {
		str = strings.Replace(str, key, val, -1)
	}
	return str
}

// DictToJSON sanitizes a python dict string representation to valid JSON.
func DictToJSON(dict string) string {
	dict = strings.Replace(dict, "'", "\"", -1)
	dict = strings.Replace(dict, "False", "false", -1)
	dict = strings.Replace(dict, "True", "true", -1)
	dict = strings.Replace(dict, "(", "[", -1)
	dict = strings.Replace(dict, ")", "]", -1)
	return dict
}

// MakeUnique returns an unique string in reference of the given pool
// its made of the base string plus a number if it exists within the pool
func MakeUnique(str string, pool []string) string {
	var nb int
	tested := str
	for tested == "" || IsIn(tested, pool...) {
		nb++
		tested = str + strconv.Itoa(nb)
	}
	return tested
}

// Reverse returns s reversed
func Reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

// SplitEveryN retuns a slice containing str in multiple parts
// each part has its size to n (or less if last part)
func SplitEveryN(str string, n int) []string {
	var out []string
	var bSl []byte
	bStr := []byte(str)
	for i, b := range bStr {
		if (i)%n == 0 && i != 0 {
			out = append(out, string(bSl))
			bSl = nil
		}
		bSl = append(bSl, b)
	}
	out = append(out, string(bSl))
	return out
}

// SplitAtN returns the given strings separated as two string splited after the N-th character
func SplitAtN(str string, n int) (out1 string, out2 string) {
	if n > len(str) {
		out1 = str
		out2 = ""
	} else {
		out1 = str[:n]
		out2 = str[n:]
	}
	return
}

//ContainsOnly returns true if the given string only contains the characters given
func ContainsOnly(str string, lst ...byte) bool {
	for _, b := range []byte(str) {
		for _, l := range lst {
			if l != b {
				return false
			}
		}
	}
	return true
}

// IsIn returns true if the given str is the same as one of the strings given in lst
func IsIn(str string, lst ...string) bool {
	for _, l := range lst {
		if str == l {
			return true
		}
	}
	return false
}

// LeftJustify returns a string with pad string at right side if str's rune length is smaller than length.
// If str's rune length is larger than length, str itself will be returned.
//
// If pad is an empty string, str will be returned.
//
// Samples:
//     LeftJustify("hello", 4, " ")    => "hello"
//     LeftJustify("hello", 10, " ")   => "hello     "
//     LeftJustify("hello", 10, "123") => "hello12312"
func LeftJustify(str string, length int, pad string) string {
	l := len(str)

	if l >= length || pad == "" {
		return str
	}

	remains := length - l
	padLen := len(pad)

	output := &bytes.Buffer{}
	output.Grow(len(str) + (remains/padLen+1)*len(pad))
	output.WriteString(str)
	writePadString(output, pad, padLen, remains)
	return output.String()
}

// RightJustify returns a string with pad string at left side if str's rune length is smaller than length.
// If str's rune length is larger than length, str itself will be returned.
//
// If pad is an empty string, str will be returned.
//
// Samples:
//     RightJustify("hello", 4, " ")    => "hello"
//     RightJustify("hello", 10, " ")   => "     hello"
//     RightJustify("hello", 10, "123") => "12312hello"
func RightJustify(str string, length int, pad string) string {
	l := len(str)

	if l >= length || pad == "" {
		return str
	}

	remains := length - l
	padLen := len(pad)

	output := &bytes.Buffer{}
	output.Grow(len(str) + (remains/padLen+1)*len(pad))
	writePadString(output, pad, padLen, remains)
	output.WriteString(str)
	return output.String()
}

// CenterJustify returns a string with pad string at both side if str's rune length is smaller than length.
// If str's rune length is larger than length, str itself will be returned.
//
// If pad is an empty string, str will be returned.
//
// Samples:
//     Center("hello", 4, " ")    => "hello"
//     Center("hello", 10, " ")   => "  hello   "
//     Center("hello", 10, "123") => "12hello123"
func CenterJustify(str string, length int, pad string) string {
	l := len(str)

	if l >= length || pad == "" {
		return str
	}

	remains := length - l
	padLen := len(pad)

	output := &bytes.Buffer{}
	output.Grow(len(str) + (remains/padLen+1)*len(pad))
	writePadString(output, pad, padLen, remains/2)
	output.WriteString(str)
	writePadString(output, pad, padLen, (remains+1)/2)
	return output.String()
}

func writePadString(output *bytes.Buffer, pad string, padLen, remains int) {
	var r rune
	var size int

	repeats := remains / padLen

	for i := 0; i < repeats; i++ {
		output.WriteString(pad)
	}

	remains = remains % padLen

	if remains != 0 {
		for i := 0; i < remains; i++ {
			r, size = utf8.DecodeRuneInString(pad)
			output.WriteRune(r)
			pad = pad[size:]
		}
	}
}
