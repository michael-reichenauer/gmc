package viewmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel/gitrepo"
	"github.com/michael-reichenauer/gmc/utils/gitlib"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"github.com/thoas/go-funk"
	"hash/fnv"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	masterName       = "master"
	remoteMasterName = "origin/master"
)

var branchColors = []ui.Color{
	ui.CRed,
	ui.CBlue,
	ui.CYellow,
	ui.CGreen,
	ui.CCyan,
	ui.CRedDk,
	ui.CGreenDk,
	ui.CYellowDk,
	//ui.CBlueDk,
	ui.CMagenta,
	ui.CMagentaDk,
	ui.CCyanDk,
}

type Status struct {
	AllChanges int
	GraphWidth int
}

type Service struct {
	ChangedEvents chan interface{}

	gitRepoService *gitrepo.Service
	configService  *config.Service
	branchesGraph  *branchesGraph

	lock               sync.Mutex
	currentViewModel   *repo
	gmRepo             gitrepo.Repo
	gmStatus           gitrepo.Status
	customBranchColors map[string]int
}

func NewModel(configService *config.Service) *Service {
	return &Service{
		ChangedEvents:      make(chan interface{}),
		branchesGraph:      newBranchesGraph(),
		gitRepoService:     gitrepo.NewService(),
		configService:      configService,
		currentViewModel:   newRepo(),
		customBranchColors: make(map[string]int),
	}
}

func ToSid(commitID string) string {
	return gitrepo.ToSid(commitID)
}

func (s *Service) Start() {
	go s.monitorGitModelRoutine()
}

func (s *Service) OpenRepo(workingFolder string) error {
	log.Eventf("repo-open", "open %q", workingFolder)
	s.gitRepoService.Open(workingFolder)
	repo := s.configService.GetRepo(s.gitRepoService.RepoPath())
	for _, b := range repo.Branches {
		s.customBranchColors[b.DisplayName] = b.Color
	}
	s.TriggerRefreshModel()
	return nil
}

func (s *Service) TriggerRefreshModel() {
	log.Event("vms-refresh")
	s.gitRepoService.TriggerRefreshRepo()
}

func (s *Service) LoadRepo(branchIds []string) {
	log.Event("vms-load-repo")
	gmRepo, err := s.gitRepoService.GetFreshRepo()
	if err != nil {
		log.Infof("Detected repo error: %v", err)
		return
	}
	s.lock.Lock()
	s.gmRepo = gmRepo
	s.gmStatus = gmRepo.Status
	s.lock.Unlock()
	repo := s.getViewModel(branchIds)
	s.lock.Lock()
	s.currentViewModel = repo
	s.lock.Unlock()
}

func (s *Service) showBranches(branchNames []string) {
	log.Event("vms-show-branches")
	t := time.Now()
	repo := s.getViewModel(branchNames)
	log.Infof("LoadBranches time %v", time.Since(t))

	s.lock.Lock()
	s.currentViewModel = repo
	s.lock.Unlock()
	s.ChangedEvents <- nil
	s.configService.SetRepo(s.RepoPath(), func(r *config.Repo) {
		r.ShownBranches = s.CurrentBranchNames()
	})
}

func (s *Service) monitorGitModelRoutine() {
	for {
		select {
		case gmRepo := <-s.gitRepoService.RepoEvents:
			log.Infof("Detected repo change")
			s.lock.Lock()
			s.gmRepo = gmRepo
			s.gmStatus = gmRepo.Status
			s.lock.Unlock()
			log.Event("vms-changed-repo")

		case gmStatus := <-s.gitRepoService.StatusEvents:
			log.Infof("Detected status change")
			s.lock.Lock()
			s.gmStatus = gmStatus
			s.lock.Unlock()
			log.Event("vms-changed-status")
		case err := <-s.gitRepoService.ErrorEvents:
			log.Infof("Detected repo error: %v", err)
			log.Event("vms-repo-error")
		}

		// Refresh model
		s.showBranches(s.CurrentBranchNames())
	}
}

func (s *Service) CurrentBranchNames() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.currentViewModel == nil {
		return []string{}
	}

	var branchNames []string
	for _, b := range s.currentViewModel.Branches {
		branchNames = append(branchNames, b.name)
	}
	return branchNames
}

func (s *Service) GetCommitByIndex(index int) (Commit, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if index < 0 || index >= len(s.currentViewModel.Commits) {
		return Commit{}, fmt.Errorf("no commit")
	}
	return toCommit(s.currentViewModel.Commits[index]), nil
}

