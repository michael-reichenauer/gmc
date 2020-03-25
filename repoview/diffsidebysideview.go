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

type DiffSideBySideView struct {
	vm          *diffSideVM
	mainService mainService
	leftSide    *DiffSideView
	rightSide   *DiffSideView
	commitID    string
	isUnified   bool
	lastBounds  ui.Rect
}

func (t *DiffSideBySideView) PostOnUIThread(f func()) {
	t.leftSide.PostOnUIThread(f)
}

func NewSideBySideView(
	uiHandler *ui.UI,
	mainService mainService,
	diffGetter DiffGetter,
	commitID string,
) DiffView {
	t := &DiffSideBySideView{
		mainService: mainService,
		commitID:    commitID,
	}
	t.vm = NewDiffSideVM(t, diffGetter, commitID)
	t.vm.setUnified(t.isUnified)
	t.leftSide = newDiffSideView(uiHandler, t.viewDataLeft, t.onLoadLeft, t.onMovedLeft)
	t.rightSide = newDiffSideView(uiHandler, t.viewDataRight, nil, t.onMovedRight)

	t.leftSide.Properties().HideScrollbar = true
	t.leftSide.Properties().Title = "Before " + commitID[:6]
	t.rightSide.Properties().Title = "After " + commitID[:6]
	return t
}

func (t *DiffSideBySideView) onLoadLeft() {
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

func (t *DiffSideBySideView) viewDataLeft(viewPort ui.ViewPage) ui.ViewPageData {
	diff, err := t.vm.getCommitDiffLeft(viewPort)
	if err != nil {
		return ui.ViewPageData{}
	}
	return ui.ViewPageData{Lines: diff.lines, FirstIndex: diff.firstIndex, Total: diff.total}
}

func (t *DiffSideBySideView) viewDataRight(viewPort ui.ViewPage) ui.ViewPageData {
	diff, err := t.vm.getCommitDiffRight(viewPort)
	if err != nil {
		return ui.ViewPageData{}
	}
	return ui.ViewPageData{Lines: diff.lines, FirstIndex: diff.firstIndex, Total: diff.total}
}

func (t *DiffSideBySideView) Show(bounds ui.Rect) {
	lb, rb := t.getSplitBounds(bounds)
	t.leftSide.Show(lb)
	t.rightSide.Show(rb)
}

func (t *DiffSideBySideView) SetTop() {
	t.rightSide.SetTop()
	t.leftSide.SetTop()
}

func (t *DiffSideBySideView) SetCurrentView() {
	t.leftSide.SetCurrentView()
}

func (t *DiffSideBySideView) Close() {
	t.leftSide.Close()
	t.rightSide.Close()
}

func (t *DiffSideBySideView) SetBounds(bounds ui.Rect) {
	lb, rb := t.getSplitBounds(bounds)
	t.leftSide.SetBounds(lb)
	t.rightSide.SetBounds(rb)
}

func (t *DiffSideBySideView) NotifyChanged() {
	t.leftSide.NotifyChanged()
	t.rightSide.NotifyChanged()
}

func (t *DiffSideBySideView) getSplitBounds(bounds ui.Rect) (ui.Rect, ui.Rect) {
	t.lastBounds = bounds
	if t.isUnified {
		return bounds, ui.Rect{W: 1, H: 1}
	}
	wl := bounds.W / 2
	wr := bounds.W - wl - 1
	lb := ui.Rect{X: bounds.X, Y: bounds.Y, W: wl, H: bounds.H}
	rb := ui.Rect{X: bounds.X + wl + 1, Y: bounds.Y, W: wr, H: bounds.H}
	return lb, rb
}

func (t *DiffSideBySideView) onMovedLeft() {
	p := t.leftSide.ViewPage()
	t.rightSide.SetPage(p.FirstLine, p.CurrentLine, p.FirstCharIndex)
}

func (t *DiffSideBySideView) onMovedRight() {
	p := t.rightSide.ViewPage()
	t.leftSide.SetPage(p.FirstLine, p.CurrentLine, p.FirstCharIndex)
}

func (t *DiffSideBySideView) ToUnified() {
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

func (t *DiffSideBySideView) ToSideBySide() {
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

func (t *DiffSideBySideView) scrollHorizontalLeft() {
	t.leftSide.ScrollHorizontal(-1)
}

func (t *DiffSideBySideView) scrollHorizontalRight() {
	t.leftSide.ScrollHorizontal(1)
}
