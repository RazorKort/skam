package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

func NewImportScreeen(msgs chan<- Msg) *ImportScreen {
	var pathKey widget.Editor
	pathKey.SingleLine = true

	var password widget.Editor
	password.SingleLine = true
	password.Mask = '*'

	var password2 widget.Editor
	password2.SingleLine = true
	password2.Mask = '*'

	return &ImportScreen{
		Password:  password,
		Password2: password2,
		Path:      pathKey,
		msgs:      msgs,
		inset:     layout.UniformInset(unit.Dp(16)),
	}
}

func (is *ImportScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {
	gtx = th.Background(gtx)

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceEvenly,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := th.Button(&is.BackBtn, "Placeholder back")
				return is.inset.Layout(gtx, btn.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Start,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						input := th.Input(&is.Path, "Path to file with keys")
						return is.inset.Layout(gtx, input.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						input := th.Input(&is.Password, "New password")
						return is.inset.Layout(gtx, input.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						input := th.Input(&is.Password2, "Confirm password")
						return is.inset.Layout(gtx, input.Layout)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := th.Button(&is.ImportKeyBtn, "Import")
				return is.inset.Layout(gtx, btn.Layout)
			}),
		)
	})
}

func (is *ImportScreen) Update(gtx layout.Context) bool {
	changed := false

	pswd := is.Password.Text()
	cpswd := is.Password2.Text()

	match := pswd == cpswd && pswd != ""
	if match != is.paswordsMatch {
		is.paswordsMatch = match
		changed = true
	}

	if is.BackBtn.Clicked(gtx) && !is.IsLoading {
		is.msgs <- NavigateToLogin{}
		changed = true
	}

	if is.ImportKeyBtn.Clicked(gtx) && !is.IsLoading && is.paswordsMatch {
		is.msgs <- ImportAttempt{
			Path:     is.Path.Text(),
			Password: pswd,
		}
		changed = true
	}

	return changed
}
