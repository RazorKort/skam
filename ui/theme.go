package ui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type AppTheme struct {
	*material.Theme
	Colors AppColors
}

type AppColors struct {
	Primary             color.NRGBA
	Secondary           color.NRGBA
	Error               color.NRGBA
	Success             color.NRGBA
	Warning             color.NRGBA
	Background          color.NRGBA
	BackgroundSecondary color.NRGBA
	Surface             color.NRGBA
	Border              color.NRGBA
	OnPrimary           color.NRGBA
	OnSecondary         color.NRGBA
	TextPrimary         color.NRGBA
	TextSecondary       color.NRGBA
}

func NewAppTheme() *AppTheme {
	base := material.NewTheme()

	colors := AppColors{
		Primary:             rgba(96, 202, 160, 255),
		Secondary:           rgba(52, 211, 153, 255),
		Error:               rgba(248, 113, 113, 255),
		Success:             rgba(110, 231, 183, 255),
		Warning:             rgba(251, 191, 36, 255),
		Background:          rgba(23, 23, 23, 255),
		BackgroundSecondary: rgba(35, 35, 35, 255),
		Surface:             rgba(38, 38, 38, 255),
		OnPrimary:           rgba(0, 0, 0, 230),
		OnSecondary:         rgba(0, 0, 0, 230),
		TextPrimary:         rgba(250, 250, 250, 255),
		TextSecondary:       rgba(163, 163, 163, 255),
		Border:              rgba(55, 55, 55, 127),
	}

	base.Palette = material.Palette{
		Bg:         colors.Background,
		Fg:         colors.TextPrimary,
		ContrastBg: colors.Primary,
		ContrastFg: colors.OnPrimary,
	}
	return &AppTheme{
		Theme:  base,
		Colors: colors,
	}
}

// Button returns a primary button styled with the app colors.
func (t *AppTheme) Button(click *widget.Clickable, label string) material.ButtonStyle {
	btn := material.Button(t.Theme, click, label)
	btn.Background = t.Colors.Primary
	btn.Color = t.Colors.OnPrimary
	return btn
}

// SecondaryButton returns a secondary/accent button.
func (t *AppTheme) SecondaryButton(click *widget.Clickable, label string) material.ButtonStyle {
	btn := material.Button(t.Theme, click, label)
	btn.Background = t.Colors.Secondary
	btn.Color = t.Colors.OnSecondary
	return btn
}

// OutlinedButton returns an outlined button using text colors.
func (t *AppTheme) OutlinedButton(click *widget.Clickable, label string) material.ButtonStyle {
	btn := material.Button(t.Theme, click, label)
	btn.Background = color.NRGBA{0, 0, 0, 0} // прозрачный фон
	btn.Color = t.Colors.TextPrimary         // цвет текста
	//btn.Border = widget.Border{Color: t.Colors.Primary, Width: unit.Dp(1)}
	return btn
}

// Input returns a styled editor for single or multi-line text.
func (t *AppTheme) Input(ed *widget.Editor, hint string) material.EditorStyle {
	e := material.Editor(t.Theme, ed, hint)
	e.Color = t.Colors.TextPrimary
	e.HintColor = t.Colors.TextSecondary
	return e
}

// Card lays out content with a standard surface background and padding.
func (t *AppTheme) Card(gtx layout.Context, w layout.Widget) layout.Dimensions {
	inset := layout.UniformInset(unit.Dp(16)) // стандартный паддинг

	// клип на всю доступную область
	r := image.Rectangle{Max: gtx.Constraints.Max}
	rr := clip.RRect{
		Rect: r,
		SE:   gtx.Dp(unit.Dp(8)), // скругление углов
		SW:   gtx.Dp(unit.Dp(8)),
		NW:   gtx.Dp(unit.Dp(8)),
		NE:   gtx.Dp(unit.Dp(8)),
	}.Push(gtx.Ops)

	// фон карточки
	paint.Fill(gtx.Ops, t.Colors.Surface)
	rr.Pop()

	// паддинг + дочерний виджет
	return inset.Layout(gtx, w)
}

// TitleText returns a styled title text.
func (t *AppTheme) TitleText(text string) material.LabelStyle {
	lbl := material.H5(t.Theme, text)
	lbl.Color = t.Colors.TextPrimary
	return lbl
}

// BodyText returns standard body text.
func (t *AppTheme) BodyText(text string) material.LabelStyle {
	lbl := material.Body1(t.Theme, text)
	lbl.Color = t.Colors.TextPrimary
	return lbl
}

// BodyTextOnPrimary returns body text for primary-colored surfaces.
func (t *AppTheme) BodyTextOnPrimary(text string) material.LabelStyle {
	lbl := material.Body1(t.Theme, text)
	lbl.Color = t.Colors.OnPrimary
	return lbl
}

// CaptionSecondary returns caption text with secondary (muted) color.
func (t *AppTheme) CaptionSecondary(text string) material.LabelStyle {
	lbl := material.Caption(t.Theme, text)
	lbl.Color = t.Colors.TextSecondary
	return lbl
}

func (t *AppTheme) IconButtonSecondary(button *widget.Clickable, icon *widget.Icon, description string) material.IconButtonStyle {
	btn := material.IconButton(t.Theme, button, icon, description)
	btn.Size = unit.Dp(20)
	btn.Background = t.Bg
	btn.Color = t.Colors.Secondary
	btn.Inset = layout.UniformInset(unit.Dp(10))
	return btn
}

func (t *AppTheme) Background(gtx layout.Context) layout.Context {
	paint.FillShape(gtx.Ops, t.Colors.Background,
		clip.Rect{Max: gtx.Constraints.Max}.Op(),
	)

	return gtx
}

func rgba(r, g, b, a int) color.NRGBA {
	return color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}
