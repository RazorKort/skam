package ui

import (
	"image"
	"image/color"
	"skam/back"
	"skam/messages"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"
)

func NewApplication(w *app.Window) *Application {
	return &Application{
		Window: w,
		Msgs:   make(chan messages.Msg, 32),
	}
}

func (a *Application) DrawError(gtx layout.Context) {
	// Стек с центрированием слоёв
	layout.Stack{Alignment: layout.Center}.Layout(gtx,
		// Затемнение на весь экран
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			defer pointer.PassOp{}.Push(gtx.Ops).Pop()
			event.Op(gtx.Ops, a)
			paint.FillShape(gtx.Ops, color.NRGBA{A: 200, R: 0, G: 0, B: 0},
				clip.Rect{Max: gtx.Constraints.Max}.Op(),
			)
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		// Центрированный попап
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			const popupW, popupH = 400, 200
			popupSize := image.Pt(popupW, popupH)
			// Фиксируем размер попапа
			gtx.Constraints = layout.Exact(popupSize)

			paint.FillShape(gtx.Ops, a.Theme.Bg,
				clip.Rect{Max: popupSize}.Op(),
			)

			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(10).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						msg := a.Theme.BodyText(a.ErrorMsg)
						msg.Alignment = text.Middle
						return msg.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(a.Theme.Theme, &a.ErrorOkBtn, "OK")
						if a.ErrorOkBtn.Clicked(gtx) {
							a.ShowError = false
							a.ErrorMsg = ""
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),
	)
}
func (a *Application) HandleMessage(msg messages.Msg) {
	switch m := msg.(type) {

	case messages.ShowError:
		a.ErrorMsg = m.ErrorMessage
		a.ShowError = true

	case messages.HideError:
		a.ShowError = false
		a.ErrorMsg = ""

	// --- Навигация ---
	case messages.NavigateToLogin:
		a.CurrentScreen = NewLoginScreen(a.Msgs)

	case messages.NavigateToRegister:
		a.CurrentScreen = NewRegisterScreen(a.Msgs)
	case messages.NavigateToImport:
		a.CurrentScreen = NewImportScreeen(a.Msgs)
	case messages.NavigateToMain:
		//new ws connection
		//load friends and messages
		//load messages only when clicked on friend and check if loaded
		a.CurrentScreen = NewMainScreen(a.Msgs)

	case messages.LoginAttempt:
		go func() {
			err := a.Client.LoadKeys(a.KEY_PATH, m.Password)
			if err != nil {
				a.Msgs <- messages.LoginFailed{}
				a.Msgs <- messages.ShowError{ErrorMessage: err.Error()}
			} else {
				a.Msgs <- messages.LoginSuccess{}
			}
		}()
	case messages.RegisterAttempt:
		go func() {
			err := a.Client.Register(m.Name, m.Password)
			if err != nil {
				a.Msgs <- messages.RegisterFailed{}
				a.Msgs <- messages.ShowError{ErrorMessage: err.Error()}

			} else {
				a.Msgs <- messages.RegisterSuccess{}
			}
		}()
	case messages.ImportAttempt:
		go func() {
			status := back.CheckPath(m.Path)
			if !status {
				a.Msgs <- messages.ImportFailed{}
				a.Msgs <- messages.ShowError{ErrorMessage: "No such file"}
				return
			}
			err := a.Client.LoadKeys(m.Path, "")
			if err != nil {
				a.Msgs <- messages.ImportFailed{}
				a.Msgs <- messages.ShowError{ErrorMessage: err.Error()}
				return
			}
			err = a.Client.EncryptKey(m.Password)
			if err != nil {
				a.Msgs <- messages.ImportFailed{}
				a.Msgs <- messages.ShowError{ErrorMessage: err.Error()}
				return
			}
			a.Msgs <- messages.ImportSuccess{}

		}()

		//navigate to main screen
	case messages.LoginSuccess:
		if screen, ok := (a.CurrentScreen).(*LoginScreen); ok {
			screen.IsLoading = false
			a.Msgs <- messages.NavigateToMain{}
		}
	case messages.RegisterSuccess:
		if screen, ok := (a.CurrentScreen).(*LoginScreen); ok {
			screen.IsLoading = false
			a.Msgs <- messages.NavigateToMain{}
		}
	case messages.ImportSuccess:
		if screen, ok := (a.CurrentScreen).(*LoginScreen); ok {
			screen.IsLoading = false
			a.Msgs <- messages.NavigateToMain{}
		}

		//drop loading flag
	case messages.LoginFailed:
		if screen, ok := (a.CurrentScreen).(*LoginScreen); ok {
			screen.IsLoading = false
		}
	case messages.RegisterFailed:
		if screen, ok := (a.CurrentScreen).(*LoginScreen); ok {
			screen.IsLoading = false
		}
	case messages.ImportFailed:
		if screen, ok := (a.CurrentScreen).(*LoginScreen); ok {
			screen.IsLoading = false
		}
	}
}
