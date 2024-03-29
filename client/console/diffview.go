package console

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/cui"
)

type DiffView interface {
	Show()
	SetTop()
	SetCurrentView()
	Close()
	NotifyChanged()
}

type diffView struct {
	ui        cui.UI
	vm        *diffVM
	leftSide  cui.View
	rightSide cui.View
	path      string
	isUnified bool
}

func (t *diffView) PostOnUIThread(f func()) {
	t.leftSide.PostOnUIThread(f)
}

func NewCommitDiffView(ui cui.UI, diffGetter DiffGetter, repoID string, commitID string) DiffView {
	t := &diffView{
		ui: ui,
	}
	t.vm = newCommitDiffVM(t.ui, t, diffGetter, repoID, commitID)
	t.vm.setUnified(t.isUnified)
	t.leftSide = t.newLeftSide()
	t.rightSide = t.newRightSide()
	return t
}

func NewFileDiffView(ui cui.UI, diffGetter DiffGetter, repoID string, path string) DiffView {
	t := &diffView{
		ui:   ui,
		path: path,
	}
	t.vm = newFileDiffVM(t.ui, t, diffGetter, repoID, path)
	t.vm.setUnified(t.isUnified)
	t.leftSide = t.newLeftSide()
	t.rightSide = t.newRightSide()
	return t
}

func (t *diffView) newLeftSide() cui.View {
	view := t.ui.NewViewFromPageFunc(t.viewDataLeft)
	view.Properties().OnLoad = t.vm.load
	view.Properties().OnMoved = t.onMovedLeft
	view.Properties().Name = "DiffViewLeft"
	view.Properties().HasFrame = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	view.Properties().OnMouseRight = t.showContextMenu
	view.Properties().Title = "Before"

	// Only need to set key on left side since left side is always current
	view.SetKey(gocui.KeyEsc, t.Close)
	view.SetKey(gocui.KeyCtrlC, t.Close)
	view.SetKey(gocui.KeyCtrlQ, t.Close)
	view.SetKey('q', t.Close)
	view.SetKey('1', t.ToUnified)
	view.SetKey('2', t.ToSideBySide)
	view.SetKey(gocui.KeyArrowLeft, t.scrollHorizontalLeft)
	view.SetKey(gocui.KeyArrowRight, t.scrollHorizontalRight)

	return view
}

func (t *diffView) newRightSide() cui.View {
	view := t.ui.NewViewFromPageFunc(t.viewDataRight)
	view.Properties().OnMoved = t.onMovedRight
	view.Properties().Name = "DiffViewRight"
	view.Properties().HasFrame = true
	view.Properties().Title = "After"
	view.Properties().OnMouseRight = t.showContextMenu

	// No need to set key on right side since left side is always current
	return view
}

func (t *diffView) viewDataLeft(viewPort cui.ViewPage) cui.ViewText {
	diff, err := t.vm.getCommitDiffLeft(viewPort)
	if err != nil {
		return cui.ViewText{}
	}
	return diff
}

func (t *diffView) viewDataRight(viewPort cui.ViewPage) cui.ViewText {
	diff, err := t.vm.getCommitDiffRight(viewPort)
	if err != nil {
		return cui.ViewText{}
	}
	return diff
}

func (t *diffView) Show() {
	lbf, rbf := t.getBounds()
	t.leftSide.Show(lbf)
	t.rightSide.Show(rbf)
	t.SetTop()
	t.SetCurrentView()
}

func (t *diffView) SetTop() {
	if t.isUnified {
		t.rightSide.SetTop()
		t.leftSide.SetTop()
		return
	}

	t.leftSide.SetTop()
	t.rightSide.SetTop()
}

func (t *diffView) SetCurrentView() {
	t.leftSide.SetCurrentView()
}

func (t *diffView) Close() {
	t.leftSide.Close()
	t.rightSide.Close()
}

func (t *diffView) NotifyChanged() {
	t.leftSide.NotifyChanged()
	t.rightSide.NotifyChanged()
}

func (t *diffView) getBounds() (cui.BoundFunc, cui.BoundFunc) {
	left := func(w, h int) cui.Rect {
		if t.isUnified {
			return cui.Rect{X: 0, Y: 1, W: w - 1, H: h - 1}
		}
		wl := w/2 - 2
		return cui.Rect{X: 0, Y: 1, W: wl, H: h - 1}
	}
	right := func(w, h int) cui.Rect {
		if t.isUnified {
			return cui.Rect{W: 1, H: 1}
		}
		wl := w/2 - 2
		wr := w - wl - 3
		return cui.Rect{X: wl + 2, Y: 1, W: wr, H: h - 1}
	}
	return left, right
}

func (t *diffView) onMovedLeft() {
	t.rightSide.SyncWithView(t.leftSide)
}

func (t *diffView) onMovedRight() {
	t.leftSide.SyncWithView(t.rightSide)
}

func (t *diffView) ToUnified() {
	if t.isUnified {
		return
	}
	t.isUnified = true
	t.leftSide.Properties().HideVerticalScrollbar = false
	t.leftSide.Properties().Title = "Unified diff"
	t.vm.setUnified(t.isUnified)
	t.SetTop()
	t.ui.ResizeAllViews()
}

func (t *diffView) ToSideBySide() {
	if !t.isUnified {
		return
	}
	t.isUnified = false
	t.leftSide.Properties().HideVerticalScrollbar = true
	t.leftSide.Properties().Title = "Before"
	t.rightSide.Properties().Title = "After"
	t.vm.setUnified(t.isUnified)
	t.SetTop()
	t.ui.ResizeAllViews()
}

func (t *diffView) scrollHorizontalLeft() {
	t.leftSide.ScrollHorizontal(-1)
}

func (t *diffView) scrollHorizontalRight() {
	t.leftSide.ScrollHorizontal(1)
}

func (t *diffView) showContextMenu(x int, y int) {
	cm := t.ui.NewMenu("")
	if t.isUnified {
		cm.Add(cui.MenuItem{Text: "Show Split Diff", Key: "2", Action: func() { t.ToSideBySide() }})
	} else {
		cm.Add(cui.MenuItem{Text: "Show Unified Diff", Key: "1", Action: func() { t.ToUnified() }})
	}

	cm.Add(cui.MenuItem{Text: "Close", Key: "Esc", Action: t.Close})
	cm.Show(x+3, y+2)
}
