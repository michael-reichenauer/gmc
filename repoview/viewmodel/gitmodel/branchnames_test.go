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
}

func TestParseSubject(t *testing.T) {
	h := newBranchNames()

	fi := h.parseMergeBranchNames("Merge branch 'develop' into master")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "master", fi.into)

	fi = h.parseMergeBranchNames("Merged branch 'develop' into master")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "master", fi.into)

	fi = h.parseMergeBranchNames("Merged commit 'develop' into master")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "master", fi.into)

	fi = h.parseMergeBranchNames("Merged 'develop' into master")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "master", fi.into)

	fi = h.parseMergeBranchNames("Merge branch 'develop'")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "", fi.into)

	fi = h.parseMergeBranchNames("Merge branch develop")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "", fi.into)

	// Pull merge
	fi = h.parseMergeBranchNames("Merge branch 'master' of https://sa-git/git/sa/Products/AcmAcs")
	assert.Equal(t, "master", fi.from)
	assert.Equal(t, "master", fi.into)
}

func c(id, subject string, parents ...string) *Commit {
	return &Commit{Id: id, Subject: subject, ParentIDs: parents}
}
