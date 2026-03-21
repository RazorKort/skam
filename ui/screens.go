package ui

import (
	"skam/back"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
)

type Application struct {
	Window *app.Window

	Msgs          chan Msg
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

type LoginScreen struct {
	Password     widget.Editor
	LoginButton  widget.Clickable
	ImportKeyBtn widget.Clickable
	RegisterBtn  widget.Clickable
	inset        layout.Inset

	msgs chan<- Msg

	IsLoading bool
}

type RegisterScreen struct {
	BackBtn         widget.Clickable
	Name            widget.Editor
	ConfirmPassword widget.Editor
	Password        widget.Editor
	RegisterBtn     widget.Clickable

	inset layout.Inset

	msgs chan<- Msg

	IsLoading     bool
	paswordsMatch bool
}

type ImportScreen struct {
	BackBtn       widget.Clickable
	Path          widget.Editor
	Password      widget.Editor
	Password2     widget.Editor
	ImportKeyBtn  widget.Clickable
	msgs          chan<- Msg
	IsLoading     bool
	inset         layout.Inset
	paswordsMatch bool
}

type MainScreen struct {
	Client           *back.Client
	FriendsList      layout.List
	friendClickables map[int]*widget.Clickable
	SelectedFriend   *back.User
	ProfileBtn       widget.Clickable
	Message          widget.Editor
	SendBtn          widget.Clickable
	Search           widget.Editor
	SearchBtn        widget.Clickable
	inset            layout.Inset
	IsLoading        bool
	msgs             chan<- Msg
}
