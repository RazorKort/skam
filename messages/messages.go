package messages

// базовый интерфейс для сообщений...
// что бы это не значило
type Msg interface {
	IsMsg()
}

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
