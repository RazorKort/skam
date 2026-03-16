package main

import (
	"fmt"
	"os"
	"skam/back"
	"skam/messages"
	"skam/ui"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
)

const URL string = "skam.su:10000"
const KEY_PATH string = "session.key"

func main() {
	go func() {
		window := new(app.Window)
		window.Option(
			app.Title("Skam"),
			app.MinSize(unit.Dp(400), unit.Dp(600)),
		)

		err := RunApp(window)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

// главный луп. даже слушает сообщения которые перекидываются от окон
func RunApp(w *app.Window) error {
	client, err := back.NewClient(URL)
	if err != nil {
		panic(err)
	}

	msgs := make(chan messages.Msg, 32)

	//вот тут мы чекаем существует ли session.key и вызываем регу или ауууф
	currentScreen := ui.NewLoginScreen(msgs)
	th := ui.NewAppTheme()

	var ops op.Ops
	for {
		select {
		case msg := <-msgs:
			HandleMessage(msg, currentScreen, client, msgs)
			w.Invalidate()
		default:
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				return e.Err
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				currentScreen.Update(gtx)
				currentScreen.Layout(gtx, th)
				e.Frame(gtx.Ops)

			}
		}
	}
}

func HandleMessage(msg messages.Msg, currentScreen interface{}, client *back.Client, msgs chan messages.Msg) {
	switch m := msg.(type) {

	// --- Навигация ---
	case messages.NavigateToLogin:
		//*currentScreen = ui.NewLogin(msgs, client)

	case messages.NavigateToRegister:
		//*currentScreen = ui.NewRegister(msgs, client)

	case messages.LoginAttempt:
		go func() {
			fmt.Println(m.Password)
			err := client.LoadKeys(KEY_PATH, m.Password)
			if err != nil {
				msgs <- messages.LoginFailed{}
				panic(err)
			}
			msgs <- messages.LoginSuccess{}
		}()

	case messages.LoginSuccess:
		if screen, ok := currentScreen.(*ui.LoginScreen); ok {
			screen.IsLoading = false // ← сбрасываем при успехе
		}
	}
}
