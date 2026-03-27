package ui

import (
	"skam/back"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

func NewProfileScreen(msgs chan<- Msg, c *back.Client) *ProfileScreen {
	var backbtn widget.Clickable

	ps := &ProfileScreen{
		BackBtn: backbtn,
		inset:   layout.UniformInset(unit.Dp(16)),
		msgs:    msgs,
	}
	return ps
}

func (ps *ProfileScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {
	gtx = th.Background(gtx)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := th.Button(&ps.BackBtn, "Back arrow")
				return ps.inset.Layout(gtx, btn.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := th.BodyText(ps.Client.Name)
				return ps.inset.Layout(gtx, text.Layout)
			}),
		)
	})

}

func (ps *ProfileScreen) Update(gtx layout.Context) bool {
	changed := false
	if ps.BackBtn.Clicked(gtx) {
		ps.msgs <- NavigateToMain{}
		changed = true
	}
	return changed
}
