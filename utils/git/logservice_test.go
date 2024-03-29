package git

import (
	"strings"
	"testing"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
)

var logText = strings.Join(strings.Split(`
e8cbef1cf080fe4b102482157000468fffe45e67|2019-10-12 18:48:43 +0200|2019-10-12 18:48:43 +0200|Michael Reichenauer|73e01bf288805d247e5c805238e936c6706373cd|Fix columns
73e01bf288805d247e5c805238e936c6706373cd|2019-10-12 10:47:44 +0200|2019-10-12 10:47:44 +0200|Michael Reichenauer|1dcc6410b61d20a4b15861437cd7c78409bed8fc|fix
1dcc6410b61d20a4b15861437cd7c78409bed8fc|2019-10-12 09:03:37 +0200|2019-10-12 09:03:37 +0200|Michael Reichenauer||Initial commit
`, "\n"), "\x00")

func TestSomeCommits(t *testing.T) {
	wf := tests.CreateTempFolder()
	defer tests.CleanTemp()

	git := New(wf.Path())
	assert.NoError(t, git.InitRepo())
	assert.NoError(t, git.ConfigUser("test", "test@test.com"))

	wf.File("a.txt").Write("1")
	assert.NoError(t, git.Commit("initial"))

	l, err := git.GetLog()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(l))
	assert.Equal(t, "initial", l[0].Subject)

	wf.File("a.txt").Write("2")
	assert.NoError(t, git.Commit("second"))

	l, _ = git.GetLog()
	assert.Equal(t, 2, len(l))
	assert.Equal(t, "second", l[0].Subject)
	assert.Equal(t, "initial", l[1].Subject)
	assert.Equal(t, l[1].ID, l[0].ParentIDs[0])
}

func TestLogFromCurrentDir_Manual(t *testing.T) {
	tests.ManualTest(t)

	log, err := newLog(newGitCmd(utils.CurrentDir())).getLog(100)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(log))

	t.Logf("%d commits:\n%v", len(log), log)
}
