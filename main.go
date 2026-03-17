package main

import (
	"os"
	"skam/back"
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
		appState := ui.NewApplication(window)
		appState.KEY_PATH = KEY_PATH
		appState.URL = URL
		err := RunApp(appState)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

// главный луп. даже слушает сообщения которые перекидываются от окон
func RunApp(a *ui.Application) error {
	var err error
	a.Client, err = back.NewClient(a.URL)
	if err != nil {
		panic(err)
	}

	a.Theme = ui.NewAppTheme()

	if back.CheckPath(KEY_PATH) {
		a.CurrentScreen = ui.NewLoginScreen(a.Msgs)
	} else {
		a.CurrentScreen = ui.NewRegisterScreen(a.Msgs)
	}

	var ops op.Ops
	for {
		select {
		case msg := <-a.Msgs:
			a.HandleMessage(msg)
			a.Window.Invalidate()
		default:
			switch e := a.Window.Event().(type) {
			case app.DestroyEvent:
				return e.Err
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				a.CurrentScreen.Update(gtx)
				a.CurrentScreen.Layout(gtx, a.Theme)

				if a.ShowError {
					a.DrawError(gtx)
				}

				e.Frame(gtx.Ops)

			}
		}
	}
}
