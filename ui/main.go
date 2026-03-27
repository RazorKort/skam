package ui

import (
	"image"
	"skam/back"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func NewMainScreen(msgs chan<- Msg, c *back.Client) *MainScreen {
	var message widget.Editor
	var search widget.Editor
	search.SingleLine = true
	var searchBtn widget.Clickable
	var profileBtn widget.Clickable
	var send widget.Clickable
	sendIcon, err := widget.NewIcon(icons.ContentSend)
	if err != nil {
		sendIcon = nil
	}
	sendingIcon, err := widget.NewIcon(icons.ActionSchedule)
	if err != nil {
		sendingIcon = nil
	}
	sendedIcon, err := widget.NewIcon(icons.ActionDone)
	if err != nil {
		sendedIcon = nil
	}
	profileIcon, err := widget.NewIcon(icons.SocialPerson)
	if err != nil {
		profileIcon = nil
	}
	searchIcon, err := widget.NewIcon(icons.ActionSearch)
	if err != nil {
		searchIcon = nil
	}
	addIcon, err := widget.NewIcon(icons.ContentAdd)
	if err != nil {
		addIcon = nil
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
		SendingIcon:      *sendingIcon,
		SendedIcon:       *sendedIcon,
		SearchBtn:        searchBtn,
		ProfileBtn:       profileBtn,
		SearchIcon:       *searchIcon,
		ProfileIcon:      *profileIcon,
		AddIcon:          *addIcon,
		friendClickables: make(map[int]*widget.Clickable),
		inset:            layout.UniformInset(unit.Dp(16)),
		msgs:             msgs,
		friendSub:        false,
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
						return layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								btn := material.IconButton(th.Theme, &ms.ProfileBtn, &ms.ProfileIcon, "Navigate to profile")
								btn.Size = unit.Dp(20)
								btn.Background = th.Bg
								btn.Color = th.Colors.Secondary
								btn.Inset = layout.UniformInset(unit.Dp(10))
								inset := layout.UniformInset(unit.Dp(10))
								return inset.Layout(gtx, btn.Layout)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								input := th.Input(&ms.Search, "Find friends")
								return ms.inset.Layout(gtx, input.Layout)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								btn := material.IconButton(th.Theme, &ms.SearchBtn, &ms.SearchIcon, "Find smbd")
								btn.Size = unit.Dp(20)
								btn.Background = th.Bg
								btn.Color = th.Colors.Secondary
								btn.Inset = layout.UniformInset(unit.Dp(10))
								inset := layout.UniformInset(unit.Dp(10))
								return inset.Layout(gtx, btn.Layout)
							}),
						)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						if ms.friendSub {
							return ms.FriendsList.Layout(gtx, len(ms.Client.Find), func(gtx layout.Context, index int) layout.Dimensions {
								friend := ms.Client.Find[index]
								click := ms.friendClickables[friend.Id]
								return layout.Flex{
									Axis: layout.Horizontal,
								}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										text := th.BodyText(friend.Name)
										return ms.inset.Layout(gtx, text.Layout)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										btn := material.IconButton(th.Theme, click, &ms.AddIcon, "add friend to friends")
										return ms.inset.Layout(gtx, btn.Layout)
									}),
								)
							})
						} else {
							return ms.FriendsList.Layout(gtx, len(ms.Client.Friends), func(gtx layout.Context, index int) layout.Dimensions {
								friend := ms.Client.Friends[index]
								click := ms.friendClickables[friend.Id]
								return material.Clickable(gtx, click, func(gtx layout.Context) layout.Dimensions {
									text := th.BodyText(friend.Name)
									return ms.inset.Layout(gtx, text.Layout)
								})
							})
						}
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

								ms.MessagesList.ScrollToEnd = true
								return ms.MessagesList.Layout(gtx, len(ms.Client.SelectedFriend.Messages), func(gtx layout.Context, index int) layout.Dimensions {
									msg := ms.Client.SelectedFriend.Messages[index]
									text := th.BodyText(msg.Plaintext)
									aligment := layout.W
									if msg.Sender_id == ms.Client.Id {
										aligment = layout.E
									}
									return aligment.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										if aligment == layout.E {
											if msg.Sended {
												return layout.Flex{
													Axis: layout.Horizontal,
												}.Layout(gtx,
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														return ms.inset.Layout(gtx, text.Layout)
													}),
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														return layout.Flex{
															Axis: layout.Horizontal,
														}.Layout(gtx,
															layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
																return layout.Dimensions{}
															}),
															layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
																gtx.Constraints.Max = image.Pt(20, 20)
																return ms.SendedIcon.Layout(gtx, th.Colors.Secondary)
															}),
														)

													}))
											} else {
												return layout.Flex{
													Axis: layout.Horizontal,
												}.Layout(gtx,
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														return ms.inset.Layout(gtx, text.Layout)
													}),
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														gtx.Constraints.Max = image.Pt(20, 20)
														return ms.SendingIcon.Layout(gtx, th.Colors.Secondary)
													}),
												)
											}
										} else {
											return layout.Flex{
												Axis: layout.Horizontal,
											}.Layout(gtx,
												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
													return ms.inset.Layout(gtx, text.Layout)
												}),
											)
										}

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
				ms.friendSub = false
				ms.msgs <- FriendClicked{}
			}
		}
	}
	for _, user := range ms.Client.Find {
		if clickable, exists := ms.friendClickables[user.Id]; exists {
			if clickable.Clicked(gtx) && !ms.IsLoading {
				ms.msgs <- AddFriend{Id: user.Id}
				ms.IsLoading = true
				ms.friendSub = false
			}
		}
	}
	if ms.SendBtn.Clicked(gtx) {
		text := strings.TrimSpace(ms.Message.Text())
		if text != "" {
			changed = true
			ms.msgs <- SendMessage{Text: text}
			ms.MessagesList.Position.BeforeEnd = false
			ms.Message.SetText("")
		}

	}

	if ms.SearchBtn.Clicked(gtx) && !ms.IsLoading {
		text := strings.TrimSpace(ms.Search.Text())
		if text != "" {
			changed = true
			ms.IsLoading = true
			ms.msgs <- SearchUser{Text: text}
		}

	}

	if ms.ProfileBtn.Clicked(gtx) {
		changed = true
		ms.msgs <- NavigateToProfile{}

	}

	return changed
}
