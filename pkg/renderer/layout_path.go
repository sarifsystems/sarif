// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package renderer

import "math"

type PathLayout struct {
	Path   [][]float64
	LatLng bool
}

func (l *PathLayout) Render(ctx *Context, bounds Rect) error {
	ctx.SetStrokeColor(ctx.Style.ColorAccent)

	if l.LatLng {
		for i, p := range l.Path {
			lat, lng := p[0], p[1]
			x := (lng + 180) * (bounds.Width() / 360)
			latRad := lat * math.Pi / 180
			mercN := math.Log(math.Tan((math.Pi / 4) + (latRad / 2)))
			y := (bounds.Height() / 2) - (bounds.Width() * mercN / (2 * math.Pi))
			p[0], p[1] = x, y
			l.Path[i] = p
		}
	}

	size := Rect{1e9, 1e9, 0, 0}
	for _, p := range l.Path {
		if len(p) < 2 {
			continue
		}
		if p[0] < size.Left {
			size.Left = p[0]
		}
		if p[0] > size.Right {
			size.Right = p[0]
		}
		if p[1] < size.Top {
			size.Top = p[1]
		}
		if p[1] > size.Bottom {
			size.Bottom = p[1]
		}
	}
	dw, dh := bounds.Width(), bounds.Height()
	sw, sh := size.Width(), size.Height()
	scale := math.Min(dw/sw, dh/sh)
	ctx.Save()
	ctx.Translate(bounds.Left-size.Left*scale+(dw-sw*scale)/2, bounds.Top-size.Top*scale+(dh-sh*scale)/2)
	ctx.Scale(scale, scale)

	ctx.SetLineWidth(5 / scale)
	first := true
	for _, p := range l.Path {
		if len(p) < 2 {
			continue
		}
		if first {
			ctx.MoveTo(p[0], p[1])
			first = false
		} else {
			ctx.LineTo(p[0], p[1])
		}
	}
	ctx.Stroke()

	ctx.Restore()
	return nil
}
