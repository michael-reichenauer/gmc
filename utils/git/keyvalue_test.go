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
	err := git.SetKeyValue("keyname", "value1")
	assert.NoError(t, err)

	v, err := git.GetKeyValue("keyname")
	assert.NoError(t, err)
	assert.Equal(t, "value1", v)

	err = git.SetKeyValue("keyname", "value1")
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
	assert.NoError(t, git2.ConfigUser("test", "test@test.com"))
	assert.NoError(t, git2.Clone(git1.RepoPath(), wf2.Path()))

	// Prepare for cloned repo 3
	wf3 := tests.CreateTempFolder()
	git3 := New(wf3.Path())
	assert.NoError(t, git3.ConfigUser("test", "test@test.com"))
	assert.NoError(t, git3.Clone(git1.RepoPath(), wf3.Path()))

	// Set key in repo 2
	assert.NoError(t, git2.SetKeyValue("keyname", "value1"))
	v2, err := git2.GetKeyValue("keyname")
	assert.NoError(t, err)
	assert.Equal(t, "value1", v2)

	// Push key from repo 2 to server repo 1
	assert.NoError(t, git2.PushKeyValue("keyname"))

	// Verify that repo 3 does not have the key yet
	_, err = git3.GetKeyValue("keyname")
	assert.True(t, err != nil)

	// Pull the key from server repo into repo 3 and verify value
	assert.NoError(t, git3.PullKeyValue("keyname"))
	v3, err := git3.GetKeyValue("keyname")
	assert.NoError(t, err)
	assert.Equal(t, "value1", v3)

	// Update value in repo 3 and push to server repo
	assert.NoError(t, git3.SetKeyValue("keyname", "value2"))
	assert.NoError(t, git3.PushKeyValue("keyname"))

	// Pull the key to repo 3 and verify new value
	assert.NoError(t, git2.PullKeyValue("keyname"))
	v2, err = git2.GetKeyValue("keyname")
	assert.NoError(t, err)
	assert.Equal(t, "value2", v2)
}
