package viewrepo

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils"
	"time"
)

type commit struct {
	Index        int
	ID           string
	SID          string
	Subject      string
	Message      string
	Author       string
	AuthorTime   time.Time
	Parent       *commit
	MergeParent  *commit
	ParentIDs    []string
	ChildIDs     []string
	IsCurrent    bool
	Tags         []string
	More         utils.Bitmask
	Branch       *branch
	graph        []api.GraphColumn
	BranchTips   []string
	IsLocalOnly  bool
	IsRemoteOnly bool
}

func (c *commit) String() string {
	return fmt.Sprintf("%s %s (%s)", c.SID, c.Subject, c.Branch.displayName)
}
