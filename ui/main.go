package ui

import (
	"skam/back"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func NewMainScreen(msgs chan<- Msg, c *back.Client) *MainScreen {
	var message widget.Editor
	var search widget.Editor
	var send widget.Clickable
	sendIcon, err := widget.NewIcon(icons.ContentSend)
	if err != nil {
		sendIcon = nil
	}

	ms := &MainScreen{
		Client:  c,
		Message: message,
		Search:  search,
		FriendsList: layout.List{
			Axis: layout.Vertical,
		},
		MessagesList: layout.List{
			Axis: layout.Vertical,
		},
		SendBtn:          send,
		SendIcon:         *sendIcon,
		friendClickables: make(map[int]*widget.Clickable),
		inset:            layout.UniformInset(unit.Dp(16)),
		msgs:             msgs,
	}
	for _, friend := range c.Friends {
		ms.friendClickables[friend.Id] = &widget.Clickable{}
	}
	return ms
}

func (ms *MainScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {
	gtx = th.Background(gtx)

	return layout.NW.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Start,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = 400
				gtx.Constraints.Max.X = 400
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := th.TitleText("поиск")
						return ms.inset.Layout(gtx, text.Layout)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {

						return ms.FriendsList.Layout(gtx, len(ms.Client.Friends), func(gtx layout.Context, index int) layout.Dimensions {
							friend := ms.Client.Friends[index]
							click := ms.friendClickables[friend.Id]
							return material.Clickable(gtx, click, func(gtx layout.Context) layout.Dimensions {
								text := th.BodyText(friend.Name)
								return ms.inset.Layout(gtx, text.Layout)
							})
						})
					}),
				)
			}),
			layout.Flexed(7, func(gtx layout.Context) layout.Dimensions {
				if ms.Client.SelectedFriend != nil {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							text := th.TitleText(ms.Client.SelectedFriend.Name)
							return ms.inset.Layout(gtx, text.Layout)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {

							if ms.IsLoading {
								text := th.BodyText("Загрузка сообщений...")
								return ms.inset.Layout(gtx, text.Layout)
							} else {

								ms.MessagesList.Position.BeforeEnd = true
								return ms.MessagesList.Layout(gtx, len(ms.Client.SelectedFriend.Messages), func(gtx layout.Context, index int) layout.Dimensions {
									msg := ms.Client.SelectedFriend.Messages[index]
									text := th.BodyText(msg.Plaintext)
									aligment := layout.W
									if msg.Sender_id == ms.Client.Id {
										aligment = layout.E
									}
									return aligment.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return ms.inset.Layout(gtx, text.Layout)
									})

								})
							}
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis:      layout.Horizontal,
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									input := th.Input(&ms.Message, "Message")
									return ms.inset.Layout(gtx, input.Layout)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									btn := material.IconButton(th.Theme, &ms.SendBtn, &ms.SendIcon, "Send")
									btn.Size = unit.Dp(20)
									return ms.inset.Layout(gtx, btn.Layout)
								}),
							)
						}),
					)
				} else {
					return layout.Flex{
						Axis:      layout.Vertical,
						Alignment: layout.Middle,
					}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								text := th.BodyText("Выберите друга для начала чата")
								return ms.inset.Layout(gtx, text.Layout)
							})
						}),
					)
				}

			}),
		)
	})
}

func (ms *MainScreen) Update(gtx layout.Context) bool {
	changed := false

	for i, friend := range ms.Client.Friends {
		if clickable, exists := ms.friendClickables[friend.Id]; exists {
			if clickable.Clicked(gtx) && !ms.IsLoading {
				ms.Client.SelectedFriend = &ms.Client.Friends[i]
				changed = true
				ms.IsLoading = true
				ms.msgs <- FriendClicked{}
			}
		}
	}
	if ms.SendBtn.Clicked(gtx) {
		text := ms.Message.Text()
		if text != "" {
			ms.msgs <- SendMessage{Text: text}
			ms.Message.SetText("")
		}
	}

	return changed
}
