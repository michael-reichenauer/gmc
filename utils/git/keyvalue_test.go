package git

import (
	"testing"

	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestKeyValue(t *testing.T) {
	// Setup git repo
	wf := tests.CreateTempFolder()
	defer tests.CleanTemp()
	git := New(wf.Path())
	assert.NoError(t, git.InitRepo())
	assert.NoError(t, git.ConfigUser("test", "test@test.com"))

	// Test set/get value
	err := git.SetValue("keyname", "value1")
	assert.NoError(t, err)

	v, err := git.GetValue("keyname")
	assert.NoError(t, err)
	assert.Equal(t, "value1", v)

	err = git.SetValue("keyname", "value1")
	assert.NoError(t, err)
}

func TestSyncViaServer(t *testing.T) {
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

	// Prepare for cloned repo 3
	wf3 := tests.CreateTempFolder()
	git3 := New(wf3.Path())
	assert.NoError(t, git3.Clone(git1.RepoPath(), wf3.Path()))

	// Set key in repo 2
	assert.NoError(t, git2.SetValue("keyname", "value1"))
	v2, err := git2.GetValue("keyname")
	assert.NoError(t, err)
	assert.Equal(t, "value1", v2)

	// // Pull in repo 3 and verify
	// assert.NoError(t, git3.Fetch())
	// assert.NoError(t, git3.PullCurrentBranch())
	// l3, err = git3.GetLog()
	// assert.NoError(t, err)
	// assert.Equal(t, 1, len(l3))
	// assert.Equal(t, "initial", l3[0].Subject)
	// assert.Equal(t, "1", wf3.File("a.txt").Read())

	// // Chang file and push repo 3 to server repo 1
	// wf3.File("a.txt").Write("2")
	// assert.NoError(t, git3.Commit("second"))
	// assert.NoError(t, git3.PushBranch("master"))

	// // Update repo 2 from server and verify that commit from repo 3 reached repo 2 via server
	// assert.NoError(t, git2.PullCurrentBranch())
	// l2, err = git2.GetLog()
	// assert.NoError(t, err)
	// assert.Equal(t, 2, len(l2))
	// assert.Equal(t, "second", l2[0].Subject)
	// assert.Equal(t, "initial", l2[1].Subject)
}
