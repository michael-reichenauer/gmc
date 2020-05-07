package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBranches(t *testing.T) {
	wf := tests.CreateTempFolder()
	defer tests.CleanTemp()
	file1 := "a.txt"
	git := New(wf.Path())
	assert.NoError(t, git.InitRepo())

	bs, err := git.GetBranches()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(bs))

	wf.File(file1).Write("1")
	assert.NoError(t, git.Commit("initial"))
	assert.Equal(t, "1", wf.File(file1).Read())

	bs, _ = git.GetBranches()
	assert.Equal(t, 1, len(bs))
	assert.Equal(t, "master", bs[0].Name)
	assert.Equal(t, "master", bs.MustCurrent().Name)

	assert.NoError(t, git.CreateBranch("develop"))
	bs, _ = git.GetBranches()
	assert.Equal(t, 2, len(bs))
	assert.Equal(t, "develop", bs.MustCurrent().Name)
	wf.File(file1).Write("2")
	assert.NoError(t, git.Commit("second"))

	assert.NoError(t, git.Checkout("master"))
	assert.Equal(t, "1", wf.File(file1).Read())
	assert.NoError(t, git.MergeBranch("develop"))
	assert.NoError(t, git.Commit("merged"))
	assert.Equal(t, "2", wf.File(file1).Read())

	cs, _ := git.GetLog()

	assert.NoError(t, git.CreateBranchAt("feature", cs.MustBySubject("second").ID))
	bs, _ = git.GetBranches()
	assert.Equal(t, 3, len(bs))

	wf.File(file1).Write("3")
	assert.NoError(t, git.Commit("third"))

	assert.NoError(t, git.Checkout("master"))
	assert.Equal(t, "2", wf.File(file1).Read())

	assert.NoError(t, git.MergeBranch("feature"))
	assert.NoError(t, git.Commit("merged2"))
	assert.Equal(t, "3", wf.File(file1).Read())

	cs, _ = git.GetLog()
	assert.Equal(t, cs.MustBySubject("merged").ID, cs.MustBySubject("merged2").ParentIDs[0])
	assert.Equal(t, cs.MustBySubject("third").ID, cs.MustBySubject("merged2").ParentIDs[1])
	assert.Equal(t, cs.MustBySubject("second").ID, cs.MustBySubject("third").ParentIDs[0])
	assert.Equal(t, cs.MustBySubject("initial").ID, cs.MustBySubject("merged").ParentIDs[0])
	assert.Equal(t, cs.MustBySubject("second").ID, cs.MustBySubject("merged").ParentIDs[1])
	assert.Equal(t, cs.MustBySubject("initial").ID, cs.MustBySubject("second").ParentIDs[0])
	assert.Equal(t, 5, len(cs))
	// t.Logf("\n%s", cs)
}

func TestListBranches_Manual(t *testing.T) {
	tests.ManualTest(t)

	branches, err := newBranchService(newGitCmd(utils.CurrentDir())).getBranches()
	assert.NoError(t, err)
	for _, b := range branches {
		log.Infof("%+v", b)
		t.Logf("%v", b)
	}
}
