package repoview

import (
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"github.com/thoas/go-funk"
	"path/filepath"
	"strings"
)

const (
	RFC3339Small = "2006-01-02 15:04"
	markerWidth  = 5
)

type repoPage struct {
	lines              []string
	firstIndex         int
	currentIndex       int
	total              int
	repoPath           string
	currentBranchName  string
	uncommittedChanges int
}

type repoVM struct {
	notifier         notifier
	viewModelService *viewmodel.Service
	isDetails        bool
}

type trace struct {
	RepoPath    string
	ViewPage    ui.ViewPage
	BranchNames []string
}

type notifier interface {
	NotifyChanged()
}

func newRepoVM(model *viewmodel.Service, notifier notifier) *repoVM {
	return &repoVM{
		notifier:         notifier,
		viewModelService: model,
	}
}

func (h *repoVM) onLoad() {
	h.viewModelService.Start()
	go h.monitorModelRoutine()
}

func (h *repoVM) LoadWithBranches(branchNames []string) {
	h.viewModelService.LoadRepo(branchNames)
}

func (h *repoVM) monitorModelRoutine() {
	for range h.viewModelService.ChangedEvents {
		log.Infof("Detected model change")
		h.notifier.NotifyChanged()
	}
}

func (h *repoVM) GetRepoPage(viewPort ui.ViewPage) (repoPage, error) {
	rvp, err := h.viewModelService.GetRepoViewPort(viewPort.FirstLine, viewPort.Height)
	if err != nil {
		return repoPage{}, err
	}
	messageLength, _, authorLength, timeLength := columnWidths(rvp.GraphWidth+markerWidth, viewPort.Width)

	commits := rvp.Commits

	var currentLineCommit viewmodel.Commit
	if viewPort.CurrentLine-viewPort.FirstLine < len(commits) && viewPort.CurrentLine-viewPort.FirstLine >= 0 {
		currentLineCommit = commits[viewPort.CurrentLine-viewPort.FirstLine]
	}

	var lines []string
	for _, c := range commits {
		var sb strings.Builder
		h.writeGraph(&sb, c)
		sb.WriteString(" ")
		writeMoreMarker(&sb, c)
		writeCurrentMarker(&sb, c)
		//sb.WriteString(" ")
		writeAheadBehindMarker(&sb, c)
		h.writeSubject(&sb, c, currentLineCommit, messageLength)
		sb.WriteString(" ")
		//	writeSid(&sb, c, sidLength)
		//	sb.WriteString(" ")
		writeAuthor(&sb, c, authorLength)
		sb.WriteString(" ")
		writeAuthorTime(&sb, c, timeLength)
		lines = append(lines, sb.String())
	}

	return repoPage{
		repoPath:           rvp.RepoPath,
		lines:              lines,
		total:              rvp.TotalCommits,
		firstIndex:         rvp.FirstIndex,
		currentIndex:       viewPort.CurrentLine,
		uncommittedChanges: rvp.UncommittedChanges,
		currentBranchName:  rvp.CurrentBranchName,
	}, nil
}

func (h *repoVM) GetOpenBranchMenuItems(index int) []ui.MenuItem {
	branches := h.viewModelService.GetCommitOpenBranches(index)

	current, ok := h.viewModelService.CurrentBranch()
	if ok {
		if nil == funk.Find(branches, func(b viewmodel.Branch) bool {
			return current.DisplayName == b.DisplayName
		}) {
			branches = append(branches, current)
		}
	}

	var items []ui.MenuItem
	for _, b := range branches {
		items = append(items, h.toOpenBranchMenuItem(b))
	}

	if len(items) > 0 {
		items = append(items, ui.SeparatorMenuItem)
	}

	var activeSubItems []ui.MenuItem
	for _, b := range h.viewModelService.GetActiveBranches() {
		activeSubItems = append(activeSubItems, h.toOpenBranchMenuItem(b))
	}
	items = append(items, ui.MenuItem{Text: "Active Branches", SubItems: activeSubItems})

	var allGitSubItems []ui.MenuItem
	for _, b := range h.viewModelService.GetAllBranches() {
		if b.IsGitBranch {
			allGitSubItems = append(allGitSubItems, h.toOpenBranchMenuItem(b))
		}
	}
	items = append(items, ui.MenuItem{Text: "All Git Branches", SubItems: allGitSubItems})

	var allSubItems []ui.MenuItem
	for _, b := range h.viewModelService.GetAllBranches() {
		allSubItems = append(allSubItems, h.toOpenBranchMenuItem(b))
	}
	items = append(items, ui.MenuItem{Text: "All Branches", SubItems: allSubItems})

	return items
}

func (h *repoVM) GetCloseBranchMenuItems() []ui.MenuItem {
	var items []ui.MenuItem
	commitBranches := h.viewModelService.GetShownBranches()
	for _, b := range commitBranches {
		items = append(items, h.toCloseBranchMenuItem(b))
	}
	return items
}

func (h *repoVM) GetSwitchBranchMenuItems() []ui.MenuItem {
	var items []ui.MenuItem
	commitBranches := h.viewModelService.GetShownBranches()
	for _, b := range commitBranches {
		items = append(items, h.toCloseBranchMenuItem(b))
	}
	return items
}

func (h *repoVM) CloseBranch(index int) {
	h.viewModelService.CloseBranch(index)
}

func (h *repoVM) toOpenBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
	return ui.MenuItem{Text: h.branchItemText(branch), Action: func() {
		h.viewModelService.ShowBranch(branch.Name)
	}}
}

