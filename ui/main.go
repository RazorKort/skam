package ui

import (
	"image"
	"skam/back"
	"strings"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func NewMainScreen(msgs chan<- Msg, c *back.Client) *MainScreen {
	var message widget.Editor
	var search widget.Editor
	search.SingleLine = true
	search.Submit = true

	closeIcon, err := widget.NewIcon(icons.NavigationClose)
	if err != nil {
		closeIcon = nil
	}
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
		SendIcon:         *sendIcon,
		SendingIcon:      *sendingIcon,
		SendedIcon:       *sendedIcon,
		SearchIcon:       *searchIcon,
		ProfileIcon:      *profileIcon,
		AddIcon:          *addIcon,
		CloseIcon:        *closeIcon,
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

				//попиксельное рисование, пиздец...
				var path clip.Path
				path.Begin(gtx.Ops)
				path.MoveTo(f32.Pt(float32(gtx.Constraints.Max.X), 0))
				path.LineTo(f32.Pt(float32(gtx.Constraints.Max.X), float32(gtx.Constraints.Max.Y)))
				pathSpec := path.End()
				paint.FillShape(gtx.Ops, th.Colors.Border, clip.Stroke{
					Path:  pathSpec,
					Width: 1,
				}.Op())

				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {

						dimis := layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(gtx,

							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								var button *widget.Clickable
								var icon *widget.Icon
								if ms.friendSub {
									button = &ms.CloseBtn
									icon = &ms.CloseIcon
								} else {
									button = &ms.ProfileBtn
									icon = &ms.ProfileIcon
								}
								btn := th.IconButtonSecondary(button, icon, "Navigate to profile or close")
								inset := layout.UniformInset(unit.Dp(10))
								inset.Bottom = 6
								return inset.Layout(gtx, btn.Layout)
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								input := th.Input(&ms.Search, "Find friends")
								return input.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								btn := th.IconButtonSecondary(&ms.SearchBtn, &ms.SearchIcon, "Find smbd")
								inset := layout.UniformInset(unit.Dp(10))
								inset.Bottom = 6
								return inset.Layout(gtx, btn.Layout)
							}),
						)

						var path clip.Path
						path.Begin(gtx.Ops)
						path.MoveTo(f32.Pt(0, float32(dimis.Size.Y)))
						path.LineTo(f32.Pt(float32(dimis.Size.X), float32(dimis.Size.Y)))
						pathSpec := path.End()
						paint.FillShape(gtx.Ops, th.Colors.Border, clip.Stroke{
							Path:  pathSpec,
							Width: 1,
						}.Op())

						return dimis
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						if ms.friendSub {
							if len(ms.Client.Find) == 0 {
								return layout.Flex{
									Axis: layout.Vertical,
								}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return layout.Inset{
											Top: unit.Dp(20),
										}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											text := th.BodyText("Nothing found")
											return layout.Center.Layout(gtx, text.Layout)
										})

									}),
								)
							} else {
								return ms.FriendsList.Layout(gtx, len(ms.Client.Find), func(gtx layout.Context, index int) layout.Dimensions {
									friend := ms.Client.Find[index]
									click := ms.friendClickables[friend.Id]
									return layout.Flex{
										Axis:      layout.Horizontal,
										Alignment: layout.Middle,
									}.Layout(gtx,
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											text := th.BodyText(friend.Name)
											return ms.inset.Layout(gtx, text.Layout)
										}),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											btn := th.IconButtonSecondary(click, &ms.AddIcon, "add friend to friends")
											inset := layout.UniformInset(unit.Dp(10))
											return inset.Layout(gtx, btn.Layout)
										}),
									)
								})
							}
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
								aligment := layout.Center
								return aligment.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									text := th.BodyText("Загрузка сообщений...")
									return ms.inset.Layout(gtx, text.Layout)
								})

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
													Axis:      layout.Horizontal,
													Alignment: layout.End,
												}.Layout(gtx,
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														return ms.inset.Layout(gtx, text.Layout)
													}),
													layout.Rigid(func(gtx layout.Context) layout.Dimensions {
														var inset layout.Inset
														inset = layout.Inset{Bottom: unit.Dp(10), Right: unit.Dp(6)}
														return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
															gtx.Constraints.Max = image.Pt(20, 20)
															return ms.SendedIcon.Layout(gtx, th.Colors.Secondary)
														})

													}))
											} else {
												return layout.Flex{
													Axis:      layout.Horizontal,
													Alignment: layout.End,
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
								Alignment: layout.End,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									inset2 := layout.UniformInset(0)
									inset2.Bottom = 14
									inset2.Left = 8
									return inset2.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return widget.Border{
											Color:        th.Colors.Border,
											CornerRadius: 8,
											Width:        1,
										}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											input := th.Input(&ms.Message, "Message")
											inset := layout.UniformInset(12)
											return inset.Layout(gtx, input.Layout)
										})
									})
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
				//он отлавливает когда ты нажимаешь на добавить в друзья уже добавленного друга...
				//не просто говнокод, а ГОВНОКОДИЩЕ
				ms.Search.SetText("")
				ms.msgs <- FriendClicked{}
			}
		}
	}
	for _, user := range ms.Client.Find {
		if clickable, exists := ms.friendClickables[user.Id]; exists {
			if clickable.Clicked(gtx) && !ms.IsLoading {
				ms.Search.SetText("")
				ms.msgs <- AddFriend{Id: user.Id}
				ms.IsLoading = true
				ms.friendSub = false
			}
		}
	}
	ev, _ := ms.Search.Update(gtx)
	if ev != nil && !ms.IsLoading {
		switch ev.(type) {
		case widget.SubmitEvent:
			text := strings.TrimSpace(ms.Search.Text())
			if text != "" {
				changed = true
				ms.IsLoading = true

				ms.msgs <- SearchUser{Text: text}
			}
		}
	}
	for {
		// Фильтруем нажатие Enter
		kev, ok := gtx.Event(key.Filter{Name: key.NameReturn})
		if !ok {
			break
		}

		if kev, ok := kev.(key.Event); ok && kev.State == key.Press {
			if kev.Modifiers.Contain(key.ModShift) {
				// Shift+Enter → вставляем новую строку
				ms.Message.Insert("\n")
				changed = true
			} else {
				text := ms.Message.Text()
				if text != "" {
					changed = true
					ms.msgs <- SendMessage{Text: text}
					ms.MessagesList.Position.BeforeEnd = false
					ms.Message.SetText("")
				}
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

	if ms.CloseBtn.Clicked(gtx) {
		changed = true
		ms.Search.SetText("")
		ms.friendSub = false
	}

	return changed
}
