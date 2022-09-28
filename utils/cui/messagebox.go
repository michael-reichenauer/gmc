package cui

import (
	"strings"

	"github.com/jroimartin/gocui"
)

type MessageBox struct {
	OnOK        func()
	OnClose     func()
	ui          *ui
	boxView     View
	textView    View
	buttonsView View
	text        string
	title       string
}

func NewMessageBox(ui *ui, text, title string) *MessageBox {
	return &MessageBox{ui: ui, text: text, title: title}
}

func (t *MessageBox) Show() {
	t.boxView = t.newBoxView()
	t.buttonsView = t.newButtonsView()
	t.textView = t.newTextView()

	bb, tb, bbb := t.getBounds()
	t.boxView.Show(bb)
	t.buttonsView.Show(bbb)
	t.textView.Show(tb)

	t.boxView.SetTop()
	t.buttonsView.SetTop()
	t.textView.SetTop()
	t.textView.SetCurrentView()
}

func (t *MessageBox) newBoxView() View {
	view := t.ui.NewView("")
	view.Properties().Title = t.title
	view.Properties().Name = "MessageBox"
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *MessageBox) newButtonsView() View {
	view := t.ui.NewView(" [OK]")
	view.Properties().Name = "MessageBoxButtons"
	view.Properties().OnMouseLeft = t.onButtonsClick
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *MessageBox) newTextView() View {
	view := t.ui.NewView(t.text)
	view.Properties().Name = "MessageBoxText"
	view.Properties().HideCurrentLineMarker = true
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().IsWrap = true
	view.SetKey(gocui.KeyEnter, t.handleOk)
	view.SetKey(gocui.KeyCtrlO, t.handleOk)
	view.SetKey(gocui.KeyEsc, t.Close)
	view.SetKey(gocui.KeyCtrlC, t.Close)
	return view
}

func (t *MessageBox) Close() {
	t.textView.Close()
	t.buttonsView.Close()
	t.boxView.Close()
	if t.OnClose != nil {
		t.OnClose()
	}
}

func (t *MessageBox) handleOk() {
	t.Close()
	if t.OnOK != nil {
		t.OnOK()
	}
}

func (t *MessageBox) getBounds() (BoundFunc, BoundFunc, BoundFunc) {
	lines := strings.Split(t.text, "\n")

	width := t.maxTextWidth(lines)
	if width < 30 {
		width = 30
	}
	if width > 100 {
		width = 100
	}

	height := len(lines) + 3
	if height < 4 {
		height = 4
	}
	if height > 40 {
		height = 40
	}

	box := CenterBounds(width, height, width, height)
	text := Relative(box, func(b Rect) Rect {
		return Rect{b.X, b.Y, b.W, b.H - 2}
	})
	buttons := Relative(box, func(b Rect) Rect {
		return Rect{b.X, b.Y + b.H - 1, b.W, 1}
	})

	return box, text, buttons
}

func (t *MessageBox) maxTextWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

func (t *MessageBox) onButtonsClick(x int, y int) {
	if x < 4 {
		t.handleOk()
	}
}
