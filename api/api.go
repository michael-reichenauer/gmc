package api

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"time"
)

type ViewRepo struct {
	Commits            []Commit
	CurrentBranchName  string
	RepoPath           string
	UncommittedChanges int
	MergeMessage       string
	Conflicts          int
	ConsoleGraph       Graph
}

type Color int

const (
	MoreNone                    = 0
	MoreMergeIn   utils.Bitmask = 1 << iota // ╮
	MoreBranchOut                           // ╭
)

const (
	BBlank                = 0         //
	BCommit utils.Bitmask = 1 << iota // ┣
	BLine                             // ┃
	BPass                             // ╂

	BTip        // ┏
	BBottom     // ┗
	BMergeLeft  // ╭
	BMergeRight // ╮

	BBranchLeft  //  ╰
	BBranchRight // ╯
	BMLine       // │
	BActiveTip   // ┣
)

type Commit struct {
	ID           string
	SID          string
	Subject      string
	Message      string
	Author       string
	AuthorTime   time.Time
	IsCurrent    bool
	Branch       Branch
	Tags         []string
	More         utils.Bitmask
	ParentIDs    []string
	ChildIDs     []string
	BranchTips   []string
	IsLocalOnly  bool
	IsRemoteOnly bool
}

type GraphColumn struct {
	Connect utils.Bitmask
	Branch  utils.Bitmask

	//BranchName        string
	BranchColor Color
	PassColor   Color
}

type GraphRow []GraphColumn
type Graph []GraphRow

type Branch struct {
	Name          string
	DisplayName   string
	Index         int
	IsMultiBranch bool
	RemoteName    string
	LocalName     string
	IsRemote      bool
	IsGitBranch   bool
	IsCurrent     bool
	TipID         string
	HasLocalOnly  bool
	HasRemoteOnly bool
	Color         Color
}

type Branches []Branch

func (bs Branches) Contains(predicate func(b Branch) bool) bool {
	for _, b := range bs {
		if predicate(b) {
			return true
		}
	}
	return false
}

type RepoChange struct {
	IsStarting bool
	ViewRepo   ViewRepo
	SearchText string
	Error      error
}

type Api interface {
	OpenRepo(path string) (Repo, error)
	GetRecentDirs() ([]string, error)
	GetSubDirs(path string) ([]string, error)
}

type Repo interface {
	TriggerRefreshModel()
	TriggerSearch(text string)
	StartMonitor()
	Close()
	RepoChanges() chan RepoChange
	GetCommitOpenInBranches(id string) []Branch
	GetCommitOpenOutBranches(id string) []Branch
	CurrentNotShownBranch() (Branch, bool)
	CurrentBranch() (Branch, bool)
	GetLatestBranches(shown bool) []Branch
	GetAllBranches(shown bool) []Branch
	GetShownBranches(master bool) []Branch
	ShowBranch(name string)
	HideBranch(name string)
	SwitchToBranch(name string, name2 string) error
	PushBranch(name string) error
	PullBranch() error
	MergeBranch(name string) error
	CreateBranch(name string) error
	DeleteBranch(name string) error
	GetCommitDiff(id string) (git.CommitDiff, error)
	Commit(message string) error
}
