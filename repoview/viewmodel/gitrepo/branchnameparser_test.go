package gitrepo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBranchNames(t *testing.T) {
	h := newBranchNameParser()

	fi := h.parseCommit(c("1", "Merge branch 'develop' into master", "2", "3"))
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "master", fi.into)
}

func TestParseSubject(t *testing.T) {
	h := newBranchNameParser()

	fi := h.parseMergeBranchNames("Merge branch 'develop' into master")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "master", fi.into)

	fi = h.parseMergeBranchNames("Merge from branch 'develop' into master")
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

	fi = h.parseMergeBranchNames("Merge remote-tracking branch 'refs/remotes/origin/branches/fetch' into branches/fetch")
	assert.Equal(t, "branches/fetch", fi.from)
	assert.Equal(t, "branches/fetch", fi.into)

	fi = h.parseMergeBranchNames("Merge branch 'develop'")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "", fi.into)

	fi = h.parseMergeBranchNames("Merge branch develop")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "", fi.into)

	// Optional source repo
	fi = h.parseMergeBranchNames("Merge branch 'develop' of https://github.com/michael-reichenauer/gmc into branches/fetch")
	assert.Equal(t, "develop", fi.from)
	assert.Equal(t, "branches/fetch", fi.into)

	// Pull merge
	fi = h.parseMergeBranchNames("Merge branch 'branches/fetch' of https://github.com/michael-reichenauer/gmc")
	assert.Equal(t, "branches/fetch", fi.from)
	assert.Equal(t, "branches/fetch", fi.into)

}

func c(id, subject string, parents ...string) *Commit {
	return &Commit{Id: id, Subject: subject, ParentIDs: parents}
}
