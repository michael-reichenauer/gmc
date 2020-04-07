package repoview

import (
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type DiffView interface {
	Show(rect ui.Rect)
	SetTop()
	SetCurrentView()
	Close()
	SetBounds(rect ui.Rect)
	NotifyChanged()
}

type diffView struct {
	ui          *ui.UI
	vm          *diffVM
	mainService mainService
	leftSide    *DiffSideView
	rightSide   *DiffSideView
	commitID    string
	isUnified   bool
	lastBounds  ui.Rect
}

func (t *diffView) PostOnUIThread(f func()) {
	t.leftSide.PostOnUIThread(f)
}

func NewDiffView(
	ui *ui.UI,
	mainService mainService,
	diffGetter DiffGetter,
	commitID string,
) DiffView {
	t := &diffView{
		ui:          ui,
		mainService: mainService,
		commitID:    commitID,
	}
	t.vm = newDiffVM(t, diffGetter, commitID)
	t.vm.setUnified(t.isUnified)
	t.leftSide = t.newLeftSide()
	t.rightSide = t.newRightSide()
	return t
}

func (t *diffView) newLeftSide() *DiffSideView {
	view := newDiffSideView(t.ui, t.viewDataLeft, t.vm.load, t.onMovedLeft)

	view.Properties().HideVerticalScrollbar = true
	view.Properties().OnMouseRight = t.showContextMenu
	view.Properties().Title = "Before " + t.commitID[:6]

	view.SetKey(gocui.KeyEsc, t.mainService.HideDiff)
	view.SetKey(gocui.KeyCtrlC, t.mainService.HideDiff)
	view.SetKey(gocui.KeyCtrlC, t.mainService.HideDiff)
	view.SetKey(gocui.KeyCtrlQ, t.mainService.HideDiff)
	view.SetKey('q', t.mainService.HideDiff)
	view.SetKey('1', t.ToUnified)
	view.SetKey('2', t.ToSideBySide)
	view.SetKey(gocui.KeyArrowLeft, t.scrollHorizontalLeft)
	view.SetKey(gocui.KeyArrowRight, t.scrollHorizontalRight)

	return view
}

func (t *diffView) newRightSide() *DiffSideView {
	view := newDiffSideView(t.ui, t.viewDataRight, nil, t.onMovedRight)
	view.Properties().Title = "After " + t.commitID[:6]
	return view
}

func (t *diffView) viewDataLeft(viewPort ui.ViewPage) ui.ViewText {
	diff, err := t.vm.getCommitDiffLeft(viewPort)
	if err != nil {
		return ui.ViewText{}
	}
	return diff
}

func (t *diffView) viewDataRight(viewPort ui.ViewPage) ui.ViewText {
	diff, err := t.vm.getCommitDiffRight(viewPort)
	if err != nil {
		return ui.ViewText{}
	}
	return diff
}

func (t *diffView) Show(bounds ui.Rect) {
	lb, rb := t.getSplitBounds(bounds)
	t.leftSide.Show(lb)
	t.rightSide.Show(rb)
}

func (t *diffView) SetTop() {
	t.rightSide.SetTop()
	t.leftSide.SetTop()
}

func (t *diffView) SetCurrentView() {
	t.leftSide.SetCurrentView()
}

func (t *diffView) Close() {
	t.leftSide.Close()
	t.rightSide.Close()
}

func (t *diffView) SetBounds(bounds ui.Rect) {
	lb, rb := t.getSplitBounds(bounds)
	t.leftSide.SetBounds(lb)
	t.rightSide.SetBounds(rb)
}

func (t *diffView) NotifyChanged() {
	t.leftSide.NotifyChanged()
	t.rightSide.NotifyChanged()
}

func (t *diffView) getSplitBounds(bounds ui.Rect) (ui.Rect, ui.Rect) {
	t.lastBounds = bounds
	if t.isUnified {
		return ui.Rect{X: 0, Y: 1, W: bounds.W + 2, H: bounds.H + 1}, ui.Rect{W: 1, H: 1}
	}
	wl := bounds.W/2 - 3
	wr := bounds.W - wl
	lb := ui.Rect{X: bounds.X - 1, Y: bounds.Y, W: wl, H: bounds.H + 1}
	rb := ui.Rect{X: bounds.X + wl + 1, Y: bounds.Y, W: wr, H: bounds.H + 1}
	return lb, rb
}

func (t *diffView) onMovedLeft() {
	t.rightSide.SyncWithView(t.leftSide.View)
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
	t.leftSide.Properties().Title = "Unified diff " + t.commitID[:6]
	t.vm.setUnified(t.isUnified)
	t.SetBounds(t.lastBounds)
	t.NotifyChanged()
}

func (t *diffView) ToSideBySide() {
	if !t.isUnified {
		return
	}
	t.isUnified = false
	t.leftSide.Properties().HideVerticalScrollbar = true
	t.leftSide.Properties().Title = "Before " + t.commitID[:6]
	t.rightSide.Properties().Title = "After " + t.commitID[:6]
	t.vm.setUnified(t.isUnified)
	t.SetBounds(t.lastBounds)
	t.NotifyChanged()
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
		cm.Add(ui.MenuItem{Text: "Show Split Diff", Key: "2", Action: func() { t.ToSideBySide() }})
	} else {
		cm.Add(ui.MenuItem{Text: "Show Unified Diff", Key: "1", Action: func() { t.ToUnified() }})
	}

	cm.Add(ui.MenuItem{Text: "Close", Key: "Esc", Action: func() { t.mainService.HideDiff() }})
	cm.Show(x+3, y+2)
}
