package git

import (
	"testing"

	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestClonePushPull(t *testing.T) {
	defer tests.CleanTemp()

	// Prepare server repo 1
	wf1 := tests.CreateTempFolder()
	git1 := New(wf1.Path())
	assert.NoError(t, git1.InitRepoBare())
	assert.NoError(t, git1.ConfigUser("test", "test@test.com"))

	// Prepare for cloned repo 2
	wf2 := tests.CreateTempFolder()
	git2 := New(wf2.Path())
	assert.NoError(t, git2.Clone(git1.RepoPath(), wf2.Path()))
	l2, err := git2.GetLog()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(l2))

	// Prepare for cloned repo 3
	wf3 := tests.CreateTempFolder()
	git3 := New(wf3.Path())
	assert.NoError(t, git3.Clone(git1.RepoPath(), wf3.Path()))
	l3, err := git3.GetLog()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(l3))

	// Commit first commit in repo 2 and push to server repo 1
	wf2.File("a.txt").Write("1")
	assert.NoError(t, git2.Commit("initial"))
	assert.NoError(t, git2.PushBranch("master"))

	// Pull in repo 3 and verify
	assert.NoError(t, git3.Fetch())
	assert.NoError(t, git3.PullCurrentBranch())
	l3, err = git3.GetLog()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(l3))
	assert.Equal(t, "initial", l3[0].Subject)
	assert.Equal(t, "1", wf3.File("a.txt").Read())

	// Chang file and push repo 3 to server repo 1
	wf3.File("a.txt").Write("2")
	assert.NoError(t, git3.Commit("second"))
	assert.NoError(t, git3.PushBranch("master"))

	// Update repo 2 from server and verify that commit from repo 3 reached repo 2 via server
	assert.NoError(t, git2.PullCurrentBranch())
	l2, err = git2.GetLog()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(l2))
	assert.Equal(t, "second", l2[0].Subject)
	assert.Equal(t, "initial", l2[1].Subject)
}
