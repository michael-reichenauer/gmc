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
