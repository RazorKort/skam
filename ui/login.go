package ui

import (
	"skam/messages"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

// LoginScreen holds the state for the login UI.
type LoginScreen struct {
	Password     widget.Editor
	LoginButton  widget.Clickable
	ImportKeyBtn widget.Clickable
	RegisterBtn  widget.Clickable
	inset        layout.Inset

	msgs chan<- messages.Msg

	IsLoading bool
}

// NewLoginScreen creates a new login screen with reasonable defaults.
func NewLoginScreen(msgs chan<- messages.Msg) *LoginScreen {
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
				// Import key button (outlined)
				btn := th.OutlinedButton(&ls.ImportKeyBtn, "Import key")
				return ls.inset.Layout(gtx, btn.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Registration button (outlined)
				btn := th.OutlinedButton(&ls.RegisterBtn, "Register")
				return ls.inset.Layout(gtx, btn.Layout)
			}),
		)
	})
}

func (ls *LoginScreen) Update(gtx layout.Context) bool {
	changed := false
	// Обработка кнопки логина
	if ls.LoginButton.Clicked(gtx) && !ls.IsLoading {
		ls.IsLoading = true
		changed = true
		ls.msgs <- messages.LoginAttempt{
			Password: ls.Password.Text(),
		}
	}

	// Обработка кнопки регистрации
	if ls.RegisterBtn.Clicked(gtx) {
		ls.msgs <- messages.NavigateToRegister{}
		changed = true
	}

	// // Обработка импорта ключа
	// if ls.ImportKeyBtn.Clicked(gtx) {
	//     ls.msgs <- messages.NavigateToImport{}
	//     changed = true
	// }

	return changed
}
