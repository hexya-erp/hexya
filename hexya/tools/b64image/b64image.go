// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

/*
Package imageutils provides helper functions for manipulating
base64 encoded PNG or JPEG images
*/
package b64image

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math/rand"
	"strings"
)

// Colorize adds a color to the transparent background of the orginal image
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
