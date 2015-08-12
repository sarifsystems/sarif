// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package renderer

type OverviewLayout struct {
	Title         string
	TitleIcon     rune
	Subtitle      string
	FootnoteLeft  string
	FootnoteRight string

	LeftColumn  Layout
	RightColumn Layout
	Background  Layout
}

func (l *OverviewLayout) Render(ctx *Context, bounds Rect) error {
	ctx.Save()
	defer ctx.Restore()

	// Padding
	bounds.Left += 40
	bounds.Top += 40
	bounds.Right -= 40
	bounds.Bottom -= 40

	// Sub layouts
	if l.LeftColumn != nil {
		colBounds := bounds
		colBounds.Right = colBounds.Left + 200
		bounds.Left = 240 + 30
		if err := l.LeftColumn.Render(ctx, colBounds); err != nil {
			return err
		}
	}
	if l.RightColumn != nil {
		colBounds := bounds
		colBounds.Left = colBounds.Right - 200
		bounds.Right = 640 - 240 - 30
		if err := l.RightColumn.Render(ctx, colBounds); err != nil {
			return err
		}
	}

	// Main content
	if l.Background != nil {
		if err := l.Background.Render(ctx, bounds); err != nil {
			return err
		}
	}

	if l.Title != "" {
		left := float64(0)
		if l.TitleIcon > 0 {
			left += 36 + 18
			ctx.SetFillColor(ctx.Style.ColorAccent)
			ctx.DrawIcon(l.TitleIcon, 36, bounds.Left, bounds.Top+36)
		}
		ctx.SetFillColor(ctx.Style.ColorText)
		ctx.SetFontSize(36)
		ctx.FillStringAt(l.Title, bounds.Left+left, bounds.Top+36)
	}

	ctx.SetFillColor(ctx.Style.ColorLightPrimary)
	if l.Subtitle != "" {
		ctx.SetFontSize(24)
		ctx.FillStringAt(l.Subtitle, bounds.Left, bounds.Top+36+24*2)
	}
	if l.FootnoteLeft != "" {
		ctx.SetFontSize(20)
		ctx.FillStringAt(l.FootnoteLeft, bounds.Left, bounds.Bottom)
	}
	if l.FootnoteRight != "" {
		ctx.SetFontSize(20)
		_, _, right, _ := ctx.GetStringBounds(l.FootnoteRight)
		ctx.FillStringAt(l.FootnoteRight, bounds.Right-right, bounds.Bottom)
	}

	return nil
}
