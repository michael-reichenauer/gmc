package repoview

import (
	"context"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"github.com/thoas/go-funk"
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
	repoViewer       ui.Viewer
	mainService      mainService
	viewModelService *viewmodel.Service
	isDetails        bool
	workingFolder    string
	cancel           context.CancelFunc
	repo             viewmodel.ViewRepo
	firstIndex       int
	currentIndex     int
}

type trace struct {
	RepoPath    string
	ViewPage    ui.ViewPage
	BranchNames []string
}

func newRepoVM(
	repoViewer ui.Viewer,
	mainService mainService,
	configService *config.Service,
	workingFolder string) *repoVM {
	return &repoVM{
		repoViewer:       repoViewer,
		mainService:      mainService,
		viewModelService: viewmodel.NewService(configService, workingFolder),
		workingFolder:    workingFolder,
	}
}

func (h *repoVM) load() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	go h.monitorModelRoutine(ctx)
}

func (h *repoVM) close() {
	h.cancel()
}

func (h *repoVM) monitorModelRoutine(ctx context.Context) {
	h.viewModelService.StartMonitor(ctx)

	for vr := range h.viewModelService.ViewRepos {
		log.Infof("Detected model change")
		h.repoViewer.PostOnUIThread(func() {
			h.repo = vr
			h.repoViewer.NotifyChanged()
		})
	}
}

func (h *repoVM) GetRepoPage(viewPage ui.ViewPage) (repoPage, error) {
	firstIndex, lines := h.getLines(viewPage)
	h.firstIndex = firstIndex
	h.currentIndex = viewPage.CurrentLine
	return repoPage{
		repoPath:           h.repo.RepoPath,
		lines:              lines,
		total:              len(h.repo.Commits),
		firstIndex:         firstIndex,
		currentIndex:       viewPage.CurrentLine,
		uncommittedChanges: h.repo.UncommittedChanges,
		currentBranchName:  h.repo.CurrentBranchName,
	}, nil
}

func (h *repoVM) getLines(viewPage ui.ViewPage) (int, []string) {
	firstIndex, commits := h.getCommits(viewPage)

	graphWidth := h.repo.GraphWidth + markerWidth
	messageLength, authorLength, timeLength := columnWidths(graphWidth, viewPage.Width)

	var currentLineCommit viewmodel.Commit
	if len(commits) > 0 {
		currentLineCommit = commits[viewPage.CurrentLine-viewPage.FirstLine]
	}

	var lines []string
	for _, c := range commits {
		var sb strings.Builder
		h.writeGraph(&sb, c)
		sb.WriteString(" ")
		writeMoreMarker(&sb, c)
		writeCurrentMarker(&sb, c)
		writeAheadBehindMarker(&sb, c)
		h.writeSubject(&sb, c, currentLineCommit, messageLength)
		sb.WriteString(" ")
		writeAuthor(&sb, c, authorLength)
		sb.WriteString(" ")
		writeAuthorTime(&sb, c, timeLength)
		lines = append(lines, sb.String())
	}
	return firstIndex, lines

}

func (h *repoVM) getCommits(viewPage ui.ViewPage) (int, []viewmodel.Commit) {
	firstIndex := viewPage.FirstLine
	count := viewPage.Height
	if count > len(h.repo.Commits) {
		// Requested count larger than available, return just all available commits
		count = len(h.repo.Commits)
	}

	if firstIndex+count >= len(h.repo.Commits) {
		// Requested commits past available, adjust to return available commits
		firstIndex = len(h.repo.Commits) - count
	}
	return firstIndex, h.repo.Commits[firstIndex : firstIndex+count]
}

func (h *repoVM) showOpenMenu(x, y int) {
	menu := h.mainService.NewMenu("Show")

	items := h.GetOpenBranchMenuItems()
	menu.AddItems(items)

	menu.Add(ui.SeparatorMenuItem)
	menu.Add(ui.MenuItem{Text: "Commit Diff ...", Key: "Ctrl-D", Action: func() {
		h.mainService.ShowDiff(h.currentIndex)
	}})
	// switchItems := h.vm.GetSwitchBranchMenuItems()
	//	menu.Add(ui.MenuItem{Text: "Switch/Checkout", SubItems: switchItems})
	menu.Add(h.mainService.RecentReposMenuItem())
	menu.Add(h.mainService.MainMenuItem())
	//
	menu.Show(x+1, y+2)
}