func (s *Service) GetRepoViewPort(firstIndex, count int) (ViewPort, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if count > len(s.currentViewModel.Commits) {
		// Requested count larger than available, return just all available commits
		count = len(s.currentViewModel.Commits)
	}

	if firstIndex+count >= len(s.currentViewModel.Commits) {
		// Requested commits past available, adjust to return available commits
		firstIndex = len(s.currentViewModel.Commits) - count
	}

	return newViewPort(s.currentViewModel, firstIndex, count), nil
}

func (s *Service) CurrentBranch() (Branch, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	current, ok := s.gmRepo.CurrentBranch()
	if !ok {
		return Branch{}, false
	}
	if containsBranch(s.currentViewModel.Branches, current.Name) {
		return Branch{}, false
	}

	return toBranch(s.currentViewModel.toBranch(current, 0)), true
}

func (s *Service) GetAllBranches() []Branch {
	s.lock.Lock()
	defer s.lock.Unlock()
	var branches []Branch
	for _, b := range s.gmRepo.Branches {
		if containsDisplayNameBranch(branches, b.DisplayName) {
			continue
		}
		if containsBranch(s.currentViewModel.Branches, b.Name) {
			continue
		}
		branches = append(branches, toBranch(s.currentViewModel.toBranch(b, 0)))
	}
	sort.SliceStable(branches, func(i, j int) bool {
		return -1 == strings.Compare(branches[i].DisplayName, branches[j].DisplayName)
	})

	return branches
}

func (s *Service) GetActiveBranches() []Branch {
	s.lock.Lock()
	defer s.lock.Unlock()
	var branches []Branch
	for _, b := range s.gmRepo.Branches {
		if containsDisplayNameBranch(branches, b.DisplayName) {
			continue
		}
		if containsBranch(s.currentViewModel.Branches, b.Name) {
			continue
		}
		branches = append(branches, toBranch(s.currentViewModel.toBranch(b, 0)))
	}
	sort.SliceStable(branches, func(i, j int) bool {
		return s.gmRepo.CommitById[branches[i].TipID].AuthorTime.After(s.gmRepo.CommitById[branches[j].TipID].AuthorTime)
	})

	return branches
}

func containsDisplayNameBranch(branches []Branch, displayName string) bool {
	for _, b := range branches {
		if displayName == b.DisplayName {
			return true
		}
	}
	return false
}

func (s *Service) GetCommitOpenBranches(index int) []Branch {
	s.lock.Lock()
	defer s.lock.Unlock()
	if index >= len(s.currentViewModel.Commits) {
		// Repo must just have changed, just ignore
		return nil
	}
	c := s.currentViewModel.Commits[index]
	if !c.IsMore {
		// Not a graph point that can be expanded  (no branches at this point)
		return nil
	}

	var branches []*gitrepo.Branch

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it
		mergeParent := s.gmRepo.CommitById[c.ParentIDs[1]]
		branches = append(branches, mergeParent.Branch)
	}

	for _, ccId := range c.ChildIDs {
		// commit has children (i.e.), other branches have merged from this branch
		if ccId == StatusID {
			continue
		}
		cc := s.gmRepo.CommitById[ccId]
		if cc.Branch.Name != c.Branch.name {
			branches = append(branches, cc.Branch)
		}
	}

	for _, b := range s.gmRepo.Branches {
		if b.TipID == b.BottomID && b.BottomID == c.ID && b.ParentBranch.Name == c.Branch.name {
			// empty branch with no own branch commit, (branch start)
			branches = append(branches, b)
		}
	}

	var bs []Branch
	for _, b := range branches {
		// if Branches(bs).Contains(func(bsb Branch) bool {
		// 	return b.Name == bsb.Name || b.RemoteName == b.Name ||
		// 		b.Name == bsb.RemoteName
		// }) {
		// 	// Skip duplicates
		// 	continue
		// }
		if nil != funk.Find(bs, func(bsb Branch) bool {
			return b.Name == bsb.Name || b.RemoteName == b.Name ||
				b.Name == bsb.RemoteName
		}) {
			// Skip duplicates
			continue
		}
		if containsBranch(s.currentViewModel.Branches, b.Name) {
			// Skip branches already shown
			continue
		}
		bs = append(bs, toBranch(s.currentViewModel.toBranch(b, 0)))
	}

	return bs
}

