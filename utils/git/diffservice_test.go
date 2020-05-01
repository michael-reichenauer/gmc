package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommitDiff(t *testing.T) {
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

func TestCommitDiff2(t *testing.T) {
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