func (h *repoVM) showCloseMenu() {
	// menu := ui.NewMenu(h.uiHandler, "Hide Branch")
	// items := h.vm.GetCloseBranchMenuItems()
	// if len(items) == 0 {
	// 	return
	// }
	// menu.AddItems(items)
	//
	// y := h.ViewPage().CurrentLine - h.ViewPage().FirstLine + 2
	// menu.Show(10, y)
}

func (h *repoVM) GetOpenBranchMenuItems() []ui.MenuItem {

	branches := h.viewModelService.GetCommitOpenBranches(h.currentIndex, h.repo)

	current, ok := h.viewModelService.CurrentBranch(h.repo)
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
	for _, b := range h.viewModelService.GetActiveBranches(h.repo) {
		activeSubItems = append(activeSubItems, h.toOpenBranchMenuItem(b))
	}
	items = append(items, ui.MenuItem{Text: "Active Branches", SubItems: activeSubItems})

	var allGitSubItems []ui.MenuItem
	for _, b := range h.viewModelService.GetAllBranches(h.repo) {
		if b.IsGitBranch {
			allGitSubItems = append(allGitSubItems, h.toOpenBranchMenuItem(b))
		}
	}
	items = append(items, ui.MenuItem{Text: "All Git Branches", SubItems: allGitSubItems})

	var allSubItems []ui.MenuItem
	for _, b := range h.viewModelService.GetAllBranches(h.repo) {
		allSubItems = append(allSubItems, h.toOpenBranchMenuItem(b))
	}
	items = append(items, ui.MenuItem{Text: "All Branches", SubItems: allSubItems})

	return items

}

func (h *repoVM) GetCloseBranchMenuItems() []ui.MenuItem {
	return nil
	// var items []ui.MenuItem
	// commitBranches := h.viewModelService.GetShownBranches(true)
	// for _, b := range commitBranches {
	// 	items = append(items, h.toCloseBranchMenuItem(b))
	// }
	// return items
}

func (h *repoVM) GetSwitchBranchMenuItems() []ui.MenuItem {
	return nil
	// var items []ui.MenuItem
	// commitBranches := h.viewModelService.GetShownBranches(false)
	// for _, b := range commitBranches {
	// 	items = append(items, h.toSwitchBranchMenuItem(b))
	// }
	// return items
}

func (h *repoVM) CloseBranch(index int) {
	//	h.viewModelService.CloseBranch(index)
}

func (h *repoVM) toOpenBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
	return ui.MenuItem{Text: h.branchItemText(branch), Action: func() {
		h.viewModelService.ShowBranch(branch.Name, h.repo)
	}}
}

// func (h *repoVM) toCloseBranchMenuItem(branch viewmodel.Branch) ui.MenuItem {
// 	return ui.MenuItem{Text: h.branchItemText(branch), Action: func() {
// 		h.viewModelService.HideBranch(branch.Name)
// 	}}
// }

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

func (h *repoVM) refresh() {
	log.Infof("refresh")
	h.viewModelService.TriggerRefreshModel()
}

func (h *repoVM) RefreshTrace(viewPage ui.ViewPage) {
	git.EnableTracing("")
	// traceBytes := utils.MustJsonMarshal(trace{
	// 	RepoPath:    h.viewModelService.RepoPath(),
	// 	ViewPage:    viewPage,
	// 	BranchNames: h.viewModelService.CurrentBranchNames(),
	// })
	// utils.MustFileWrite(filepath.Join(git.CurrentTracePath(), "repovm"), traceBytes)

	//h.viewModelService.TriggerRefreshModel()
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

func (h *repoVM) ChangeBranchColor() {
	//h.viewModelService.ChangeBranchColor(index)
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

func columnWidths(graphWidth, viewWidth int) (msgLength, authorLength, timeLength int) {
	width := viewWidth - graphWidth
	authorLength = 15
	timeLength = 12
	if width < 90 {
		authorLength = 10
		timeLength = 6
	}
	if width < 60 {
		authorLength = 0
		timeLength = 0
	}
	msgLength = viewWidth - graphWidth - authorLength - timeLength
	if msgLength < 0 {
		msgLength = 0
	}
	return
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

func (h *repoVM) showDiff() {

}

func (h *repoVM) saveTotalDebugState() {
	//	h.vm.RefreshTrace(h.ViewPage())
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
