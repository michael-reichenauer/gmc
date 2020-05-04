package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommitDiff(t *testing.T) {
	wf := tests.CreateTempFolder()
	file1 := "a.txt"

	git := New(wf.Path())
	assert.NoError(t, git.InitRepo(wf.Path()))

	wf.File(file1).Write("1")
	assert.NoError(t, git.Commit("initial"))

	assert.NoError(t, git.CreateBranch("develop"))
	wf.File(file1).Write("2")
	assert.NoError(t, git.Commit("second"))

	assert.NoError(t, git.Checkout("master"))
	assert.NoError(t, git.MergeBranch("develop"))
	assert.NoError(t, git.Commit("merged"))
	assert.Equal(t, "2", wf.File(file1).Read())

	log, _ := git.GetLog()

	// Get diff of first commit
	diff, err := git.CommitDiff(log.MustBySubject("initial").ID)
	assert.NoError(t, err)

	// Verify one added file1 with one added line "1"
	assert.Equal(t, 1, len(diff.FileDiffs))
	assert.Equal(t, DiffAdded, diff.FileDiffs[0].DiffMode)
	assert.Equal(t, file1, diff.FileDiffs[0].PathAfter)
	assert.Equal(t, 1, len(diff.FileDiffs[0].SectionDiffs))
	assert.Equal(t, DiffAdded, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].DiffMode)
	assert.Equal(t, "1", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].Line)

	//  Verify one modified file1 with one line removed "1"  and one line added "2"
	diff, _ = git.CommitDiff(log.MustBySubject("second").ID)
	assert.Equal(t, 1, len(diff.FileDiffs))
	assert.Equal(t, DiffModified, diff.FileDiffs[0].DiffMode)
	assert.Equal(t, 1, len(diff.FileDiffs[0].SectionDiffs))
	assert.Equal(t, DiffRemoved, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].DiffMode)
	assert.Equal(t, "1", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].Line)
	assert.Equal(t, DiffAdded, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[1].DiffMode)
	assert.Equal(t, "2", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[1].Line)

	//  Verify one modified file1 with one line removed "1"  and one line added "2"
	diff, _ = git.CommitDiff(log.MustBySubject("merged").ID)
	assert.Equal(t, 1, len(diff.FileDiffs))
	assert.Equal(t, DiffModified, diff.FileDiffs[0].DiffMode)
	assert.Equal(t, 1, len(diff.FileDiffs[0].SectionDiffs))
	assert.Equal(t, DiffRemoved, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].DiffMode)
	assert.Equal(t, "1", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[0].Line)
	assert.Equal(t, DiffAdded, diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[1].DiffMode)
	assert.Equal(t, "2", diff.FileDiffs[0].SectionDiffs[0].LinesDiffs[1].Line)
}

func TestCommitDiff_Manual(t *testing.T) {
	tests.ManualTest(t)
	diffService := newDiff(newGitCmd(utils.CurrentDir()), newStatus(newGitCmd(utils.CurrentDir())))
	diff, err := diffService.commitDiff("31a1ff3c7dfe6bc90776aeb75e4b27eaf361c706")
	assert.NoError(t, err)
	for i, df := range diff.FileDiffs {
		if i != 0 {
			t.Logf("")
		}
		t.Logf("------------------------------")
		t.Logf("File: %s", df.PathAfter)
		if df.IsRenamed {
			t.Logf("Renamed: %s -> %s", df.PathBefore, df.PathAfter)
		}

		for j, ds := range df.SectionDiffs {
			if j != 0 {
				t.Logf("")
			}
			t.Logf("------------------------------")
			t.Logf("Lines: %s", ds.ChangedIndexes)
			t.Logf("------------------------------")
			for _, dl := range ds.LinesDiffs {
				switch dl.DiffMode {
				case DiffSame:
					t.Logf("  %s", dl.Line)
				case DiffAdded:
					t.Logf("+ %s", dl.Line)
				case DiffRemoved:
					t.Logf("- %s", dl.Line)
				}
			}
		}
	}
}

func TestCommitDiff2_Manual(t *testing.T) {
	tests.ManualTest(t)
	diffService := newDiff(newGitCmd(utils.CurrentDir()), newStatus(newGitCmd(utils.CurrentDir())))
	diff, err := diffService.commitDiff(UncommittedID)
	assert.NoError(t, err)
	for i, df := range diff.FileDiffs {
		if i != 0 {
			t.Logf("")
		}
		t.Logf("------------------------------")
		t.Logf("File: %s", df.PathAfter)
		if df.IsRenamed {
			t.Logf("Renamed: %s -> %s", df.PathBefore, df.PathAfter)
		}

		for j, ds := range df.SectionDiffs {
			if j != 0 {
				t.Logf("")
				t.Logf("------------------------------")
			}

			t.Logf("Lines: %s", ds.ChangedIndexes)
			t.Logf("------------------------------")
			for _, dl := range ds.LinesDiffs {
				switch dl.DiffMode {
				case DiffSame:
					t.Logf("  %s", dl.Line)
				case DiffAdded:
					t.Logf("+ %s", dl.Line)
				case DiffRemoved:
					t.Logf("- %s", dl.Line)
				}
			}
		}
	}
}
