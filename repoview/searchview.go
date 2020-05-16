package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

type Searcher interface {
	Search(text string)
	CloseSearch()
	ScrollVertical(scroll int)
	SetCurrentView()
}

func NewSearchView(ui ui.UI, searcher Searcher) *SearchView {
	h := &SearchView{ui: ui, searcher: searcher}
	return h
}

type SearchView struct {
	ui         ui.UI
	boxView    ui.View
	textView   ui.View
	searcher   Searcher
	lastSearch string
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
	view.SetKey(gocui.KeyArrowUp, t.scrollUpp)
	view.SetKey(gocui.KeyArrowDown, t.scrollDown)
	view.SetKey(gocui.KeyTab, t.searcher.SetCurrentView)
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().OnEdit = t.onEdit
	return view
}

func (t *SearchView) onEdit() {
	text := strings.TrimSpace(t.textView.ReadLines()[0])
	if text == t.lastSearch {
		return
	}
	t.lastSearch = text
	t.searcher.Search(text)
}

func (t *SearchView) Close() {
	t.textView.Close()
	t.boxView.Close()
	t.searcher.CloseSearch()
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

func (t *SearchView) onCancel() {
	log.Event("search-cancel")
	t.Close()
}

func (t *SearchView) onOk() {
}

func (t *SearchView) scrollUpp() {
	t.searcher.ScrollVertical(-1)
}

func (t *SearchView) scrollDown() {
	t.searcher.ScrollVertical(1)
}

func (t *SearchView) SetCurrentView() {
	t.textView.SetCurrentView()
	t.textView.NotifyChanged()
}
