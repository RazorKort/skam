package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

// NewLoginScreen creates a new login screen with reasonable defaults.
func NewLoginScreen(msgs chan<- Msg) *LoginScreen {
	var password widget.Editor
	password.SingleLine = true
	password.Mask = '*'

	return &LoginScreen{
		Password: password,
		msgs:     msgs,
		inset:    layout.UniformInset(unit.Dp(16)),
	}
}

func (ls *LoginScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {

	gtx = th.Background(gtx)

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceEvenly,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Title
				title := th.TitleText("Login")
				return ls.inset.Layout(gtx, title.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Password input
				editor := th.Input(&ls.Password, "Password")
				return ls.inset.Layout(gtx, editor.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Login button (primary)
				btn := th.Button(&ls.LoginButton, "Login")
				return ls.inset.Layout(gtx, btn.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceAround,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := th.OutlinedButton(&ls.ImportKeyBtn, "Import key")
						return ls.inset.Layout(gtx, btn.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := th.OutlinedButton(&ls.RegisterBtn, "Register")
						return ls.inset.Layout(gtx, btn.Layout)
					}),
				)
			}),
		)
	})
}

func (ls *LoginScreen) Update(gtx layout.Context) bool {
	changed := false
	// Обработка кнопки логина
	if ls.LoginButton.Clicked(gtx) && !ls.IsLoading && ls.Password.Text() != "" {
		ls.IsLoading = true
		changed = true
		ls.msgs <- LoginAttempt{
			Password: ls.Password.Text(),
		}
	}

	// Обработка кнопки регистрации
	if ls.RegisterBtn.Clicked(gtx) {
		changed = true
		ls.msgs <- NavigateToRegister{}

	}

	//Обработка импорта ключа
	if ls.ImportKeyBtn.Clicked(gtx) {
		ls.msgs <- NavigateToImport{}
		changed = true
	}

	return changed
}
