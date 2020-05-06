package viewmodel

import (
	"github.com/michael-reichenauer/gmc/utils/tests"
	"testing"
)

func Test(t *testing.T) {
	tests.ManualTest(t)
	// model := NewService(`C:\Work Files\GitMind`)
	// model.Load()
	// vp, err := model.GetRepoViewPort(0, 18, 0)
	// assert.NoError(t, err)
	// assert.NotEqual(t, len(vp.Commits), 0)
	// for _, c := range vp.Commits {
	// 	t.Logf("%s", c.String())
	// }
}
