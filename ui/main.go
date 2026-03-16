package ui

import (
	"errors"
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Chat represents a single chat in the sidebar.
type Chat struct {
	Name         string
	Participants string
	LastMessage  string
}

// Message represents a single message in the chat.
type Message struct {
	FromMe  bool
	Content string
}

// MainScreen holds the UI state for the main chat layout.
type MainScreen struct {
	Chats          []Chat
	Messages       []Message
	SelectedChat   int
	ChatListClicks []widget.Clickable

	MessageEditor widget.Editor
	SendButton    widget.Clickable
}

// RunMain is a minimal GioUI window loop that renders the main chat screen.
func RunMain(w *app.Window) error {
	if w == nil {
		return errors.New("nil window")
	}
	th := NewAppTheme()
	screen := NewMainScreen()

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			screen.Layout(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}

// NewMainScreen initializes a basic main chat UI with placeholder data.
func NewMainScreen() *MainScreen {
	ms := &MainScreen{
		Chats: []Chat{
			{Name: "General", Participants: "Alice, Bob, You", LastMessage: "See you later!"},
			{Name: "Project X", Participants: "Team", LastMessage: "Deadline is tomorrow."},
			{Name: "Random", Participants: "Friends", LastMessage: "Check this out."},
		},
		Messages: []Message{
			{FromMe: false, Content: "Hey, how's it going?"},
			{FromMe: true, Content: "All good here, working on the new UI."},
			{FromMe: false, Content: "Nice! Can't wait to see it."},
		},
		SelectedChat: 0,
	}

	ms.MessageEditor.SingleLine = true

	ms.ChatListClicks = make([]widget.Clickable, len(ms.Chats))

	return ms
}

// Layout renders the main chat application layout.
func (ms *MainScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		// Left sidebar: chat list
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			sidebarWidth := gtx.Constraints.Max.X / 4 // responsive: ~25% width
			if sidebarWidth < gtx.Dp(unit.Dp(200)) {
				sidebarWidth = gtx.Dp(unit.Dp(200))
			}
			if sidebarWidth > gtx.Dp(unit.Dp(320)) {
				sidebarWidth = gtx.Dp(unit.Dp(320))
			}

			gtx.Constraints.Min.X = sidebarWidth
			gtx.Constraints.Max.X = sidebarWidth

			return layout.Stack{}.Layout(gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					r := image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
					paintFill(gtx, th.Colors.Surface, r)
					return layout.Dimensions{Size: gtx.Constraints.Max}
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					inset := layout.UniformInset(unit.Dp(8))
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis:      layout.Vertical,
							Alignment: layout.Start,
						}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								header := th.TitleText("Chats")
								return layout.UniformInset(unit.Dp(8)).Layout(gtx, header.Layout)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								// Simple vertical list of chats.
								return layout.Flex{
									Axis:      layout.Vertical,
									Alignment: layout.Start,
								}.Layout(gtx, ms.chatListItems(gtx, th)...)
							}),
						)
					})
				}),
			)
		}),

		// Main area: header + message list + input
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Start,
			}.Layout(gtx,
				// Header
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ms.layoutHeader(gtx, th)
				}),
				// Message list
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return ms.layoutMessages(gtx, th)
				}),
				// Input area
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ms.layoutInput(gtx, th)
				}),
			)
		}),
	)
}

func (ms *MainScreen) chatListItems(gtx layout.Context, th *AppTheme) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(ms.Chats))
	for i := range ms.Chats {
		idx := i
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.ButtonLayout(th.Theme, &ms.ChatListClicks[idx])
				return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// Highlight selected chat
					bg := th.Colors.Surface
					if idx == ms.SelectedChat {
						bg = th.Colors.Primary
					}
					r := image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Dp(unit.Dp(56)))
					paintFill(gtx, bg, r)

					inset := layout.UniformInset(unit.Dp(8))
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						chat := ms.Chats[idx]
						name := th.BodyText(chat.Name)
						participants := th.CaptionSecondary(chat.Participants)
						return layout.Flex{
							Axis:      layout.Vertical,
							Alignment: layout.Start,
						}.Layout(gtx,
							layout.Rigid(name.Layout),
							layout.Rigid(participants.Layout),
						)
					})
				})
			}),
		)
	}
	return children
}

func (ms *MainScreen) layoutHeader(gtx layout.Context, th *AppTheme) layout.Dimensions {
	if len(ms.Chats) == 0 {
		return layout.Dimensions{}
	}
	chat := ms.Chats[ms.SelectedChat]

	headerHeight := gtx.Dp(unit.Dp(56))
	r := image.Rect(0, 0, gtx.Constraints.Max.X, headerHeight)

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			paintFill(gtx, th.Colors.Surface, r)
			return layout.Dimensions{Size: r.Size()}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			inset := layout.UniformInset(unit.Dp(12))
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				title := th.BodyText(chat.Name)
				sub := th.CaptionSecondary(chat.Participants)
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Start,
				}.Layout(gtx,
					layout.Rigid(title.Layout),
					layout.Rigid(sub.Layout),
				)
			})
		}),
	)
}

func (ms *MainScreen) layoutMessages(gtx layout.Context, th *AppTheme) layout.Dimensions {
	inset := layout.UniformInset(unit.Dp(8))

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Start,
		}.Layout(gtx, ms.messageItems(gtx, th)...)
	})
}

func (ms *MainScreen) messageItems(gtx layout.Context, th *AppTheme) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(ms.Messages))
	for i := range ms.Messages {
		msg := ms.Messages[i]
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(4),
					Bottom: unit.Dp(4),
					Left:   unit.Dp(4),
					Right:  unit.Dp(4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// Align bubbles left/right based on sender.
					align := layout.W
					bg := th.Colors.Surface
					if msg.FromMe {
						align = layout.E
						bg = th.Colors.Primary
					}

					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Start,
					}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Stack{Alignment: align}.Layout(gtx, layout.Stacked(func(gtx layout.Context) layout.Dimensions {
								bubbleInset := layout.UniformInset(unit.Dp(8))
								return bubbleInset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									// Bubble background.
									defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Max}, gtx.Dp(unit.Dp(12))).Push(gtx.Ops).Pop()
									paintFill(gtx, bg, image.Rectangle{Max: gtx.Constraints.Max})

									label := th.BodyText(msg.Content)
									if msg.FromMe {
										label = th.BodyTextOnPrimary(msg.Content)
									}
									return label.Layout(gtx)
								})
							}))
						}),
					)
				})
			}),
		)
	}
	return children
}

func (ms *MainScreen) layoutInput(gtx layout.Context, th *AppTheme) layout.Dimensions {
	inset := layout.UniformInset(unit.Dp(8))

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				editor := th.Input(&ms.MessageEditor, "Type a message")
				return editor.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := th.SecondaryButton(&ms.SendButton, "Send")
				inset := layout.Inset{Left: unit.Dp(8)}
				return inset.Layout(gtx, btn.Layout)
			}),
		)
	})
}

// paintFill fills the given rectangle with a solid color.

func paintFill(gtx layout.Context, c color.NRGBA, r image.Rectangle) {
	op := clip.Rect(r).Push(gtx.Ops) // Push возвращает ClipOp
	paint.Fill(gtx.Ops, c)
	op.Pop() // снимаем клип
}
