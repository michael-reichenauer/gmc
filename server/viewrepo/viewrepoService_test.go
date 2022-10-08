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

func TestCurrentRepo(t *testing.T) {
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
}

func TestSpecial(t *testing.T) {
	//tests.ManualTest(t)
	// /workspaces/Dependitor
	repoPath := ""

	repoService := augmented.NewRepoService(nil, repoPath)
	repo, err := repoService.GetFreshRepo()
	assert.NoError(t, err)
	assert.Greater(t, len(repo.Commits), 0)

	viewRepoService := NewViewRepoService(nil, repoPath)
	cb, ok := repo.CurrentBranch()
	assert.True(t, ok)
	viewRepo := viewRepoService.GetViewModel(repo, []string{cb.Name})
	assert.Greater(t, len(viewRepo.Commits), 0)

	graph := console.NewRepoGraph()

	for _, c := range viewRepo.Commits {
		var sb strings.Builder
		graph.WriteGraph(&sb, c.graph)
		tt := c.AuthorTime.Format("2006-01-02 15:04")
		tt = tt[2:]
		t.Logf("%s %s %s %s %s", sb.String(), c.SID, utils.Text(c.Subject, 30), utils.Text(c.Author, 10), utils.Text(tt, 14))
	}
}

func CurrentRoot() string {
	root, err := git.WorkingTreeRoot(utils.CurrentDir())
	if err != nil {
		panic(log.Fatal(err))
	}
	return root
}
