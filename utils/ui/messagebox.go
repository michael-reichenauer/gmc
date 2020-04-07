package ui

import (
	"github.com/jroimartin/gocui"
	"strings"
)

type MessageBox struct {
	ui          *UI
	boxView     View
	textView    View
	buttonsView View
	text        string
	title       string
}

func NewMessageBox(ui *UI, text, title string) *MessageBox {
	return &MessageBox{ui: ui, text: text, title: title}
}

func (h *MessageBox) Show() {
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
	view.Properties().Name = "MessageBox"
	return view
}

func (h *MessageBox) newButtonsView() View {
	view := h.ui.NewView("[OK]")
	view.Properties().Name = "MessageBoxButtons"
	view.Properties().OnMouseLeft = h.onButtonsClick
	return view
}

func (h *MessageBox) newTextView() View {
	view := h.ui.NewView(h.text)
	view.Properties().Name = "MessageBoxText"
	view.Properties().HideCurrentLineMarker = true
	view.SetKey(gocui.KeyEsc, h.Close)
	view.SetKey(gocui.KeyEnter, h.Close)
	return view
}

func (h *MessageBox) Close() {
	h.textView.Close()
	h.buttonsView.Close()
	h.boxView.Close()
}

func (h *MessageBox) getBounds() (Rect, Rect, Rect) {
	lines := strings.Split(h.text, "\n")

	width := h.maxTextWidth(lines)
	if width < 30 {
		width = 30
	}
	if width > 70 {
		width = 70
	}

	height := len(lines)
	if height < 4 {
		height = 4
	}
	if height > 20 {
		height = 20
	}

	vb := h.ui.CenterBounds(width, height)
	return vb, Rect{vb.X, vb.Y, vb.W, vb.H - 2}, Rect{vb.X, vb.Y + vb.H - 1, vb.W, 1}
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
