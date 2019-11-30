package model

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"time"
)

type GraphColumn struct {
	Connect  utils.Bitmask
	Branch   utils.Bitmask
	BranchId string
}

type commit struct {
	Index       int
	ID          string
	SID         string
	Subject     string
	Message     string
	Author      string
	AuthorTime  time.Time
	Parent      *commit
	MergeParent *commit
	//Children        []*commit
	//Children      []*commit
	ParentIDs []string
	ChildIDs  []string
	IsCurrent bool
	IsMore    bool
	Branch    *branch
	graph     []GraphColumn
}

func (c *commit) String() string {
	return fmt.Sprintf("%s %s (%s)", c.SID, c.Subject, c.Branch.displayName)
}

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
