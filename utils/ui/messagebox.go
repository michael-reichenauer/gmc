package ui

import (
	"github.com/jroimartin/gocui"
	"strings"
)

type MessageBox struct {
	ui           *UI
	boxView      View
	textView     View
	buttonsView  View
	text         string
	title        string
	previousView View
}

func NewMessageBox(ui *UI, text, title string) *MessageBox {
	return &MessageBox{ui: ui, text: text, title: title}
}

func (h *MessageBox) Show() {
	h.previousView = h.ui.CurrentView()

	h.boxView = h.newBoxView()
	h.buttonsView = h.newButtonsView()
	h.textView = h.newTextView()

	bb, tb, bbb := h.getBounds()
	h.boxView.Show(bb)
	h.buttonsView.Show(bbb)
	h.textView.Show(tb)

	h.boxView.SetTop()
	h.buttonsView.SetTop()
	h.textView.SetTop()
	h.textView.SetCurrentView()
}

func (h *MessageBox) newBoxView() View {
	view := h.ui.NewView("")
	view.Properties().Title = h.title
	return view
}

func (h *MessageBox) newButtonsView() View {
	view := h.ui.NewView("[OK]")
	view.Properties().OnMouseLeft = h.onButtonsClick
	return view
}

func (h *MessageBox) newTextView() View {
	view := h.ui.NewView(h.text)
	view.Properties().HideCurrentLineMarker = true
	view.SetKey(gocui.KeyEsc, gocui.ModNone, h.Close)
	view.SetKey(gocui.KeyEnter, gocui.ModNone, h.Close)
	return view
}

func (h *MessageBox) Close() {
	h.textView.Close()
	h.buttonsView.Close()
	h.boxView.Close()
	h.ui.SetCurrentView(h.previousView)
}

func (h *MessageBox) getBounds() (Rect, Rect, Rect) {
	lines := strings.Split(h.text, "\n")
	windowWidth, windowHeight := h.ui.WindowSize()

	width := h.maxTextWidth(lines)
	if width < 30 {
		width = 30
	}
	if width > 70 {
		width = 70
	}

	if width > windowWidth-4 {
		width = windowWidth - 4
	}
	height := len(lines)
	if height < 4 {
		height = 4
	}
	if height > 20 {
		height = 20
	}
	if height > windowHeight-4 {
		height = windowHeight - 4
	}
	x := (windowWidth - width) / 2
	y := (windowHeight - height) / 2

	return Rect{x, y, width, height}, Rect{x, y, width, height - 2}, Rect{x, y + height - 1, width, 1}
}

func (h *MessageBox) maxTextWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

func (h *MessageBox) onButtonsClick(x int, y int) {
	if x < 4 {
		h.Close()
	}
}
