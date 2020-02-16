package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

var branchesText = `
* develop 12898c42e3c919ba6ef6b679a7adb07339213148 Support for git log
  master  73e01bf288805d247e5c805238e936c6706373cd fix"
`

func Test(t *testing.T) {
	branches, err := newBranches(newGitCmd(utils.CurrentDir())).getBranches()
	assert.NoError(t, err)
	for _, b := range branches {
		log.Infof("%+v", b)
		t.Logf("%v", b)
	}
}

func TestParseBranchesText(t *testing.T) {
	branches, err := newBranches(newGitCmd(utils.CurrentDir())).parseBranchesOutput(branchesText)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(branches))
}
