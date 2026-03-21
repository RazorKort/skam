package ui

import (
	"skam/back"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func NewMainScreen(msgs chan<- Msg, c *back.Client) *MainScreen {
	var message widget.Editor
	var search widget.Editor

	ms := &MainScreen{
		Client:  c,
		Message: message,
		Search:  search,
		FriendsList: layout.List{
			Axis: layout.Vertical,
		},
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
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := th.TitleText("имя")
						return ms.inset.Layout(gtx, text.Layout)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						text := th.TitleText("чат")
						return ms.inset.Layout(gtx, text.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := th.TitleText("send msg")
						return ms.inset.Layout(gtx, text.Layout)
					}),
				)
			}),
		)
	})
}

func (ms *MainScreen) Update(gtx layout.Context) bool {
	changed := false

	for _, friend := range ms.Client.Friends {
		if clickable, exists := ms.friendClickables[friend.Id]; exists {
			if clickable.Clicked(gtx) {
				ms.SelectedFriend = &friend
				changed = true
				ms.IsLoading = true
				ms.msgs <- FriendClicked{}
			}
		}
	}

	return changed
}
