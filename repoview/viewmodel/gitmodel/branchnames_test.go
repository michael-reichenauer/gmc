package gitmodel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBranchNames(t *testing.T) {
	h := newBranchNames()
	fi := h.parseCommit(c("1", "Merge branch 'develop' into master", "2", "3"))
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "master", fi.into)
	fi = h.parseCommit(c("2", "Merge branch 'develop'", "3", "4"))
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "", fi.into)
}

func c(id, subject string, parents ...string) *Commit {
	return &Commit{Id: id, Subject: subject, ParentIDs: parents}
}
