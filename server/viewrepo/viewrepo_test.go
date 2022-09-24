package viewrepo

import (
	"testing"

	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/samber/lo"
)

func Test2(t *testing.T) {
	t.Logf("Test")

	even := lo.Filter[int]([]int{1, 2, 3, 4}, func(x int, _ int) bool {
		return x%2 == 0
	})
	t.Logf("array %v", even)
	// []int{2, 4}
}

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
