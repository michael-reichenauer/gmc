package console

import (
	"testing"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/server/viewrepo"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
)

type getterMock struct {
	git git.Git
}

func (t getterMock) GetCommitDiff(id string, rsp *api.CommitDiff) error {
	diff, err := t.git.CommitDiff(id)
	if err != nil {
		return err
	}
	*rsp = viewrepo.ToCommitDiff(diff)
	return nil
}

func TestConflicts_Manual(t *testing.T) {
	tests.ManualTest(t)
	wf := tests.CreateTempFolder()
	g := git.New(wf.Path())
	// viewer := newViewerMock(newUIMock(), nil)
	// getter := &getterMock{git: g}

	g.InitRepo()
	g.ConfigRepoUser("test", "test@test.com")

	file1 := "a.txt"
	file2 := "b.txt"
	assert.NoError(t, g.InitRepo())
	wf.File(file1).Write("1\n2\n3\n")
	wf.File(file2).Write("5\n6\n7\n")
	assert.NoError(t, g.Commit("initial"))

	assert.NoError(t, g.CreateBranch("develop"))
	wf.File(file1).Write("1\n22\n3\n")
	wf.File(file2).Write("5\n62\n7\n")
	assert.NoError(t, g.Commit("second"))

	assert.NoError(t, g.Checkout("master"))
	assert.NoError(t, g.MergeBranch("develop"))
	assert.NoError(t, g.Commit("merged"))
	assert.Equal(t, "1\n22\n3\n", wf.File(file1).Read())

	wf.File(file1).Write("1\nx2\n3\n")
	assert.NoError(t, g.Commit("commitonmaster"))

	assert.NoError(t, g.Checkout("develop"))
	wf.File(file1).Write("1\ny2\n3\n")
	wf.File(file2).Write("5\n63\n7\n")
	assert.NoError(t, g.Commit("commitondevelop"))

	assert.NoError(t, g.Checkout("master"))

	assert.Equal(t, git.ErrConflicts, g.MergeBranch("develop"))

	//assert.NoError(t, g.MergeBranch("develop"))

	// vm := newDiffVM(newUIMock(), viewer, getter, git.UncommittedID)
	// vm.load()
	// viewer.Run()
	// vtl, err := vm.getCommitDiffLeft(cui.ViewPage{Width: 10, Height: 1000})
	// assert.NoError(t, err)
	// vtr, err := vm.getCommitDiffRight(cui.ViewPage{Width: 10, Height: 1000})
	// assert.NoError(t, err)
	// t.Logf("Left:\n%s", strings.Join(vtl.Lines, "\n"))
	// t.Logf("Right:\n%s", strings.Join(vtr.Lines, "\n"))
}

func TestDiffVM_Manual(t *testing.T) {
	tests.ManualTest(t)

	// g := git.New(git.CurrentRoot())
	// viewer := newViewerMock(newUIMock(), nil)
	// getter := &getterMock{git: g}
	// id := git.UncommittedID
	// vm := newDiffVM(newUIMock(), viewer, getter, id)
	// vm.load()
	// viewer.Run()
	// vt, err := vm.getCommitDiffLeft(cui.ViewPage{Width: 100, Height: 1000})
	// assert.NoError(t, err)
	// t.Logf("%s", strings.Join(vt.Lines, "\n"))
}
