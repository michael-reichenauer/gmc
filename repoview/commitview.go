package repoview

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
)

type Committer interface {
	GetCommitDiff(id string) (git.CommitDiff, error)
	Commit(message string) error
}

func NewCommitView(ui *ui.UI, committer Committer, repoView *repoVM) *CommitView {
	h := &CommitView{ui: ui, vm: NewCommitVM(), committer: committer, repoView: repoView}
	return h
}

type CommitView struct {
	ui          *ui.UI
	vm          *commitVM
	repoView    *repoVM
	committer   Committer
	boxView     ui.View
	textView    ui.View
	buttonsView ui.View
}

func (h *CommitView) Show(message string) {
	log.Infof("Commit message %q", message)
	h.boxView = h.newCommitView()
	h.buttonsView = h.newButtonsView()
	h.textView = h.newTextView(message)

	bb, tb, bbb := h.getBounds()
	h.boxView.Show(bb)
	h.buttonsView.Show(bbb)
	h.textView.Show(tb)

	h.boxView.SetTop()
	h.buttonsView.SetTop()
	h.textView.SetTop()
	h.textView.SetCurrentView()
}

func (h *CommitView) newCommitView() ui.View {
	view := h.ui.NewView("")
	view.Properties().Title = "Commit on: " + "branch name"
	view.Properties().Name = "CommitView"
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	return view
}

func (h *CommitView) newButtonsView() ui.View {
	view := h.ui.NewView("[OK] [Cancel]")
	view.Properties().OnMouseLeft = h.onButtonsClick
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	return view
}

func (h *CommitView) newTextView(text string) ui.View {
	view := h.ui.NewView(text)
	view.Properties().HideCurrentLineMarker = true
	view.Properties().IsEditable = true
	view.SetKey(gocui.KeyEsc, h.onCancel)
	view.SetKey(gocui.KeyCtrlSpace, h.onOk)
	view.SetKey(gocui.KeyCtrlD, h.showDiff)
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideHorizontalScrollbar = true
	return view
}

func (h *CommitView) Close() {
	h.textView.Close()
	h.buttonsView.Close()
	h.boxView.Close()
}

func (h *CommitView) getBounds() (ui.BoundFunc, ui.BoundFunc, ui.BoundFunc) {
	box := ui.CenterBounds(10, 5, 70, 15)
	text := ui.Relative(box, func(b ui.Rect) ui.Rect {
		return ui.Rect{X: b.X, Y: b.Y, W: b.W, H: b.H - 2}
	})
	buttons := ui.Relative(box, func(b ui.Rect) ui.Rect {
		return ui.Rect{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1}
	})
	return box, text, buttons
}

func (h *CommitView) onButtonsClick(x int, y int) {
	if x > 0 && x < 5 {
		h.onOk()
	}
	if x > 5 && x < 14 {
		h.onCancel()
	}
}

func (h *CommitView) onCancel() {
	log.Event("commit-cancel")
	h.Close()
}

func (h *CommitView) onOk() {
	msg := strings.Join(h.textView.ReadLines(), "\n")
	err := h.committer.Commit(msg)
	if err != nil {
		log.Eventf("commit-error", "failed to commit, %v", err)
		h.ui.ShowErrorMessageBox(fmt.Sprintf("Failed to commit,\n%v", err))
		h.Close()
		return
	}

	log.Event("commit-ok")
	h.Close()
}

func (h *CommitView) showDiff() {
	log.Event("commit-show-diff")
	diffView := NewDiffView(h.ui, h.committer, git.UncommittedID)
	diffView.Show()
}