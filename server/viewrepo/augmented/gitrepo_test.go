package augmented

import (
	"context"
	"testing"

	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/michael-reichenauer/gmc/utils/timer"
)

func TestCurrentRepo_Manual(t *testing.T) {
	// tests.ManualTest(t)
	gr := NewRepoService(nil, git.CurrentRoot())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gr.StartMonitor(ctx)

	gr.TriggerManualRefresh()
	for r := range gr.RepoChanges() {
		if r.IsStarting {
			continue
		}

		st := timer.Start()
		commits := r.Repo.SearchCommits("v0.22")
		t.Logf("Commits: %d of %d %s", len(commits), len(r.Repo.Commits), st)
		break
	}
}

func TestAcsRepo_Manual(t *testing.T) {
	tests.ManualTest(t)
	gr := NewRepoService(nil, "C:\\Work Files\\AcmAcs")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gr.StartMonitor(ctx)

	gr.TriggerManualRefresh()
	var repo Repo
	for r := range gr.RepoChanges() {
		if r.IsStarting {
			continue
		}
		repo = r.Repo
		break
	}

	st := timer.Start()
	commits := repo.SearchCommits("1011")
	t.Logf("Commits: %d of %d %s", len(commits), len(repo.Commits), st)
}
