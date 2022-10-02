package augmented

import (
	"context"
	"testing"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/michael-reichenauer/gmc/utils/timer"
	"github.com/stretchr/testify/assert"
)

func TestCurrentAugmentedRepo_Manual(t *testing.T) {
	//tests.ManualTest(t)
	repoService := NewRepoService(nil, CurrentRoot())
	repo, err := repoService.GetFreshRepo()
	assert.NoError(t, err)
	assert.Greater(t, len(repo.Commits), 0)
}

func TestSpecialAugmentedRepo_Manual(t *testing.T) {
	//tests.ManualTest(t)
	// repoService := NewRepoService(nil, "")
	// repo, err := repoService.GetFreshRepo()
	// assert.NoError(t, err)
	// assert.Greater(t, len(repo.Commits), 0)
}

func TestCurrentRepoTrigger_Manual(t *testing.T) {
	tests.ManualTest(t)
	repoService := NewRepoService(nil, CurrentRoot())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repoService.StartMonitor(ctx)

	repoService.TriggerManualRefresh()
	for r := range repoService.RepoChanges() {
		if r.IsStarting {
			continue
		}

		st := timer.Start()
		commits := r.Repo.SearchCommits("v0.22")
		t.Logf("Commits: %d of %d %s", len(commits), len(r.Repo.Commits), st)
		break
	}
}

func CurrentRoot() string {
	root, err := git.WorkingTreeRoot(utils.CurrentDir())
	if err != nil {
		panic(log.Fatal(err))
	}
	return root
}
