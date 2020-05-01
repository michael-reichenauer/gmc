package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var logText = strings.Join(strings.Split(`
e8cbef1cf080fe4b102482157000468fffe45e67|2019-10-12 18:48:43 +0200|2019-10-12 18:48:43 +0200|Michael Reichenauer|73e01bf288805d247e5c805238e936c6706373cd|Fix columns
73e01bf288805d247e5c805238e936c6706373cd|2019-10-12 10:47:44 +0200|2019-10-12 10:47:44 +0200|Michael Reichenauer|1dcc6410b61d20a4b15861437cd7c78409bed8fc|fix
1dcc6410b61d20a4b15861437cd7c78409bed8fc|2019-10-12 09:03:37 +0200|2019-10-12 09:03:37 +0200|Michael Reichenauer||Initial commit
`, "\n"), "\x00")

func TestLogFromCurrentDir(t *testing.T) {
	tests.ManualTest(t)

	log, err := newLog(newGitCmd(utils.CurrentDir())).getLog()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(log))

	t.Logf("%v", log)
}

func TestLog(t *testing.T) {
	commits, err := newLog(newMockCmd("")).parseCommits(logText)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(commits))
	assert.Equal(t, "e8cbef1cf080fe4b102482157000468fffe45e67", commits[0].ID)
	assert.Equal(t, "73e01bf288805d247e5c805238e936c6706373cd", commits[0].ParentIDs[0])
}
