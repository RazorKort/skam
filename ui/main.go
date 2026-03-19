package ui

import (
	"skam/messages"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

func NewMainScreen(msgs <-chan messages.Msg) *MainScreen {
	var message widget.Editor
	var search widget.Editor

	return &MainScreen{
		Message: message,
		Search:  search,
		inset:   layout.UniformInset(unit.Dp(16)),
	}
}

func (ms *MainScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {
	gtx = th.Background(gtx)

	return layout.NW.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Baseline,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := th.TitleText("поиск")
						return ms.inset.Layout(gtx, text.Layout)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						text := th.BodyText("друзья")
						return ms.inset.Layout(gtx, text.Layout)
					}),
				)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
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

	return changed
}
