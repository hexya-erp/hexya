// Copyright 2020 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package fileutils_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/tools/fileutils"
)

func TestCopy(t *testing.T) {
	t.Run("Testing Copy", func(t *testing.T) {
		srcFileName := filepath.Join(os.TempDir(), "fileutils-input")
		dstFileName := filepath.Join(os.TempDir(), "fileutils-output")
		s, err := os.Create(srcFileName)
		assert.Nil(t, err)
		s.WriteString("This is the file's content")
		s.Close()
		err = fileutils.Copy(srcFileName, dstFileName)
		assert.Nil(t, err)
		fs, err := os.Stat(srcFileName)
		assert.Nil(t, err)
		fd, err := os.Stat(dstFileName)
		assert.Nil(t, err)
		assert.EqualValues(t, fd.Size(), fs.Size())
		d, err := os.Open(dstFileName)
		assert.Nil(t, err)
		data, err := io.ReadAll(d)
		assert.Nil(t, err)
		assert.EqualValues(t, string(data), "This is the file's content")
	})

}
