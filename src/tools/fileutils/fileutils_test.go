// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package fileutils_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexya-erp/hexya/src/tools/fileutils"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCopy(t *testing.T) {
	Convey("Testing Copy", t, func() {
		srcFileName := filepath.Join(os.TempDir(), "fileutils-input")
		dstFileName := filepath.Join(os.TempDir(), "fileutils-output")
		s, err := os.Create(srcFileName)
		So(err, ShouldBeNil)
		s.WriteString("This is the file's content")
		s.Close()
		err = fileutils.Copy(srcFileName, dstFileName)
		So(err, ShouldBeNil)
		fs, err := os.Stat(srcFileName)
		So(err, ShouldBeNil)
		fd, err := os.Stat(dstFileName)
		So(err, ShouldBeNil)
		So(fd.Size(), ShouldEqual, fs.Size())
		d, err := os.Open(dstFileName)
		So(err, ShouldBeNil)
		data, err := ioutil.ReadAll(d)
		So(err, ShouldBeNil)
		So(string(data), ShouldEqual, "This is the file's content")
	})

}
