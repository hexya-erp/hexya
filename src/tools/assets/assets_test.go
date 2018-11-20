// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package assets

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCompileLessFiles(t *testing.T) {
	Convey("Testing less compilation", t, func() {
		input := strings.NewReader(".class { width: (1 + 1) }")
		output := bytes.Buffer{}
		err := CompileLess(input, &output)
		So(err, ShouldBeNil)
		data, err := ioutil.ReadAll(&output)
		So(err, ShouldBeNil)
		So(string(data), ShouldEqual, ".class {\n  width: 2;\n}\n")
	})

}
