package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestBranches(t *testing.T) {
	wf := tests.CreateTempFolder()
	defer tests.CleanTemp()
	file1 := path.Join(wf, "a.txt")
	assert.NoError(t, InitRepo(wf))
	gr := OpenRepo(wf)

	bs, err := gr.GetBranches()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(bs))

	assert.NoError(t, utils.FileWrite(file1, []byte("1")))
	assert.NoError(t, gr.Commit("initial"))
	assert.Equal(t, "1", string(utils.MustFileRead(file1)))

	bs, _ = gr.GetBranches()
	assert.Equal(t, 1, len(bs))
	assert.Equal(t, "master", bs[0].Name)
	assert.Equal(t, "master", bs.MustCurrent().Name)

	err = gr.CreateBranch("develop")
	bs, _ = gr.GetBranches()
	assert.Equal(t, 2, len(bs))
	assert.Equal(t, "develop", bs.MustCurrent().Name)
	assert.NoError(t, utils.FileWrite(file1, []byte("2")))
	assert.NoError(t, gr.Commit("second"))

	assert.NoError(t, gr.Checkout("master"))
	assert.Equal(t, "1", string(utils.MustFileRead(file1)))
	assert.NoError(t, gr.MergeBranch("develop"))
	assert.NoError(t, gr.Commit("merged"))
	assert.Equal(t, "2", string(utils.MustFileRead(file1)))

	cs, _ := gr.GetLog()

	err = gr.CreateBranchAt("feature", cs.MustBySubject("second").ID)
	bs, _ = gr.GetBranches()
	assert.Equal(t, 3, len(bs))

	assert.NoError(t, utils.FileWrite(file1, []byte("3")))
	assert.NoError(t, gr.Commit("third"))

	assert.NoError(t, gr.Checkout("master"))
	assert.Equal(t, "2", string(utils.MustFileRead(file1)))

	assert.NoError(t, gr.MergeBranch("feature"))
	assert.NoError(t, gr.Commit("merged2"))
	assert.Equal(t, "3", string(utils.MustFileRead(file1)))

	cs, _ = gr.GetLog()
	assert.Equal(t, cs.MustBySubject("merged").ID, cs.MustBySubject("merged2").ParentIDs[0])
	assert.Equal(t, cs.MustBySubject("third").ID, cs.MustBySubject("merged2").ParentIDs[1])
	assert.Equal(t, cs.MustBySubject("second").ID, cs.MustBySubject("third").ParentIDs[0])
	assert.Equal(t, cs.MustBySubject("initial").ID, cs.MustBySubject("merged").ParentIDs[0])
	assert.Equal(t, cs.MustBySubject("second").ID, cs.MustBySubject("merged").ParentIDs[1])
	assert.Equal(t, cs.MustBySubject("initial").ID, cs.MustBySubject("second").ParentIDs[0])
	assert.Equal(t, 5, len(cs))
	// t.Logf("\n%s", cs)
}

func TestListBranches(t *testing.T) {
	tests.ManualTest(t)
	branches, err := newBranchService(newGitCmd(utils.CurrentDir())).getBranches()
	assert.NoError(t, err)
	for _, b := range branches {
		log.Infof("%+v", b)
		t.Logf("%v", b)
	}
}