func (s *Service) GetCommitCloseBranches(index int) []Branch {
	s.lock.Lock()
	defer s.lock.Unlock()

	if index >= len(s.currentViewModel.Commits) {
		// Repo must just have changed, just ignore
		return nil
	}
	c := s.currentViewModel.Commits[index]

	var branches []*gitrepo.Branch

	if len(c.ParentIDs) > 1 {
		// commit has branch merged into this commit add it
		mergeParent := s.gmRepo.CommitById[c.ParentIDs[1]]
		branches = append(branches, mergeParent.Branch)
	}

	for _, ccId := range c.ChildIDs {
		// commit has children (i.e.), other branches have merged from this branch
		if ccId == StatusID {
			continue
		}
		cc := s.gmRepo.CommitById[ccId]
		if cc.Branch.Name != c.Branch.name {
			branches = append(branches, cc.Branch)
		}
	}

	for _, b := range s.gmRepo.Branches {
		if b.TipID == b.BottomID && b.BottomID == c.ID &&
			b.ParentBranch != nil && b.ParentBranch.Name == c.Branch.name {
			// empty branch with no own branch commit, (branch start)
			branches = append(branches, b)
		}
	}

	if c.ID != StatusID {
		commit := s.gmRepo.CommitById[c.ID]
		branches = append(branches, commit.Branch)
	}

	var bs []Branch
	for _, b := range branches {
		if b.Name == masterName || b.Name == remoteMasterName {
			continue
		}
		if containsDisplayNameBranch(bs, b.DisplayName) {
			continue
		}
		if containsViewBranch(bs, b.Name) {
			// Skip duplicates
			continue
		}
		if !containsBranch(s.currentViewModel.Branches, b.Name) {
			// Skip branches not shown
			continue
		}
		bs = append(bs, toBranch(s.currentViewModel.toBranch(b, 0)))
	}

	for _, b := range s.currentViewModel.Branches {
		if b.name == masterName || b.name == remoteMasterName {
			continue
		}
		if containsDisplayNameBranch(bs, b.displayName) {
			continue
		}
		if containsViewBranch(bs, b.name) {
			// Skip duplicates
			continue
		}
		if containsViewBranch(bs, b.name) {
			// Skip duplicates
			continue
		}
		bs = append(bs, toBranch(b))
	}
	return bs
}

func containsViewBranch(branches []Branch, name string) bool {
	for _, b := range branches {
		if name == b.Name {
			return true
		}
	}
	return false
}

func containsBranch(branches []*branch, name string) bool {
	for _, b := range branches {
		if name == b.name {
			return true
		}
	}
	return false
}

func (s *Service) ShowBranch(name string) {
	s.lock.Lock()

	branchNames := s.toBranchNames(s.currentViewModel.Branches)

	branch, ok := s.gmRepo.BranchByName(name)
	if !ok {
		s.lock.Unlock()
		return
	}

	branchNames = s.addBranchWithAncestors(branchNames, branch)
	s.lock.Unlock()

	log.Event("vms-branches-open")
	s.showBranches(branchNames)
}

func (s *Service) HideBranch(name string) {
	s.lock.Lock()

	branch := s.currentViewModel.BranchByName(name)
	if branch == nil {
		s.lock.Unlock()
		return
	}

	var branchNames []string
	for _, b := range s.currentViewModel.Branches {
		if b.name != branch.name && !branch.isAncestor(b) {
			branchNames = append(branchNames, b.name)
		}
	}

	s.lock.Unlock()

	log.Event("vms-branches-close")
	s.showBranches(branchNames)
}

func (s *Service) RepoPath() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.gmRepo.RepoPath
}

func (s *Service) CloseBranch(index int) {
	s.lock.Lock()
	if index >= len(s.currentViewModel.Commits) {
		// Repo must just have changed, just ignore
		s.lock.Unlock()
		return
	}
	c := s.currentViewModel.Commits[index]
	if c.Branch.name == masterName || c.Branch.name == remoteMasterName {
		// Cannot close master
		s.lock.Unlock()
		return
	}

	// get branch ids except for the commit branch or decedent branches
	var branchNames []string
	for _, b := range s.currentViewModel.Branches {
		if b.name != c.Branch.name && !c.Branch.isAncestor(b) {
			branchNames = append(branchNames, b.name)
		}
	}
	s.lock.Unlock()
	log.Event("vms-branches-close")
	s.showBranches(branchNames)
}

