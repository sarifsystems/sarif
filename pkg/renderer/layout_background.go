// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package renderer

import (
	"image"
	"image/color"
	"math"
	"os"

	"github.com/llgcode/draw2d/draw2dkit"
)

type BackgroundLayout struct {
	Image     string
	Grayscale bool
	Color     bool
}

func (l *BackgroundLayout) Render(ctx *Context, bounds Rect) error {
	dw, dh := bounds.Width(), bounds.Height()

	if l.Image != "" {
		f, err := os.Open(l.Image)
		if err != nil {
			return err
		}
		defer f.Close()

		src, _, err := image.Decode(f)
		if err != nil {
			return err
		}

		if l.Grayscale {
			src = &Converted{src, color.GrayModel}
		}

		sw, sh := float64(src.Bounds().Dx()), float64(src.Bounds().Dy())
		scale := math.Max(dw/sw, dh/sh)
		ctx.Save()
		ctx.Translate(bounds.Left+(dw-sw*scale)/2, bounds.Top+(dh-sh*scale)/2)
		ctx.Scale(scale, scale)
		ctx.DrawImage(src)
		ctx.Restore()
	}

	if l.Color {
		ctx.SetFillColor(ctx.Style.ColorDarkPrimary)
		draw2dkit.Rectangle(ctx, bounds.Left, bounds.Top, bounds.Right, bounds.Bottom)
		ctx.Fill()
	}

	return nil
}
