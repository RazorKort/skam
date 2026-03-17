package messages

// базовый интерфейс для сообщений...
// что бы это не значило
type Msg interface {
	IsMsg()
}

//error popup
type ShowError struct {
	ErrorMessage string
	Msg
}
type HideError struct{ Msg }

// навигация
type NavigateToLogin struct{ Msg }
type NavigateToRegister struct{ Msg }
type NavigateToMain struct{ Msg }

//статусы логина
type LoginAttempt struct {
	Password string
	Msg
}
type LoginSuccess struct{ Msg }
type LoginFailed struct{ Msg }

//статусы окна регистрации
type RegisterAttempt struct {
	Password string
	Name     string
	Msg
}
type RegisterSuccess struct{ Msg }
type RegisterFailed struct{ Msg }
