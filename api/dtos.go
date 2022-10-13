package api

import (
	"time"

	"github.com/michael-reichenauer/gmc/utils"
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

type NoArg int
type NoRsp *int

var (
	NilArg NoArg = 0
	NilRsp NoRsp = &no
	no     int
)

type CheckoutReq struct {
	RepoID      string
	Name        string
	DisplayName string
}

type GetBranchesReq struct {
	RepoID                    string
	IncludeOnlyCurrent        bool
	IncludeOnlyGitBranches    bool
	IncludeOnlyCommitBranches string
	IncludeOnlyShown          bool
	IncludeOnlyNotShown       bool
	SkipMaster                bool
	SortOnLatest              bool
}

type BranchName struct {
	RepoID     string
	BranchName string
}

type Search struct {
	RepoID string
	Text   string
}

type FilesReq struct {
	RepoID string
	Ref    string
}

type CommitDetailsReq struct {
	RepoID   string
	CommitID string
}

type CommitDetailsRsp struct {
	Id         string
	BranchName string
	Message    string
	Files      []string
}

type CommitDiffInfoReq struct {
	RepoID   string
	CommitID string
}

type FileDiffInfoReq struct {
	RepoID string
	Path   string
}

type AmbiguousBranchBranchesReq struct {
	RepoID   string
	CommitID string
}

type CommitInfoReq struct {
	RepoID  string
	Message string
}

type Repo struct {
	Commits            []Commit
	Branches           []Branch
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
	BBlank                   = 0         //
	BCommit    utils.Bitmask = 1 << iota // ┣
	BLine                                // ┃
	BTip                                 // ┏
	BBottom                              // ┗
	BActiveTip                           // ┣

	// Connections
	Pass           // ─ or ╂ for branch positions
	MergeFromLeft  // ╭
	MergeFromRight // ╮
	BranchToLeft   // ╰
	BranchToRight  // ╯
	ConnectLine    // │
)

type Commit struct {
	ID                 string
	SID                string
	Subject            string
	Message            string
	Author             string
	AuthorTime         time.Time
	IsCurrent          bool
	BranchIndex        int
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
	Name                 string
	DisplayName          string
	Index                int
	IsAmbiguousBranch    bool
	RemoteName           string
	LocalName            string
	IsRemote             bool
	IsGitBranch          bool
	IsCurrent            bool
	IsSetAsParent        bool
	IsMainBranch         bool
	TipID                string
	HasLocalOnly         bool
	HasRemoteOnly        bool
	Color                Color
	AmbiguousBranchNames []string

	IsShown bool
	IsIn    bool
	IsOut   bool
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
	ViewRepo   Repo
	SearchText string
	Error      error
}

type CommitDiff struct {
	Id        string
	Author    string
	Date      string
	Message   string
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
