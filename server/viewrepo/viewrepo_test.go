package viewrepo

import (
	"testing"

	"github.com/michael-reichenauer/gmc/utils/tests"
)

func Test2(t *testing.T) {
	// even := lo.Filter([]int{1, 2, 3, 4, 6}, func(x int, _ int) bool {
	// 	return x%2 == 0
	// })
	// t.Logf("array %v", even)
	// []int{2, 4, 6}
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