func (s *Service) getViewModel(branchNames []string) *repo {
	repo := newRepo()
	repo.gmRepo = s.gmRepo
	repo.gmStatus = s.gmStatus

	branches := s.getGitModelBranches(branchNames, s.gmRepo)
	for _, b := range branches {
		repo.addBranch(b)
	}
	currentBranch, ok := s.gmRepo.CurrentBranch()
	if ok {
		repo.CurrentBranchName = currentBranch.Name
	}

	repo.addVirtualStatusCommit()
	for _, c := range repo.gmRepo.Commits {
		repo.addGitCommit(c)
	}
	s.adjustCurrentBranchIfStatus(repo)
	s.setBranchParentChildRelations(repo)
	s.setParentChildRelations(repo)

	// Draw branch lines
	s.branchesGraph.drawBranchLines(repo)

	// Draw branch connector lines
	s.branchesGraph.drawConnectorLines(repo)
	return repo
}

func (s *Service) setBranchParentChildRelations(repo *repo) {
	for _, b := range repo.Branches {
		b.tip = repo.commitById[b.tipId]
		b.bottom = repo.commitById[b.bottomId]
		if b.parentBranchName != "" {
			b.parentBranch = repo.BranchByName(b.parentBranchName)
		}
	}
}

func (s *Service) setParentChildRelations(repo *repo) {
	for _, c := range repo.Commits {
		if len(c.ParentIDs) > 0 {
			// Commit has a parent
			c.Parent = repo.commitById[c.ParentIDs[0]]
			if len(c.ParentIDs) > 1 {
				// Merge parent can be nil if not the merge parent branch is included in the repo view model as well
				c.MergeParent = repo.commitById[c.ParentIDs[1]]
				if c.MergeParent == nil {
					// commit has a merge parent, that is not visible, mark the commit as expandable
					c.IsMore = true
				}
			}
		}

		for _, childID := range c.ChildIDs {
			if _, ok := repo.commitById[childID]; !ok {
				// commit has a child, that is not visible, mark the commit as expandable
				c.IsMore = true
			}
		}

		if !c.IsMore {
			for _, branchName := range c.BranchTips {
				if !repo.containsBranchName(branchName) {
					c.IsMore = true
					break
				}
			}

		}
	}
}

func (s *Service) toBranchNames(branches []*branch) []string {
	var ids []string
	for _, b := range branches {
		ids = append(ids, b.name)
	}
	return ids
}

func (s *Service) ChangeBranchColor(index int) {
	log.Event("vms-branches-change-color")
	var branchName string
	s.lock.Lock()
	if index >= len(s.currentViewModel.Commits) {
		// Repo must just have changed, just ignore
		s.lock.Unlock()
		return
	}
	c := s.currentViewModel.Commits[index]
	branchName = c.Branch.displayName
	s.lock.Unlock()

	color := s.BranchColor(branchName)
	for i, c := range branchColors {
		if color == c {
			index := int((i + 1) % len(branchColors))
			color = branchColors[index]
			s.customBranchColors[branchName] = int(color)
			break
		}
	}

	s.configService.SetRepo(s.gmRepo.RepoPath, func(r *config.Repo) {
		isSet := false
		cb := config.Branch{DisplayName: branchName, Color: int(color)}

		for i, b := range r.Branches {
			if branchName == b.DisplayName {
				r.Branches[i] = cb
				isSet = true
				break
			}
		}
		if !isSet {
			r.Branches = append(r.Branches, cb)
		}
	})
}

func (s *Service) BranchColor(name string) ui.Color {
	if strings.HasPrefix(name, "multi:") {
		return ui.CWhite
	}
	color, ok := s.customBranchColors[name]
	if ok {
		return ui.Color(color)
	}
	if name == "master" {
		return ui.CMagenta
	}
	if name == "develop" {
		return ui.CRedDk
	}

	h := fnv.New32a()
	h.Write([]byte(name))
	index := int(h.Sum32()) % len(branchColors)
	return branchColors[index]
}

func (s *Service) adjustCurrentBranchIfStatus(repo *repo) {
	if len(repo.Commits) < 2 || repo.Commits[0].ID != StatusID || repo.CurrentCommit == nil {
		return
	}

	// Link status commit with current commit and adjust branch
	statusCommit := repo.Commits[0]
	current := repo.CurrentCommit
	statusCommit.Parent = current
	current.ChildIDs = append([]string{statusCommit.ID}, current.ChildIDs...)
	statusCommit.Branch.tip = statusCommit
	statusCommit.Branch.tipId = statusCommit.ID

	if statusCommit.Branch.name != current.Branch.name {
		// Status commit is first local commit, lets adjust branch
		statusCommit.Branch.bottom = statusCommit
		statusCommit.Branch.bottomId = statusCommit.ID
	}
}

func (s *Service) GetCommitDiff(id string) ([]gitlib.FileDiff, error) {
	return s.gitRepoService.GetCommitDiff(id)
}
