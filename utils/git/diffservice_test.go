package git

import (
	"testing"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestCurrentUncommittedRepo(t *testing.T) {
	tests.ManualTest(t)

	g := New("")
	diff, err := g.CommitDiff(UncommittedID)
	assert.NoError(t, err)
	t.Logf("Diff: %#v", diff)
}

func TestSpecialRepo(t *testing.T) {
	tests.ManualTest(t)

	g := New("/workspaces/gmd")
	diff, err := g.CommitDiff("b472e9694556a33c8e7ffc4714cd7df12fa1ca1c")
	assert.NoError(t, err)
	t.Logf("Diff: %s", utils.PrettyString(diff))
}

func TestCommitDiff(t *testing.T) {
	wf := tests.CreateTempFolder()
	file1 := "a.txt"

	g := New(wf.Path())
	assert.NoError(t, g.InitRepo())
	assert.NoError(t, g.ConfigUser("test", "test@test.com"))

	wf.File(file1).Write("1")
	assert.NoError(t, g.Commit("initial"))

	assert.NoError(t, g.CreateBranch("develop"))
	wf.File(file1).Write("2")
	assert.NoError(t, g.Commit("second"))

	assert.NoError(t, g.Checkout("master"))
	assert.NoError(t, g.MergeBranch("develop"))
	assert.NoError(t, g.Commit("merged"))
	assert.Equal(t, "2", wf.File(file1).Read())

	log, _ := g.GetLog()

	// Get diff of first commit
	diff, err := g.CommitDiff(log.MustBySubject("initial").ID)
	assert.NoError(t, err)

	// Verify one added file1 with one added line "1"
	assert.Equal(t, 1, len(diff.FileDiffs))
	assert.Equal(t, DiffAdded, diff.FileDiffs[0].DiffMode)
	assert.Equal(t, file1, diff.FileDiffs[0].PathAfter)
	assert.Equal(t, 1, len(diff.FileDiffs[0].SectionDiffs))
	assert.Equal(t, DiffAdded, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].DiffMode)
	assert.Equal(t, "1", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].Line)

	//  Verify one modified file1 with one line removed "1"  and one line added "2"
	diff, _ = g.CommitDiff(log.MustBySubject("second").ID)
	assert.Equal(t, 1, len(diff.FileDiffs))
	assert.Equal(t, DiffModified, diff.FileDiffs[0].DiffMode)
	assert.Equal(t, 1, len(diff.FileDiffs[0].SectionDiffs))
	assert.Equal(t, DiffRemoved, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].DiffMode)
	assert.Equal(t, "1", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].Line)
	assert.Equal(t, DiffAdded, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[1].DiffMode)
	assert.Equal(t, "2", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[1].Line)

	//  Verify one modified file1 with one line removed "1"  and one line added "2"
	diff, _ = g.CommitDiff(log.MustBySubject("merged").ID)
	assert.Equal(t, 1, len(diff.FileDiffs))
	assert.Equal(t, DiffModified, diff.FileDiffs[0].DiffMode)
	assert.Equal(t, 1, len(diff.FileDiffs[0].SectionDiffs))
	assert.Equal(t, DiffRemoved, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].DiffMode)
	assert.Equal(t, "1", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].Line)
	assert.Equal(t, DiffAdded, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[1].DiffMode)
	assert.Equal(t, "2", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[1].Line)

	wf.File(file1).Write("4")
	assert.NoError(t, g.Commit("commitonmaster"))

	assert.NoError(t, g.Checkout("develop"))
	wf.File(file1).Write("5")
	diff, err = g.CommitDiff(UncommittedID)
	assert.NoError(t, err)

	assert.NoError(t, g.Commit("commitondevelop"))
	assert.NoError(t, g.Checkout("master"))

	assert.Error(t, ErrConflicts, g.MergeBranch("develop"))

	diff, err = g.CommitDiff(UncommittedID)
	assert.NoError(t, err)
}

func TestFileDiff(t *testing.T) {
	wf := tests.CreateTempFolder()
	file1 := "a.txt"

	g := New(wf.Path())
	assert.NoError(t, g.InitRepo())
	assert.NoError(t, g.ConfigUser("test", "test@test.com"))

	wf.File(file1).Write("1")
	assert.NoError(t, g.Commit("initial"))

	wf.File(file1).Write("2")
	assert.NoError(t, g.Commit("second"))

	wf.File(file1).Write("3")
	assert.NoError(t, g.Commit("third"))

	_, err := g.FileDiff(file1)
	assert.NoError(t, err)
	//t.Logf("diff:\n%s", utils.PrettyString(diff))
}
