package gitlib

import (
	"fmt"
	"time"
)

type Commit struct {
	ID         string
	SID        string
	ParentIDs  []string
	Subject    string
	Message    string
	Author     string
	AuthorTime time.Time
	CommitTime time.Time
}

func (c *Commit) String() string {
	return fmt.Sprintf("%s %s", c.SID, c.Subject)
}
