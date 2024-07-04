package editor

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"unicode"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type Editor struct {
	content      []rune
	cursor       int
	scrollOffset int
	fontSize     unit.Sp
	lineHeight   unit.Sp
	textColor    color.NRGBA
	bgColor      color.NRGBA
	lineNumColor color.NRGBA
	shaper       *text.Shaper
	focused      bool
}

func NewEditor(shaper *text.Shaper) *Editor {
	return &Editor{
		content:      []rune{},
		cursor:       0,
		fontSize:     unit.Sp(22),
		lineHeight:   unit.Sp(26),
		textColor:    color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		bgColor:      color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		lineNumColor: color.NRGBA{R: 150, G: 150, B: 150, A: 255},
		shaper:       shaper,
		focused:      true,
	}
}

func (e *Editor) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	event.Op(gtx.Ops, e)
	for {
		ev, ok := gtx.Event(key.Filter{Focus: nil, Optional: key.ModShift})
		if !ok {
			break
		}
		switch ev := ev.(type) {
		case key.FocusEvent:
			e.focused = ev.Focus
		case key.Event:
			e.HandleKey(ev)
		}
	}

	paint.Fill(gtx.Ops, e.bgColor)

	lines := e.getLines()
	visibleLines := e.getVisibleLines()
	startLine := e.scrollOffset
	endLine := min(startLine+visibleLines, len(lines))

	lineNumWidth := e.drawLineNumbers(gtx, th, startLine, endLine)

	contentOffset := lineNumWidth + 20 // TODO: Maybe make this configurable
	e.drawContent(gtx, th, lines, startLine, endLine, contentOffset)

	e.drawCursor(gtx, th, float32(contentOffset))

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (e *Editor) drawLineNumbers(gtx layout.Context, th *material.Theme, startLine, endLine int) int {
	maxWidth := 0
	for lineNum := startLine; lineNum < endLine; lineNum++ {
		lineNumStr := fmt.Sprintf("%d", lineNum+1)
		lbl := material.Label(th, e.fontSize, lineNumStr)
		lbl.Color = e.lineNumColor
		stack := op.Offset(image.Point{Y: int(e.lineHeight) * (lineNum - startLine)}).Push(gtx.Ops)
		dims := lbl.Layout(gtx)
		stack.Pop()
		maxWidth = max(maxWidth, dims.Size.X)
	}
	return maxWidth
}

func (e *Editor) drawContent(gtx layout.Context, th *material.Theme, lines []string, startLine, endLine, xOffset int) {
	for lineNum := startLine; lineNum < endLine; lineNum++ {
		if lineNum >= len(lines) {
			break
		}
		line := lines[lineNum]
		lbl := material.Label(th, e.fontSize, line)
		lbl.Color = e.textColor

		// Create a new context with adjusted constraints
		/* textGtx := gtx
		textGtx.Constraints = layout.Constraints{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: gtx.Constraints.Max.X - xOffset, Y: int(e.lineHeight)},
		} */

		stack := op.Offset(image.Point{X: xOffset - gtx.Constraints.Max.X, Y: (lineNum - startLine) * int(e.lineHeight)}).Push(gtx.Ops)
		lbl.Layout(gtx)
		stack.Pop()
	}
}

func (e *Editor) drawCursor(gtx layout.Context, th *material.Theme, xOffset float32) {
	cursorLine, cursorCol := e.getCursorPosition()
	if cursorLine < e.scrollOffset || cursorLine >= e.scrollOffset+int(gtx.Constraints.Max.Y/int(e.lineHeight)) {
		return // Cursor is not in view
	}

	lines := e.getLines()
	cursorX := int(xOffset)
	if cursorCol > 0 && cursorLine < len(lines) {
		line := lines[cursorLine][:cursorCol]
		lbl := material.Label(th, e.fontSize, line)
		stack := op.Offset(image.Point{X: cursorX - gtx.Constraints.Max.X, Y: (cursorLine - e.scrollOffset) * int(e.lineHeight)}).Push(gtx.Ops)
		dims := lbl.Layout(gtx)
		stack.Pop()
		cursorX += dims.Size.X
	}

	cursorY := (cursorLine - e.scrollOffset) * int(e.lineHeight)

	cursorColor := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	paint.FillShape(gtx.Ops,
		cursorColor,
		clip.Rect{
			Min: image.Point{X: cursorX, Y: cursorY},
			Max: image.Point{X: cursorX + 2, Y: cursorY + int(e.lineHeight)},
		}.Op(),
	)
}

