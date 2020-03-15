// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

/*
Package b64image provides helper functions for manipulating
base64 encoded PNG or JPEG images
*/
package b64image

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	// Load JPEG driver
	_ "image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/disintegration/imaging"
)

// Colorize adds a color to the transparent background of the original image
// and returns the result. If color is the zero value, then a random color
// is applied.
//
// The original must be a base64 encoded image, either JPEG or PNG.
// The result is a base64 encoded PNG image.
func Colorize(original string, clr color.Color) string {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(original))
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Warn("Unable to read image for colorizing")
		return original
	}
	if ColorsEqual(clr, color.RGBA{}) {
		vr := rand.Intn(8)
		vg := rand.Intn(8)
		vb := rand.Intn(8)
		va := rand.Intn(8)
		clr = color.RGBA{
			R: uint8(vr*24 + 32),
			G: uint8(vg*24 + 32),
			B: uint8(vb*24 + 32),
			A: uint8(va*24 + 32),
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(clr), dst.Bounds().Min, draw.Src)
	draw.Draw(dst, dst.Bounds(), img, img.Bounds().Min, draw.Over)
	var buf bytes.Buffer
	png.Encode(&buf, dst)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// ColorsEqual returns true if both colors are the same
func ColorsEqual(clr1, clr2 color.Color) bool {
	r1, g1, b1, a1 := clr1.RGBA()
	r2, g2, b2, a2 := clr2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// Resize a base64 encoded image. The image will be resized to the given
// size, while keeping the aspect ratios, and holes in the image will be
// filled with transparent background. The image will not be stretched if
// smaller than the expected size.
//
// A None value for any of width or height mean an automatically computed
// value based respectively on height or width of the source image.
func Resize(original string, width, height int, avoidIfSmall bool) string {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(original))
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Warn("Unable to read image for colorizing", "err", err)
		return original
	}
	if width == 0 {
		width = int(float64(img.Bounds().Dx()*height) / float64(img.Bounds().Dy()))
	}
	if height == 0 {
		height = int(float64(img.Bounds().Dy()*width) / float64(img.Bounds().Dx()))
	}
	if avoidIfSmall && img.Bounds().Dx() <= width && img.Bounds().Dy() <= height {
		return original
	}
	if img.Bounds().Dx() != width && img.Bounds().Dy() != height {
		img = imaging.Fit(img, width, height, imaging.Linear)
		img = imaging.Sharpen(img, 2.0)
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	minPoint := image.Pt((img.Bounds().Dx()-width)/2, (img.Bounds().Dy()-height)/2)
	draw.Draw(dst, dst.Bounds(), img, minPoint, draw.Over)

	var buf bytes.Buffer
	png.Encode(&buf, dst)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// ReadAll opens the given file which must be an image and returns its content as base64
func ReadAll(fileName string) (string, error) {
	imgFile, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer imgFile.Close()
	buf := bytes.Buffer{}
	w := base64.NewEncoder(base64.StdEncoding, &buf)
	_, err = io.Copy(w, imgFile)
	if err != nil {
		return "", err
	}
	w.Close()
	return buf.String(), nil
}
