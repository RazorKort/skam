package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func NewRegisterScreen(msgs chan<- Msg) *RegisterScreen {
	backIcon, err := widget.NewIcon(icons.NavigationArrowBack)
	if err != nil {
		backIcon = nil
	}
	var name widget.Editor
	name.SingleLine = true
	name.Submit = true

	var password widget.Editor
	password.SingleLine = true
	password.Submit = true
	password.Mask = '*'

	var confirmPassword widget.Editor
	confirmPassword.SingleLine = true
	confirmPassword.Submit = true
	confirmPassword.Mask = '*'

	return &RegisterScreen{
		Name:            name,
		msgs:            msgs,
		Password:        password,
		ConfirmPassword: confirmPassword,
		BackIcon:        *backIcon,
		inset:           layout.UniformInset(unit.Dp(16)),
		paswordsMatch:   false,
	}
}

func (rs *RegisterScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {

	gtx = th.Background(gtx)
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = gtx.Constraints.Max
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Pt(0, 0)
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceAround,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// Title
						title := th.TitleText("Register in Skam")
						return rs.inset.Layout(gtx, title.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis:      layout.Vertical,
							Alignment: layout.Start,
						}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								// Password input
								editor := th.Input(&rs.Name, "Enter your Name")
								return rs.inset.Layout(gtx, editor.Layout)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								// Password input
								editor := th.Input(&rs.Password, "Password")
								return rs.inset.Layout(gtx, editor.Layout)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								// Password input
								editor := th.Input(&rs.ConfirmPassword, "Confirm Password")
								return rs.inset.Layout(gtx, editor.Layout)
							}),
						)
					}),

					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// Login button (primary)
						btn := th.Button(&rs.RegisterBtn, "Register")
						return rs.inset.Layout(gtx, btn.Layout)
					}),
				)
			})
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			btn := material.IconButton(th.Theme, &rs.BackBtn, &rs.BackIcon, "back arrow")
			btn.Size = unit.Dp(20)
			btn.Background = th.Bg
			btn.Color = th.Colors.Secondary
			btn.Inset = layout.UniformInset(unit.Dp(10))
			inset := layout.UniformInset(unit.Dp(10))
			return inset.Layout(gtx, btn.Layout)
		}),
	)

}

func (rs *RegisterScreen) Update(gtx layout.Context) bool {
	changed := false
	//проверка совпадения паролей
	pswd := rs.Password.Text()
	cpswd := rs.ConfirmPassword.Text()

	match := pswd == cpswd && pswd != ""
	if match != rs.paswordsMatch {
		rs.paswordsMatch = match
		changed = true
	}
	// Обработка кнопки логина
	if rs.RegisterBtn.Clicked(gtx) && !rs.IsLoading && rs.paswordsMatch {
		rs.IsLoading = true
		changed = true
		rs.msgs <- RegisterAttempt{
			Name:     rs.Name.Text(),
			Password: pswd,
		}
	}
	ev, _ := rs.ConfirmPassword.Update(gtx)
	if ev != nil && !rs.IsLoading && rs.paswordsMatch {
		switch ev.(type) {
		case widget.SubmitEvent:
			rs.IsLoading = true
			changed = true
			rs.msgs <- RegisterAttempt{
				Name:     rs.Name.Text(),
				Password: pswd,
			}
		}
	}

	// Обработка кнопки назад
	if rs.BackBtn.Clicked(gtx) && !rs.IsLoading {
		rs.msgs <- NavigateToLogin{}
		changed = true
	}

	return changed
}
