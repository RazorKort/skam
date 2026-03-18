package ui

import (
	"skam/messages"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

func NewRegisterScreen(msgs chan<- messages.Msg) *RegisterScreen {
	var name widget.Editor
	name.SingleLine = true

	var password widget.Editor
	password.SingleLine = true
	password.Mask = '*'

	var confirmPassword widget.Editor
	confirmPassword.SingleLine = true
	confirmPassword.Mask = '*'

	return &RegisterScreen{
		Name:            name,
		msgs:            msgs,
		Password:        password,
		ConfirmPassword: confirmPassword,
		inset:           layout.UniformInset(unit.Dp(16)),
		paswordsMatch:   false,
	}
}

func (rs *RegisterScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {

	gtx = th.Background(gtx)

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceEvenly,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Back button
				btn := th.Button(&rs.BackBtn, "Placeholder. will be back arrow")
				return rs.inset.Layout(gtx, btn.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Title
				title := th.TitleText("Reister in Skam")
				return rs.inset.Layout(gtx, title.Layout)
			}),
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
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Login button (primary)
				btn := th.Button(&rs.RegisterBtn, "Register")
				return rs.inset.Layout(gtx, btn.Layout)
			}),
		)
	})
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
		rs.msgs <- messages.RegisterAttempt{
			Name:     rs.Name.Text(),
			Password: pswd,
		}
	}

	// Обработка кнопки назад
	if rs.BackBtn.Clicked(gtx) && !rs.IsLoading {
		rs.msgs <- messages.NavigateToLogin{}
		changed = true
	}

	return changed
}
