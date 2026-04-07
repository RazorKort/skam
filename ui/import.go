package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"golang.org/x/exp/shiny/materialdesign/icons"
)

func NewImportScreeen(msgs chan<- Msg) *ImportScreen {
	backIcon, err := widget.NewIcon(icons.NavigationArrowBack)
	if err != nil {
		backIcon = nil
	}
	var pathKey widget.Editor
	pathKey.SingleLine = true
	pathKey.Submit = true

	var password widget.Editor
	password.SingleLine = true
	password.Submit = true
	password.Mask = '*'

	var password2 widget.Editor
	password2.SingleLine = true
	password2.Submit = true
	password2.Mask = '*'

	return &ImportScreen{
		Password:  password,
		Password2: password2,
		Path:      pathKey,
		msgs:      msgs,
		BackIcon:  *backIcon,
		inset:     layout.UniformInset(unit.Dp(16)),
	}
}

func (is *ImportScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {
	gtx = th.Background(gtx)

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			//какой криворукий уебан писал layout.Center, что он центрирует по min
			//почему блять в Expanded передается min 0,0
			//почему я блять руками должен менять ограничения чтобы нормально центрировать это говно
			//ответов не будет блять
			gtx.Constraints.Min = gtx.Constraints.Max
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Pt(0, 0)
				inset := layout.UniformInset(10)

				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceAround,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := th.TitleText("Import keys")
						return is.inset.Layout(gtx, text.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X = 320
						gtx.Constraints.Min.X = 320
						return layout.Flex{
							Axis:      layout.Vertical,
							Alignment: layout.Start,
						}.Layout(gtx,

							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return widget.Border{
									Color:        th.Colors.Border,
									CornerRadius: 8,
									Width:        1,
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									input := th.Input(&is.Path, "Path to file with keys")
									return inset.Layout(gtx, input.Layout)
								})
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								inset2 := layout.UniformInset(10)
								inset2.Left = 0
								inset2.Right = 0
								return inset2.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return widget.Border{
										Color:        th.Colors.Border,
										CornerRadius: 8,
										Width:        1,
									}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										input := th.Input(&is.Password, "New password")
										return inset.Layout(gtx, input.Layout)
									})
								})
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return widget.Border{
									Color:        th.Colors.Border,
									CornerRadius: 8,
									Width:        1,
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									input := th.Input(&is.Password2, "Confirm password")
									return inset.Layout(gtx, input.Layout)
								})
							}),
						)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := th.Button(&is.ImportKeyBtn, "Import")
						return is.inset.Layout(gtx, btn.Layout)
					}),
				)
			})

		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			btn := th.IconButtonSecondary(&is.BackBtn, &is.BackIcon, "back arrow")
			inset := layout.UniformInset(unit.Dp(10))
			return inset.Layout(gtx, btn.Layout)
		}),
	)

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
	ev, _ := is.Password2.Update(gtx)
	if ev != nil && !is.IsLoading && is.paswordsMatch {
		switch ev.(type) {
		case widget.SubmitEvent:
			is.IsLoading = true
			is.msgs <- ImportAttempt{
				Path:     is.Path.Text(),
				Password: pswd,
			}
			changed = true
		}
	}

	return changed
}
