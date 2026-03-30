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
	BackIcon        widget.Icon
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
	BackIcon      widget.Icon
	Path          widget.Editor
	Password      widget.Editor
	Password2     widget.Editor
	ImportKeyBtn  widget.Clickable
	msgs          chan<- Msg
	IsLoading     bool
	inset         layout.Inset
	paswordsMatch bool
	needToScroll  bool
}

type MainScreen struct {
	Client           *back.Client
	FriendsList      layout.List
	friendClickables map[int]*widget.Clickable
	MessagesList     layout.List
	ProfileBtn       widget.Clickable
	ProfileIcon      widget.Icon
	Message          widget.Editor
	SendBtn          widget.Clickable
	SendIcon         widget.Icon
	SendingIcon      widget.Icon
	SendedIcon       widget.Icon
	Search           widget.Editor
	SearchBtn        widget.Clickable
	SearchIcon       widget.Icon
	CloseBtn         widget.Clickable
	CloseIcon        widget.Icon
	inset            layout.Inset
	AddIcon          widget.Icon
	IsLoading        bool
	friendSub        bool

	msgs chan<- Msg
}

type ProfileScreen struct {
	Client   *back.Client
	BackIcon widget.Icon
	BackBtn  widget.Clickable
	inset    layout.Inset
	msgs     chan<- Msg
}
