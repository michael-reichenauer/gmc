package git

import (
	"gmc/utils"
	"testing"
)

var branchesText = `
* develop 12898c42e3c919ba6ef6b679a7adb07339213148 Support for git log
  master  73e01bf288805d247e5c805238e936c6706373cd fix"
`

func Test(t *testing.T) {
	branchesText, err := getGitBranches(utils.CurrentDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("branches %q", branchesText)
}

func TestParseBranchesText(t *testing.T) {
	branches, err := parseBranches(branchesText)
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 2 {
		t.Fatalf("Unexpeted branches count: %d", len(branches))
	}
	t.Logf("branches %d", len(branches))
}
