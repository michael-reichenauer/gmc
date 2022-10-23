package console

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/async"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type Committer interface {
	GetCommitDiff(info api.CommitDiffInfoReq, diff *api.CommitDiff) error
	GetFileDiff(info api.FileDiffInfoReq, diff *[]api.CommitDiff) error
	Commit(info api.CommitInfoReq, rsp api.NoRsp) error
}

func NewCommitView(ui cui.UI, committer Committer, repoID, branchName string, changes int) *CommitView {
	h := &CommitView{ui: ui, committer: committer, repoID: repoID, branchName: branchName, changes: changes}
	return h
}

type CommitView struct {
	ui          cui.UI
	committer   Committer
	commitView  cui.View
	messageView cui.View
	buttonsView cui.View
	repoID      string
	branchName  string
	changes     int
}

func (h *CommitView) Show(message string) {
	log.Infof("Commit message %q", message)
	h.commitView = h.newCommitView()
	h.buttonsView = h.newButtonsView()
	h.messageView = h.newMessageView(message)

	bb, tb, bbb := h.getBounds()
	h.commitView.Show(bb)
	h.buttonsView.Show(bbb)
	h.messageView.Show(tb)

	h.commitView.SetTop()
	h.messageView.SetTop()
	h.buttonsView.SetTop()
	h.commitView.SetCurrentView()
}

// The total dialog with title and frame
func (h *CommitView) newCommitView() cui.View {
	view := h.ui.NewView("")
	view.Properties().Title = fmt.Sprintf("Commit %d files on: %s", h.changes, h.branchName)
	view.Properties().Name = "CommitView"
	view.Properties().IsEditable = true
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	view.SetKey(gocui.KeyEnter, h.onOk)
	view.SetKey(gocui.KeyCtrlO, h.onOk)
	view.SetKey(gocui.KeyCtrlC, h.onCancel)
	view.SetKey(gocui.KeyEsc, h.onCancel)
	view.SetKey(gocui.KeyCtrlD, h.showDiff)
	view.SetKey(gocui.KeyTab, h.goToMessage)
	view.SetKey(gocui.KeyArrowDown, h.goToMessage)
	return view
}

func (h *CommitView) newMessageView(text string) cui.View {
	view := h.ui.NewView(text)
	view.Properties().Title = "..."
	view.Properties().IsEditable = true
	view.Properties().HasFrame = false
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	view.SetKey(gocui.KeyCtrlO, h.onOk)
	view.SetKey(gocui.KeyCtrlC, h.onCancel)
	view.SetKey(gocui.KeyEsc, h.onCancel)
	view.SetKey(gocui.KeyCtrlD, h.showDiff)
	view.SetKey(gocui.KeyTab, h.goToSubject)
	return view
}

// The OK/Cancel buttons
func (h *CommitView) newButtonsView() cui.View {
	view := h.ui.NewView(" [OK] [Cancel]")
	view.Properties().HasFrame = true
	view.Properties().OnMouseLeft = h.onButtonsClick
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	view.Properties().HideHorizontalScrollbar = true
	return view
}

func (h *CommitView) Close() {
	h.messageView.Close()
	h.buttonsView.Close()
	h.commitView.Close()
}

func (h *CommitView) goToMessage() {
	h.messageView.SetCurrentView()
}
func (h *CommitView) goToSubject() {
	h.commitView.SetCurrentView()
}

func (h *CommitView) getBounds() (cui.BoundFunc, cui.BoundFunc, cui.BoundFunc) {
	box := cui.CenterBounds(10, 5, 70, 15)
	msg := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y + 2, W: b.W, H: b.H - 4}
	})
	buttons := cui.Relative(box, func(b cui.Rect) cui.Rect {
		return cui.Rect{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1}
	})
	return box, msg, buttons
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
	subject := strings.Trim(h.commitView.ReadLines()[0], "\n")
	msg := strings.TrimRight(strings.Join(h.messageView.ReadLines(), "\n"), "\n")
	total := subject
	if len(msg) > 0 {
		total = total + "\n" + msg
	}

	progress := h.ui.ShowProgress("Committing ...")
	req := api.CommitInfoReq{RepoID: h.repoID, Message: total}
	async.RunE(func() error { return h.committer.Commit(req, api.NilRsp) }).
		Then(func(r any) {
			progress.Close()
			log.Event("commit-ok")
		}).
		Catch(func(e error) {
			progress.Close()
			log.Eventf("commit-error", "failed to commit, %v", e)
			h.ui.ShowErrorMessageBox("Failed to commit,\n%v", e)
		}).
		Finally(func() { h.Close() })
}

func (h *CommitView) showDiff() {
	log.Event("commit-show-diff")
	diffView := NewCommitDiffView(h.ui, h.committer, h.repoID, git.UncommittedID)
	diffView.Show()
}
