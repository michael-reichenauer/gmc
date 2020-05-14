package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

func NewSearchView(ui ui.UI) *SearchView {
	h := &SearchView{ui: ui}
	return h
}

type SearchView struct {
	ui       ui.UI
	boxView  ui.View
	textView ui.View
}

func (t *SearchView) Show() {
	t.boxView = t.newBoxView()
	t.textView = t.newTextView()

	bb, tb := t.getBounds()
	t.boxView.Show(bb)
	t.textView.Show(tb)

	t.boxView.SetTop()
	t.textView.SetTop()
	t.textView.SetCurrentView()
}

func (t *SearchView) newBoxView() ui.View {
	view := t.ui.NewView(" \n Search:")
	view.Properties().Name = "SearchView"
	view.Properties().HasFrame = true
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *SearchView) newTextView() ui.View {
	view := t.ui.NewView("")
	view.Properties().HideCurrentLineMarker = true
	view.Properties().IsEditable = true
	view.SetKey(gocui.KeyCtrlO, t.onOk)
	view.SetKey(gocui.KeyEnter, t.onOk)
	view.SetKey(gocui.KeyCtrlC, t.onCancel)
	view.SetKey(gocui.KeyEsc, t.onCancel)
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().OnEdit = t.onEdit
	return view
}

func (t *SearchView) onEdit() {
	text := strings.TrimSpace(t.textView.ReadLines()[0])
	log.Infof("Search in search %q", text)
}

func (t *SearchView) Close() {
	t.textView.Close()
	t.boxView.Close()
}

func (t *SearchView) getBounds() (ui.BoundFunc, ui.BoundFunc) {
	box := ui.Relative(ui.FullScreen(), func(b ui.Rect) ui.Rect {
		return ui.Rect{X: b.X, Y: b.Y - 1, W: b.W, H: 2}
	})
	text := ui.Relative(box, func(b ui.Rect) ui.Rect {
		return ui.Rect{X: b.X + 9, Y: b.Y + 1, W: b.W - 10, H: 1}
	})

	return box, text
}

func (t *SearchView) onButtonsClick(x int, y int) {
	if x > 0 && x < 5 {
		t.onOk()
	}
	if x > 5 && x < 14 {
		t.onCancel()
	}
}

func (t *SearchView) onCancel() {
	log.Event("search-cancel")
	t.Close()
}

func (t *SearchView) onOk() {
	log.Event("search-ok")
	name := t.textView.ReadLines()[0]
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	t.Close()
}