func (h *repoVM) toCloseBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
	return ui.MenuItem{Text: h.branchItemText(branch), Action: func() {
		h.viewModelService.HideBranch(branch.Name)
	}}
}

func (h *repoVM) toSwitchBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
	return ui.MenuItem{Text: h.branchItemText(branch), Action: func() {
		h.viewModelService.SwitchToBranch(branch.Name)
	}}
}

func (h *repoVM) branchItemText(branch viewmodel.Branch) string {
	if branch.IsCurrent {
		return "●" + branch.DisplayName
	} else {
		return " " + branch.DisplayName
	}
}

func (h *repoVM) Refresh() {
	h.viewModelService.TriggerRefreshModel()
}

func (h *repoVM) RefreshTrace(viewPage ui.ViewPage) {
	git.EnableTracing("")
	traceBytes := utils.MustJsonMarshal(trace{
		RepoPath:    h.viewModelService.RepoPath(),
		ViewPage:    viewPage,
		BranchNames: h.viewModelService.CurrentBranchNames(),
	})
	utils.MustFileWrite(filepath.Join(git.CurrentTracePath(), "repovm"), traceBytes)

	h.viewModelService.TriggerRefreshModel()
}

func writeMoreMarker(sb *strings.Builder, c viewmodel.Commit) {
	if c.IsMore {
		sb.WriteString(moreMarker)
	} else {
		sb.WriteString(" ")
	}
}

func (h *repoVM) writeGraph(sb *strings.Builder, c viewmodel.Commit) {
	for i := 0; i < len(c.Graph); i++ {
		bColor := h.viewModelService.BranchColor(c.Graph[i].BranchDisplayName)

		if i != 0 {
			cColor := bColor
			if c.Graph[i].Connect == viewmodel.BPass &&
				c.Graph[i].PassName != "" &&
				c.Graph[i].PassName != "-" {
				cColor = h.viewModelService.BranchColor(c.Graph[i].PassName)
			} else if c.Graph[i].Connect.Has(viewmodel.BPass) {
				cColor = ui.CWhite
			}
			sb.WriteString(ui.ColorRune(cColor, graphConnectRune(c.Graph[i].Connect)))
		}
		if c.Graph[i].Branch == viewmodel.BPass &&
			c.Graph[i].PassName != "" &&
			c.Graph[i].PassName != "-" {
			bColor = h.viewModelService.BranchColor(c.Graph[i].PassName)
		} else if c.Graph[i].Branch == viewmodel.BPass {
			bColor = ui.CWhite
		}
		sb.WriteString(ui.ColorRune(bColor, graphBranchRune(c.Graph[i].Branch)))
	}
}

func (h *repoVM) ChangeBranchColor(index int) {
	h.viewModelService.ChangeBranchColor(index)
}

func (h *repoVM) ToggleDetails() {
	h.isDetails = !h.isDetails
	if h.isDetails {
		log.Event("details-view-show")
	} else {
		log.Event("details-view-hide")
	}
}

func writeCurrentMarker(sb *strings.Builder, c viewmodel.Commit) {
	if c.IsCurrent {
		sb.WriteString(currentCommitMarker)
	} else {
		sb.WriteString(" ")
	}
}

func columnWidths(graphWidth, viewWidth int) (msgLength, sidLength, authorLength, timeLength int) {
	width := viewWidth - graphWidth
	sidLength = 0
	authorLength = 15
	timeLength = 12
	if width < 90 {
		authorLength = 10
		timeLength = 6
	}
	if width < 60 {
		sidLength = 0
		authorLength = 0
		timeLength = 0
	}
	msgLength = viewWidth - graphWidth - authorLength - timeLength - sidLength
	if msgLength < 0 {
		msgLength = 0
	}
	return
}

func writeSid(sb *strings.Builder, commit viewmodel.Commit, length int) {
	sid := commit.SID
	if commit.ID == viewmodel.StatusID {
		sid = " "
	}
	sb.WriteString(ui.Dark(utils.Text(sid, length)))
}

func writeAuthor(sb *strings.Builder, commit viewmodel.Commit, length int) {
	sb.WriteString(ui.Dark(utils.Text(commit.Author, length)))
}

func writeAuthorTime(sb *strings.Builder, c viewmodel.Commit, length int) {
	if c.ID == viewmodel.StatusID {
		sb.WriteString(ui.Dark(utils.Text("", length)))
		return
	}
	tt := c.AuthorTime.Format(RFC3339Small)
	tt = strings.Replace(tt, "-", "", -1)

	tt = tt[2:]
	sb.WriteString(ui.Dark(utils.Text(tt, length)))
}

func (h *repoVM) writeSubject(sb *strings.Builder, c viewmodel.Commit, selectedCommit viewmodel.Commit, length int) {
	subject := utils.Text(c.Subject, length)
	if c.ID == viewmodel.StatusID {
		sb.WriteString(ui.YellowDk(subject))
		return
	}
	color := ui.CWhite
	if c.IsLocalOnly {
		color = ui.CGreenDk
	} else if c.IsRemoteOnly {
		color = ui.CBlue
	}
	if h.isDetails &&
		c.Branch.DisplayName != selectedCommit.Branch.DisplayName {
		color = ui.CDark
	}
	sb.WriteString(ui.ColorText(color, subject))
}

func writeAheadBehindMarker(sb *strings.Builder, c viewmodel.Commit) {
	if c.IsLocalOnly {
		sb.WriteString(ui.GreenDk("▲"))
	} else if c.IsRemoteOnly {
		sb.WriteString(ui.Blue("▼"))
	} else {
		sb.WriteString(" ")
	}
}
