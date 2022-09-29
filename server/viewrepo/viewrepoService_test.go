package viewrepo

import (
	"strings"
	"testing"

	"github.com/michael-reichenauer/gmc/client/console"
	"github.com/michael-reichenauer/gmc/server/viewrepo/augmented"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/stretchr/testify/assert"
)

func Test2(t *testing.T) {
	// even := lo.Filter([]int{1, 2, 3, 4, 6}, func(x int, _ int) bool {
	// 	return x%2 == 0
	// })
	// t.Logf("array %v", even)
	// []int{2, 4, 6}
}

func Test(t *testing.T) {
	//tests.ManualTest(t)
	repoService := augmented.NewRepoService(nil, CurrentRoot())
	repo, err := repoService.GetFreshRepo()
	assert.NoError(t, err)
	assert.Greater(t, len(repo.Commits), 0)

	viewRepoService := NewViewRepoService(nil, CurrentRoot())
	cb, ok := repo.CurrentBranch()
	assert.True(t, ok)
	viewRepo := viewRepoService.GetViewModel(repo, []string{cb.Name})
	assert.Greater(t, len(viewRepo.Commits), 0)

	graph := console.NewRepoGraph()

	for _, c := range viewRepo.Commits {
		var sb strings.Builder
		graph.WriteGraph(&sb, c.graph)
		t.Logf("%s %s %s", sb.String(), c.SID, c.Subject)
	}

	// model := NewService(`C:\Work Files\GitMind`)
	// model.Load()
	// vp, err := model.GetRepoViewPort(0, 18, 0)
	// assert.NoError(t, err)
	// assert.NotEqual(t, len(vp.Commits), 0)
	// for _, c := range vp.Commits {
	// 	t.Logf("%s", c.String())
	// }
}

func CurrentRoot() string {
	root, err := git.WorkingTreeRoot(utils.CurrentDir())
	if err != nil {
		panic(log.Fatal(err))
	}
	return root
}
