package gitmodel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBranchNames(t *testing.T) {
	f, i := parseBranchNames(c("1", "Merge branch 'develop' into master", "2", "3"))
	assert.Equal(t, "develop", f)
	assert.Equal(t, "master", i)
	f, i = parseBranchNames(c("1", "Merge branch 'develop'", "2", "3"))
	assert.Equal(t, "develop", f)
	assert.Equal(t, "", i)
}

func c(id, subject string, parents ...string) *Commit {
	return &Commit{Id: id, Subject: subject, ParentIDs: parents}
}
