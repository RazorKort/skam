package ui

import (
	"image"
	"skam/back"
	"strconv"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func NewProfileScreen(msgs chan<- Msg, c *back.Client) *ProfileScreen {
	backIcon, err := widget.NewIcon(icons.NavigationArrowBack)
	if err != nil {
		backIcon = nil
	}
	ps := &ProfileScreen{
		Client:   c,
		BackIcon: *backIcon,

		inset: layout.UniformInset(unit.Dp(16)),
		msgs:  msgs,
	}
	return ps
}

func (ps *ProfileScreen) Layout(gtx layout.Context, th *AppTheme) layout.Dimensions {
	gtx = th.Background(gtx)
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = gtx.Constraints.Max
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Pt(0, 0)
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := th.BodyText("Name " + ps.Client.Name)
						return ps.inset.Layout(gtx, text.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := th.BodyText("ID " + strconv.Itoa(ps.Client.Id))
						return ps.inset.Layout(gtx, text.Layout)
					}),
				)
			})
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			btn := material.IconButton(th.Theme, &ps.BackBtn, &ps.BackIcon, "back arrow")
			btn.Size = unit.Dp(20)
			btn.Background = th.Bg
			btn.Color = th.Colors.Secondary
			btn.Inset = layout.UniformInset(unit.Dp(10))
			inset := layout.UniformInset(unit.Dp(10))
			return inset.Layout(gtx, btn.Layout)
		}),
	)
}

func (ps *ProfileScreen) Update(gtx layout.Context) bool {
	changed := false
	if ps.BackBtn.Clicked(gtx) {
		ps.msgs <- NavigateToMain{}
		changed = true
	}
	return changed
}
