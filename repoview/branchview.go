package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"strings"
)

type Brancher interface {
	CreateBranch(name string)
}

func NewBranchView(ui cui.UI, brancher Brancher) *BranchView {
	h := &BranchView{ui: ui, brancher: brancher}
	return h
}

type BranchView struct {
	ui          cui.UI
	brancher    Brancher
	boxView     cui.View
	textView    cui.View
	buttonsView cui.View
}

func (h *BranchView) Show() {
	h.boxView = h.newBranchView()
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

func (h *BranchView) newBranchView() cui.View {
	view := h.ui.NewView("\n\nName:")
	view.Properties().Title = "Create Branch"
	view.Properties().Name = "CreateBranchDlg"
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (h *BranchView) newButtonsView() cui.View {
	view := h.ui.NewView(" [OK] [Cancel]")
	view.Properties().OnMouseLeft = h.onButtonsClick
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (h *BranchView) newTextView() cui.View {
	view := h.ui.NewView("")
	view.Properties().HasFrame = true
	view.Properties().HideCurrentLineMarker = true
	view.Properties().IsEditable = true
	view.SetKey(gocui.KeyCtrlO, h.onOk)
	view.SetKey(gocui.KeyEnter, h.onOk)
	view.SetKey(gocui.KeyCtrlC, h.onCancel)
	view.SetKey(gocui.KeyEsc, h.onCancel)
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideHorizontalScrollbar = true
	return view
}

func (h *BranchView) Close() {
	h.textView.Close()
	h.buttonsView.Close()
	h.boxView.Close()
}

func (h *BranchView) getBounds() (cui.BoundFunc, cui.BoundFunc, cui.BoundFunc) {
	box := cui.CenterBounds(35, 5, 35, 5)
	text := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X + 7, Y: b.Y + 1, W: b.W - 9, H: 1}
	})
	buttons := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1}
	})
	return box, text, buttons
}

func (h *BranchView) onButtonsClick(x int, y int) {
	if x > 0 && x < 5 {
		h.onOk()
	}
	if x > 5 && x < 14 {
		h.onCancel()
	}
}

func (h *BranchView) onCancel() {
	log.Event("branch-cancel")
	h.Close()
}

func (h *BranchView) onOk() {
	name := h.textView.ReadLines()[0]
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	h.brancher.CreateBranch(name)
	h.Close()
}
