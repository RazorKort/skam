package ui

import (
	"image"
	"image/color"
	"log"
	"skam/back"
	"sort"

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
		Msgs:   make(chan Msg, 32),
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
func (a *Application) HandleMessage(msg Msg) {
	switch m := msg.(type) {

	case ShowError:
		a.ErrorMsg = m.ErrorMessage
		a.ShowError = true

	case HideError:
		a.ShowError = false
		a.ErrorMsg = ""

	// --- Навигация ---
	case NavigateToLogin:
		a.CurrentScreen = NewLoginScreen(a.Msgs)

	case NavigateToRegister:
		a.CurrentScreen = NewRegisterScreen(a.Msgs)
	case NavigateToImport:
		a.CurrentScreen = NewImportScreeen(a.Msgs)
	case NavigateToMain:
		a.CurrentScreen = NewMainScreen(a.Msgs, a.Client)

	case LoginAttempt:
		go func() {
			err := a.Client.LoadKeys(a.KEY_PATH, m.Password)
			if err != nil {
				a.Msgs <- LoginFailed{}
				a.Msgs <- ShowError{ErrorMessage: err.Error()}
				return
			}
			if ok := a.AuthandLoad(); ok {
				a.Msgs <- LoginSuccess{}
				return
			}
			a.Msgs <- LoginFailed{}
		}()
	case RegisterAttempt:
		go func() {
			err := a.Client.Register(m.Name, m.Password)
			if err != nil {
				a.Msgs <- RegisterFailed{}
				a.Msgs <- ShowError{ErrorMessage: err.Error()}
				return
			}
			if ok := a.AuthandLoad(); ok {
				a.Msgs <- RegisterSuccess{}
				return
			}
			a.Msgs <- RegisterFailed{}
		}()
	case ImportAttempt:
		go func() {
			status := back.CheckPath(m.Path)
			if !status {
				a.Msgs <- ImportFailed{}
				a.Msgs <- ShowError{ErrorMessage: "No such file"}
				return
			}
			err := a.Client.LoadKeys(m.Path, "")
			if err != nil {
				a.Msgs <- ImportFailed{}
				a.Msgs <- ShowError{ErrorMessage: err.Error()}
				return
			}
			err = a.Client.EncryptKey(m.Password)
			if err != nil {
				a.Msgs <- ImportFailed{}
				a.Msgs <- ShowError{ErrorMessage: err.Error()}
				return
			}
			if ok := a.AuthandLoad(); ok {
				a.Msgs <- ImportSuccess{}
				return
			}
			a.Msgs <- ImportFailed{}

		}()

		//navigate to main screen
	case LoginSuccess:
		if screen, ok := (a.CurrentScreen).(*LoginScreen); ok {
			screen.IsLoading = false
			a.Msgs <- NavigateToMain{}
		}
	case RegisterSuccess:
		if screen, ok := (a.CurrentScreen).(*RegisterScreen); ok {
			screen.IsLoading = false
			a.Msgs <- NavigateToMain{}
		}
	case ImportSuccess:
		if screen, ok := (a.CurrentScreen).(*ImportScreen); ok {
			screen.IsLoading = false
			a.Msgs <- NavigateToMain{}
		}
		//drop loading flag
	case LoginFailed:
		if screen, ok := (a.CurrentScreen).(*LoginScreen); ok {
			screen.IsLoading = false
		}
	case RegisterFailed:
		if screen, ok := (a.CurrentScreen).(*RegisterScreen); ok {
			screen.IsLoading = false
		}
	case ImportFailed:
		if screen, ok := (a.CurrentScreen).(*ImportScreen); ok {
			screen.IsLoading = false
		}

	case FriendClicked:
		go func() {
			if screen, ok := (a.CurrentScreen).(*MainScreen); ok {
				if !a.Client.SelectedFriend.Loaded {
					err := a.Client.LoadMessages(a.Client.SelectedFriend.Id)
					if err != nil {
						screen.IsLoading = false
						a.Client.SelectedFriend.Loaded = true
						a.Msgs <- ShowError{ErrorMessage: err.Error()}
						return
					}
				}
				for i := range a.Client.SelectedFriend.Messages {
					err := back.DecryptMessage(&a.Client.SelectedFriend.Messages[i], *a.Client.SelectedFriend)
					if err != nil {
						screen.IsLoading = false
						a.Msgs <- ShowError{ErrorMessage: err.Error()}
						return
					}
				}
				screen.MessagesList.Position.First = len(a.Client.SelectedFriend.Messages) - 1
				screen.IsLoading = false
			}
		}()

	case SendMessage:
		go func() {
			msg, err := a.Client.AddMessage(m.Text)
			if err != nil {
				a.Msgs <- ShowError{ErrorMessage: err.Error()}
			}
			a.Window.Invalidate()
			a.Client.SendMessage(*msg)
		}()
	}
}

func (a *Application) AuthandLoad() bool {

	err := a.Client.Auth()
	if err != nil {
		a.Msgs <- ShowError{ErrorMessage: err.Error()}
		return false
	}
	err = a.Client.GetFriends()
	if err != nil {
		a.Msgs <- ShowError{ErrorMessage: err.Error()}
		return false
	}
	wsClient, err := a.Client.NewWSClient()
	if err != nil {
		a.Msgs <- ShowError{ErrorMessage: err.Error()}
		return false
	} else {
		a.Client.WS = wsClient
		a.Client.WsMsgChan = wsClient.MsgChan
		go a.handleWebSocketMessages()
	}
	return true
}

// приходит сообщение, декриптим, заставляем перерисовать gui
// если не firend не selected friend можно отправлять сообщение на pop up
func (a *Application) handleWebSocketMessages() {
	for msg := range a.Client.WsMsgChan {
		switch msg.Type {
		case "ack":
			friend_indx := a.Client.FriendsById[msg.Receiver_id]
			for i := len(a.Client.Friends[friend_indx].Messages) - 1; i >= 0; i-- {
				message := &a.Client.Friends[friend_indx].Messages[i]
				if message.Created_at == msg.Created_at && message.Id < 0 {
					message.Id = msg.Id
					message.Sended = true
				}
			}
		case "message":
			friend_indx := a.Client.FriendsById[msg.Sender_id]
			err := back.DecryptMessage(&msg, a.Client.Friends[friend_indx])
			if err != nil {
				log.Printf("Failed to decrypt Ws message: %v", err)
				continue
			}
			a.Client.Friends[friend_indx].Messages = append(a.Client.Friends[friend_indx].Messages, msg)
			sort.Slice(a.Client.Friends[friend_indx].Messages, func(i, j int) bool {
				return a.Client.Friends[friend_indx].Messages[i].Created_at < a.Client.Friends[friend_indx].Messages[j].Created_at
			})
			if a.Window != nil {
				a.Window.Invalidate()
			}
		}

	}
}
