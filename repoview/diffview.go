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
	t.leftSide = newDiffSideView(ui, t.viewDataLeft, t.onLoadLeft, t.onMovedLeft)
	t.rightSide = newDiffSideView(ui, t.viewDataRight, nil, t.onMovedRight)

	t.leftSide.Properties().HideScrollbar = true
	t.leftSide.Properties().OnMouseRight = t.showContextMenu
	t.leftSide.Properties().Title = "Before " + commitID[:6]
	t.rightSide.Properties().Title = "After " + commitID[:6]
	return t
}

func (t *diffView) onLoadLeft() {
	t.leftSide.SetKey(gocui.KeyEsc, gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey(gocui.KeyCtrlC, gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey(gocui.KeyCtrlC, gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey(gocui.KeyCtrlQ, gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey('q', gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey('1', gocui.ModNone, t.ToUnified)
	t.leftSide.SetKey('2', gocui.ModNone, t.ToSideBySide)
	t.leftSide.SetKey(gocui.KeyArrowLeft, gocui.ModNone, t.scrollHorizontalLeft)
	t.leftSide.SetKey(gocui.KeyArrowRight, gocui.ModNone, t.scrollHorizontalRight)

	t.vm.load()
}

func (t *diffView) viewDataLeft(viewPort ui.ViewPage) ui.ViewPageData {
	diff, err := t.vm.getCommitDiffLeft(viewPort)
	if err != nil {
		return ui.ViewPageData{}
	}
	return diff
}

func (t *diffView) viewDataRight(viewPort ui.ViewPage) ui.ViewPageData {
	diff, err := t.vm.getCommitDiffRight(viewPort)
	if err != nil {
		return ui.ViewPageData{}
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
		return bounds, ui.Rect{W: 1, H: 1}
	}
	wl := bounds.W/2 - 3
	wr := bounds.W - wl
	lb := ui.Rect{X: bounds.X - 1, Y: bounds.Y, W: wl, H: bounds.H + 1}
	rb := ui.Rect{X: bounds.X + wl + 1, Y: bounds.Y, W: wr, H: bounds.H + 1}
	return lb, rb
}

func (t *diffView) onMovedLeft() {
	t.rightSide.SetPage(t.leftSide.ViewPage())
}

func (t *diffView) onMovedRight() {
	t.leftSide.SetPage(t.rightSide.ViewPage())
}

func (t *diffView) ToUnified() {
	if t.isUnified {
		return
	}
	t.isUnified = true
	t.leftSide.Properties().HideScrollbar = false
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
	t.leftSide.Properties().HideScrollbar = true
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
