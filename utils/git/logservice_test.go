package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

var logText = `
e8cbef1cf080fe4b102482157000468fffe45e67|2019-10-12 18:48:43 +0200|2019-10-12 18:48:43 +0200|Michael Reichenauer|73e01bf288805d247e5c805238e936c6706373cd|Fix columns
73e01bf288805d247e5c805238e936c6706373cd|2019-10-12 10:47:44 +0200|2019-10-12 10:47:44 +0200|Michael Reichenauer|1dcc6410b61d20a4b15861437cd7c78409bed8fc|fix
1dcc6410b61d20a4b15861437cd7c78409bed8fc|2019-10-12 09:03:37 +0200|2019-10-12 09:03:37 +0200|Michael Reichenauer||Initial commit
`

func TestSome(t *testing.T) {
	tests.ManualTest(t)
	t.Log("Running a manual test")
}

func TestLogFromCurrentDir(t *testing.T) {
	tests.ManualTest(t)
	log, err := newLog(newGitCmd(utils.CurrentDir())).getLog()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, 0, len(log))
}

func TestLog(t *testing.T) {
	tests.ManualTest(t)
	commits, err := newLog(newGitCmd(utils.CurrentDir())).parseCommits(logText)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 3 {
		t.Errorf("Unexpexted commits count %d", len(commits))
	}
	if commits[0].ID != "e8cbef1cf080fe4b102482157000468fffe45e67" ||
		commits[0].ParentIDs[0] != "73e01bf288805d247e5c805238e936c6706373cd" {
		t.Errorf("Unexpexted commit  %+v", commits[0])
	}
}
