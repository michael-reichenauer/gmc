package console

import (
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type CloneDlg interface {
	Show()
}

func newCloneDlg(ui cui.UI, clone func(uri, path string)) CloneDlg {
	h := &cloneDlg{ui: ui, clone: clone}
	return h
}

type cloneDlg struct {
	ui          cui.UI
	clone       func(uri, path string)
	boxView     cui.View
	uriView     cui.View
	pathView    cui.View
	buttonsView cui.View
}

func (t *cloneDlg) Show() {
	t.boxView = t.newCloneView()
	t.buttonsView = t.newButtonsView()
	t.uriView = t.newUriView()
	t.pathView = t.newPathView()

	bb, ub, pb, bbb := t.getBounds()
	t.boxView.Show(bb)
	t.buttonsView.Show(bbb)
	t.uriView.Show(ub)
	t.pathView.Show(pb)

	t.boxView.SetTop()
	t.buttonsView.SetTop()
	t.uriView.SetTop()
	t.pathView.SetTop()
	t.uriView.SetCurrentView()
}

func (t *cloneDlg) newCloneView() cui.View {
	view := t.ui.NewView("\n\nUri:\n\n\nPath:")
	view.Properties().Title = "Clone Repo"
	view.Properties().Name = "CloneDlg"
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *cloneDlg) newButtonsView() cui.View {
	view := t.ui.NewView(" [OK] [Cancel]")
	view.Properties().OnMouseLeft = t.onButtonsClick
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *cloneDlg) newUriView() cui.View {
	view := t.ui.NewView("")
	view.Properties().HasFrame = true
	view.Properties().HideCurrentLineMarker = true
	view.Properties().IsEditable = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().OnMouseLeft = func(_, _ int) { t.uriView.SetCurrentView() }
	view.SetKey(gocui.KeyCtrlO, t.onOk)
	view.SetKey(gocui.KeyEnter, t.onOk)
	view.SetKey(gocui.KeyCtrlC, t.onCancel)
	view.SetKey(gocui.KeyEsc, t.onCancel)
	view.SetKey(gocui.KeyTab, func() { t.pathView.SetCurrentView() })
	view.SetKey(gocui.KeyCtrlV, func() { log.Infof("ctrl v") })
	return view
}

func (t *cloneDlg) newPathView() cui.View {
	view := t.ui.NewView("")
	view.Properties().HasFrame = true
	view.Properties().HideCurrentLineMarker = true
	view.Properties().IsEditable = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().OnMouseLeft = func(_, _ int) { t.pathView.SetCurrentView() }
	view.SetKey(gocui.KeyCtrlO, t.onOk)
	view.SetKey(gocui.KeyEnter, t.onOk)
	view.SetKey(gocui.KeyCtrlC, t.onCancel)
	view.SetKey(gocui.KeyEsc, t.onCancel)
	view.SetKey(gocui.KeyTab, func() { t.uriView.SetCurrentView() })
	return view
}

func (t *cloneDlg) Close() {
	t.uriView.Close()
	t.pathView.Close()
	t.buttonsView.Close()
	t.boxView.Close()
}

func (t *cloneDlg) getBounds() (cui.BoundFunc, cui.BoundFunc, cui.BoundFunc, cui.BoundFunc) {
	box := cui.CenterBounds(45, 8, 70, 8)
	uri := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X + 7, Y: b.Y + 1, W: b.W - 9, H: 1}
	})
	path := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X + 7, Y: b.Y + 4, W: b.W - 9, H: 1}
	})
	buttons := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1}
	})
	return box, uri, path, buttons
}

func (t *cloneDlg) onButtonsClick(x int, y int) {
	if x > 0 && x < 5 {
		t.onOk()
	}
	if x > 5 && x < 14 {
		t.onCancel()
	}
}

func (t *cloneDlg) onCancel() {
	t.Close()
}

func (t *cloneDlg) onOk() {
	uri := strings.TrimSpace(t.uriView.ReadLines()[0])
	path := strings.TrimSpace(t.pathView.ReadLines()[0])
	if uri == "" || path == "" {
		t.ui.ShowErrorMessageBox("Error", "Empty uri or path is not allowed.")
		return
	}

	t.clone(uri, path)
	t.Close()
}
