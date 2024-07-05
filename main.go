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
	"github.com/vypal/vedit/ui/toolbar"
	"github.com/vypal/vedit/ui/widgets"
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

var LayoutManager = widgets.NewLayoutManager()
var edit *editor.Editor

func exampleSplit(th *material.Theme) {
	edit = editor.NewEditor(th.Shaper)
	toolbar := toolbar.ToolBar{
		Items: []toolbar.ToolBarItem{
			&toolbar.Button{Text: "New", Theme: th},
		},
		BackgroundColor: color.NRGBA{R: 0x1a, G: 0x1b, B: 0x1b, A: 0xff},
	}
	// Vytvoření kořenového rozdělení
	rootsplit := LayoutManager.AddSplit(nil, widgets.Horizontal, 0.075, nil)

	// Přidání editoru do kořenového rozdělení
	LayoutManager.AddSplit(rootsplit, widgets.Vertical, 0, func(gtx layout.Context) layout.Dimensions {
		return toolbar.Layout(gtx)
	})

	// Vytvoření a přidání dalších rozdělení a komponent
	bsplit := LayoutManager.AddSplit(rootsplit, widgets.Vertical, -0.5, nil)
	LayoutManager.AddSplit(bsplit, widgets.Horizontal, 0.5, func(gtx layout.Context) layout.Dimensions {
		return FillWithLabel(gtx, th, "Files", color.NRGBA{R: 0x1a, G: 0x1b, B: 0x1b, A: 0xff})
	})
	LayoutManager.AddSplit(bsplit, widgets.Horizontal, 0.5, func(gtx layout.Context) layout.Dimensions {
		return edit.Layout(gtx, th)
	})

	// Nastavení minimální velikosti pro kořenové rozdělení
	rootsplit.Fixed = true
}

func run(window *app.Window) error {
	theme := material.NewTheme()
	exampleSplit(theme)
	var ops op.Ops
	for {

		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			paint.Fill(&ops, color.NRGBA{R: 0x1a, G: 0x1b, B: 0x1b, A: 0xff})
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			LayoutManager.Layout(gtx)

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
