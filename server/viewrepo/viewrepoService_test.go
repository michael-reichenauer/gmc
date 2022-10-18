package viewrepo

import (
	"strings"
	"testing"

	"github.com/michael-reichenauer/gmc/client/console"
	"github.com/michael-reichenauer/gmc/server/viewrepo/augmented"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestShowBranchColors(t *testing.T) {
	tests.ManualTest(t)
	for i := 0; i < len(branchColors); i++ {
		t.Log(cui.ColorText(branchColors[i], strings.Repeat("━", 20)))
	}
}

func TestShowAllColors(t *testing.T) {
	tests.ManualTest(t)
	for i := 0; i < len(cui.AllColors); i++ {
		t.Log(cui.ColorText(cui.AllColors[i], strings.Repeat("━", 20)))
	}
}

func TestCurrentRepo(t *testing.T) {
	tests.ManualTest(t)
	repoService := augmented.NewRepoService(CurrentRoot())
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
	tests.ManualTest(t)
	// /workspaces/Dependitor
	// /workspaces/gg
	//repoPath := "/workspaces/gg"
	repoPath := ""

	repoService := augmented.NewRepoService(repoPath)
	repo, err := repoService.GetFreshRepo()
	assert.NoError(t, err)
	assert.Greater(t, len(repo.Commits), 0)

	viewRepoService := NewViewRepoService(nil, repoPath)
	cb, ok := repo.CurrentBranch()
	assert.True(t, ok)
	viewRepo := viewRepoService.GetViewModel(repo, []string{cb.Name, "origin/develop"})
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
