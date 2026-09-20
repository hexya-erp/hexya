// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package assets

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileLessFiles(t *testing.T) {
	t.Run("Testing less compilation", func(t *testing.T) {
		input := strings.NewReader(".class { width: (1 + 1) }")
		output := bytes.Buffer{}
		err := LessCompiler{}.Compile(input, &output)
		assert.Nil(t, err)
		data, err := io.ReadAll(&output)
		assert.Nil(t, err)
		assert.EqualValues(t, string(data), ".class {\n  width: 2;\n}\n")
	})
	t.Run("Testing scss compilation", func(t *testing.T) {
		input := strings.NewReader(".class { width: (1 + 1) }")
		output := bytes.Buffer{}
		err := ScssCompiler{}.Compile(input, &output)
		assert.Nil(t, err)
		data, err := io.ReadAll(&output)
		assert.Nil(t, err)
		assert.EqualValues(t, string(data), ".class {\n  width: 2;\n}\n")
	})

}
