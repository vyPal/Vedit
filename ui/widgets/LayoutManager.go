package widgets

import (
	"image"
	"image/color"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type Direction int

const (
	Vertical Direction = iota
	Horizontal
)

const defaultBarWidth = unit.Dp(2)
const defaultBarHeight = unit.Dp(2)

type Split struct {
	Direction   Direction
	Ratio       float32
	MinSize     unit.Dp
	Fixed       bool
	FirstChild  *Split
	SecondChild *Split
	Widget      layout.Widget
	dragging    bool
	dragX       float32
	dragY       float32
	dragID      pointer.ID
}

type LayoutManager struct {
	RootSplit *Split
}

func NewLayoutManager() *LayoutManager {
	return &LayoutManager{}
}

func (lm *LayoutManager) Layout(gtx layout.Context) layout.Dimensions {
	if lm.RootSplit == nil {
		return layout.Dimensions{}
	}
	return lm.layoutSplit(gtx, lm.RootSplit)
}

func (lm *LayoutManager) layoutSplit(gtx layout.Context, split *Split) layout.Dimensions {
	if split.FirstChild == nil || split.SecondChild == nil {
		if split.Widget != nil {
			stack := clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops)
			dims := split.Widget(gtx)
			stack.Pop()
			return dims
		}
		return layout.Dimensions{}
	}

	var dims layout.Dimensions
	switch split.Direction {
	case Vertical:
		dims = lm.layoutFlexSplit(gtx, layout.Vertical, split)
	case Horizontal:
		dims = lm.layoutFlexSplit(gtx, layout.Horizontal, split)
	}
	return dims
}

func (lm *LayoutManager) layoutFlexSplit(gtx layout.Context, axis layout.Axis, split *Split) layout.Dimensions {
	var barSize int
	if axis == layout.Vertical {
		barSize = gtx.Dp(defaultBarWidth)
	} else {
		barSize = gtx.Dp(defaultBarHeight)
	}

	var firstSize int
	if axis == layout.Vertical {
		firstSize = int((split.Ratio+1)/2*float32(gtx.Constraints.Max.X) - float32(barSize))
	} else {
		firstSize = int(split.Ratio*float32(gtx.Constraints.Max.Y) - float32(barSize))
	}
	secondOffset := firstSize + barSize
	var secondSize int
	if axis == layout.Vertical {
		secondSize = gtx.Constraints.Max.X - secondOffset
	} else {
		secondSize = gtx.Constraints.Max.Y - secondOffset
	}

	split.handleInput(gtx, axis, firstSize)

	// Layout first child
	{
		gtx := gtx
		if axis == layout.Vertical {
			gtx.Constraints = layout.Exact(image.Pt(firstSize, gtx.Constraints.Max.Y))
		} else {
			gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.X, firstSize))
		}
		lm.layoutSplit(gtx, split.FirstChild)
	}

	// Layout second child
	{
		var off op.TransformStack
		gtx := gtx
		if axis == layout.Vertical {
			off = op.Offset(image.Pt(secondOffset, 0)).Push(gtx.Ops)
			gtx.Constraints = layout.Exact(image.Pt(secondSize, gtx.Constraints.Max.Y))
		} else {
			off = op.Offset(image.Pt(0, secondOffset)).Push(gtx.Ops)
			gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.X, secondSize))
		}
		lm.layoutSplit(gtx, split.SecondChild)
		off.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (split *Split) handleInput(gtx layout.Context, axis layout.Axis, firstSize int) {
	var barRect image.Rectangle
	if axis == layout.Vertical {
		barRect = image.Rect(firstSize, 0, firstSize+gtx.Dp(defaultBarWidth), gtx.Constraints.Max.Y)
	} else {
		barRect = image.Rect(0, firstSize, gtx.Constraints.Max.X, firstSize+gtx.Dp(defaultBarHeight))
	}

	var barColor = color.NRGBA{R: 0x5A, G: 0x72, B: 0xB2, A: 0xFF}

	var expandedBarRect image.Rectangle
	if axis == layout.Vertical {
		expandedBarRect = image.Rect(barRect.Min.X-5, barRect.Min.Y, barRect.Max.X+5, barRect.Max.Y)
	} else {
		expandedBarRect = image.Rect(barRect.Min.X, barRect.Min.Y-5, barRect.Max.X, barRect.Max.Y+5)
	}

	area := clip.Rect(expandedBarRect).Push(gtx.Ops)

	paint.ColorOp{Color: barColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	if !split.Fixed {
		pointer.PassOp.Push(pointer.PassOp{}, gtx.Ops)
		event.Op(gtx.Ops, split)
		if axis == layout.Vertical {
			pointer.CursorColResize.Add(gtx.Ops)
		} else {
			pointer.CursorRowResize.Add(gtx.Ops)
		}
		for {
			e, ok := gtx.Event(pointer.Filter{
				Target: split,
				Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel,
			})
			if !ok {
				break
			}

			switch e := e.(type) {
			case pointer.Event:
				switch e.Kind {
				case pointer.Press:
					if split.dragging {
						break
					}
					split.dragging = true
					split.dragX = e.Position.X
					split.dragY = e.Position.Y
					split.dragID = e.PointerID
				case pointer.Drag:
					if e.PointerID != split.dragID {
						break
					}
					if axis == layout.Vertical {
						deltaX := e.Position.X - split.dragX
						split.dragX = e.Position.X

						deltaRatio := float32(deltaX) * 2 / float32(gtx.Constraints.Max.X)
						split.Ratio += deltaRatio
					} else {
						deltaY := e.Position.Y - split.dragY
						split.dragY = e.Position.Y

						deltaRatio := float32(deltaY) / float32(gtx.Constraints.Max.Y)
						split.Ratio += deltaRatio
					}
					if e.Priority < pointer.Grabbed {
						gtx.Execute(pointer.GrabCmd{
							Tag: split,
							ID:  split.dragID,
						})
					}
				case pointer.Release, pointer.Cancel:
					split.dragging = false
				}
			}
		}
	}
	area.Pop()
}

func (lm *LayoutManager) AddSplit(parent *Split, direction Direction, ratio float32, widget layout.Widget) *Split {
	newSplit := &Split{
		Direction: direction,
		Ratio:     ratio,
		Widget:    widget,
	}
	if parent == nil {
		lm.RootSplit = newSplit
	} else {
		if parent.FirstChild == nil {
			parent.FirstChild = newSplit
		} else if parent.SecondChild == nil {
			parent.SecondChild = newSplit
		} else {
			// TODO: handle error
		}
	}
	return newSplit
}
