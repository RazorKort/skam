package ui

import "skam/back"

// базовый интерфейс для сообщений...
// что бы это не значило
type Msg interface {
	IsMsg()
}

// error popup
type ShowError struct {
	ErrorMessage string
	Msg
}
type HideError struct{ Msg }

// навигация
type NavigateToLogin struct{ Msg }
type NavigateToRegister struct{ Msg }
type NavigateToMain struct {
	c *back.Client
	Msg
}
type NavigateToImport struct{ Msg }

type MainFailed struct{ Msg }

// статусы логина
type LoginAttempt struct {
	Password string
	Msg
}
type LoginSuccess struct{ Msg }
type LoginFailed struct{ Msg }

// статусы окна регистрации
type RegisterAttempt struct {
	Password string
	Name     string
	Msg
}
type RegisterSuccess struct{ Msg }
type RegisterFailed struct{ Msg }

type ImportAttempt struct {
	Path     string
	Password string
	Msg
}
type ImportSuccess struct{ Msg }
type ImportFailed struct{ Msg }

type FriendClicked struct {
	Msg
}
type ShowDialog struct{ Msg }

type SendMessage struct {
	Text string
	Msg
}
