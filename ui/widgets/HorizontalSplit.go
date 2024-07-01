package widgets

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

type HSplit struct {
	Ratio float32 // -1 to 1

	Bar unit.Dp

	drag   bool
	dragID pointer.ID
	dragY  float32
}

func (s *HSplit) Layout(gtx layout.Context, top, bottom layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarWidth)
	}

	proportion := (s.Ratio + 1) / 2
	topsize := int(proportion*float32(gtx.Constraints.Max.Y) - float32(bar))

	bottomoffset := topsize + bar
	bottomsize := gtx.Constraints.Max.Y - bottomoffset

	{ // handle input
		barRect := image.Rect(0, topsize, gtx.Constraints.Max.Y, bottomoffset)
		area := clip.Rect(barRect).Push(gtx.Ops)

		// register for input
		event.Op(gtx.Ops, s)
		pointer.CursorRowResize.Add(gtx.Ops)

		for {
			ev, ok := gtx.Event(pointer.Filter{
				Target: s,
				Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel,
			})
			if !ok {
				break
			}

			e, ok := ev.(pointer.Event)
			if !ok {
				continue
			}

			switch e.Kind {
			case pointer.Press:
				if s.drag {
					break
				}

				s.dragID = e.PointerID
				s.dragY = e.Position.Y
				s.drag = true

			case pointer.Drag:
				if s.dragID != e.PointerID {
					break
				}

				deltaY := e.Position.Y - s.dragY
				s.dragY = e.Position.Y

				deltaRatio := deltaY * 2 / float32(gtx.Constraints.Max.Y)
				s.Ratio += deltaRatio

				if e.Priority < pointer.Grabbed {
					gtx.Execute(pointer.GrabCmd{
						Tag: s,
						ID:  s.dragID,
					})
				}

			case pointer.Release:
				fallthrough
			case pointer.Cancel:
				s.drag = false
			}
		}

		area.Pop()
	}

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.Y, topsize))
		top(gtx)
	}

	{
		off := op.Offset(image.Pt(0, bottomoffset)).Push(gtx.Ops)
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.Y, bottomsize))
		bottom(gtx)
		off.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}
