package augmented

import (
	"fmt"

	"github.com/michael-reichenauer/gmc/utils/git"
)

type Status struct {
	Modified     int
	Added        int
	Deleted      int
	Conflicted   int
	IsMerging    bool
	MergeMessage string
}

func newStatus(gs git.Status) Status {
	return Status{
		Modified:     gs.Modified,
		Added:        gs.Added,
		Deleted:      gs.Deleted,
		Conflicted:   gs.Conflicted,
		IsMerging:    gs.IsMerging,
		MergeMessage: gs.MergeMessage,
	}
}

func (s Status) OK() bool {
	return s.AllChanges() == 0 && !s.IsMerging
}

func (s Status) AllChanges() int {
	return s.Modified + s.Added + s.Deleted + s.Conflicted
}

func (s *Status) String() string {
	return fmt.Sprintf("M:%d,A:%d,D:%d,C:%d", s.Modified, s.Added, s.Deleted, s.Conflicted)
}
