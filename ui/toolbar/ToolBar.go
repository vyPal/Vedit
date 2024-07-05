package toolbar

import (
	"image"
	"image/color"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type ToolBar struct {
	Items           []ToolBarItem
	BackgroundColor color.NRGBA
}

type ToolBarItem interface {
	Layout(gtx layout.Context) layout.Dimensions
}

func (tb *ToolBar) Layout(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, tb.BackgroundColor)
	children := make([]layout.FlexChild, len(tb.Items)+1)
	for i, item := range tb.Items {
		children[i] = layout.Rigid(item.Layout)
	}
	children[len(tb.Items)] = layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return layout.Spacer{}.Layout(gtx)
	})
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
}

type Button struct {
	Text    string
	Theme   *material.Theme
	OnClick func()
	hovered bool
}

func (b *Button) Layout(gtx layout.Context) layout.Dimensions {
	// Create an area for pointer input.
	rect := clip.Rect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y}})
	pointerArea := rect.Push(gtx.Ops)

	// Listen for pointer events.
	event.Op(gtx.Ops, b)
	pointerArea.Pop()

	// Process pointer events.
	for {
		e, ok := gtx.Event(pointer.Filter{Target: b, Kinds: pointer.Press | pointer.Enter | pointer.Leave})
		if !ok {
			break
		}
		switch e := e.(type) {
		case pointer.Event:
			switch e.Kind {
			case pointer.Enter:
				b.hovered = true
			case pointer.Leave:
				b.hovered = false
			case pointer.Press:
				if b.OnClick != nil {
					b.OnClick()
				}
			}
		}
	}

	/* return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if b.hovered {
				paint.ColorOp{Color: b.Theme.ContrastBg}.Add(gtx.Ops)
				defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
				paint.PaintOp{}.Add(gtx.Ops)
			}

			lbl := material.Label(b.Theme, unit.Sp(16), b.Text)
			lbl.Color = b.Theme.ContrastBg
			return lbl.Layout(gtx)
		})
	}) */

	// Use Stack to overlay the bounding box and label.
	return layout.Stack{}.Layout(gtx,
		// Draw bounding box if hovered.
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			if b.hovered {
				paint.ColorOp{Color: b.Theme.ContrastBg}.Add(gtx.Ops)
				defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
				paint.PaintOp{}.Add(gtx.Ops)
			}
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		// Layout the label with padding.
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(b.Theme, unit.Sp(16), b.Text)
				lbl.Color = b.Theme.ContrastBg
				return lbl.Layout(gtx)
			})
		}),
	)
}