/* func (e *Editor) drawCursor(gtx layout.Context, th *material.Theme, xOffset float32) {
	if e.cursor.Y < e.scrollOffset || e.cursor.Y >= e.scrollOffset+int(gtx.Constraints.Max.Y/int(e.lineHeight)) {
		return // Cursor is not in view
	}

	cursorX := int(xOffset)
	if e.cursor.X > 0 && e.cursor.Y < len(e.content) {
		line := string(e.content[e.cursor.Y][:e.cursor.X])
		lbl := material.Label(th, e.fontSize, line)
		dims := lbl.Layout(gtx)
		cursorX += dims.Size.X
	}

	cursorY := (e.cursor.Y - e.scrollOffset) * int(e.lineHeight)

	cursorColor := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	paint.FillShape(gtx.Ops,
		cursorColor,
		clip.Rect{
			Min: image.Point{X: cursorX, Y: cursorY},
			Max: image.Point{X: cursorX + 2, Y: cursorY + int(e.lineHeight)},
		}.Op(),
	)
} */

func (e *Editor) Insert(text string) {
	runes := []rune(text)
	e.content = append(e.content[:e.cursor], append(runes, e.content[e.cursor:]...)...)
	e.cursor += len(runes)
	if text == "\n" {
		curLine, _ := e.getCursorPosition()
		if curLine >= e.scrollOffset+e.getVisibleLines() {
			e.scrollOffset = curLine - e.getVisibleLines() + 1
		}
	}
}

func (e *Editor) HandleKey(ev key.Event) {
	if !e.focused {
		return
	}

	switch ev.State {
	case key.Press:
		switch ev.Name {
		case key.NameLeftArrow:
			e.MoveCursor(e.cursor - 1)
		case key.NameRightArrow:
			e.MoveCursor(e.cursor + 1)
		case key.NameUpArrow:
			e.moveCursorUp()
		case key.NameDownArrow:
			e.moveCursorDown()
		case key.NameReturn:
			e.Insert("\n")
		case key.NameDeleteBackward:
			e.backspace()
		case key.NameDeleteForward:
			e.delete()
		default:
			if ev.Modifiers == 0 && unicode.IsPrint(rune(ev.Name[0])) && ev.Name != "Shift" && ev.Name != "Space" {
				e.Insert(strings.ToLower(string(ev.Name)))
			} else if ev.Modifiers == 0 && ev.Name == "Space" {
				e.Insert(" ")

			} else if ev.Modifiers.Contain(key.ModShift) && unicode.IsPrint(rune(ev.Name[0])) {
				e.Insert(strings.ToUpper(string(ev.Name)))
			}
		}
	}
}

func (e *Editor) MoveCursor(pos int) {
	e.cursor = max(0, min(pos, len(e.content)))
	e.adjustScrollOffset()
}

func (e *Editor) moveCursorUp() {
	curLine, curCol := e.getCursorPosition()
	if curLine > 0 {
		prevLineStart := e.getLineStart(curLine - 1)
		prevLineEnd := e.getLineEnd(curLine - 1)
		e.MoveCursor(prevLineStart + min(curCol, prevLineEnd-prevLineStart))
	}
}

func (e *Editor) moveCursorDown() {
	curLine, curCol := e.getCursorPosition()
	lines := e.getLines()
	if curLine < len(lines)-1 {
		nextLineStart := e.getLineStart(curLine + 1)
		nextLineEnd := e.getLineEnd(curLine + 1)
		e.MoveCursor(nextLineStart + min(curCol, nextLineEnd-nextLineStart))
	}
}

func (e *Editor) adjustScrollOffset() {
	curLine, _ := e.getCursorPosition()
	visibleLines := e.getVisibleLines()

	if curLine < e.scrollOffset {
		e.scrollOffset = curLine
	} else if curLine >= e.scrollOffset+visibleLines {
		e.scrollOffset = curLine - visibleLines + 1
	}
}

func (e *Editor) getVisibleLines() int {
	return 20 // TODO: Calculate
}

func (e *Editor) getLineStart(lineNum int) int {
	lines := e.getLines()
	if lineNum < 0 || lineNum >= len(lines) {
		return 0
	}
	return strings.Index(string(e.content), strings.Join(lines[:lineNum], "\n")) + 1
}

func (e *Editor) getLineEnd(lineNum int) int {
	lines := e.getLines()
	if lineNum < 0 || lineNum >= len(lines) {
		return 0
	}
	return strings.Index(string(e.content), strings.Join(lines[:lineNum+1], "\n"))
}

func (e *Editor) getLines() []string {
	return strings.Split(string(e.content), "\n")
}

func (e *Editor) getCursorPosition() (int, int) {
	//lines := e.getLines()
	curLine := 0
	curCol := 0
	for _, ch := range e.content[:e.cursor] {
		if ch == '\n' {
			curLine++
			curCol = 0
		} else {
			curCol++
		}
	}
	return curLine, curCol
}

func (e *Editor) Delete(start, end int) {
	if start < end {
		e.content = append(e.content[:start], e.content[end:]...)
		e.cursor = start
	}
}

func (e *Editor) backspace() {
	if e.cursor > 0 {
		e.Delete(e.cursor-1, e.cursor)
	}
}

func (e *Editor) delete() {
	if e.cursor < len(e.content) {
		e.Delete(e.cursor, e.cursor+1)
	}
}
