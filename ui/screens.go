package ui

import (
	"skam/back"
	"skam/messages"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
)

type Application struct {
	Window *app.Window

	Msgs          chan messages.Msg
	CurrentScreen Screen
	Client        *back.Client
	Theme         *AppTheme
	ErrorMsg      string
	ShowError     bool
	ErrorOkBtn    widget.Clickable
	KEY_PATH      string
	URL           string
}

type Screen interface {
	Update(gtx layout.Context) bool
	Layout(gtx layout.Context, th *AppTheme) layout.Dimensions
}

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

type RegisterScreen struct {
	BackBtn         widget.Clickable
	Name            widget.Editor
	ConfirmPassword widget.Editor
	Password        widget.Editor
	RegisterBtn     widget.Clickable

	inset layout.Inset

	msgs chan<- messages.Msg

	IsLoading     bool
	paswordsMatch bool
}

type ImportScreen struct {
	BackBtn       widget.Clickable
	Path          widget.Editor
	Password      widget.Editor
	Password2     widget.Editor
	ImportKeyBtn  widget.Clickable
	msgs          chan<- messages.Msg
	IsLoading     bool
	inset         layout.Inset
	paswordsMatch bool
}
