package console

import (
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type Committer interface {
	GetCommitDiff(info api.CommitDiffInfoReq, diff *api.CommitDiff) error
	Commit(info api.CommitInfoReq, rsp api.NoRsp) error
}

func NewCommitView(ui cui.UI, committer Committer, repoID, branchName string) *CommitView {
	h := &CommitView{ui: ui, committer: committer, repoID: repoID, branchName: branchName}
	return h
}

type CommitView struct {
	ui          cui.UI
	committer   Committer
	boxView     cui.View
	textView    cui.View
	buttonsView cui.View
	repoID      string
	branchName  string
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

func (h *CommitView) newCommitView() cui.View {
	view := h.ui.NewView("")
	view.Properties().Title = "Commit on: " + h.branchName
	view.Properties().Name = "CommitView"
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (h *CommitView) newButtonsView() cui.View {
	view := h.ui.NewView(" [OK] [Cancel]")
	view.Properties().OnMouseLeft = h.onButtonsClick
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (h *CommitView) newTextView(text string) cui.View {
	view := h.ui.NewView(text)
	view.Properties().HideCurrentLineMarker = true
	view.Properties().IsEditable = true
	view.SetKey(gocui.KeyCtrlO, h.onOk)
	view.SetKey(gocui.KeyCtrlC, h.onCancel)
	view.SetKey(gocui.KeyEsc, h.onCancel)
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

func (h *CommitView) getBounds() (cui.BoundFunc, cui.BoundFunc, cui.BoundFunc) {
	box := cui.CenterBounds(10, 5, 70, 15)
	text := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y, W: b.W, H: b.H - 2}
	})
	buttons := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1}
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
	progress := h.ui.ShowProgress("Committing ...")
	go func() {
		err := h.committer.Commit(api.CommitInfoReq{RepoID: h.repoID, Message: msg}, api.NilRsp)
		h.ui.Post(func() {
			progress.Close()
			if err != nil {
				log.Eventf("commit-error", "failed to commit, %v", err)
				h.ui.ShowErrorMessageBox("Failed to commit,\n%v", err)
				h.Close()
				return
			}

			log.Event("commit-ok")
			h.Close()
		})
	}()
}

func (h *CommitView) showDiff() {
	log.Event("commit-show-diff")
	diffView := NewDiffView(h.ui, h.committer, h.repoID, git.UncommittedID)
	diffView.Show()
}
