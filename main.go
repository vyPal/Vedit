package main

import (
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/vypal/vedit/ui/editor"
)

func main() {
	go func() {
		window := new(app.Window)
		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

/*
var LayoutManager = widgets.NewLayoutManager()

	func exampleSplit(th *material.Theme) {
		rootsplit := LayoutManager.AddSplit(nil, widgets.Vertical, 0.7, nil)
		LayoutManager.AddSplit(LayoutManager.RootSplit, widgets.Vertical, 0.5, func(gtx layout.Context) layout.Dimensions {
			editor := editor.NewEditor(th.Shaper)
			return editor.Layout(gtx, th)
		})
		rsplit := LayoutManager.AddSplit(LayoutManager.RootSplit, widgets.Horizontal, 0.5, nil)
		LayoutManager.AddSplit(rsplit, widgets.Horizontal, 0.5, func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, "Top", color.NRGBA{G: 0x80, A: 0xff})
		})
		LayoutManager.AddSplit(rsplit, widgets.Horizontal, 0.5, func(gtx layout.Context) layout.Dimensions {
			return FillWithLabel(gtx, th, "Bottom", color.NRGBA{B: 0x80, A: 0xff})
		})
		rootsplit.MinSize = 100
	}
*/
func run(window *app.Window) error {
	theme := material.NewTheme()
	editor := editor.NewEditor(theme.Shaper)
	//exampleSplit(theme)
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			editor.Layout(gtx, theme)

			//LayoutManager.Layout(gtx)

			// Pass the drawing operations to the GPU.
			e.Frame(&ops)
		}
	}
}

func FillWithLabel(gtx layout.Context, th *material.Theme, text string, backgroundColor color.NRGBA) layout.Dimensions {
	ColorBox(gtx, gtx.Constraints.Max, backgroundColor)
	return layout.Center.Layout(gtx, material.H3(th, text).Layout)
}

func ColorBox(gtx layout.Context, size image.Point, c color.NRGBA) layout.Dimensions {
	d := layout.Dimensions{Size: size}
	paintRect(gtx.Ops, c, d.Size)
	return d
}

func paintRect(ops *op.Ops, c color.NRGBA, size image.Point) {
	defer clip.Rect{Max: size}.Push(ops).Pop()
	paint.ColorOp{Color: c}.Add(ops)
	paint.PaintOp{}.Add(ops)
}
