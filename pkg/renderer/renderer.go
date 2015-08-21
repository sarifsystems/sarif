// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package renderer

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"os"

	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

type Context struct {
	Image draw.Image
	*draw2dimg.GraphicContext
	Style Style
}

type Layout interface {
	Render(*Context, Rect) error
}

type OutputFormat struct {
	Format string
	Width  int
	Height int
}

type Renderer struct {
	Image   draw.Image
	Context *draw2d.GraphicContext
}

type Style struct {
	ColorDarkPrimary  color.Color
	ColorPrimary      color.Color
	ColorLightPrimary color.Color
	ColorText         color.Color

	ColorAccent        color.Color
	ColorTextPrimary   color.Color
	ColorTextSecondary color.Color
	ColorDivider       color.Color

	ColorSecondary color.Color
}

func NewContext() *Context {
	img := image.NewRGBA(image.Rect(0, 0, 640, 360))
	ctx := &Context{
		img,
		draw2dimg.NewGraphicContext(img),
		StyleDefault,
	}
	ctx.SetFontData(draw2d.FontData{Name: "Default"})

	return ctx
}

func (ctx *Context) Render(layouts ...Layout) error {
	bounds := Rect{0, 0, 640, 360}
	for _, l := range layouts {
		if err := l.Render(ctx, bounds); err != nil {
			return err
		}
	}
	return nil
}

func LoadFont(name, file string) error {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	font, err := truetype.Parse(f)
	if err != nil {
		return err
	}
	draw2d.RegisterFont(draw2d.FontData{Name: name}, font)
	return nil
}

func (ctx *Context) SaveToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, ctx.Image)
}

func (ctx *Context) DrawIcon(icon rune, size, x, y float64) error {
	ctx.SetFontData(draw2d.FontData{Name: "FontAwesome"})
	ctx.SetFontSize(size)
	ctx.FillStringAt(string([]rune{icon}), x, y)
	ctx.SetFontData(draw2d.FontData{Name: "Default"})
	return nil
}

// Converted implements image.Image, so you can
// pretend that it is the converted image.
type Converted struct {
	Img image.Image
	Mod color.Model
}

// We return the new color model...
func (c *Converted) ColorModel() color.Model {
	return c.Mod
}

// ... but the original bounds
func (c *Converted) Bounds() image.Rectangle {
	return c.Img.Bounds()
}

// At forwards the call to the original image and
// then asks the color model to convert it.
func (c *Converted) At(x, y int) color.Color {
	return c.Mod.Convert(c.Img.At(x, y))
}
