package console

import (
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

type BranchDlg interface {
	Show()
}

func newBranchDlg(ui cui.UI, createBranch func(name string)) BranchDlg {
	h := &branchDlg{ui: ui, createBranch: createBranch}
	return h
}

type branchDlg struct {
	ui           cui.UI
	createBranch func(name string)
	boxView      cui.View
	textView     cui.View
	buttonsView  cui.View
}

func (t *branchDlg) Show() {
	t.boxView = t.newBranchView()
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

func (t *branchDlg) newBranchView() cui.View {
	view := t.ui.NewView("\n\nName:")
	view.Properties().Title = "Create Branch"
	view.Properties().Name = "CreateBranchDlg"
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *branchDlg) newButtonsView() cui.View {
	view := t.ui.NewView(" [OK] [Cancel]")
	view.Properties().OnMouseLeft = t.onButtonsClick
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *branchDlg) newTextView() cui.View {
	view := t.ui.NewView("")
	view.Properties().HasFrame = true
	view.Properties().HideCurrentLineMarker = true
	view.Properties().IsEditable = true
	view.SetKey(gocui.KeyCtrlO, t.onOk)
	view.SetKey(gocui.KeyEnter, t.onOk)
	view.SetKey(gocui.KeyCtrlC, t.onCancel)
	view.SetKey(gocui.KeyEsc, t.onCancel)
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideHorizontalScrollbar = true
	return view
}

func (t *branchDlg) Close() {
	t.textView.Close()
	t.buttonsView.Close()
	t.boxView.Close()
}

func (t *branchDlg) getBounds() (cui.BoundFunc, cui.BoundFunc, cui.BoundFunc) {
	box := cui.CenterBounds(35, 5, 35, 5)
	text := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X + 7, Y: b.Y + 1, W: b.W - 9, H: 1}
	})
	buttons := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1}
	})
	return box, text, buttons
}

func (t *branchDlg) onButtonsClick(x int, y int) {
	if x > 0 && x < 5 {
		t.onOk()
	}
	if x > 5 && x < 14 {
		t.onCancel()
	}
}

func (t *branchDlg) onCancel() {
	t.Close()
}

func (t *branchDlg) onOk() {
	name := t.textView.ReadLines()[0]
	name = strings.TrimSpace(name)
	if name == "" {
		t.ui.ShowErrorMessageBox("Error", "Empty branch name not allowed.")
		return
	}

	t.createBranch(name)
	t.Close()
}
