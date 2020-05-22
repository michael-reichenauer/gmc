package api

import (
	"github.com/michael-reichenauer/gmc/utils"
	"time"
)

type DiffMode int

const (
	DiffModified DiffMode = iota
	DiffAdded
	DiffRemoved
	DiffSame
	DiffConflicts
	DiffConflictStart
	DiffConflictSplit
	DiffConflictEnd
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
	ID                 string
	SID                string
	Subject            string
	Message            string
	Author             string
	AuthorTime         time.Time
	IsCurrent          bool
	Branch             Branch
	Tags               []string
	More               utils.Bitmask
	ParentIDs          []string
	ChildIDs           []string
	BranchTips         []string
	IsLocalOnly        bool
	IsRemoteOnly       bool
	IsUncommitted      bool
	IsPartialLogCommit bool
}

type GraphColumn struct {
	Connect     utils.Bitmask
	Branch      utils.Bitmask
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

type CommitDiff struct {
	FileDiffs []FileDiff
}

type FileDiff struct {
	PathBefore   string
	PathAfter    string
	IsRenamed    bool
	DiffMode     DiffMode
	SectionDiffs []SectionDiff
}

type SectionDiff struct {
	ChangedIndexes string
	LinesDiffs     []LinesDiff
}

type LinesDiff struct {
	DiffMode DiffMode
	Line     string
}
