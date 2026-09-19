// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package b64image

import (
	"encoding/base64"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorize(t *testing.T) {
	t.Run("Testing Colorize function", func(t *testing.T) {
		imgString, err := ReadAll("testdata/avatar.png")
		assert.Nil(t, err)
		t.Run("Applying a fully opaque color", func(t *testing.T) {
			clr := color.RGBA{R: 32, G: 224, B: 224, A: 255}
			dstImageString := Colorize(imgString, clr)
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(dstImageString))
			destImg, _, _ := image.Decode(reader)
			t.Run("Result image should not be the original one", func(t *testing.T) {
				assert.NotEqualValues(t, dstImageString, imgString)
			})
			t.Run("The target image should have the same dimensions", func(t *testing.T) {
				assert.EqualValues(t, destImg.Bounds().Dx(), 180)
				assert.EqualValues(t, destImg.Bounds().Dy(), 180)
			})
			t.Run("The color at 2,2 should be the given color", func(t *testing.T) {
				assert.True(t, ColorsEqual(destImg.At(2, 2), clr))
			})
			t.Run("The color at 90,90 should be the original color", func(t *testing.T) {
				assert.True(t, ColorsEqual(destImg.At(90, 90), color.RGBA{R: 217, G: 222, B: 226, A: 255}))
			})
		})
		t.Run("Unreadable image should be returned as is", func(t *testing.T) {
			clr := color.RGBA{R: 32, G: 224, B: 224, A: 255}
			dstImageString := Colorize("foo bar", clr)
			assert.EqualValues(t, dstImageString, "foo bar")
		})
		t.Run("Testing random color", func(t *testing.T) {
			dstImageString := Colorize(imgString, color.RGBA{})
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(dstImageString))
			destImg, _, _ := image.Decode(reader)
			t.Run("Result image should not be the original one", func(t *testing.T) {
				assert.NotEqualValues(t, dstImageString, imgString)
			})
			t.Run("The target image should have the same dimensions", func(t *testing.T) {
				assert.EqualValues(t, destImg.Bounds().Dx(), 180)
				assert.EqualValues(t, destImg.Bounds().Dy(), 180)
			})
			t.Run("The color at 2,2 should be the same as 4,4 and not the empty color", func(t *testing.T) {
				assert.True(t, ColorsEqual(destImg.At(2, 2), destImg.At(4, 4)))
				assert.False(t, ColorsEqual(destImg.At(2, 2), color.RGBA{}))
			})
			t.Run("The color at 90,90 should be the original color", func(t *testing.T) {
				assert.True(t, ColorsEqual(destImg.At(90, 90), color.RGBA{R: 217, G: 222, B: 226, A: 255}))
			})

		})
	})
}

func TestResize(t *testing.T) {
	t.Run("Testing Resize function", func(t *testing.T) {
		imgString, err := ReadAll("testdata/avatar.png")
		assert.Nil(t, err)
		t.Run("Resizing smaller should create a smaller image", func(t *testing.T) {
			smallImg := Resize(imgString, 100, 150, false)
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(smallImg))
			destImg, _, _ := image.Decode(reader)
			assert.EqualValues(t, destImg.Bounds().Dx(), 100)
			assert.EqualValues(t, destImg.Bounds().Dy(), 150)
		})
		t.Run("Resizing bigger should create a bigger image", func(t *testing.T) {
			bigImg := Resize(imgString, 300, 400, false)
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(bigImg))
			destImg, _, _ := image.Decode(reader)
			assert.EqualValues(t, destImg.Bounds().Dx(), 300)
			assert.EqualValues(t, destImg.Bounds().Dy(), 400)
		})
		t.Run("Resizing bigger, with avoid, should not create a bigger image", func(t *testing.T) {
			bigImg := Resize(imgString, 300, 400, true)
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(bigImg))
			destImg, _, _ := image.Decode(reader)
			assert.EqualValues(t, destImg.Bounds().Dx(), 180)
			assert.EqualValues(t, destImg.Bounds().Dy(), 180)
		})
	})
}
