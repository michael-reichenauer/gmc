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
}

func (t *DiffSideBySideView) PostOnUIThread(f func()) {
	t.leftSide.PostOnUIThread(f)
}

func NewSideBySideView(uiHandler *ui.UI,
	mainService mainService,
	diffGetter DiffGetter,
	commitID string) DiffView {

	t := &DiffSideBySideView{
		mainService: mainService,
	}
	t.vm = NewDiffSideVM(t, diffGetter, commitID)
	t.leftSide = newDiffSideView(uiHandler, t.viewDataLeft, t.onLoadLeft, t.onMovedLeft)
	t.leftSide.Properties().HideScrollbar = true
	t.leftSide.Properties().Title = "Before " + commitID[:6]
	t.rightSide = newDiffSideView(uiHandler, t.viewDataRight, nil, t.onMovedRight)
	t.rightSide.Properties().Title = "After " + commitID[:6]
	return t
}

func (t *DiffSideBySideView) onLoadLeft() {
	t.leftSide.SetKey(gocui.KeyEsc, gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey(gocui.KeyCtrlC, gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey(gocui.KeyCtrlC, gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey(gocui.KeyCtrlQ, gocui.ModNone, t.mainService.HideDiff)
	t.leftSide.SetKey('q', gocui.ModNone, t.mainService.HideDiff)

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
	wl := bounds.W / 2
	wr := bounds.W - wl - 1
	lb := ui.Rect{X: bounds.X, Y: bounds.Y, W: wl, H: bounds.H}
	rb := ui.Rect{X: bounds.X + wl + 1, Y: bounds.Y, W: wr, H: bounds.H}
	return lb, rb
}

func (t *DiffSideBySideView) onMovedLeft() {
	p := t.leftSide.ViewPage()
	t.rightSide.SetPage(p.FirstLine, p.CurrentLine)
}

func (t *DiffSideBySideView) onMovedRight() {
	p := t.rightSide.ViewPage()
	t.leftSide.SetPage(p.FirstLine, p.CurrentLine)
}
